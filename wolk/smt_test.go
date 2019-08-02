// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	common "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

const (
	keysPerBlock = 1000
	storageType  = "remote"
)

func TestConcurrencySMT(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	offset := 0
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace%d", offset))
	pcs, err := NewMockStorage()
	if err != nil {
		t.Fatalf("NewLevelDBChunkStore %v", err)
	}
	smt0 := NewSparseMerkleTree(NumBitsAddress, pcs)

	// this gets recorded in blocks
	chunkHash := make(map[uint64]common.Hash)

	// make a bunch of keys
	nblocks := uint64(100)
	nkeys := nblocks * keysPerBlock
	key := make(map[uint64]common.Address)
	val := make(map[uint64]common.Hash)
	for i := uint64(0); i < nkeys; i++ {
		b := make([]byte, 20)
		rand.Read(b)
		key[i] = common.BytesToAddress(b)
		val[i] = common.BytesToHash(wolkcommon.Computehash(key[i].Bytes()))
	}
	maxkey := uint64(0)
	for block := uint64(0); block < nblocks; block++ {
		wg := new(sync.WaitGroup)
		st := time.Now()
		var failsMu sync.RWMutex
		fails := 0
		success := 0
		if block > 1 {
			for j := uint64(0); j < keysPerBlock; j++ {
				wg.Add(1)
			}
			// these reads are happening concurrently with the writes below
			go func(maxkey uint64) {
				for j := uint64(0); j < keysPerBlock; j++ {
					n := uint64(maxkey - keysPerBlock - j - 1)
					k := key[n]
					expectedv := val[n]
					var v []byte
					var ok bool
					v, ok, _, _, _, err = smt0.Get(context.TODO(), k.Bytes(), false)
					if err != nil {
						t.Fatalf("[smt_test:TestConcurrencySMT] Get ERR %v\n", err)
					} else if !ok {
						failsMu.Lock()
						fails++
						failsMu.Unlock()
						log.Error("[smt_test:TestConcurrencySMT] fail", "n", n, "k", k)
						t.Fatalf("ABANDON n=%d k=%x\n", n, k)
					} else if bytes.Compare(v, expectedv.Bytes()) != 0 {
						t.Fatalf("[smt_test:TestConcurrencySMT] Get(%x) incorrect value got %x but expected %x", k, v, expectedv)
					}
					failsMu.Lock()
					success++
					failsMu.Unlock()
					wg.Done()
				}
			}(maxkey)
		}
		// insert keysPerBlock keys
		for j := uint64(0); j < keysPerBlock; j++ {
			k := key[maxkey+j]
			v := val[maxkey+j]
			err = smt0.Insert(context.TODO(), k.Bytes(), v.Bytes(), uint64(32), false)
			if err != nil {
				log.Error(fmt.Sprintf("[smt_test:TestConcurrencySMT] Insert ERR %v\n", err))
			}
		}
		smt0.Flush(context.TODO(), wg, true)
		wg.Wait()
		chunkHash[block] = smt0.ChunkHash()
		failsMu.RLock()
		maxkey = maxkey + keysPerBlock
		log.Info("[smt_test:TestConcurrencySMT] New block", "block", block, "chunkHash", smt0.ChunkHash(), "success", success, "fails", fails, "maxkey", maxkey, "tm", time.Since(st))
		failsMu.RUnlock()
		smt0 = NewSparseMerkleTree(NumBitsAddress, pcs)
		smt0.Init(chunkHash[block])
	}
}

