// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
)

const warningTxTime = 1000 * time.Millisecond

// NoSQL objects
type NoSQLStateDB struct {
	muCollections sync.RWMutex

	collectionOwner map[common.Address]string
	systemProof     map[common.Address]*Proof
	storageBytes    map[common.Address]uint64

	// smt objects - txn hashes of bucket definitions (TxBuckets)
	collections     map[common.Address]*SparseMerkleTree // common.Address is owner + bucket
	collectionProof map[common.Address]*Proof            // proofs that the roothash of collections smt exist in keystorage

	// avl objects - txn hashes of rows of data (TxKeys)
	indexedCollections     map[common.Address]*AVLTree // common.Address is owner + bucket + index
	indexedCollectionProof map[common.Address]*Proof   // proofs that the roothash of data avltree exists in KeyStorage
}

func CollectionHash(owner string, collection string) common.Address {
	return common.BytesToAddress(wolkcommon.Computehash(append([]byte(owner), collection...))[0:20])
}

func IndexedCollectionHash(owner string, collection string, index string) common.Address {
	addr := append([]byte(owner), collection...)
	addr = append(addr, index...)
	return common.BytesToAddress(wolkcommon.Computehash(addr)[0:20])
}

func KeyToAddress(k []byte) common.Address {
	return common.BytesToAddress(wolkcommon.Computehash(k)[0:20])
}

func IndexedKeyToAddress(k []byte) common.Address {
	paddedKey := padBytes(k, 20)
	//dprint("[statedb_nosql:IndexedKeyToAddress] k(%v) paddedKey(%v)", k, paddedKey)
	return common.BytesToAddress(paddedKey[0:20])
}

// CommitNoSQL inserts new SMT roots for all the collections that were updated from the keyobjects into keyStorage
func (statedb *StateDB) CommitNoSQL(ctx context.Context, wg *sync.WaitGroup, writeToCloudstore bool) (err error) {

	statedb.nosql.muCollections.Lock()
	defer statedb.nosql.muCollections.Unlock()
	// log.Info("CommitNoSQL keyStorage Dump prior to smt flushes")
	// statedb.keyStorage.Dump()

	// flush roothashes of collection buckets
	for addr, collSMT := range statedb.nosql.collections { // addr is owner + bucket

		// check collection's owner
		collectionOwner, ok := statedb.nosql.collectionOwner[addr]
		if !ok {
			log.Error("[statedb_nosql:CommitNoSQL] NO OWNER")
			return fmt.Errorf("[statedb_nosql:CommitNoSQL] NO collection owner! owner-bucket key(%x)", addr)
		}

		// flush collection smt
		_, err = collSMT.Flush(ctx, wg, writeToCloudstore)
		if err != nil {
			log.Error("[statedb_nosql:CommitNoSQL] FLUSH", "err", err)
			return fmt.Errorf("[statedb_nosql:CommitNoSQL] %s", err) // this?
		}

		// tally up storage bytes for the collection - should be in data trees
		oldStorageBytes := uint64(0)
		if old, ok := statedb.nosql.storageBytes[addr]; ok {
			oldStorageBytes = old
		}
		newStorageBytes, err := collSMT.StorageBytes()
		if err != nil {
			return fmt.Errorf("statedb_nosql:CommitNoSQL] %s", err)
		}
		diff := newStorageBytes - oldStorageBytes
		err = statedb.tallyStorageUsage(ctx, collectionOwner, diff)
		if err != nil {
			log.Error("[statedb_nosql:CommitNoSQL] FAILURE", "addr key", addr, "v", collSMT.ChunkHash(), "oldStorageBytes", oldStorageBytes, "newStorageBytes", newStorageBytes, "diff", diff, "err", err)
			return fmt.Errorf("[statedb_nosql:CommitNoSQL] %s", err)
		}
		log.Trace("**** CommitNoSQL Insert 2: colleqctionOwner ****", "oldStorageBytes", oldStorageBytes, "newStorageBytes", newStorageBytes, "k", addr, "v", collSMT.ChunkHash())

		//smt.Dump()

		// insert into keyStorage key:<owner+bucket addr> val:<root of collectionSMT>
		err = statedb.keyStorage.Insert(ctx, addr.Bytes(), collSMT.ChunkHash().Bytes(), diff, false)
		if err != nil {
			return fmt.Errorf("[statedb_nosql:CommitNoSQL] %s", err) // this?
		}

		// clean up the collections cache
		delete(statedb.nosql.collections, addr)
	}

	// log.Info("CommitNoSQL keyStorage Dump after to smt flushes")
	// statedb.keyStorage.Dump()

	// flush roothashes of indexed collection trees
	for addr, icTree := range statedb.nosql.indexedCollections { // addr is owner+bucket+index

		// flush the tree
		ok, err := icTree.Flush(ctx, wg, writeToCloudstore)
		if err != nil {
			log.Error("[statedb_nosql:CommitNoSQL] FLUSH", "err", err)
			return fmt.Errorf("[statedb_nosql:CommitNoSQL] %s", err)
		}
		if !ok {
			log.Info("[statedb_nosql:CommitNoSQL] flush not needed, continue")
			continue
		}
		// tally storagebytes  - TODO:
		/*
			oldStorageBytes := uint64(0)
			if old, ok := statedb.nosql.storageBytes[addr]; ok {
				oldStorageBytes = old
			}
			newStorageBytes, err := icTree.StorageBytes()
			if err != nil {
				log.Error("[statedb_nosql:CommitNoSQL] StorageBytes", "err", err)
				return err
			}
			diff := newStorageBytes - oldStorageBytes
			err = statedb.tallyStorageUsage(ctx, collectionOwner, diff)
			if err != nil {
				log.Error("[statedb_nosql:CommitNoSQL] FAILURE", "addr key", addr, "v", collSMT.ChunkHash(), "oldStorageBytes", oldStorageBytes, "newStorageBytes", newStorageBytes, "diff", diff, "err", err)
				return fmt.Errorf("[statedb_nosql:CommitNoSQL] %s", err)
			}
			log.Trace("**** CommitNoSQL Insert 2: collectionOwner ****", "oldStorageBytes", oldStorageBytes, "newStorageBytes", newStorageBytes, "k", addr, "v", collSMT.ChunkHash())
		*/
		// insert into keyStorage k<owner+bucket+index>, v<roothash of icTree>
		// TODO: chg 0 to storageBytes tally diff
		//log.Info("[statedb_nosql:CommitNoSQL] keystorage insert", "icHash", addr, "icTree.ChunkHash", icTree.ChunkHash())
		err = statedb.keyStorage.Insert(ctx, addr.Bytes(), icTree.ChunkHash().Bytes(), 0, false)
		if err != nil {
			return fmt.Errorf("[statedb_nosql:CommitNoSQL] %s", err)
		}
	}

	return nil
}

