// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"

	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"

	// "math/rand"
	"net/http"
	"path"
	"testing"
	"time"
)

const (
	defaultConsensus  = "singlenode"
	defaultN          = 5
	defaultIterations = 10
	usePreemptive     = false
	useResume         = false
)

func getReceivingNode(consensusAlgorithm string, i int, offset int) (node int) {
	if consensusAlgorithm == consensusSingleNode {
		return 0
	}
	return (offset + i) % MAX_VALIDATORS
}

func testTransferConsensus(t *testing.T, wolk []*WolkStore, consensusAlgorithm string, N int) {

	// create N accounts
	txs := make([]*Transaction, N)
	privateKeys := make([]*ecdsa.PrivateKey, N)
	for ownernumber := 0; ownernumber < N; ownernumber++ {
		owner := fmt.Sprintf("owner%d", ownernumber)
		k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(owner)))
		privateKeys[ownernumber], _ = ethcrypto.HexToECDSA(k_str)
		addr := crypto.GetECDSAAddress(privateKeys[ownernumber])
		tx, err := NewTransaction(privateKeys[ownernumber], http.MethodPut, owner, NewTxAccount(owner, []byte("FakeRSAPublicKeyForAccounts")))
		if err != nil {
			t.Fatal(err)
		}

		// submit the SetName transaction
		node := getReceivingNode(consensusAlgorithm, ownernumber, 0)
		txhash, err := wolk[node].SendRawTransaction(tx)
		if err != nil {
			t.Fatal(err)
		}
		log.Trace("[backend_consensus_test:testTransferConsensus] **** SENT SETNAME TX ****", "receivingNode", node, "txhash", txhash, "tx", tx, "addr", addr)
		txs[ownernumber] = tx
	}
	log.Info("[backend_consensus_test:testTransferConsensus] SENT SETNAME TX", "len(txs)", len(txs))

	start := time.Now()
	for ownernumber := 0; ownernumber < N; {
		tx := txs[ownernumber]
		missing := 0
		for i, wolknode := range wolk {
			_, bn, _, ok, err := wolknode.GetTransaction(context.TODO(), tx.Hash())
			if err != nil || !ok {
				missing++
			} else if ok && bn > 0 {
				log.Info(fmt.Sprintf("[backend_consensus_test:testTransferConsensus] Included: tx(%x) in BLOCK %d [Node%d] %s", tx.Hash(), bn, i, time.Since(start)))
			} else {
				log.Trace(fmt.Sprintf("[backend_consensus_test:testTransferConsensus] Not included yet: tx(%x)  [Node%d] %s", tx.Hash(), i, time.Since(start)))
				missing++
			}
		}
		if missing == 0 {
			ownernumber++
		} else {
			time.Sleep(1000 * time.Millisecond)
		}
	}
	log.Info("[backend_consensus_test:testTransferConsensus] TX all found", "N", N)

	// do a getname call and see if it matches addr
	for ownernumber := 0; ownernumber < N; ownernumber++ {
		owner := fmt.Sprintf("owner%d", ownernumber)
		pk := privateKeys[ownernumber]
		addr := crypto.GetECDSAAddress(pk)
		count := 0
		for i, wolknode := range wolk {
			var options RequestOptions
			receivedaddr, ok, _, err := wolknode.GetName(owner, &options)
			if err != nil {
				t.Fatalf("GetName ERR %v", err)
			} else if ok {
				log.Trace("[backend_consensus_test:testTransferConsensus] GetName Success", "node", i, "name", owner)
			} else {
				log.Info("[backend_consensus_test:testTransferConsensus] GetName NOT FOUND", "node", i, "name", owner)
			}
			if bytes.Compare(addr.Bytes(), receivedaddr.Bytes()) != 0 {
				t.Fatalf("GetName INCORRECT addr %v != receivedaddr %v", addr, receivedaddr)
			} else {
				count++
			}
		}
		if count != len(wolk) {
			t.Fatalf("[backend_consensus_test:testTransferConsensus] INCONSISTENCY ACROSS nodes!")
		}
	}
	log.Info("[backend_consensus_test:testTransferConsensus] GetName CONSISTENT", "N", N)

	// TEST 1: owner0 is transferring 15 WOLK to addr1..addrN-1
	amount := uint64(15)
	txs = make([]*Transaction, N)
	balance_start := make([]uint64, N)
	addr0 := crypto.GetECDSAAddress(privateKeys[0])
	node := getReceivingNode(consensusAlgorithm, 0, 0)
	balance_start[0], _, _ = wolk[node].GetBalance(addr0, 0)
	for r := 1; r < N; r++ {
		node = getReceivingNode(consensusAlgorithm, r, 0)
		addr := crypto.GetECDSAAddress(privateKeys[r])
		balance_start[r], _, _ = wolk[node].GetBalance(addr, 0)
		tx, err := NewTransaction(privateKeys[0], http.MethodPost, path.Join(ProtocolName, "transfer"), NewTxTransfer(amount, fmt.Sprintf("owner%d", r)))
		if err != nil {
			t.Fatalf("NewTransaction %v", err)
		}
		// submit the Transfer transaction
		txhash, err := wolk[node].SendRawTransaction(tx)
		if err != nil {
			t.Fatal(err)
		}
		txs[r] = tx
		log.Info("[backend_consensus_test:testTransferConsensus] **** SENT TRANSFER TX ****", "receivingNode", node, "txhash", txhash, "tx", tx)
	}

	// CHECK for the presence of N-1 transactions across all nodes
	for r := 1; r < N; r++ {
		tx := txs[r]
		hastx := make(map[int]bool)
		for i, _ := range wolk {
			hastx[i] = false
		}
		start := time.Now()
		for done := false; !done; {
			missing := 0
			for i, wolknode := range wolk {
				if hastx[i] == false {
					_, bn, _, ok, err := wolknode.GetTransaction(context.TODO(), tx.Hash())
					if err != nil || !ok {
						log.Trace(fmt.Sprintf("[backend_consensus_test:testTransferConsensus] Not included yet: tx(%x)  [Node%d] %s", tx.Hash(), i, time.Since(start)))
						missing++
					} else if ok && bn > 0 {
						log.Trace(fmt.Sprintf("[backend_consensus_test:testTransferConsensus] Included: tx(%x) in BLOCK %d [Node%d] %s", tx.Hash(), bn, i, time.Since(start)))
						hastx[i] = true
					} else {
						missing++
					}
				}
			}
			log.Trace(fmt.Sprintf("[backend_consensus_test:testTransferConsensus] %d/%d nodes are *missing* transfer transaction %d [%s]\n", missing, MAX_VALIDATORS, r, time.Since(start)))
			if missing == 0 {
				done = true
			} else {
				time.Sleep(1000 * time.Millisecond)
			}
		}
	}

	// Check balances across all nodes
	for i, wolknode := range wolk {
		//  1. balance of addr0 should be 15 more
		balance0_end, ok, err := wolknode.GetBalance(addr0, 0)
		if err != nil {
			t.Fatalf("[backend_consensus_test:testTransferConsensus] GetBalance0 ERR %v", err)
		} else if ok {
			log.Trace("[backend_consensus_test:testTransferConsensus] GetBalance0 Success", "node", i, "addr", addr0, "balance", balance0_end)
		} else {
			log.Trace("[backend_consensus_test:testTransferConsensus] GetBalance NOT FOUND", "node", i, "addr", addr0)
		}
		expected_diff := int64(-amount) * int64(N-1)
		diff := int64(balance0_end) - int64(balance_start[0])
		if diff != expected_diff {
			t.Fatalf("[backend_consensus_test:testTransferConsensus] TX TEST Balance0 FAILURE %d != %d", diff, expected_diff)
		} else {
			log.Trace("[backend_consensus_test:testTransferConsensus] TX TEST Balance0 SUCCESS")
		}

		//  2. balance of addr1..N-1 should 15 less!
		count := 0
		for r := 1; r < N; r++ {
			addr := crypto.GetECDSAAddress(privateKeys[r])
			balance_end, ok1, err := wolknode.GetBalance(addr, 0)
			if err != nil {
				t.Fatalf("[backend_consensus_test:testTransferConsensus] GetBalance1 ERR %v", err)
			} else if !ok1 {
				log.Info("[backend_consensus_test:testTransferConsensus] GetBalance1 NOT FOUND", "node", i, "addr", addr)
			}
			diff := int64(balance_end) - int64(balance_start[r])
			if diff != int64(amount) {
				t.Fatalf("[backend_consensus_test:testTransferConsensus] TX TEST Node%d Balance%d FAILURE %d != %d", i, r, diff, amount)
			} else {
				log.Trace("[backend_consensus_test:testTransferConsensus] TX TEST GetBalance SUCCESS", "Node", i, "Balance", balance_end, "addr", addr)
				count++
			}
		}
		if count != N-1 {
			t.Fatalf("[backend_consensus_test:testTransferConsensus] Balance INCONSISTENCY ACROSS nodes!")
		}
		log.Info("[backend_consensus_test:testTransferConsensus] Balance CONSISTENT", "Node", i)
	}
}

