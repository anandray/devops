// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// checkObject represents a checkID which is being cashed
type checkObject struct {
	checkID   common.Hash
	db        *StateDB
	blockHash common.Hash
}

//NewAccountObject creates an account object.
func NewCheckObject(db *StateDB, checkID common.Hash, blockHash common.Hash) *checkObject {
	return &checkObject{
		db:        db,
		checkID:   checkID,
		blockHash: blockHash,
	}
}

func (self *checkObject) SetBlockHash(blockHash common.Hash) {
	self.blockHash = blockHash
}

//
// Attribute accessors
//
func (self *checkObject) CheckID() common.Hash {
	return self.checkID
}

func (self *checkObject) BlockHash() common.Hash {
	return self.blockHash
}

func (self *checkObject) String() string {
	s := fmt.Sprintf("{\"checkID\": \"%x\", \"blockHash\":\"%x\"}", self.checkID, self.blockHash)
	return s
}

func CheckIDToAddress(checkID common.Hash) common.Address {
	str := fmt.Sprintf("%x", checkID.Bytes())
	str = str[:40]
	return common.HexToAddress(str)
}