// TODO: check this
func (statedb *StateDB) StorageCollection(ctx context.Context, owner string, collection string) (storageBytes uint64, ok bool, err error) {
	smt, ok, err := statedb.getCollectionSMT(ctx, owner, collection, nil)
	if err != nil {
		return storageBytes, false, err
	} else if ok {
		storageBytes, err = smt.StorageBytes()
		return storageBytes, true, err
	}
	return 0, false, nil
}

// The owner string, when hashed to compute systemHash references an SMT whose root can be found in keyStorage.
// In this SMT is the list of collections the owner actually owns.
func (statedb *StateDB) getOwnerSMT(ctx context.Context, owner string, proof *NoSQLProof) (smt *SparseMerkleTree, ok bool, err error) {
	systemHash := NameToAddress(owner)
	//log.Info(fmt.Sprintf("[statedb_nosql:getOwnerSMT] FETCH owner [%s] with hash [%x]", owner, systemHash))
	statedb.nosql.muCollections.Lock()
	defer statedb.nosql.muCollections.Unlock()
	if systemSMT, ok := statedb.nosql.collections[systemHash]; ok {
		if proof != nil {
			// retrieve from cache
			if systemProof, okp := statedb.nosql.systemProof[systemHash]; okp {
				proof.SystemProof = NewSerializedProof(systemProof)
				log.Trace("[statedb_nosql:getOwnerSMT] OwnerProof found in cache", "ownerproof", systemProof.String())
			} else {
				log.Error("[statedb_nosql:getOwnerSMT] OwnerProof NOT found", "systemHash", fmt.Sprintf("%x", systemHash))
				proof.SystemProof = NewSerializedProof(new(Proof))
			}
			proof.KeyChunkHash = statedb.keyStorage.ChunkHash()
			proof.KeyMerkleRoot = statedb.keyStorage.MerkleRoot(ctx)
			proof.SystemChunkHash = systemSMT.ChunkHash()
			proof.SystemMerkleRoot = systemSMT.MerkleRoot(ctx)
			proof.SystemHash = systemHash
			proof.Owner = owner
			log.Trace("[statedb_nosql:getOwnerSMT]", "OwnerProof DONE from cache", proof)
		}
		return systemSMT, true, nil
	}
	systemSMT := NewSparseMerkleTree(NumBitsAddress, statedb.Storage)
	// this SMT holds all the buckets the user has created
	systemChunkHashBytes, ok, deleted, systemProof, _, err := statedb.keyStorage.Get(ctx, systemHash.Bytes(), true)
	if err != nil {
		log.Error("[statedb_nosql:getOwnerSMT] keyStorage.Get err", "err", err)
		return smt, false, err
	} else if !ok || deleted {
		log.Error(fmt.Sprintf("[statedb_nosql:getOwnerSMT] systemSMT NOT ok - hash: %x", systemHash))
		return smt, ok, nil
	}

	systemSMT.Init(common.BytesToHash(systemChunkHashBytes))
	storageBytes, err := systemSMT.StorageBytes()
	if err != nil {
		return systemSMT, true, err
	}
	statedb.nosql.collectionOwner[systemHash] = owner
	statedb.nosql.collections[systemHash] = systemSMT
	statedb.nosql.storageBytes[systemHash] = storageBytes

	log.Trace(fmt.Sprintf("[statedb_nosql:getOwnerSMT] nosql.collections[%x] SMT created for owner [%s] len %d => %v", systemHash, owner, len(statedb.nosql.collections), systemSMT))
	// for h0, smtj := range statedb.nosql.collections {
	// 	fmt.Printf("  -> %x - %v\n", h0, smtj)
	// }
	if proof != nil {
		statedb.nosql.systemProof[systemHash] = systemProof // keep this for cache hits
		// Verify proof process: at position proof.SystemHash there is the value proof.SystemChunkHash, which will hash up to proof.KeyMerkleRoot
		proof.KeyChunkHash = statedb.keyStorage.ChunkHash()
		proof.KeyMerkleRoot = statedb.keyStorage.MerkleRoot(ctx)
		proof.SystemChunkHash = systemSMT.ChunkHash()
		proof.SystemMerkleRoot = systemSMT.MerkleRoot(ctx)
		proof.SystemHash = systemHash
		proof.SystemProof = NewSerializedProof(systemProof)
	}
	return systemSMT, true, nil
}

