// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// productsObject represents a wallet address which is being modified.
type productsObject struct {
	key     common.Address
	val     common.Hash
	db      *StateDB
	deleted bool
}

//NewProductsObjects creates an collections object.
func NewProductsObjects(db *StateDB, addr common.Address, tx *Transaction) *productsObject {
	return &productsObject{
		db:      db,
		key:     addr,
		val:     tx.Hash(),
		deleted: false,
	}
}

//
// Attribute accessors
//

func (s *productsObject) String() string {
	if s == nil {
		var empty common.Hash
		return fmt.Sprintf("key=%x val=%x txhash=%x", empty, empty, empty)
	}
	return fmt.Sprintf("val=%x", s.val)
}

func (self *productsObject) SetTx(tx *Transaction) {
	self.db.journal.entries = append(self.db.journal.entries, bucketsChange{
		val:  tx.Hash(),
		prev: self.val,
	})
	self.val = tx.Hash()
}