func TestTransferConsensus(t *testing.T) {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, _, _, _, _ := newWolkPeer(t, defaultConsensus)
	testTransferConsensus(t, wolk, defaultConsensus, 8)
}

func testFatNameBlocksLoop(t *testing.T, wolk []*WolkStore, nAccounts int) {
	// TODO: repeat numIterations times
	txs := make([]*Transaction, nAccounts)
	txs_setname := make([]*Transaction, 0)
	for ownernumber := 0; ownernumber < nAccounts; ownernumber++ {
		//owner := fmt.Sprintf("owner%d", ownernumber)
		r, _ := generateRandomBytes(32)
		owner := fmt.Sprintf("owner%d%x", ownernumber, r)
		k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(owner)))
		pk, _ := ethcrypto.HexToECDSA(k_str)
		addr := crypto.GetECDSAAddress(pk)
		tx, err := NewTransaction(pk, http.MethodPut, owner, NewTxAccount(owner, []byte("FakeRSAPublicKeyForAccounts")))
		if err != nil {
			t.Fatal(err)
		}

		// submit the SetName transaction
		/*
			node := 0
				for i := 0; i < MAX_VALIDATORS; i++ {
					j := (ownernumber + 4 + i) % MAX_VALIDATORS
					if wolk[j].NumPeers() > 2 {
						node = j
						break
					}
				} */
		txs_setname = append(txs_setname, tx)
		txhash := tx.Hash()
		//txhash, err := wolk[node].SendRawTransaction(tx)
		if err != nil {
			t.Fatal(err)
		}
		log.Trace("[backend_consensus_test:testFatNameBlocksLoop] **** SENT SETNAME TX ****", "txhash", txhash, "tx", tx, "addr", addr)
		txs[ownernumber] = tx
	}

	node := 0
	err := wolk[node].SendRawTransactions(txs_setname)
	if err != nil {
		t.Fatal(err)
	}
	/*
		node := 0
		for _, tx := range txs_setname {
			wolk[node].addTransactionToPool(tx)
		}
	*/

	foundbtx := make(map[common.Hash]bool)
	for _, tx := range txs {
		foundbtx[tx.Hash()] = false
	}

	blocks := make(map[uint64]*Block)
	start := time.Now()
	for done := false; !done; {
		txhash := get_random_key(foundbtx)
		node := 0
		_, bn, _, ok, err := wolk[node].GetTransaction(context.TODO(), txhash)
		if err != nil || !ok {

		} else if bn > 0 {
			_, ok := blocks[bn]
			if !ok {
				b, found, err2 := wolk[node].GetBlockByNumber(bn)
				if err2 != nil {
					log.Error("getBlockNumber", "err", err)
				} else if found {
					blocks[bn] = b
					for _, tx2 := range b.Transactions {
						delete(foundbtx, tx2.Hash())
					}
					log.Info("[backend_consensus_test:testFatNameBlocksLoop] GOT BLOCK with tx", "bn", bn, "len(b.Transactions)", len(b.Transactions), "NEW len(foundbtx)", len(foundbtx))
				}
			}
		}
		if len(foundbtx) == 0 {
			done = true
			log.Info("[backend_consensus_test:testFatNameBlocksLoop] Finished.", "Node", node, "tm", time.Since(start))
		}
		if !done {
			if time.Since(start) > time.Second*600 {
				log.Error("[backend_consensus_test:testFatNameBlocksLoop] ABANDONED test")
				return
			}
			//log.Warn("[backend_consensus_test:testFatNameBlocksLoop] Did not find sample tx ...", "Node", node, "len(foundbtx)", len(foundbtx), "tm", time.Since(start))
			// time.Sleep(20 * time.Second)
		}
	}

}