func (statedb *StateDB) getCollectionSMT(ctx context.Context, owner string, collection string, proof *NoSQLProof) (smt *SparseMerkleTree, ok bool, err error) {
	// if we previously opened the collection SMT (caused through state changes in this block)
	collectionHash := CollectionHash(owner, collection)
	var collectionProof *Proof
	statedb.nosql.muCollections.Lock()
	defer statedb.nosql.muCollections.Unlock()

	if cachedSMT, ok := statedb.nosql.collections[collectionHash]; ok {
		smt = cachedSMT
		if p, ok := statedb.nosql.collectionProof[collectionHash]; ok {
			collectionProof = p
		}
		log.Trace("[statedb_nosql:getCollectionSMT] FOUND collectionSMT in cache", "collection", collection, "collectionHash", collectionHash)
	} else {

		smt = NewSparseMerkleTree(NumBitsAddress, statedb.Storage)
		statedb.nosql.collectionOwner[collectionHash] = owner
		statedb.nosql.collections[collectionHash] = smt
		// check if we have a collection in keyStorage (caused through state changes in previous blocks)
		var collectionChunkHashBytes []byte
		var ok bool
		var deleted bool
		var storageBytes uint64
		collectionChunkHashBytes, ok, deleted, collectionProof, storageBytes, err = statedb.keyStorage.Get(ctx, collectionHash.Bytes(), true)
		if err != nil {
			log.Error("[statedb_nosql:getCollectionSMT] keyStorage.Get ERR", "collectionHash", collectionHash, "err", err)
			return smt, ok, fmt.Errorf("[statedb_nosql:getCollectionSMT] %s", err)
		}
		if !ok {
			return smt, false, fmt.Errorf("[statedb_nosql:getCollectionSMT] Collection(%s) with owner(%s) not found", collection, owner)
		}
		smt.Init(common.BytesToHash(collectionChunkHashBytes))
		statedb.nosql.collectionProof[collectionHash] = collectionProof
		statedb.nosql.storageBytes[collectionHash] = storageBytes // storagebytes[h] = storage used by address across specific collection // instead, calculate storage bytes for collection + index?

		// we have created this collection before!
		if deleted {
			log.Warn("[statedb_nosql:getCollectionSMT] this collection has been deleted. What to do?", "collection", collection)
		}
		log.Trace("[statedb_nosql:getCollectionSMT] created collection", "collectionHash", collectionHash, "collectionChunkHashBytes", fmt.Sprintf("%x", collectionChunkHashBytes),
			"owner", owner, "collection", collection, "statedb.nosql.storageBytes[collectionHash]", statedb.nosql.storageBytes[collectionHash], "storageBytes", storageBytes)
	}
	if proof != nil {
		// Verify Proof Process: at proof.CollectionHash position is proof.CollectionChunkHash, which hashes up to proof.KeyMerkleRoot
		proof.CollectionHash = collectionHash
		proof.CollectionProof = NewSerializedProof(collectionProof)
		proof.CollectionChunkHash = smt.ChunkHash()
		proof.CollectionMerkleRoot = smt.MerkleRoot(ctx)
		log.Trace("[statedb_nosql:getCollectionSMT] CollectionProof DONE", "proof", proof)
	}
	return smt, true, nil
}

