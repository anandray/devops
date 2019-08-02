// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"sync"
)

// accountObject represents a wallet address which is being modified.
type accountObject struct {
	mu      sync.RWMutex
	address common.Address
	db      *StateDB
	account Account
	deleted bool
}

//NewAccountObject creates an account object.
func NewAccountObject(db *StateDB, addr common.Address, a Account) *accountObject {
	return &accountObject{
		db:      db,
		address: addr,
		account: a,
		deleted: false,
	}
}

// EncodeRLP implements rlp.Encoder.
func (ao *accountObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, ao.account)
}

func (ao *accountObject) deepCopy(db *StateDB) *accountObject {
	accountObject := NewAccountObject(db, ao.address, ao.account)
	return accountObject
}

//
// Attribute accessors
//

func (ao *accountObject) Address() (a common.Address) {
	ao.mu.RLock()
	a = ao.address
	ao.mu.RUnlock()
	return a
}

func (ao *accountObject) Balance() (balance uint64) {
	ao.mu.RLock()
	balance = ao.account.Balance
	ao.mu.RUnlock()
	return balance
}

func (ao *accountObject) Usage() (usage uint64) {
	ao.mu.RLock()
	usage = ao.account.Usage
	ao.mu.RUnlock()
	return usage
}

func (ao *accountObject) Quota() (quota uint64) {
	ao.mu.RLock()
	quota = ao.account.Quota
	ao.mu.RUnlock()
	return quota
}

func (ao *accountObject) String() string {
	ao.mu.RLock()
	s := fmt.Sprintf("{\"addr\":\"%x\", \"balance\":%v, \"quota\":%v}", ao.address, ao.account.Balance, ao.account.Quota)
	ao.mu.RUnlock()
	return s
}

func (ao *accountObject) SetBalance(amount uint64) {
	ao.mu.Lock()
	ao.db.journal.entries = append(ao.db.journal.entries, balanceChange{
		account: &ao.address,
		prev:    ao.account.Balance,
	})
	ao.account.Balance = amount
	ao.mu.Unlock()
}

func (ao *accountObject) SetRSAPublicKey(rsaPublicKey []byte) {
	ao.mu.Lock()
	ao.db.journal.entries = append(ao.db.journal.entries, rsaPublicKeyChange{
		account: &ao.address,
		prev:    ao.account.RSAPublicKey,
	})
	ao.account.RSAPublicKey = rsaPublicKey
	ao.mu.Unlock()
}

func (ao *accountObject) SetQuota(quota uint64) {
	ao.mu.Lock()
	ao.db.journal.entries = append(ao.db.journal.entries, quotaChange{
		account: &ao.address,
		prev:    ao.account.Quota,
	})
	ao.account.Quota = quota
	ao.mu.Unlock()
}

func (ao *accountObject) SetShimURL(shimURL string) {
	ao.mu.Lock()
	ao.db.journal.entries = append(ao.db.journal.entries, shimURLChange{
		account: &ao.address,
		prev:    ao.account.ShimURL,
	})
	ao.account.ShimURL = shimURL
	ao.mu.Unlock()
}

func (ao *accountObject) TallyReward(rewards uint64) {
	ao.mu.Lock()
	ao.db.journal.entries = append(ao.db.journal.entries, balanceChange{
		account: &ao.address,
		prev:    ao.account.Balance,
	})
	ao.account.Balance += rewards
	ao.mu.Unlock()
}

func (ao *accountObject) TallyStorageUsage(usage uint64) {
	ao.mu.Lock()
	ao.db.journal.entries = append(ao.db.journal.entries, usageChange{
		account: &ao.address,
		prev:    ao.account.Usage,
	})
	ao.account.Usage += usage
	ao.mu.Unlock()
}

func (ao *accountObject) ChargeForStorageClaims(currentBlock uint64, storageBeta uint64) uint64 {
	ao.mu.Lock()
	ao.db.journal.entries = append(ao.db.journal.entries, storeClaimsChange{
		account: &ao.address,
		prev:    ao.account.Balance,
		// prevLastClaim: ao.account.LastClaim
	})

	charges := ao.account.Quota * (currentBlock - ao.account.LastClaim) / storageBeta
	ao.account.LastClaim = currentBlock
	if charges > ao.account.Balance {
		charges = ao.account.Balance
		ao.account.Balance = 0
	} else {
		ao.account.Balance -= charges
	}
	ao.mu.Unlock()
	return charges
}
