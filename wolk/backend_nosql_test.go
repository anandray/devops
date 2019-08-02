// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
)

type testKey struct {
	owner      string
	collection string
	key        string
	val        string
}

func testNoSQLKeyWrite(t *testing.T, wolkStore *WolkStore, owner string, privateKey *ecdsa.PrivateKey, nIterations int, nKeys int) {
	// SetName
	tx := setName(t, owner, privateKey, wolkStore)
	log.Info("setName", "tx", tx.Hash().Hex())

	// SetBucket
	bucketName := "accounts"
	tx, err := NewTransaction(privateKey, http.MethodPost, path.Join(owner, bucketName), NewTxBucket(bucketName, BucketNoSQL, nil))
	if err != nil {
		t.Fatal(err)
	}
	doTransaction(t, wolkStore, tx)
	log.Info("[backend_nosql_test:testNoSQLKeyWrite] setBucket", "tx", tx.Hash().Hex())

	// GetBuckets
	txhashList, ok, _, err := wolkStore.GetBuckets(owner, 0, nil)
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatalf("could not GetBuckets, not OK! owner(%v)", owner)
	}
	// Compare bucket names
	matched := false
	for i := 0; i < len(txhashList); i++ {
		if bytes.Compare(txhashList[i].Bytes(), tx.Hash().Bytes()) == 0 {
			matched = true
		}
	}
	if matched == false {
		t.Fatalf("GetBuckets MISMATCH: none of actual == expected(%x)\n", tx.Hash().Bytes())
	}

	for iteration := (0); iteration < nIterations; iteration++ {
		st := time.Now()
		var txhashes []common.Hash
		var txs []*Transaction
		currentTS := time.Now().Unix()

		// SetKey for a few keys which are each changing every block
		for kn := 0; kn < nKeys; kn++ {
			rawKey := rawKeyString(t, kn, iteration, currentTS)
			rawVal := rawValBytes(t, kn, iteration, currentTS)

			valChunkHash := common.BytesToHash(wolkcommon.Computehash(rawVal))
			_, err := wolkStore.Storage.SetChunk(nil, rawVal)
			if err != nil {
				t.Fatal(err)
			}
			tx, err = NewTransaction(privateKey, http.MethodPost, path.Join(owner, bucketName, rawKey), NewTxKey(valChunkHash, uint64(len(rawVal))))
			if err != nil {
				t.Fatal(err)
			}
			// do not use doTransaction here b/c will want to blast a bunch of sends without waiting for minting
			txs = append(txs, tx)
			txhashes = append(txhashes, tx.Hash())
		}

		err = wolkStore.SendRawTransactions(txs)
		if err != nil {
			t.Fatal(err)
		}

		waitForTxsDone(t, txhashes, wolkStore)
		log.Trace("[backend_nosql_test:testNoSQLKeyWrite] iteration(%v). Set (%d) keys in %s\n", iteration, len(txhashes), time.Since(st))

		// GetKey for the keys just set
		success := 0
		for kn := 0; kn < nKeys; kn++ {
			rawKey := rawKeyString(t, kn, iteration, currentTS)
			expectedVal := rawValBytes(t, kn, iteration, currentTS)
			actualVal := getKey(t, owner, bucketName, wolkStore, rawKey)

			if !bytes.Equal(actualVal, expectedVal) {
				t.Fatalf("[backend_nosql_test:testNoSQLKeyWrite] iteration(%v) with key(%s) actualVal(%s) != expectedVal(%s)", iteration, rawKey, string(actualVal), string(expectedVal))
			} else {
				success++
			}
		}
		log.Info("[backend_nosql_test:testNoSQLKeyWrite]", "iteration", iteration, "Success", success, "nKeys", nKeys)

		// Delete all the keys of before!
		txs = make([]*Transaction, 0)
		txhashes = make([]common.Hash, 0)
		for kn := 0; kn < nKeys; kn++ {
			rawKey := rawKeyString(t, kn, iteration, currentTS)
			tx, err = NewTransaction(privateKey, http.MethodDelete, path.Join(owner, bucketName, rawKey), &TxKey{})
			if err != nil {
				t.Fatal(err)
			}
			// do not use doTransaction here b/c will want to blast a bunch of sends without waiting for minting
			txs = append(txs, tx)
			txhashes = append(txhashes, tx.Hash())
		}

		err = wolkStore.SendRawTransactions(txs)
		if err != nil {
			t.Fatal(err)
		}

		waitForTxsDone(t, txhashes, wolkStore)

		options := NewRequestOptions()
		for kn := 0; kn < nKeys; kn++ {
			rawKey := rawKeyString(t, kn, iteration, currentTS)
			txhash, ok, deleted, _, err := wolkStore.GetKey(owner, bucketName, rawKey, options)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatalf("getKey:GetKey key(%v) not OK", rawKey)
			}
			if !deleted {
				t.Fatalf("getKey:GetKey key(%v) **NOT** deleted", rawKey)
			}
			log.Info("[backend_nosql_test:testNoSQLKeyWrite] GetKey Delete check PASS", "txhash", txhash)
		}
		log.Trace("[backend_nosql_test:testNoSQLKeyWrite] iteration(%v). Deleted (%d) keys in %s\n", iteration, len(txhashes), time.Since(st))
	}

	// Delete Bucket
	tx, err = NewTransaction(privateKey, http.MethodDelete, path.Join(owner, bucketName), &TxBucket{})
	if err != nil {
		t.Fatal(err)
	}
	doTransaction(t, wolkStore, tx)
	log.Info("[backend_nosql_test:testNoSQLKeyWrite] Delete Bucket", "tx", tx.Hash().Hex())
	options := NewRequestOptions()
	txhash, ok, deleted, _, err := wolkStore.GetBucket(owner, bucketName, options)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("[backend_nosql_test:testNoSQLKeyWrite] GetBucket bucket(%v) not OK", bucketName)
	}
	if !deleted {
		t.Fatalf("[backend_nosql_test:testNoSQLKeyWrite] GetBucket bucket(%v) **NOT** deleted", bucketName)
	}
	log.Info("[backend_nosql_test:testNoSQLKeyWrite] GetBucket Delete check PASS", "txhash", txhash)

	return
}

