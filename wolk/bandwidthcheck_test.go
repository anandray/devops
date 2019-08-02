// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	//"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/wolkdb/cloudstore/crypto"
)

func TestBandwidthCheck(t *testing.T) {
	amount := uint64(1600501)
	nodeID := uint64(13)
	recipient := common.HexToAddress("0x3088666E05794d2498D9d98326c1b426c9950767")
	var pkey *crypto.PrivateKey
	pkey, err := crypto.HexToPrivateKey(crypto.TestPrivateKey)
	if err != nil {
		t.Fatalf("HexToPrivateKey err %v", err)
	}

	c := NewBandwidthCheck(nodeID, recipient, amount)
	err = c.SignCheck(pkey)
	if err != nil {
		t.Fatalf("SignTx err %v", err)
	}

	encoded, _ := rlp.EncodeToBytes(c)
	var c2 BandwidthCheck // []interface{}
	err = rlp.Decode(bytes.NewReader(encoded), &c2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if strings.Compare(c.String(), c2.String()) != 0 {
		t.Fatalf("RLP serialization mismatch")
	}
	fmt.Printf("check: %s\n", c.String())
}