// gets the owner + collection + index AVL tree by looking up the roothash of the AVL tree in the collection SMT.
// note: setIndexedCollectionTree (init of new data trees) is inside SetBucket
func (statedb *StateDB) getIndexedCollectionTree(ctx context.Context, owner string, collection string, index string, proof *NoSQLProof) (tree *AVLTree, ok bool, err error) {

	// get the key to find the tree
	icHash := IndexedCollectionHash(owner, collection, index)
	var treeRootProof *Proof // proof that the treeRootHash exists in keystorage at the icHash key
	statedb.nosql.muCollections.Lock()
	defer statedb.nosql.muCollections.Unlock()

	//log.Info("[statedb_nosql:getIndexedCollectionTree] checking cache for", "o", owner, "coll", collection, "idx", index, "icHash", icHash)
	if cachedTree, ok := statedb.nosql.indexedCollections[icHash]; ok {
		// tree is cached - note: if not committed yet in CommitNoSQL, always working off of cache
		//log.Info("[statedb_nosql:getIndexedCollectionTree] tree is cached!")
		tree = cachedTree
		if p, ok := statedb.nosql.indexedCollectionProof[icHash]; ok {
			treeRootProof = p
		}
		log.Trace("[statedb_nosql:getIndexedCollectionTree] FOUND collectionTree in cache", "collection", collection, "index", index, "icHash", icHash)
	} else {

		// pull tree from storage - or make a new one
		//log.Info("[statedb_nosql:getIndexedCollectionTree] tree is not cached")

		//statedb.nosql.collectionOwner[icHash] = owner // TODO: need something like this?

		// check if the data tree exists in storage
		treeRootHash, ok, deleted, treeRootProof, storageBytes, err := statedb.keyStorage.Get(ctx, icHash.Bytes(), true) // checks the keystorage SMT for the AVL tree roothash
		if err != nil {
			return tree, false, fmt.Errorf("[statedb_nosql:getCollectionAVLTree] %s", err)
		}
		if !ok {
			return tree, false, fmt.Errorf("[statedb_nosql:getIndexedCollectionTree] collection(%s) + index(%s) belonging to owner(%s) not found. make it first?", collection, index, owner)
		}
		// we have created this collection before!
		tree = NewAVLTree(statedb.Storage)
		tree.Init(common.BytesToHash(treeRootHash))                  // load up the AVL tree
		statedb.nosql.indexedCollectionProof[icHash] = treeRootProof // proof for getting the roothash
		statedb.nosql.storageBytes[icHash] = storageBytes            // storagebytes[h] = storage used by address across specific collection + index - note: may not need this for collectionHash then.
		if deleted {
			// TODO
			log.Warn("[statedb_nosql:getIndexedCollectionTree] this collection has been deleted. What to do?", "collection", collection)
		}

		statedb.nosql.indexedCollections[icHash] = tree
		log.Trace("[statedb_nosql:getIndexedCollectionTree] created collection", "collectionHash", icHash, "avl tree root", fmt.Sprintf("%x", treeRootHash), "owner", owner, "collection", collection, "statedb.nosql.storageBytes[collectionHash]", statedb.nosql.storageBytes[icHash], "storageBytes", storageBytes)
	}

	// verifying the collection's proof
	if proof != nil { // Verify Proof Process: at proof.CollectionHash position is proof.CollectionChunkHash, which hashes up to proof.KeyMerkleRoot
		// Verify Proof Process: at proof.CollectionHash position is proof.CollectionChunkHash, which hashes up to proof.KeyMerkleRoot
		proof.CollectionHash = icHash
		proof.CollectionProof = NewSerializedProof(treeRootProof)
		proof.CollectionChunkHash = tree.ChunkHash()
		proof.CollectionMerkleRoot = tree.MerkleRoot(ctx)
		log.Trace("[statedb_nosql:getIndexedCollectionTree] CollectionProof DONE", "proof", proof)
	}

	return tree, true, nil
}

// SetBucket inserts the txBucket hash into the Owner SMT
// TODO: why doesn't SetBucket check if the collectionSMT exists already?
func (statedb *StateDB) SetBucket(ctx context.Context, tx *Transaction) (err error) {
	// the Owner SMT hold TxBuckets defining the owners buckets.
	var txp TxBucket
	deleted := false
	if string(tx.Method) == http.MethodDelete {
		deleted = true
	} else if string(tx.Method) == http.MethodPatch {
		// TODO
	} else {
		err = json.Unmarshal(tx.Payload, &txp)
		if err != nil {
			return err
		}
	}
	st := time.Now()

	// get Owner SMT
	ownerSMT, ok, err := statedb.getOwnerSMT(ctx, tx.Owner(), nil)
	if err != nil {
		return err
	} else if !ok {
		log.Error("[statedb_nosql:SetBucket] Owner not found", "owner", tx.Owner())
		return fmt.Errorf("Owner not found")
	}
	if time.Since(st) > warningTxTime {
		log.Info("[statedb_nosql:SetBucket] getOwnerSMT", "tm", time.Since(st))
	}

	// Make new SMT for owner-collection
	collectionHash := CollectionHash(tx.Owner(), tx.Collection())
	collectionSMT := NewSparseMerkleTree(NumBitsAddress, statedb.Storage)
	collectionSMT.MarkDirty(true)
	statedb.nosql.muCollections.Lock()
	statedb.nosql.collectionOwner[collectionHash] = tx.Owner()
	statedb.nosql.storageBytes[collectionHash] = 0
	statedb.nosql.collections[collectionHash] = collectionSMT
	statedb.nosql.muCollections.Unlock()

	// Insert new owner-collection into owner SMT
	err = ownerSMT.Insert(ctx, collectionHash.Bytes(), tx.Hash().Bytes(), 0, deleted)
	if err != nil {
		log.Error("[statedb_nosql:SetBucket] Insert", "collectionHash", collectionHash, "ERROR", err)
		return err
	}
	log.Trace("**** SetBucket - Insert ****", "owner", tx.Owner(), "collection", tx.Collection(), "k", collectionHash, "v", tx.Hash())
	//dprint("[statedb_nosql:SetBucket] insert owner(%x), collection(%x), k(%x), v(%x)", tx.Owner(), tx.Collection(), collectionHash, tx.Hash())

	// Set up the data trees for indexes
	if len(txp.Indexes) > 0 {
		//dprint("[statedb_nosql:SetBucket] setting up data trees for indexes")
		for _, idx := range txp.Indexes {
			icHash := IndexedCollectionHash(tx.Owner(), tx.Collection(), idx.IndexName)
			//log.Info("[statedb_nosql:SetBucket] setting new avltree", "o", tx.Owner(), "coll", tx.Collection(), "idx", idx.IndexName, "icHash", icHash)
			icTree := NewAVLTree(statedb.Storage)
			statedb.nosql.muCollections.Lock()
			statedb.nosql.indexedCollections[icHash] = icTree
			statedb.nosql.storageBytes[icHash] = 0
			statedb.nosql.muCollections.Unlock()
		}
		// TODO: do we need this?? err = ownerSMT.Insert(ctx, icHash.Bytes(), ....)
		// tx already has all indexes specified
	}

	if time.Since(st) > warningTxTime {
		log.Error("[statedb_nosql:SetBucket] Insert", "tx", tx.Hash(), "tm", time.Since(st))
	}
	return nil
}

