// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type TxPool struct {
	txpool    []*Transaction
	txhashes  map[common.Hash]time.Time
	mu        sync.RWMutex // this is for txpool only
	wolkstore *WolkStore
}

const chainHeadChanSize = 5
const statsReportInterval = 8 * time.Second
const maxTxPoolSize = 512
const maxTransactionAge = 120 * time.Second

func NewTxPool() *TxPool {
	pool := &TxPool{
		txpool:   make([]*Transaction, 0),
		txhashes: make(map[common.Hash]time.Time),
	}
	return pool
}

func (pool *TxPool) Start(wolkstore *WolkStore) {
	pool.wolkstore = wolkstore
}

func (self *TxPool) Reset(block *Block) {
	btx := make(map[common.Hash]bool)
	if block != nil {
		if len(block.Transactions) == 0 {
			return
		}
		for _, tx := range block.Transactions {
			btx[tx.Hash()] = true
		}
	}
	oldTxes := make(map[common.Hash]bool)
	self.mu.RLock()
	for txhash, addTime := range self.txhashes {
		if time.Since(addTime) > maxTransactionAge {
			oldTxes[txhash] = true
		}
	}
	self.mu.RUnlock()

	self.mu.Lock()
	defer self.mu.Unlock()
	newpool := make([]*Transaction, 0)
	for _, tx := range self.txpool {
		txhash := tx.Hash()
		_, inBlock := btx[txhash]
		_, isOld := oldTxes[txhash]
		if inBlock || isOld {
			delete(self.txhashes, txhash)
		} else {
			newpool = append(newpool, tx)
		}
	}
	self.txpool = newpool
}

func (self *TxPool) Pending() (pending map[common.Address]Transactions, err error) {
	pending = make(map[common.Address]Transactions)
	self.mu.RLock()
	defer self.mu.RUnlock()
	for _, tx := range self.txpool {
		addr, _ := tx.GetSignerAddress()
		pending[addr] = append(pending[addr], tx)
	}
	return pending, nil
}

func (self *TxPool) PendingTxlist() (pending []*Transaction, err error) {
	//fmt.Println("txpool Pending txlist", self.txpool)
	self.mu.RLock()
	pending = self.txpool
	self.mu.RUnlock()
	return pending, nil
}

func (self *TxPool) TxCount() int64 {
	self.mu.RLock()
	txcnt := int64(len(self.txpool))
	self.mu.RUnlock()
	return txcnt
}

func (self *TxPool) Stop() {
}

func (self *TxPool) Len() int {
	return len(self.txpool)
}

func (self *TxPool) addTransactionToPool(tx *Transaction) (err error) {
	// check if the transaction is in the pool already [use map instead of this?]
	if len(self.txpool) > maxTxPoolSize {
		err = fmt.Errorf("Transaction Pool is full: maximum %d (current size %d)", maxTxPoolSize, len(self.txpool))
		self.Reset(nil)
		return err
	}
	h := tx.Hash()
	self.mu.Lock()
	self.txhashes[h] = time.Now()
	self.mu.Unlock()
	self.mu.RLock()
	for _, tx0 := range self.txpool {
		if bytes.Compare(tx0.Hash().Bytes(), h.Bytes()) == 0 {
			self.mu.RUnlock()
			return nil
		}
	}
	self.mu.RUnlock()
	self.mu.Lock()
	self.txpool = append(self.txpool, tx)
	self.mu.Unlock()
	return nil
}

func (self *TxPool) HasTransaction(txhash common.Hash) bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	if _, ok := self.txhashes[txhash]; ok {
		return true
	}
	return false
}

func (self *TxPool) GetTransaction(txhash common.Hash) (tx *Transaction) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	for _, tx := range self.txpool {
		if tx.Hash() == txhash {
			return tx
		}
	}
	return tx
}