func genConfigs(t *testing.T) (cfg *cloud.Config, genesisConfig *GenesisConfig) {
	nAccounts := MAX_VALIDATORS // TODO: BUG FIX -- this is a requirement right now because counting is done by address signing proposals
	genesisFileName := "genesis-1.json"
	privateKeys := make([]*crypto.PrivateKey, nAccounts)
	pubKeys := make([]string, nAccounts)
	privateKeysECDSA := make([]*ecdsa.PrivateKey, nAccounts)
	pubKeysECDSA := make([]string, nAccounts)
	accounts := make(map[common.Address]Account)
	addr := make([]common.Address, nAccounts)
	for i := 0; i < nAccounts; i++ {
		keyString := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i))))
		var err error
		privateKeys[i], err = crypto.HexToPrivateKey(keyString)
		if err != nil {
			t.Fatalf("[backend_test:newWolkPeer] HexToEdwardsPrivateKey %v", err)
		}

		// TODO:
		// this wont work for p2p peering, unless you make a change in the registry[i].PubKey line ( which is Edwards based ) --
		// privateKeysECDSA[i], err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

		// this is necessary for p2p to work since the public keys in the genesis file match up with these determinstic keys --
		privateKeysECDSA[i], err = ethcrypto.HexToECDSA(keyString)
		if err != nil {
			t.Fatalf("[backend_test:newWolkPeer] HexToECDSA %v", err)
		}
		pubKeys[i] = fmt.Sprintf("%x", privateKeys[i].PublicKey())
		pubKeysECDSA[i] = fmt.Sprintf("%x", ethcrypto.CompressPubkey(&privateKeysECDSA[i].PublicKey))
		//addr[i] = crypto.PubkeyToAddress(privateKeys[i].PublicKey())
		addr[i] = crypto.GetECDSAAddress(privateKeysECDSA[i]) //MUST FIX
		accounts[addr[i]] = Account{Balance: uint64(100000 + i)}
	}
	HTTPPort := uint16(81)
	registry := make([]SerializedNode, MAX_VALIDATORS)
	for i := 0; i < MAX_VALIDATORS; i++ {
		dns := fmt.Sprintf("c%d.wolk.com", i)
		registry[i] = SerializedNode{
			Address:     addr[i],
			PubKey:      pubKeys[i],
			ValueInt:    uint64(10000 + i),
			ValueExt:    uint64(0),
			StorageIP:   dns,
			ConsensusIP: dns,
			Region:      1,
			HTTPPort:    HTTPPort,
		}
	}
	var err error
	networkID := int(1234)
	err = CreateGenesisFile(networkID, genesisFileName, accounts, registry)
	if err != nil {
		t.Fatalf("CreateGenesisFile err: %+v\n", err)
	}
	genesisConfig, err = LoadGenesisFile(genesisFileName)
	if err != nil {
		t.Fatalf("LoadGenesisFile err: %+v\n", err)
	}

	// create nodes

	str := fmt.Sprintf("%d", int32(time.Now().Unix()))

	//	registeredNode := registry[i]
	datadir := fmt.Sprintf("/tmp/storage%s/datadir", str)
	cfg = &cloud.DefaultConfig
	cfg.GenesisFile = genesisFileName

	cfg.DataDir = datadir
	cfg.HTTPPort = int(HTTPPort)
	cfg.NodeType = "storage"
	cfg.ConsensusIdx = 3
	cfg.ConsensusAlgorithm = "poa"
	cfg.Preemptive = false
	cfg.Provider = "leveldb"
	//cfg.Address = address
	cfg.OperatorKey = privateKeys[0]
	cfg.OperatorECDSAKey = privateKeysECDSA[0]
	return cfg, genesisConfig
}

func TestSMTProofs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	offset := 0
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace%d", offset))
	pcs, err := NewMockStorage()
	if err != nil {
		t.Fatalf("[smt_test:NewMockStorage] %v", err)
	}
	smt0 := NewSparseMerkleTree(NumBitsAddress, pcs)
	nkeys := 300
	key := make([]common.Address, nkeys)
	val := make([]common.Hash, nkeys)
	for i := 0; i < nkeys; i++ {
		str := fmt.Sprintf("k%d", i)
		key[i] = NameToAddress(str)
		val[i] = common.BytesToHash(wolkcommon.Computehash(key[i].Bytes()))
		err = smt0.Insert(context.TODO(), key[i].Bytes(), val[i].Bytes(), uint64(32), false)
		if err != nil {
			t.Fatalf("[smt_test:TestSMT] Insert ERR %v\n", err)
		}
	}
	var wg sync.WaitGroup
	smt0.Flush(context.TODO(), &wg, true)

	res, proofs, err := smt0.ScanAll(context.TODO(), true)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	mr := smt0.MerkleRoot(context.TODO())
	for addr, valHash := range res {
		if !proofs[addr].Check(valHash.Bytes(), mr.Bytes(), GlobalDefaultHashes, false) {
			t.Fatalf("FAIL valhash: %x key: %x -- MR: %x -- proof: %s\n", valHash.Bytes(), addr, mr, proofs[addr].String())
		}
	}
	ctx := context.TODO()
	for i := 0; i < nkeys; i++ {
		v1, found, _, proof, _, err := smt0.Get(ctx, key[i].Bytes(), true)
		if err != nil || !found {
			t.Fatalf("Get not found %x %v\n", key, err)
		}
		if bytes.Compare(val[i].Bytes(), v1) != 0 {
			t.Fatalf("k:%x v:%x INCORRECT\n", key, v1)
		}
		mr := smt0.MerkleRoot(ctx)
		if proof == nil {
			t.Fatalf("NO PROOF\n")
		}
		// sp := NewSerializedProof(proof)
		// str, _ := json.Marshal(sp)
		// fmt.Printf("MR %x Key: %x Val: %x PROOF: %s\n", mr, key[i], v1, string(str))
		checkproof := proof.Check(v1, mr.Bytes(), GlobalDefaultHashes, false)
		if !checkproof {
			t.Fatalf("CHECK PROOF ==> FAILURE\n")
		}
	}
}

func getWolkTestDB(t *testing.T) (ChunkStore, func()) {
	err := os.RemoveAll("/tmp/testleveldb/")
	if err != nil {
		tprint("ERR: remove ldb cache err: %s", err)
	}
	d, err := NewMockStorage()
	if err != nil {
		t.Fatal(err)
	}
	return d, func() {
		d.Close()
		os.RemoveAll("/tmp/testleveldb/test.db")
	}
}

func TestSMTBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	offset := 0
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace%d", offset))
	pcs, err := NewMockStorage()
	if err != nil {
		t.Fatalf("[smt_test:newUDPStorage] %v", err)
	}

	smt0 := NewSparseMerkleTree(NumBitsAddress, pcs)

	// this gets recorded in blocks
	chunkHash := make(map[uint64]common.Hash)

	// make a bunch of keys
	nkeys := uint64(keysPerBlock)
	key := make(map[uint64]common.Address)
	for i := uint64(0); i < nkeys; i++ {
		b := make([]byte, 20)
		rand.Read(b)
		key[i] = common.BytesToAddress(b)
	}

	// for each key, keep their versions
	ctx := context.TODO()
	nversions := uint64(20)
	kv := make(map[uint64]map[common.Address]common.Hash)
	for ver := uint64(0); ver < nversions; ver++ {
		st := time.Now()
		if ver > 0 {
			smt0 = NewSparseMerkleTree(NumBitsAddress, pcs)
			smt0.Init(chunkHash[ver-1])
		}
		kv[ver] = make(map[common.Address]common.Hash)
		wg := new(sync.WaitGroup)
		for i := uint64(0); i < nkeys; i++ {
			v := fmt.Sprintf("%x%d", ver, key[i].Bytes())
			kv[ver][key[i]] = common.BytesToHash(wolkcommon.Computehash([]byte(v)))
			err = smt0.Insert(ctx, key[i].Bytes(), kv[ver][key[i]].Bytes(), uint64(32), false)
			if err != nil {
				t.Fatalf("[smt_test:TestSMT] Insert ERR %v\n", err)
			}
		}

		smt0.Flush(ctx, wg, true)
		wg.Wait()
		//		smt0.Dump()
		chunkHash[ver] = smt0.ChunkHash()
		storageBytes, _ := smt0.StorageBytes()
		log.Info("TestSMT-Generated", "Version", ver, "Hash:", chunkHash[ver], "Merkle Root", smt0.MerkleRoot(ctx), "StorageBytes", storageBytes, "tm", time.Since(st))
		for i := uint64(0); i < nkeys; i++ {
			//go func(k common.Address) {
			k := key[i]
			expectedv := kv[ver][key[i]]
			var v []byte
			var ok bool
			v, ok, _, _, _, err = smt0.Get(ctx, k.Bytes(), false)
			if err != nil {
				t.Fatalf("[smt_test:TestSMT] Get ERR %v\n", err)
			} else if !ok {
				t.Fatalf("[smt_test:TestSMT] Get(%x) not ok\n", key[i])
			} else if bytes.Compare(v, expectedv.Bytes()) != 0 {
				t.Fatalf("[smt_test:TestSMT] Get(%x) incorrect value got %x but expected %x", k, v, expectedv)
			}

			//}(key[i])
		}
	}

	for ver := uint64(0); ver < nversions; ver++ {
		st := time.Now()
		smt0 = NewSparseMerkleTree(NumBitsAddress, pcs)
		smt0.Init(chunkHash[ver])
		log.Trace("TestSMT-Init", "Version", ver, "Hash:", chunkHash[ver])
		passes := 0

		for i := uint64(0); i < nkeys; i++ {
			k := key[i]
			v1, found, _, proof, storageBytes, err := smt0.Get(ctx, k.Bytes(), true)
			if err != nil {
				t.Fatalf("Get not found %x %v \n", k, err)
			} else if found {
				if bytes.Compare(kv[ver][k].Bytes(), v1) == 0 {
					passes++
					mr := smt0.MerkleRoot(ctx)
					requireProof := false
					if requireProof {
						if proof != nil {
							fmt.Printf("PROOF: %s\n", proof.String())
						} else {
							t.Fatalf("NO PROOF\n")
						}
						checkproof := proof.Check(v1, mr.Bytes(), GlobalDefaultHashes, false) // WAS: merkleRoot[ver].Bytes()
						if !checkproof {
							fmt.Printf("k:%x v:%x storageBytes:%d ver %d -- ", k, v1, storageBytes, ver)
							t.Fatalf("CHECK PROOF ==> FAILURE\n")
						}
					}
				} else {
					t.Fatalf("k:%x v:%x  INCORRECT\n", k, v1)
				}
			} else {
				t.Fatalf("TestSMT-Get(%x) NOT Found", k)
			}
		}
		log.Info("TestSMT-PASS", "Version", ver, "passes", passes, "nkeys", nkeys, "tm", time.Since(st))
	}
}

func TestComputeDefaultHashes(t *testing.T) {
	fmt.Printf("\nDefault Hash 0: [%x]", GlobalDefaultHashes[0])
	fmt.Printf("\nDefault Hash 1: [%x]", GlobalDefaultHashes[1])
	fmt.Printf("\nDefault Hash 2: [%x]", GlobalDefaultHashes[2])

	newTestHash := make([]byte, len(GlobalDefaultHashes[0])*2)
	copy(newTestHash, GlobalDefaultHashes[0])
	copy(newTestHash[len(GlobalDefaultHashes[0]):], GlobalDefaultHashes[0])

	hashWork := wolkcommon.Computehash(newTestHash)
	fmt.Printf("\n newTestHash as bytes: [%+v]", newTestHash)
	fmt.Printf("\n hashWork : as bytes [%+v], [%x]", hashWork, hashWork)
}