func (statedb *StateDB) GetKey(ctx context.Context, owner string, collection string, key string, withProof bool) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {
	if withProof {
		proof = new(NoSQLProof)
		_, ok, err := statedb.getOwnerSMT(ctx, owner, proof)
		if err != nil {
			log.Error("[statedb_nosql:GetKey] getOwnerSMT", "err", err)
			return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetKey] getOwnerSMT %s", err)
		} else if !ok {
			log.Error("[statedb_nosql:GetKey] NOT OK on getOwnerSMT", "owner", owner)
			return txhash, false, false, proof, nil
		}
	} else {
		proof = nil
	}

	smt, ok, err := statedb.getCollectionSMT(ctx, owner, collection, proof)
	if err != nil {
		log.Error("[statedb_nosql:GetKey] getCollectionSMT", "err", err)
		return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetKey] %s", err)
	} else if !ok {
		log.Error("[statedb_nosql:GetKey] NOT OK on getCollectionSMT", "owner", owner, "collection", collection)
		return txhash, false, false, proof, nil
	}

	k := KeyToAddress([]byte(key))

	// no proof to verify
	if proof == nil {
		txhashbytes, ok, deleted, _, err := smt.GetWithoutProof(ctx, k.Bytes())
		if err != nil {
			return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetKey] %s", err)
		}
		if !ok {
			log.Error("[statedb_nosql:GetKey] collection smt didn't get key", "owner", owner, "collection", collection, "key", k)
			return txhash, false, false, proof, nil
		}
		if deleted {
			return txhash, false, false, proof, nil
		}
		return common.BytesToHash(txhashbytes), ok, deleted, proof, nil
	}

	// proof to verify
	var keyproof *Proof
	txhashbytes, ok, deleted, keyproof, _, err := smt.Get(ctx, k.Bytes(), true)
	if !ok {
		log.Error("[statedb_nosql:GetKey] Get with Proof NOT OK", "key", key, "k", fmt.Sprintf("%x", k))
		return txhash, false, deleted, proof, nil
	} else if deleted {
		log.Warn("[statedb_nosql:GetKey] Get with Proof DELETED", "key", key, "k", fmt.Sprintf("%x", k))
	}
	proof.TxHash = common.BytesToHash(txhashbytes)
	proof.KeyProof = NewSerializedProof(keyproof)
	proof.Owner = owner
	proof.Collection = collection
	proof.Key = []byte(key)
	proof.KeyHash = k
	proof.CollectionChunkHash = smt.ChunkHash()
	proof.CollectionMerkleRoot = smt.MerkleRoot(ctx)
	proof.KeyChunkHash = statedb.keyStorage.ChunkHash()
	proof.KeyMerkleRoot = statedb.keyStorage.MerkleRoot(ctx)
	log.Trace("[statedb_nosql:GetKey] finalized", "CH", proof.KeyChunkHash, "MR", proof.KeyMerkleRoot, "owner", owner, "collection", collection, "key", key, "keyhash", proof.KeyHash)

	txhash = common.BytesToHash(txhashbytes)
	return txhash, ok, deleted, proof, nil
}

