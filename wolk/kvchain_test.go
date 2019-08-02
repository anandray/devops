// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"context"
	"fmt"
	rand "math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
)

type KVTree interface {
	Insert(ctx context.Context, k []byte, v []byte, storageBytesNew uint64, deleted bool) error
	Init(hash common.Hash)
	Flush(context.Context, *sync.WaitGroup, bool) (ok bool, err error)
	GetWithProof(ctx context.Context, k []byte) (v []byte, found bool, deleted bool, p interface{}, storageBytes uint64, err error)
	ChunkHash() common.Hash
	MerkleRoot(ctx context.Context) common.Hash
}

type KVChain struct {
	blockNumber int
	maxkeys     int
	statedb     map[int]KVTree
	blockhashes map[int]common.Hash
	lastHash    common.Hash
	mu          sync.RWMutex
	active      bool
}

func NewKVChain(t *testing.T, writesPerBlock int, totalBlocks int, pcs ChunkStore, treeType string) *KVChain {
	chain := &KVChain{
		maxkeys:     0,
		statedb:     make(map[int]KVTree),
		blockhashes: make(map[int]common.Hash),
		blockNumber: 2,
		active:      true,
	}
	var h common.Hash
	chain.blockhashes[1] = h

	go func() {
		for b := 0; b < totalBlocks; b++ {
			if b > 5 {
				chain.mu.Lock()
				delete(chain.statedb, b-4)
				chain.mu.Unlock()
			}
			//tprint("start inserting into block number(%d)", chain.blockNumber)
			var kvtree KVTree

			chain.mu.RLock()
			if chain.active == false {
				return
			}
			h = chain.blockhashes[chain.blockNumber-1]
			if treeType == "smt" {
				kvtree = NewSparseMerkleTree(NumBitsAddress, pcs)
			} else if treeType == "iavl" {
				kvtree = NewAVLTree(pcs)
			} else {
				t.Fatal("no treetype")
			}
			var q common.Hash
			tprint("%s: bn(%d) maxkeys(%d) blockhash(%x)", treeType, chain.blockNumber, chain.maxkeys, h)
			if bytes.Compare(h.Bytes(), q.Bytes()) != 0 {
				kvtree.Init(h)
			}
			chain.mu.RUnlock()

			// add N more keys to the KVTree

			for i := 0; i < writesPerBlock; i++ {
				k := []byte(fmt.Sprintf("K%d", chain.maxkeys+i))
				//if treeType == "smt" {
				k = NameToAddress(string(k)).Bytes()
				//}
				v := wolkcommon.Computehash(k)
				//tprint("inserting k(%x) v(%x)", shortbytes(k), shortbytes(v))
				err := kvtree.Insert(context.TODO(), k, v, uint64(32), false)
				if err != nil {
					tprint("[NewKVChain] Insert ERR %v", err)
				}
				//tprint("after insert:")
				if treeType == "smt" {
					kvtree.(*SparseMerkleTree).Dump()
				} else if treeType == "iavl" {
					//kvtree.(*AVLTree).PrintTree(true, false)
					//kvtree.(*AVLTree).PrintTree(true, true)
				}
			}
			//kvtree.(*AVLTree).PrintTree(nil, true, true)
			st := time.Now()
			var wg sync.WaitGroup
			kvtree.Flush(context.TODO(), &wg, true)
			wg.Wait()
			log.Info("Flush", "tm", time.Since(st))
			//kvtree.(*SparseMerkleTree).Dump()

			// MINT a new block
			chain.mu.Lock()
			chain.maxkeys = chain.maxkeys + writesPerBlock
			tprint("Minted new block(%d) %d", chain.blockNumber, chain.maxkeys)
			chain.blockhashes[chain.blockNumber] = kvtree.ChunkHash()
			chain.statedb[chain.blockNumber] = kvtree
			chain.lastHash = kvtree.MerkleRoot(context.TODO())
			chain.blockNumber = chain.blockNumber + 1
			chain.mu.Unlock()
			//tprint("")
		}
	}()
	return chain
}