func get_random_key(m map[common.Hash]bool) (h common.Hash) {
	for k, _ := range m {
		return k
	}
	return h
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, nil
}

func testFatNameBlocks(t *testing.T, wolk []*WolkStore, N int, numIterations int) {
	testFatNameBlocksLoop(t, wolk, N)
	for i := 0; i < numIterations; i++ {
		testFatNameBlocksLoop(t, wolk, N*N)
		testFatNameBlocksLoop(t, wolk, N*N*N)
		//time.Sleep(60 * time.Second)
	}

}

func TestFatName(t *testing.T) {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, _, _, _, _ := newWolkPeer(t, defaultConsensus)
	testFatNameBlocks(t, wolk, defaultN, 1)
}

func testFullConsensus(t *testing.T, consensusAlg string, N int, numIterations int) {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, _, _, _, _ := newWolkPeer(t, consensusAlg)

	testProvableNoSQL(t, wolk, 1, 1)
	// testFatNameBlocks(t, wolk, defaultN, numIterations)
	// testTransferConsensus(t, wolk, consensusAlg, N)
	// testProvableNoSQL(t, wolk, defaultN, numIterations)
}

func TestFullSingleNode(t *testing.T) {
	testFullConsensus(t, consensusSingleNode, defaultN, defaultIterations)
}

func TestFullAlgorand(t *testing.T) {
	testFullConsensus(t, consensusAlgorand, defaultN, defaultIterations)
}