func (statedb *StateDB) GetIndexedKey(ctx context.Context, owner string, collection string, index string, key []byte, withProof bool) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {

	// check if owner exists
	if withProof {
		proof = new(NoSQLProof)
		_, ok, err := statedb.getOwnerSMT(ctx, owner, proof)
		if err != nil {
			log.Error("[statedb_nosql:GetIndexedKey] getOwnerSMT", "err", err)
			return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetIndexedKey] getOwnerSMT %s", err)
		} else if !ok {
			log.Error("[statedb_nosql:GetIndexedKey] NOT OK on getOwnerSMT", "owner", owner)
			return txhash, false, false, proof, nil
		}
	} else {
		proof = nil
	}

	// check if owner-collection exists
	smt, ok, err := statedb.getCollectionSMT(ctx, owner, collection, proof)
	if err != nil {
		log.Error("[statedb_nosql:GetIndexedKey] getCollectionSMT", "err", err)
		return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetIndexedKey] %s", err)
	} else if !ok {
		log.Error("[statedb_nosql:GetIndexedKey] NOT OK on getCollectionSMT", "owner", owner, "collection", collection)
		return txhash, false, false, proof, nil
	}

	// get owner-collection-index tree
	tree, ok, err := statedb.getIndexedCollectionTree(ctx, owner, collection, index, proof)
	if err != nil {
		return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetIndexedKey] %s", err)
	} else if !ok {
		log.Error("[statedb_nosql:GetIndexedKey] NOT OK on getIndexedCollectionTree", "owner", owner, "collection", collection, "index", index)
		return txhash, false, false, proof, nil
	}

	//k := KeyToAddress(key)
	k := IndexedKeyToAddress(key)

	// no proof to verify
	if proof == nil {
		txhashbytes, ok, deleted, _, err := tree.GetWithoutProof(k.Bytes())
		if err != nil {
			return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetIndexedKey] %s", err)
		}
		if !ok {
			log.Error("[statedb_nosql:GetIndexedKey] collection-index data tree didn't get key", "owner", owner, "collection", collection, "key", key, "keyaddr", k)
			return txhash, false, false, proof, nil
		}
		if deleted {
			return txhash, false, false, proof, nil
		}
		return common.BytesToHash(txhashbytes), ok, deleted, proof, nil
	}

	// proof to verify
	txhashbytes, ok, deleted, keyproof, _, err := tree.Get(k.Bytes())
	if !ok {
		log.Error("[statedb_nosql:GetIndexedKey] Get with Proof NOT OK", "key", key, "k", fmt.Sprintf("%x", k))
		return txhash, false, deleted, proof, nil
	} else if deleted {
		log.Warn("[statedb_nosql:GetIndexedKey] Get with Proof DELETED", "key", key, "k", fmt.Sprintf("%x", k))
	}
	proof.TxHash = common.BytesToHash(txhashbytes)
	proof.KeyProof = NewSerializedProof(keyproof.(*Proof))
	proof.Owner = owner
	proof.Collection = collection
	proof.Key = key
	proof.KeyHash = k
	proof.CollectionChunkHash = smt.ChunkHash()
	proof.CollectionMerkleRoot = smt.MerkleRoot(ctx)
	proof.KeyChunkHash = statedb.keyStorage.ChunkHash()
	proof.KeyMerkleRoot = statedb.keyStorage.MerkleRoot(ctx)
	log.Trace("[statedb_nosql:GetIndexedKey] finalized", "CH", proof.KeyChunkHash, "MR", proof.KeyMerkleRoot, "owner", owner, "collection", collection, "key", key, "keyhash", proof.KeyHash)

	txhash = common.BytesToHash(txhashbytes)
	return txhash, ok, deleted, proof, nil
}

func (statedb *StateDB) ScanIndexedCollection(ctx context.Context, owner string, collection string, index string, keyStart []byte, keyEnd []byte, limit int, withProof bool) (txhashes []common.Hash, ok bool, proof *NoSQLProof, err error) {
	// check if owner exists
	if withProof {
		proof = new(NoSQLProof)
		_, ok, err := statedb.getOwnerSMT(ctx, owner, proof)
		if err != nil {
			log.Error("[statedb_nosql:GetIndexedKey] getOwnerSMT", "err", err)
			return txhashes, false, proof, fmt.Errorf("[statedb_nosql:GetIndexedKey] getOwnerSMT %s", err)
		} else if !ok {
			log.Error("[statedb_nosql:GetIndexedKey] NOT OK on getOwnerSMT", "owner", owner)
			return txhashes, false, proof, nil
		}
	} else {
		proof = nil
	}

	// check if owner-collection exists
	//smt, ok ....
	_, ok, err = statedb.getCollectionSMT(ctx, owner, collection, proof)
	if err != nil {
		log.Error("[statedb_nosql:GetIndexedKey] getCollectionSMT", "err", err)
		return txhashes, false, proof, fmt.Errorf("[statedb_nosql:GetIndexedKey] %s", err)
	} else if !ok {
		log.Error("[statedb_nosql:GetIndexedKey] NOT OK on getCollectionSMT", "owner", owner, "collection", collection)
		return txhashes, false, proof, nil
	}

	// get owner-collection-index tree
	tree, ok, err := statedb.getIndexedCollectionTree(ctx, owner, collection, index, proof)
	if err != nil {
		return txhashes, false, proof, fmt.Errorf("[statedb_nosql:GetIndexedKey] %s", err)
	} else if !ok {
		log.Error("[statedb_nosql:GetIndexedKey] NOT OK on getIndexedCollectionTree", "owner", owner, "collection", collection, "index", index)
		return txhashes, false, proof, nil
	}

	kStart := IndexedKeyToAddress(keyStart)
	kEnd := IndexedKeyToAddress(keyEnd)

	// scanProof, _, ...
	_, _, txhRecs, err := tree.Scan(kStart.Bytes(), kEnd.Bytes(), limit)
	if err != nil {
		return txhashes, false, proof, fmt.Errorf("[statedb_nosql:GetIndexedKey] %s", err)
	}
	for _, txhBytes := range txhRecs {
		txhashes = append(txhashes, common.BytesToHash(txhBytes))
	}

	// no proof to verify
	if proof == nil {
		return txhashes, true, nil, nil
	}

	// proof to verify - TODO for these range proofs
	// proof.TxHash = common.BytesToHash(txhashbytes)
	// proof.KeyProof = NewSerializedProof(keyproof.(*Proof))
	// proof.Owner = owner
	// proof.Collection = collection
	// proof.Key = key
	// proof.KeyHash = k
	// proof.CollectionChunkHash = smt.ChunkHash()
	// proof.CollectionMerkleRoot = smt.MerkleRoot(ctx)
	// proof.KeyChunkHash = statedb.keyStorage.ChunkHash()
	// proof.KeyMerkleRoot = statedb.keyStorage.MerkleRoot(ctx)
	// log.Trace("[statedb_nosql:GetIndexedKey] finalized", "CH", proof.KeyChunkHash, "MR", proof.KeyMerkleRoot, "owner", owner, "collection", collection, "key", key, "keyhash", proof.KeyHash)
	//
	// txhash = common.BytesToHash(txhashbytes)
	return txhashes, true, proof, nil
}

