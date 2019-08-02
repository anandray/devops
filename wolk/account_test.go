// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
)

func TestAccount(t *testing.T) {
	testAcct := NewAccount()
	testAcct.Balance = uint64(TestBalance)

	fmt.Printf("testAcct: %v\n", testAcct)

	encoded, _ := rlp.EncodeToBytes(testAcct)
	fmt.Printf("RLP[u]: 0x%x\n", encoded)
	var a Account // []interface{}
	err := rlp.Decode(bytes.NewReader(encoded), &a)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("decode s: %v\n", a.String())

}