func setName(t *testing.T, name string, privateKey *ecdsa.PrivateKey, wolkStore *WolkStore) *Transaction {
	tx, err := NewTransaction(privateKey, http.MethodPut, name, NewTxAccount(name, []byte("FakeRSAPublicKeyForAccounts")))
	if err != nil {
		t.Fatal(err)
	}
	doTransaction(t, wolkStore, tx)
	return tx
}

func getKey(t *testing.T, owner string, bucketName string, wolkStore *WolkStore, rawKey string) (val []byte) {
	options := NewRequestOptions()
	txhash, ok, deleted, _, err := wolkStore.GetKey(owner, bucketName, rawKey, options)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("getKey:GetKey key(%v) not OK", rawKey)
	}
	if deleted {
		t.Fatalf("getKey:GetKey key(%v) deleted", rawKey)
	}
	tx, _, _, _, err := wolkStore.GetTransaction(context.TODO(), txhash)
	if err != nil {
		t.Fatal(err)
	}
	txp, err := tx.GetTxKey()
	if err != nil {
		t.Fatal(err)
	}

	val, ok, err = wolkStore.Storage.GetChunk(context.TODO(), txp.ValHash.Bytes())
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatalf("[backend_nosql_test:getKey] GetChunk (%v) not ok", txp.ValHash)
	}
	return val
}

func rawKeyString(t *testing.T, nthKey int, nthIteration int, timestamp int64) (rawKey string) {
	return fmt.Sprintf("randomKey%d-%d_%d", nthKey, nthIteration, timestamp)
}
func rawValBytes(t *testing.T, nthKey int, nthIteration int, timestamp int64) (rawVal []byte) {
	return []byte(fmt.Sprintf("randomVal%d-%d_%d", nthKey, nthIteration, timestamp))
}

func checkTransferTxs(t *testing.T, wolkstore *WolkStore, sender string, ownerAddr map[string]common.Address, expectedBalance uint64, options *RequestOptions) (success int) {
	success = 0
	for owner, addr := range ownerAddr {
		balance, ok, err := wolkstore.GetBalance(addr, 0)
		if err != nil {
			t.Fatalf("[backend_nosql_test:checkTransferTxs] GetBalance %v", err)
		} else if !ok {
			t.Fatalf("[backend_nosql_test:checkTransferTxs] GetBalance not ok")
		} else if owner != "owner0" && balance != expectedBalance {
			//			t.Fatalf("[backend_nosql_test:checkTransferTxs] GetBalance MISMATCH %d", owner, balance)
			fmt.Printf("%s => %d\n", owner, balance)
		} else {
			success++
		}
	}
	return success
}

func TestNoSQLKeyWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Setup
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, privateKeys, _, _, _ := newWolkPeer(t, consensusSingleNode)
	testNoSQLKeyWrite(t, wolk[0], "owner0", privateKeys[0], 2, 1)
	testNoSQLKeyWrite(t, wolk[0], "owner1", privateKeys[1], 2, 10)
	// testNoSQLKeyWrite(t, wolk[0], "owner2", privateKeys[2], 2, 20)
	// testNoSQLKeyWrite(t, wolk[0], "owner3", privateKeys[3], 2, 40)
	// testNoSQLKeyWrite(t, wolk[0], "owner4", privateKeys[4], 2, 80)
	// testNoSQLKeyWrite(t, wolk[0], "owner3", privateKeys[3], 5, 1000)
}

func testProvableNoSQLLoop(t *testing.T, wolk []*WolkStore, nAccounts int, nCollectionsPerAccount int, nKeysPerCollection int) {
	// TODO: repeat numIterations times
	log.Info("[backend_consensus_test:testProvableNoSQLLoop] START", "nAccounts", nAccounts, "nCollectionsPerAccount", nCollectionsPerAccount, "nKeysPerCollection", nKeysPerCollection)

	// N setName transactions, N buckets, N keys, N transfers
	txs_setname := make([]*Transaction, 0)
	txs_bucket := make([]*Transaction, 0)
	txs_key := make([]*Transaction, 0)
	testbuckets := make([]testKey, 0)
	testkeys := make([]testKey, 0)
	ownerKeys := make([]*ecdsa.PrivateKey, nAccounts)
	owner_addr := make(map[string]common.Address)
	node := 0

	var err error
	for i := 0; i < nAccounts; i++ {
		owner := fmt.Sprintf("owner%d", i)
		k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(owner)))
		ownerKeys[i], err = ethcrypto.HexToECDSA(k_str)
		if err != nil {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] HexToECDSA %v", err)
		}

		owner_addr[owner] = crypto.GetECDSAAddress(ownerKeys[i])
		tx, err := NewTransaction(ownerKeys[i], http.MethodPut, owner, NewTxAccount(owner, []byte("FakeRSAPublicKeyForAccounts")))
		if err != nil {
			t.Fatal(err)
		}
		txs_setname = append(txs_setname, tx)

		for j := 0; j < nCollectionsPerAccount; j++ {
			collection := fmt.Sprintf("bucket%d-%d", i, j)
			b := testKey{owner: owner, collection: collection}
			testbuckets = append(testbuckets, b)
			tx, err = NewTransaction(ownerKeys[i], http.MethodPost, path.Join(owner, collection), NewTxBucket(collection, BucketNoSQL, nil))
			if err != nil {
				t.Fatalf("NewTransaction %v", err)
			}
			txs_bucket = append(txs_bucket, tx)
			for c := 0; c < nKeysPerCollection; c++ {
				k := testKey{owner: owner, collection: collection, key: fmt.Sprintf("key%d-%d-%d", i, j, c), val: fmt.Sprintf("val%d-%d-%d", i, j, c)}
				valChunkHash, err := wolk[node].Storage.SetChunk(context.TODO(), []byte(k.val))
				if err != nil {
					t.Fatal(err)
				}
				testkeys = append(testkeys, k)
				tx, err = NewTransaction(ownerKeys[i], http.MethodPost, path.Join(owner, collection, k.key), NewTxKey(valChunkHash, uint64(len(k.val))))
				if err != nil {
					t.Fatal(err)
				}
				txs_key = append(txs_key, tx)
			}
		}
	}

	options := new(RequestOptions)
	options.Proof = true
	// check before a block is minted
	node = 0

	if isPreemptive {
		t.Fatalf("Not implemented")
	}
	wolk[node].SendRawTransactions(txs_setname)
	nsecs := 30
	log.Info("[backend_consensus_test:testProvableNoSQLLoop] START txs_setname")
	successName := 0
	for i, tx := range txs_setname {
		done := false
		for r := 0; r < nsecs && !done; r++ {
			_, bn, _, ok, err := wolk[node].GetTransaction(context.TODO(), tx.Hash())
			if err != nil || !ok {
				log.Info(fmt.Sprintf("Not included yet: tx(%x)  [Node%d]", tx.Hash(), wolk[node].consensusIdx))
			} else if ok && bn > 0 {
				log.Trace(fmt.Sprintf("TRANSACTION Included in %d: tx(%x)  [Node%d] %s", bn, tx.Hash(), wolk[node].consensusIdx, tx.String()))
				err = checkSetName(wolk[node], testbuckets[i], tx, options)
				if err == nil {
					successName++
					done = true
				}
			} else {
				time.Sleep(1000 * time.Millisecond)
			}
		}
		if !done {
			t.Fatalf("HEY NOT DONE %d", i)
		}
	}
	log.Info("[backend_consensus_test:testProvableNoSQLLoop] FINISH txs_setname")

	// for each account, scan the account's buckets, which should be EMPTY!
	for i := 0; i < nAccounts; i++ {
		owner := fmt.Sprintf("owner%d", i)
		txhashList, ok, proof, err := wolk[node].GetBuckets(owner, 0, options)
		if err != nil {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] Err %v", err)
		}
		if !ok {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] Err %v", err)
		}
		if len(proof.ScanProofs) > 0 || len(txhashList) > 0 {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] GetBuckets Proof failure")
		}
		if bytes.Compare(proof.SystemChunkHash.Bytes(), getEmptySMTChunkHash()) != 0 {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] GetBuckets incorrect system chunkhash")
		}
		p := DeserializeProof(proof.SystemProof)
		if !p.Check(proof.SystemChunkHash.Bytes(), proof.KeyMerkleRoot.Bytes(), GlobalDefaultHashes, false) {
			//e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
			str, _ := json.Marshal(proof)
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] GetBuckets FAIL proof %s", string(str))
		}
	}

	wolk[node].SendRawTransactions(txs_bucket)
	successBucket := 0
	for i, tx := range txs_bucket {
		done := false
		for r := 0; r < nsecs && !done; r++ {
			_, bn, _, ok, err := wolk[node].GetTransaction(context.TODO(), tx.Hash())
			if err != nil || !ok {
				log.Info(fmt.Sprintf("Not included yet: tx(%x)  [Node%d]", tx.Hash(), wolk[node].consensusIdx))
			} else if ok && bn > 0 {
				log.Trace(fmt.Sprintf("[backend_consensus_test:testProvableNoSQLLoop] TRANSACTION Included in %d: tx(%x)  [Node%d] %s", bn, tx.Hash(), wolk[node].consensusIdx, tx.String()))
				err = checkSetBucket(wolk[node], testbuckets[i], tx, options)
				if err == nil {
					successBucket++
					done = true
				}
			} else {
				time.Sleep(1000 * time.Millisecond)
			}
		}
		if !done {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] NOT DONE %d", i)
		}
	}
	log.Info("[backend_consensus_test:testProvableNoSQLLoop] FINISH txs_bucket")

	for _, b := range testbuckets {
		_, ok, proof, err := wolk[node].ScanCollection(b.owner, b.collection, options)
		if err != nil {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] ScanCollection err %v", err)
		} else if !ok {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] ScanCollection not ok")
		}
		err = proof.VerifyScanProofs()
		if err != nil {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] ScanCollection Proof Fail %v", err)
		}
	}
	wolk[node].SendRawTransactions(txs_key)

	// allow time for gossiping, producing a block
	successKeys := 0
	for i, tx := range txs_key {
		done := false
		for r := 0; r < nsecs && !done; r++ {
			_, bn, _, ok, err := wolk[node].GetTransaction(context.TODO(), tx.Hash())
			if err != nil || !ok {
				log.Info(fmt.Sprintf("Not included yet: tx(%x)  [Node%d]", tx.Hash(), wolk[node].consensusIdx))
			} else if ok && bn > 0 {
				err = checkSetKey(wolk[node], testkeys[i], tx, options)
				if err == nil {
					log.Trace(fmt.Sprintf("TRANSACTION Included in %d: tx(%x)  [Node%d] %s", bn, tx.Hash(), wolk[node].consensusIdx, tx.String()))
					successKeys++
					done = true
				}
			} else {
				time.Sleep(1000 * time.Millisecond)
			}
		}
		if !done {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] NOT DONE %d", i)
		}
	}

	// for each account, scan the account's buckets
	for i := 0; i < nAccounts; i++ {
		owner := fmt.Sprintf("owner%d", i)
		txhashList, ok, proof, err := wolk[node].GetBuckets(owner, 0, options)
		if err != nil {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] Err %v", err)
		} else if !ok {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] Err %v", err)
		}
		err = proof.VerifyScanProofs()
		if err != nil {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] GetBuckets(%s) VerifyScanProofs Err %s", owner, err)
		}
		if len(txhashList) != nCollectionsPerAccount {
			t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] GetBuckets(%s) txhashlist count incorrect (expected %d, got %d)", owner, nCollectionsPerAccount, len(txhashList))
		}
	}

	// for each account bucket, scan all keys
	for i := 0; i < nAccounts; i++ {
		owner := fmt.Sprintf("owner%d", i)
		for j := 0; j < nCollectionsPerAccount; j++ {
			collection := fmt.Sprintf("bucket%d-%d", i, j)
			txhashList, ok, proof, err := wolk[node].ScanCollection(owner, collection, options)
			if err != nil {
				t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] Err %v", err)
			} else if !ok {
				t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] Not OK")
			}
			err = proof.VerifyScanProofs()
			if err != nil {
				t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] ScanCollection(%s, %s) VerifyScanProofs Err %s", owner, collection, err)
			}
			if len(txhashList) != nKeysPerCollection {
				t.Fatalf("[backend_consensus_test:testProvableNoSQLLoop] ScanCollection(%s, %s) txhashlist count incorrect (expected %d, got %d)", owner, collection, nKeysPerCollection, len(txhashList))
			}
		}
	}
}

