// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/log"
)

type IAVLTree struct {
	ChunkStore ChunkStore
	mtree      *MutableTree
}

func NewIAVLTree(cs ChunkStore) *IAVLTree {
	var tree IAVLTree
	tree.ChunkStore = cs
	tree.mtree = NewMutableTree(cs, 0)
	return &tree
}

func (tree *IAVLTree) Init(hash common.Hash) {
	//tree.root.SetHash(hash.Bytes())
	if EmptyBytes(hash.Bytes()) || len(hash.Bytes()) == 0 {
		return
	}
	tree.mtree = LoadMutableTree(tree.ChunkStore, hash.Bytes(), 0)
}

func (tree *IAVLTree) Flush(wg *sync.WaitGroup, writeToCloudstore bool) (err error) {
	hash, err := tree.mtree.SaveTree(wg, writeToCloudstore)
	if err != nil {
		return fmt.Errorf("[iavltree:Flush] %s", err)
	}
	dprint("[iavl_tree:Flush] hash(%x)", hash)
	return nil
}

func (tree *IAVLTree) Delete(k []byte) error {
	//tree.root.delete(k, 0)
	value, removed := tree.mtree.Remove(k)
	if len(value) == 0 || removed == false {
		return fmt.Errorf("[iavltree:Delete] key(%x) not removed", k)
	}
	return nil
}

// NOTE: all Get's assume latest version of mtree
// TODO: storageBytes, like in smt
func (tree *IAVLTree) Get(k []byte) (v []byte, found bool, deleted bool, p *RangeProof, storageBytes uint64, err error) {

	valReturned, proof, err := tree.mtree.GetWithProof(k)
	if err != nil {
		return v, found, deleted, p, storageBytes, fmt.Errorf("[iavltree:Get] %s", err)
	}

	// valReturned is NIL if not found or deleted
	emptyVal := false
	if EmptyBytes(valReturned) || len(valReturned) == 0 {
		emptyVal = true
	}
	// TODO: check this:
	if proof == nil && emptyVal { // val not found
		//log.Info("[iavltree:Get] Val not found", "key", k, "val", v)
		return v, false, false, p, storageBytes, nil
	}
	// TODO: check this:
	if emptyVal { // val is deleted (?)
		return v, false, true, proof, storageBytes, nil
	}

	return valReturned, true, false, proof, storageBytes, nil

}

func (tree *IAVLTree) GetWithProof(k []byte) (v []byte, found bool, deleted bool, p interface{}, storageBytes uint64, err error) {
	return tree.Get(k)
}

func (tree *IAVLTree) ScanAll(withProof bool) (res map[common.Address]common.Hash, proof *RangeProof, err error) {

	var keyLeaves [][]byte
	var valLeaves [][]byte
	limit := 0 // this means no limit
	count := 0
	tree.mtree.Iterate(func(key []byte, val []byte) (stop bool) {
		if limit > 0 && count >= limit {
			return true
		}
		keyLeaves = append(keyLeaves, key)
		valLeaves = append(valLeaves, val)
		count++
		return false
	})

	if len(keyLeaves) == 0 || len(valLeaves) == 0 || len(keyLeaves) != len(valLeaves) {
		return res, proof, fmt.Errorf("[iavltree:ScanAll] (%d) key leaves and (%d) val leaves", len(keyLeaves), len(valLeaves))
	}

	if withProof {
		_, _, proof, err = tree.mtree.GetRangeWithProof(keyLeaves[0], keyLeaves[len(keyLeaves)], limit) // the 0 is the limit, which means no limit here
		if err != nil {
			return res, proof, fmt.Errorf("[iavltree:ScanAll] %s", err)
		}
	}

	// reformat to fit res - will lose ordering in the map
	for i := 0; i < len(keyLeaves); i++ {
		res[common.BytesToAddress(keyLeaves[i])] = common.BytesToHash(valLeaves[i])
	}
	return res, proof, nil
}

