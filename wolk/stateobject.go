// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/wolkdb/cloudstore/log"
)

// stateObject represents a key object which is being modified in Cloudstore
type stateObject struct {
	key          common.Hash
	db           *StateDB
	val          common.Hash
	storageBytes uint64
}

// newObject creates a state object.
func NewStateObject(db *StateDB, key common.Hash, val common.Hash) *stateObject {
	so := &stateObject{
		db:  db,
		key: key,
		val: val,
	}
	db.sql.stateObjects[key] = so
	return so
}

// TODO: need to manage storageBytes in these calls
func (self *stateObject) SetOwner(val common.Hash) {
	log.Info("[stateobject:SetOwner] set owner to", "val", val.Hex())
	self.db.journal.AddEntry(ownerChange{
		key:               self.key,
		prevOwnerRootHash: self.val,
	})

	self.val = val
}

func (self *stateObject) SetDatabase(val common.Hash) {
	log.Info("[stateobject:SetDatabase] set db to", "val", val.Hex())
	self.db.journal.AddEntry(databaseChange{
		key:                  self.key,
		prevDatabaseRootHash: self.val,
	})

	self.val = val
}

func (self *stateObject) SetTable(val common.Hash) {
	log.Info("[stateobject:SetTable] set table to", "val", val.Hex())
	self.db.journal.AddEntry(tableChange{
		key:               self.key,
		prevTableRootHash: self.val,
	})

	self.val = val
}

// EncodeRLP implements rlp.Encoder.
func (self *stateObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, self.val)
}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (self *stateObject) Key() common.Hash {
	return self.key
}

func (self *stateObject) Val() common.Hash {
	return self.val
}

func (self *stateObject) String() string {
	return fmt.Sprintf("key=%x val=%x", self.key, self.val)
}

func (self *stateObject) StorageBytesOld() uint64 {
	return self.storageBytes
}

func (self *stateObject) StorageBytes() uint64 {
	return self.storageBytes + 10000
}
