// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type journalEntry interface {
	undo(*StateDB)
}

//reverting tokenstate journal change. blockNumber is rolled-back
type (
	balanceChange struct {
		account *common.Address
		prev    uint64
	}
	storageIPChange struct {
		index uint64
		prev  string
	}
	consensusIPChange struct {
		index uint64
		prev  string
	}
	valueIntChange struct {
		index uint64
		prev  uint64
	}
	valueExtChange struct {
		index uint64
		prev  uint64
	}
	regionChange struct {
		index uint64
		prev  uint8
	}
	addressChange struct {
		index uint64
		prev  common.Address
	}
	quotaChange struct {
		account *common.Address
		prev    uint64
	}
	shimURLChange struct {
		account *common.Address
		prev    string
	}
	rsaPublicKeyChange struct {
		account *common.Address
		prev    []byte
	}
	usageChange struct {
		account *common.Address
		prev    uint64
	}
	httpPortChange struct {
		index uint64
		prev  uint16
	}
	// Changes to individual accounts.
	keyChange struct {
		key  common.Hash
		prev common.Hash
	}
	bucketsChange struct {
		val  common.Hash
		prev common.Hash
	}
	ownerChange struct {
		key               common.Hash
		prevOwnerRootHash common.Hash
	}
	storeClaimsChange struct {
		account *common.Address
		prev    uint64
	}
	databaseChange struct {
		key                  common.Hash
		prevDatabaseRootHash common.Hash
	}

	tableChange struct {
		key               common.Hash
		prevTableRootHash common.Hash
	}
)

type journal struct {
	entries []journalEntry // Current changes tracked by the journal
}

// newJournal create a new initialized journal.
func newJournal() *journal {
	return &journal{
		entries: make([]journalEntry, 0),
	}
}

func (j *journal) Len() int {
	return len(j.entries)
}

func (j *journal) AddEntry(e journalEntry) {
	j.entries = append(j.entries, e)
}

func (ch balanceChange) undo(s *StateDB) {
	account, _, _ := s.getAccountObject(context.TODO(), *ch.account)
	account.SetBalance(ch.prev)
}

func (ch addressChange) undo(s *StateDB) {
	// TODO
}

func (ch storageIPChange) undo(s *StateDB) {
	// TODO
}

func (ch consensusIPChange) undo(s *StateDB) {
	// TODO
}

func (ch valueIntChange) undo(s *StateDB) {
	// TODO
}

func (ch valueExtChange) undo(s *StateDB) {
	// TODO
}

func (ch regionChange) undo(s *StateDB) {
	// TODO
}

func (ch quotaChange) undo(s *StateDB) {
	// TODO
}

func (ch shimURLChange) undo(s *StateDB) {
	// TODO
}

func (ch rsaPublicKeyChange) undo(s *StateDB) {
	// TODO
}

func (ch httpPortChange) undo(s *StateDB) {
	// TODO
}

func (ch usageChange) undo(s *StateDB) {
	// TODO
}

func (ch keyChange) undo(s *StateDB) {
	//	so, err := s.getStateObject(ch.key)
	//	if err != nil {
	//		return err
	//	}
	//	so.txhash = ch.prevTxhash
}

func (ch bucketsChange) undo(s *StateDB) {
	// TODO
}

func (ch ownerChange) undo(s *StateDB) {
	// TODO
	// so, _, err := s.getStateObject(ch.key)
	// if err != nil {
	// 	return fmt.Errorf("[journal:revert] %s", err)
	// }
	// so.val = ch.prevOwnerRootHash
	// return nil
}

func (ch databaseChange) undo(s *StateDB) {
	// TODO
	// so, _, err := s.getStateObject(ch.key)
	// if err != nil {
	// 	return fmt.Errorf("[journal:revert] %s", err)
	// }
	// so.val = ch.prevDatabaseRootHash
	// return nil
}

func (ch tableChange) undo(s *StateDB) {
	// TODO
	// so, _, err := s.getStateObject(ch.key)
	// if err != nil {
	// 	return fmt.Errorf("[journal:revert] %s", err)
	// }
	// so.val = ch.prevTableRootHash
	// return nil
}

func (ch storeClaimsChange) undo(s *StateDB) {

}