func checkSetName(wolkstore *WolkStore, o testKey, tx *Transaction, options *RequestOptions) (err error) {
	_, ok, p, err := wolkstore.GetName(o.owner, options)
	if err != nil {
		return fmt.Errorf("[backend_nosql_test:checkSetName] GetName %v", err)
	} else if !ok {
		return fmt.Errorf("[backend_nosql_test:checkSetName] GetName not ok")
	}
	pr := DeserializeProof(p.Proof)
	if !pr.Check(p.TxHash.Bytes(), p.MerkleRoot.Bytes(), GlobalDefaultHashes, false) {
		return fmt.Errorf("[backend_nosql_test:checkSetName] proof failure")
	}
	return nil
}

func checkSetBucket(wolkstore *WolkStore, b testKey, tx *Transaction, options *RequestOptions) (err error) {
	txhash, ok, _, nosqlproof, err := wolkstore.GetBucket(b.owner, b.collection, options)
	if err != nil {
		return fmt.Errorf("[backend_nosql_test:checkBucketsTx] GetBucket %v", err)
	} else if !ok {
		return fmt.Errorf("[backend_nosql_test:checkBucketsTx] GetBucket not ok")
	}
	pr := DeserializeProof(nosqlproof.CollectionProof)
	if !pr.Check(nosqlproof.TxHash.Bytes(), nosqlproof.CollectionMerkleRoot.Bytes(), GlobalDefaultHashes, false) {
		owner := b.owner
		collection := b.collection
		key := CollectionHash(owner, collection)
		fmt.Printf("PROOF: %s\n", pr.String())
		fmt.Printf("PROOF KEY: %s %s => %x =?= %x\n", owner, collection, key, pr.Key)
		fmt.Printf("PROOF BITS: %x\n", pr.ProofBits)
		fmt.Printf("VAL: %x %x\n", nosqlproof.TxHash.Bytes(), txhash)
		fmt.Printf("MR: %x\n", nosqlproof.CollectionMerkleRoot)
		return fmt.Errorf("[backend_nosql_test:checkBucketsTx] collectionproof failure")
	}
	return nil
}

func checkSetKey(wolkstore *WolkStore, k testKey, tx *Transaction, options *RequestOptions) (err error) {
	// TODO: check for existence of key with GetKey, ScanAll
	_, ok, _, nosqlproof, err := wolkstore.GetKey(k.owner, k.collection, k.key, options)
	if err != nil {
		return fmt.Errorf("[backend_nosql_test:checkKeysTx] GetKey %v", err)
	} else if !ok {
		return fmt.Errorf("[backend_nosql_test:checkKeysTx] GetKey not ok")
	}
	err = nosqlproof.Verify()
	if err != nil {
		log.Error("verifyNoSQLProof", "err", err)
	}
	return nil
}

func testProvableNoSQL(t *testing.T, wolk []*WolkStore, N int, numIterations int) {
	for i := 0; i < numIterations; i++ {
		testProvableNoSQLLoop(t, wolk, 1, 1, 1)
	}
}

func TestProvableNoSQL(t *testing.T) {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, _, _, _, _ := newWolkPeer(t, consensusSingleNode)
	testProvableNoSQL(t, wolk, defaultN, 1)
}