// if limit is 0, no limit: all values found will be returned
func (tree *IAVLTree) GetRange(startkey []byte, endkey []byte, limit int) (keys [][]byte, values []common.Hash, proof *RangeProof, err error) {

	keys, values_bytes, proof, err := tree.mtree.GetRangeWithProof(startkey, endkey, limit)
	if err != nil {
		return keys, values, proof, fmt.Errorf("[iavltree:GetRange] %s", err)
	}
	if proof == nil {
		log.Info("[iavltree:GetRange] proof is nil. Continuing.", "starkey", startkey, "endkey", endkey, "limit", limit)
		// return keys, values, proof, fmt.Errorf("[iavltree:GetRange] %s", err)
	}
	for i := 0; i < len(values_bytes); i++ {
		values = append(values, common.BytesToHash(values_bytes[i]))
	}
	return keys, values, proof, nil

}

func (tree *IAVLTree) StorageBytes() (b uint64) {
	//return tree.root.getStorageBytes(tree.ChunkStore)
	return tree.mtree.root.getStorageBytesTotal()
}

func (tree *IAVLTree) Insert(k []byte, v []byte, storageBytesNew uint64, deleted bool) (err error) {
	// check if this is what is supposed to happen if deleted is true
	if deleted {
		err = tree.Delete(k)
		if err != nil {
			return fmt.Errorf("[iavl_tree:Insert] %s", err)
		}
	}

	keyExisted := tree.mtree.Set(k, v, storageBytesNew)
	if !keyExisted {
		//log.Info("[iavltree:Insert] key didn't exist before", "key", hex.EncodeToString(k))
	} else {
		log.Info("[iavltree:Insert] key existed, and is overwritten", "key", hex.EncodeToString(k), "val", v)
		// can do stuff here: if we don't want to overwrite, rtn err
	}
	return nil
}

// PrintTree from: https://github.com/tendermint/iavl/blob/master/util.go#L10
// func (tree *IAVLTree) Dump() {
// 	latestVer, _ := tree.mtree.Load()
// 	treeLatestVer, _ := tree.mtree.GetImmutable(latestVer)
// 	PrintTree(treeLatestVer, true)
// }

// duplicate of ChunkHash() ?
func (tree *IAVLTree) Hash() common.Hash {
	return tree.ChunkHash()
}

// duplicate of Hash() ?
func (tree *IAVLTree) ChunkHash() common.Hash {
	//return common.BytesToHash(self.root.chunkHash)
	return common.BytesToHash(tree.mtree.WorkingHash())
}

func (tree *IAVLTree) MerkleRoot() common.Hash {
	return tree.ChunkHash()
}

func (tree *IAVLTree) Root() common.Hash {
	return common.BytesToHash(tree.mtree.WorkingHash()) // this is current working root, which may not be the same as last saved version root
}

// verify functions - also assume latest version
func (tree *IAVLTree) VerifyProof(key []byte, valHash common.Hash, proof *RangeProof) bool {

	err := proof.Verify(tree.mtree.WorkingHash())
	if err != nil {
		log.Error("[iavltree:VerifyProof] proof of root is false", "info", err)
		return false
	}
	//dprint("[kvtree:VerifyProof] proof of root verified!")
	err = proof.VerifyItem(key, valHash.Bytes())
	if err != nil {
		log.Error("[iavltree:VerifyProof] proof of item is false", "info", err, "key", hex.EncodeToString(key), "val", valHash)
		return false
	}
	return true
}

func (tree *IAVLTree) VerifyRangeProof(keys [][]byte, vals []common.Hash, proof *RangeProof) bool {
	err := proof.Verify(tree.mtree.WorkingHash())
	if err != nil {
		log.Info("[iavltree:VerifyRangeProof] proof of root is false", "info", err)
		return false
	}
	for i := 0; i < len(keys); i++ {
		err = proof.VerifyItem(keys[i], vals[i].Bytes())
		if err != nil {
			log.Info("[iavltree:VerifyRangeProof] proof is false", "info", err)
			return false
		}
	}
	return true
}

// verifies proof for the absence of an item
func (tree *IAVLTree) VerifyAbsence(key []byte, proof *RangeProof) bool {
	err := proof.Verify(tree.mtree.WorkingHash())
	if err != nil {
		log.Info("[iavltree:VerifyAbsence] proof of root is false", "info", err)
		return false
	}
	err = proof.VerifyAbsence(key)
	if err != nil {
		log.Info("[iavltree:VerifyAbsence] proof is false", "info", err)
		return false
	}
	return true
}
