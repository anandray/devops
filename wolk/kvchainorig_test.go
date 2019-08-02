// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	wiavl "github.com/wolkdb/cloudstore/wolk/iavl"
)

func TestIAVLOrigProof(t *testing.T) {
	ldbcache, closeldb := getTendermintTestDB(t)
	defer closeldb()

	kvtree := wiavl.NewMutableTree(ldbcache, 0)
	//var wg sync.WaitGroup
	for j := 0; j < 5; j++ {
		//k := []byte(fmt.Sprintf("%d", j))
		//v := wolkcommon.Computehash([]byte(fmt.Sprintf("v%d", j)))
		v := []byte(fmt.Sprintf("%d", j))
		k := wolkcommon.Computehash([]byte(v))
		tprint("insert k(%x) v(%x)", k, v)
		//err = kvtree.Insert(k, v, uint64(32), false)
		kvtree.Set(k, v)
	}
	//kvtree.Flush(&wg, true)
	kvtree.SaveVersion()
	//wg.Wait()
	root1 := kvtree.WorkingHash()
	//observedV, ok, _, proof, _, err := kvtree.GetWithProof(wolkcommon.Computehash([]byte("1")))
	observedV, proof, err := kvtree.GetWithProof(wolkcommon.Computehash([]byte("2")))
	if err != nil {
		t.Fatalf("%v", err)
	}

	str, _ := json.Marshal(proof)
	fmt.Printf("proofmarshal: %s\n", str)

	expectedV := []byte("2")
	if bytes.Compare(bytes.Trim(observedV, "\x00"), expectedV) != 0 {
		t.Fatalf("Val mismatch observed(%x) expected(%x)", observedV, expectedV)
	}
	log.Trace("SUCCESS Observed:", "observedV", observedV, "Proof:", proof)

	tprint("\n")
	tprint("...printing kvtree:")
	wiavl.PrintTree(kvtree.ImmutableTree)
	tprint("\n")
	tprint("...printing proof path:")
	tprint("%s", proof.String())
	tprint("\n")

	err = proof.Verify(root1)
	if err != nil {
		t.Fatalf("Verify ROOT ERR: %v", err)
	}
	err = proof.VerifyItem(wolkcommon.Computehash([]byte("2")), observedV)
	if err != nil {
		t.Fatalf("VerifyItem ERR %v", err)
	}

}
