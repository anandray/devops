// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// keyObject represents a nosql key which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a key object.
// the keys new state is modified through the object wby setting the Txhash.
// Finally, call Commit to write the modified token state into cloud storage database.
type keyObject struct {
	key common.Hash
	db  *StateDB
	tx  *Transaction
}

// newObject creates a state object.
func NewKeyObject(db *StateDB, key common.Hash, tx *Transaction) *keyObject {
	return &keyObject{
		db:  db,
		key: key,
		tx:  tx,
	}
}

func (s *keyObject) SetTx(tx *Transaction) {
	// log.Info("(self *keyObject) SetVal to", "val", val, "txhash", txhash)
	s.db.journal.AddEntry(keyChange{
		key:  s.key,
		prev: s.tx.Hash(),
	})
	s.tx = tx
}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (s *keyObject) Key() (index common.Hash) {
	if s != nil {
		index = s.key
	}
	return index
}

func (s *keyObject) Tx() (tx *Transaction) {
	if s != nil {
		tx = s.tx
	}
	return tx
}

func (s *keyObject) String() string {
	if s == nil {
		var empty common.Hash
		return fmt.Sprintf("key=%x val=%x txhash=%x", empty, empty, empty)
	}
	return fmt.Sprintf("key=%x txhash=%x", s.key, s.tx.Hash())
}
