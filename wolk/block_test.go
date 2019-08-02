// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
)

func TestQBlock(t *testing.T) {
	var u *Block

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey %v\n", err)
	}

	u = &Block{
		BlockNumber:   2,
		ParentHash:    common.BytesToHash(wolkcommon.Computehash([]byte("3"))),
		Seed:          wolkcommon.Computehash([]byte("4")),
		AccountRoot:   common.BytesToHash(wolkcommon.Computehash([]byte("5"))),
		RegistryRoot:  common.BytesToHash(wolkcommon.Computehash([]byte("6"))),
		StorageBeta:   9,
		BandwidthBeta: 10,
	}
	tx, err := NewTransaction(privateKey, http.MethodPost, path.Join("owner", "fruits", "apple"), NewTxKey(common.BytesToHash(wolkcommon.Computehash([]byte("green"))), 5))
	if err != nil {
		t.Fatalf("NewTransaction %v\n", err)
	}
	u.Transactions = append(u.Transactions, tx)

	encoded, _ := rlp.EncodeToBytes(u)

	var s Block
	err = rlp.Decode(bytes.NewReader(encoded), &s)
	if err != nil {
		t.Fatalf("Decode %v\n", err)
	}
	fmt.Printf("%s\n", s.String())
	// header checks: u == s && u == s2?
	if u.BlockNumber != s.BlockNumber {
		t.Fatalf("BlockNumber failure %d != %d", u.BlockNumber, s.BlockNumber)
	}
	if u.StorageBeta != s.StorageBeta {
		t.Fatalf("storageBeta failure %d != %d", u.StorageBeta, s.StorageBeta)
	}
	if u.BandwidthBeta != s.BandwidthBeta {
		t.Fatalf("bandwidthBeta failure %d != %d", u.BandwidthBeta, s.BandwidthBeta)
	}
	if bytes.Compare(u.Seed, s.Seed) != 0 {
		t.Fatalf("Seed failure %x != %x", u.Seed, s.Seed)
	}
	if bytes.Compare(u.AccountRoot.Bytes(), s.AccountRoot.Bytes()) != 0 {
		t.Fatalf("AccountRoot failure %x != %x", u.AccountRoot.Bytes(), s.AccountRoot.Bytes())
	}
	if bytes.Compare(u.RegistryRoot.Bytes(), s.RegistryRoot.Bytes()) != 0 {
		t.Fatalf("RegistryRoot failure %x != %x", u.RegistryRoot.Bytes(), s.RegistryRoot.Bytes())
	}

}