// KVChain has an actual read+write throughput for a given chunkstore and given provable keyvalue store (SparseMerkleTree, IAVLTree),
// but FIRST goal is *consistency* + *thread-safety* NOT to maximize throughput
// Once we know we have reliable consistency and thread-safety for all chunkstores, then we can optimize to increase throughput, but all optimizations should pass this test
func testKVChain(t *testing.T, treeType string, chunkstoreType string, writesPerBlock int, totalBlocks int) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	offset := 0
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace%d", offset))

	pcs, err := NewMockStorage()
	if err != nil {
		t.Fatalf("[kvchain_test:NewLevelDBChunkStore] %v", err)
	}
	totalReads := totalBlocks * writesPerBlock

	chain := NewKVChain(t, writesPerBlock, totalBlocks, pcs, treeType)
	notfound := 0
	done := false
	st_total := time.Now()
	waittime := 0
	for succ := 0; !done; {
		st := time.Now()
		nreads := writesPerBlock
		fails := 0
		chain.mu.RLock()
		maxkeys := chain.maxkeys
		chain.mu.RUnlock()
		if maxkeys > 0 {
			for i := 0; i < nreads; i++ {
				j := rand.Intn(maxkeys)
				k := []byte(fmt.Sprintf("K%d", j))
				//if treeType == "smt" {
				k = NameToAddress(string(k)).Bytes()
				//}

				expectedV := wolkcommon.Computehash(k)
				chain.mu.RLock()
				kvtree := chain.statedb[chain.blockNumber-1] // we have a pointer to an SMT/IAVLTree here
				lastHash := chain.lastHash
				chain.mu.RUnlock()
				observedV, ok, _, proof, storageBytes, err := kvtree.GetWithProof(context.TODO(), k)
				if err != nil {
					tprint("[kvchain_test:TestKVChain] GetKey ERR: %v", err)
				} else if !ok {
					notfound++
					//kvtree.(*SparseMerkleTree).Dump()
					kvtree.(*AVLTree).PrintTree(nil, false, true)
					t.Fatalf("[kvchain_test:TestKVChain] GetKey not found: %x", k)
				} else if bytes.Compare(observedV, expectedV) == 0 {
					verifyProof := false
					if verifyProof {
						switch pr := proof.(type) {
						case *RangeProof:
							//str, _ := json.Marshal(pr)
							//tprint("Found: *RangeProof [%d] %x %s\n", storageBytes, k, str) // typ.StringIndented("")
							err := pr.Verify(lastHash.Bytes())
							if err != nil {
								t.Fatalf("iavltree proof of root is false %x %v", lastHash, err)
							} else {
								err = pr.VerifyItem(k, observedV)
								if err != nil {
									t.Fatalf("iavltree proof is false %v", err)
								} else {
									succ++
								}
							}
						case *Proof:
							tprint("Found: *Proof [%d]\n", storageBytes)
							verified := pr.Check(observedV, lastHash.Bytes(), GlobalDefaultHashes, false)
							if !verified {
								//		t.Fatalf("SMT Verification failed")
							}
							succ++
						default:
							t.Fatalf("Unknown proof %v\n", pr)
						}
					} else {
						succ++
					}
				} else {
					fails++
					t.Fatalf("[kvchain_test:TestKVChain] FAIL: %x -- %x != %x", k, expectedV, observedV)
				}
				//time.Sleep(3 * time.Second)
			}
		}
		if succ+notfound > 0 {
			tprint("%s Read Time: %s\tReads:%d\tsuccess: %d\tnotfound:%d", treeType, time.Since(st), nreads, succ, notfound)
		}
		if succ >= totalReads {
			done = true
			chain.mu.Lock()
			chain.active = false
			chain.mu.Unlock()
			tprint("total test time(%s) total wait time(%v ms)", time.Since(st_total), waittime)
			return
		} else {
			time.Sleep(10 * time.Millisecond)
			//tprint("waited 10 ms")
			waittime += 10
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestKVChainIAVLLevelDB_1(t *testing.T) {
	testKVChain(t, "iavl", "leveldb", 10000, 10000000)
}

func TestKVChainIAVLLevelDB_2(t *testing.T) {
	testKVChain(t, "iavl", "leveldb", 100, 20)
}

func TestKVChainIAVLLevelDB_3(t *testing.T) {
	testKVChain(t, "iavl", "leveldb", 20, 1000)
}

func TestKVChainIAVLRemote(t *testing.T) {
	testKVChain(t, "iavl", "remote", 20, 100)
}

func TestKVChainSMTLevelDB(t *testing.T) {
	testKVChain(t, "smt", "leveldb", 20, 100)
}

func TestKVChainSMTUDP(t *testing.T) {
	testKVChain(t, "smt", "udp", 10, 10)
}

func TestKVChainSMTRemote(t *testing.T) {
	testKVChain(t, "smt", "remote", 1000, 1000)
}

// TODO:
// need an 'update tree' test: save version/kvtree, then insert.
// deletes / orphans
// root node saving?
