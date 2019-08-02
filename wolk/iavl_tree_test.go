// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"encoding/json"
	"fmt"
	rand "math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	tenderdb "github.com/tendermint/tendermint/libs/db"
	wolkcommon "github.com/wolkdb/cloudstore/common"
)

func TestProofIAVL(t *testing.T) {
	chunkstore, err := NewMockStorage()
	if err != nil {
		t.Fatalf("[smt_test:NewLevelDBStorage] %v", err)
	}
	rand.Seed(time.Now().UTC().UnixNano())

	iavltree := NewIAVLTree(chunkstore)
	for j := 0; j < 5; j++ {
		v := []byte(fmt.Sprintf("%d", j))
		k := wolkcommon.Computehash([]byte(v))
		sb := randUInt64(5, 32)
		tprint("insert k(%x) v(%x) sb(%v)", k, v, sb)

		err = iavltree.Insert(k, v, sb, false)
		if err != nil {
			tprint("[NewKVChain] Insert ERR %v", err)
		}
		tprint("\n")
		PrintTree(iavltree.mtree.ImmutableTree, false)
		tprint("\n")
	}
	tprint("done with inserts")
	//PrintTree(iavltree.mtree.ImmutableTree, true)

	tprint("flushing")
	var wg sync.WaitGroup
	iavltree.Flush(&wg, true)
	wg.Wait()

	tprint("tree:")
	PrintTree(iavltree.mtree.ImmutableTree, true)
	//tprint("tree's total storage bytes: %v", iavltree.StorageBytes())

	// reload the tree
	hash := iavltree.ChunkHash()
	tprint("tree hash: %x\n", hash)

	iavltree = NewIAVLTree(chunkstore)
	iavltree.Init(hash)

	// add a few more inserts
	for j := 5; j < 7; j++ {
		v := []byte(fmt.Sprintf("%d", j))
		k := wolkcommon.Computehash([]byte(v))
		sb := randUInt64(5, 32)
		tprint("insert k(%x) v(%x) sb(%v)", k, v, sb)
		err = iavltree.Insert(k, v, sb, false)
		if err != nil {
			tprint("[NewKVChain] Insert ERR %v", err)
		}
	}

	iavltree.Flush(&wg, true)
	wg.Wait()

	tprint("tree:")
	PrintTree(iavltree.mtree.ImmutableTree, true)

	expectedV := []byte("3")
	key := wolkcommon.Computehash(expectedV)
	tprint("Get key(%x) expectedval(%x) With Proof...", key, expectedV)
	tprint("keyhash(%x)", wolkcommon.Computehash(key))

	observedV, ok, _, proof, _, err := iavltree.GetWithProof(key)
	if err != nil {
		t.Fatalf("ERR %v", err)
	} else if !ok {
		t.Fatalf("NOT OK")
	}
	pr := proof.(*RangeProof)
	str, _ := json.Marshal(pr)
	fmt.Printf("proofmarshal: %s\n", str)
	if bytes.Compare(bytes.Trim(observedV, "\x00"), expectedV) != 0 {
		t.Fatalf("Val mismatch observed(%x) expected(%x)", observedV, expectedV)
	}
	tprint("SUCCESS Observedv(%x)", observedV)

	tprint("proof: \n %s", pr.String())

	tprint("Verify proof...")
	if iavltree.VerifyProof(key, common.BytesToHash(observedV), pr) {
		tprint("item verified!")
	} else {
		t.Fatal("item not verified")
	}
}

func getTendermintTestDB(t *testing.T) (tenderdb.DB, func()) {
	d, err := tenderdb.NewGoLevelDB("test", "/tmp/testleveldb")
	if err != nil {
		panic(err)
	}
	return d, func() {
		d.Close()
		os.RemoveAll("/tmp/testleveldb/test.db")
	}
}

// returns [min, max]
func randUInt64(min int, max int) uint64 {
	max = max + 1
	return uint64(min + rand.Intn(max-min))
}

// TODO:
// need an 'update tree' test: save version/iavltree, then insert.
// deletes / orphans
// root node saving?