func (statedb *StateDB) SetKey(ctx context.Context, tx *Transaction) (err error) {
	var txp TxKey
	deleted := false
	if string(tx.Method) == http.MethodDelete {
		txp.Amount = 0
		deleted = true
	} else {
		err = json.Unmarshal(tx.Payload, &txp)
		if err != nil {
			return err
		}
	}
	// the Collection SMT holds TxKeys defining the key-value pair mappings.
	k := KeyToAddress([]byte(tx.Key()))
	smt, _, err := statedb.getCollectionSMT(ctx, tx.Owner(), tx.Collection(), nil)
	if err != nil {
		log.Trace("[statedb_nosql:SetKey] getOrCreateCollectionSMT", "err", err)
		return err
	}
	err = smt.Insert(ctx, k.Bytes(), tx.Hash().Bytes(), txp.Amount, deleted)
	if err != nil {
		log.Error("[statedb_nosql:SetKey] Insert", "k", tx.Key(), "ERROR", err)
		return err
	}
	log.Trace("**** SetKey - Insert ****", "k", k, "v", tx.Hash())
	return nil
}

func (statedb *StateDB) SetIndexedKey(ctx context.Context, tx *Transaction) error {
	var txp TxKey
	deleted := false
	if string(tx.Method) == http.MethodDelete {
		txp.Amount = 0
		deleted = true
	} else {
		err := json.Unmarshal(tx.Payload, &txp)
		if err != nil {
			return err
		}
	}
	// the Collection AVLTree holds TxKeys defining the key-value pair mappings.
	tree, ok, err := statedb.getIndexedCollectionTree(ctx, tx.Owner(), tx.Collection(), txp.BucketIndexName, nil)
	if err != nil {
		return fmt.Errorf("[statedb_nosql:SetIndexedKey] %s", err)
	}
	if !ok {
		return fmt.Errorf("[statedb_nosql:SetIndexedKey] getIndexedCollectionTree NOT OK")
	}

	// set the valhash chunk
	if txp.Data != nil {
		_, err := statedb.Storage.QueueChunk(ctx, txp.Data)
		if err != nil {
			return fmt.Errorf("[statedb_nosql:SetIndexedKey] QueueChunk err: %s", err)
		}
		//dprint("[statedb_nosql:SetIndexedKey] setting valhash chunk: valhash(%x) for key(%v)", txp.ValHash, txp.Key)
		// if txp.ValHash != dataChunkID { // debug check
		// 	dprint("[statedb_nosql:SetIndexedKey] we have a problem. tx.Hash.Bytes(%x) != txp.ValHash(%x)", dataChunkID, txp.ValHash)
		// }
		txp.Data = nil // clear out the data field
		tx.Payload, err = json.Marshal(txp)
		if err != nil {
			return fmt.Errorf("[statedb_nosql:SetIndexedKey] %s", err)
		}
	}

	k := IndexedKeyToAddress(txp.Key)
	err = tree.Insert(ctx, k.Bytes(), tx.Hash().Bytes(), txp.Amount, deleted)
	if err != nil {
		return fmt.Errorf("[statedb_nosql:SetIndexedKeyAVL] %s", err)
	}
	log.Trace("**** SetIndexedKey - Insert ****", "k", k, "v", tx.Hash())

	return nil
}

func (statedb *StateDB) PatchKey(ctx context.Context, tx *Transaction) (err error) {
	return statedb.SetKey(ctx, tx)
}

func (statedb *StateDB) DeleteKey(ctx context.Context, tx *Transaction) (err error) {
	return statedb.SetKey(ctx, tx)
}

func (statedb *StateDB) ScanCollection(ctx context.Context, owner string, collection string, withProof bool) (res map[common.Address]common.Hash, ok bool, proof *NoSQLProof, err error) {
	if withProof {
		proof = new(NoSQLProof)
	}

	_, ok, errSMT := statedb.getOwnerSMT(ctx, owner, proof)
	if errSMT != nil {
		log.Error("[statedb_nosql:ScanCollection] getCollectionSMT", "err", err)
		return res, false, proof, errSMT
	} else if !ok {
		return res, false, proof, nil
	}

	smt, ok, errSMT := statedb.getCollectionSMT(ctx, owner, collection, proof)
	if errSMT != nil {
		log.Error("[statedb_nosql:ScanCollection] getCollectionSMT", "err", err)
		return res, false, proof, errSMT
	} else if !ok {
		return res, false, proof, nil
	}
	var collectionProofs map[common.Address]*Proof
	res, collectionProofs, err = smt.ScanAll(ctx, withProof)
	if err != nil {
		log.Error("[statedb_nosql:ScanCollection] ScanAll", "err", err)
		return res, false, proof, errSMT
	}

	if proof != nil {
		proof.Owner = owner
		proof.Collection = collection
		proof.ScanProofs = make([]*ScanProof, 0)
		for addr, txhash := range res {
			if keyproof, ok := collectionProofs[addr]; ok {
				sp := new(ScanProof)
				// sp.Key = key
				sp.TxHash = txhash
				sp.KeyHash = addr
				sp.KeyProof = NewSerializedProof(keyproof)
				proof.ScanProofs = append(proof.ScanProofs, sp)
			}
		}
	}
	return res, true, proof, err
}

// Given the owner string, we can get the SMT that holds the buckets the user owns.
// ScanAll is called to get all these buckets, and the scan proofs.
func (statedb *StateDB) GetBuckets(ctx context.Context, owner string, withProof bool) (res map[common.Address]common.Hash, ok bool, proof *NoSQLProof, err error) {
	if withProof {
		proof = new(NoSQLProof)
		proof.Owner = owner
	}

	smt, ok, err := statedb.getOwnerSMT(ctx, owner, proof)
	if err != nil {
		log.Error("[statedb_nosql:GetBuckets] getOwnerSMT", "err", err)
		return res, false, proof, fmt.Errorf("[statedb_nosql:GetBuckets] %s", err)
	} else if !ok {
		log.Error("[statedb_nosql:GetBuckets] getOwnerSMT NOT OK")
		return res, false, proof, nil
	}
	var collectionProofs map[common.Address]*Proof
	res, collectionProofs, err = smt.ScanAll(ctx, withProof)
	if err != nil {
		log.Error("[statedb_nosql:GetBuckets] ScanAll", "err", err)
		return res, false, proof, fmt.Errorf("[statedb_nosql:GetBuckets] %s", err)
	}

	if proof != nil {
		proof.ScanProofs = make([]*ScanProof, 0)
		for addr, txhash := range res {
			if keyproof, ok := collectionProofs[addr]; ok {
				sp := new(ScanProof)
				// sp.Key = key
				sp.TxHash = txhash
				sp.KeyHash = addr
				sp.KeyProof = NewSerializedProof(keyproof)
				proof.ScanProofs = append(proof.ScanProofs, sp)
			}
		}
	}
	return res, true, proof, err
}

func (statedb *StateDB) GetBucket(ctx context.Context, owner string, collection string, withProof bool) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {
	if withProof {
		proof = new(NoSQLProof)
	} else {
		proof = nil
	}

	ownerSMT, ok, err := statedb.getOwnerSMT(ctx, owner, proof)
	if err != nil {
		log.Error("[statedb_nosql:GetBucket] getOwnerSMT", "err", err)
		return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetBucket]  getOwnerSMT %s", err)
	} else if !ok {
		log.Error("[statedb_nosql:GetBucket] NOT OK on getOwnerSMT", "owner", owner)
		return txhash, false, false, proof, nil
	}

	collectionhash := CollectionHash(owner, collection)

	// proof does not exist
	if proof == nil {
		log.Trace("[statedb_nosql:GetBucket] GetWithoutProof", "owner", owner, "collection", collection, "withProof", withProof, "collectionhash", fmt.Sprintf("%x", collectionhash))
		txhashbytes, ok, deleted, _, err := ownerSMT.GetWithoutProof(ctx, collectionhash.Bytes())
		if err != nil {
			return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetBucket] %s", err)
		}
		if !ok {
			log.Error("[statedb_nosql:GetBucket] owner smt didn't get collection", "owner", owner, "collection", collection, "collectionhash", collectionhash)
			return txhash, false, false, proof, nil
		}
		return common.BytesToHash(txhashbytes), true, deleted, proof, nil
	}

	// proof exists
	var collectionproof *Proof
	txhashbytes, ok, deleted, collectionproof, _, err := ownerSMT.Get(ctx, collectionhash.Bytes(), withProof)
	if err != nil {
		return txhash, false, false, proof, fmt.Errorf("[statedb_nosql:GetBucket] %s", err)
	}
	if !ok {
		log.Error("[statedb_nosql:GetBucket] Get with Proof NOT OK", "collectionhash", fmt.Sprintf("%x", collectionhash))
		return txhash, false, deleted, proof, nil
	}
	if deleted {
		log.Warn("[statedb_nosql:GetBucket] Get with Proof DELETED", "collectionhash", fmt.Sprintf("%x", collectionhash))
	}
	proof.Owner = owner
	proof.Collection = collection
	proof.TxHash = common.BytesToHash(txhashbytes)
	proof.CollectionHash = collectionhash
	proof.CollectionProof = NewSerializedProof(collectionproof)
	proof.CollectionChunkHash = ownerSMT.ChunkHash()
	proof.CollectionMerkleRoot = ownerSMT.MerkleRoot(ctx)

	txhash = common.BytesToHash(txhashbytes)
	return txhash, true, deleted, proof, nil
}
