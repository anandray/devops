// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
)

// Name objects
type NamesStateDB struct {
}

func NameToAddress(name string) common.Address {
	return common.BytesToAddress(wolkcommon.Computehash([]byte(name))[0:20])
}

func (statedb *StateDB) SetName(ctx context.Context, tx *Transaction, txp *TxBucket) (err error) {
	st0 := time.Now()
	st := time.Now()
	statedb.muTxCache.Lock()
	statedb.txCache[tx.Hash()] = tx
	statedb.muTxCache.Unlock()
	if time.Since(st) > warningTxTime {
		log.Error("[statedb_name:SetName] txCache ******", "tm", time.Since(st))
	} else {
		log.Trace("[statedb_name:SetName] txCache ***ABS***", "tm", time.Since(st0))
	}

	st = time.Now()
	name := string(txp.Name)
	nameaddr := NameToAddress(name)
	if time.Since(st) > warningTxTime {
		log.Error("[statedb_name:SetName] NameToAddress ******", "tm", time.Since(st))
	} else {
		log.Trace("[statedb_name:SetName] NameToAddress ***ABS***", "tm", time.Since(st0))
	}
	st = time.Now()
	address, err := tx.GetSignerAddress()
	if err != nil {
		log.Info("[statedb_names:SetName] GetSigner", "err", err)
		return err
	}
	if time.Since(st) > warningTxTime {
		log.Error("[statedb_name:SetName] GetSignerAddress ******", "tm", time.Since(st))
	} else {
		log.Trace("[statedb_name:SetName] GetSignerAddress ***ABS***", "tm", time.Since(st0))
	}
	// TODO: only one owner per name -- once its set it can't be taken again!
	txhash := tx.Hash()
	if string(tx.Method) == http.MethodPost || string(tx.Method) == http.MethodPut {
		if statedb.accountExists(ctx, nameaddr) && len(txp.RSAPublicKey) > 0 {
			return fmt.Errorf("[statedb:SetName] cannot create account")
		}
	}

	// the hash of the name is mapped to the txhash
	st = time.Now()
	err = statedb.nameStorage.Insert(ctx, nameaddr.Bytes(), txhash.Bytes(), 0, false)
	if err != nil {
		log.Error("[statedb:SetName] Insert", "ERROR", err)
		return err
	}
	if time.Since(st) > warningTxTime {
		log.Error("[statedb_name:SetName] namestorage Insert ***1***", "name", name, "address", address, "tm", time.Since(st))
	} else {
		log.Info("[statedb_name:SetName] namestorage Insert ***ABS***", "name", name, "address", address, "tm", time.Since(st0))
	}
	st = time.Now()
	account, err := statedb.getOrCreateAccount(ctx, address)
	if err != nil {
		return err
	}
	if string(tx.Method) == http.MethodPatch {
		if len(txp.ShimURL) > 0 {
			account.SetShimURL(txp.ShimURL)
		}
		if txp.Quota > 0 {
			account.SetQuota(txp.Quota)
		}
		// and all other
	} else {
		// create whole account
		account.SetBalance(TestBalance)
		account.SetQuota(MinimumQuota)
		account.SetRSAPublicKey(txp.RSAPublicKey)
		account.SetShimURL(txp.ShimURL)
	}
	statedb.muAccountObjects.Lock()
	statedb.accountObjects[address] = account
	statedb.muAccountObjects.Unlock()

	systemHash := nameaddr
	err = statedb.keyStorage.Insert(ctx, systemHash.Bytes(), getEmptySMTChunkHash(), 0, false)
	if err != nil {
		log.Error("[statedb_names:SetName]", "err", err)
		return err
	}
	if time.Since(st) > warningTxTime {
		log.Error("[statedb_names:SetName] keystorage Insert ***2***", "tm", time.Since(st))
	} else {
		log.Info("[statedb_names:SetName] keystorage Insert ***ABS***", "tm", time.Since(st0))
	}
	return nil
}

func (statedb *StateDB) getTxCache(ctx context.Context, txhash common.Hash) (tx *Transaction, ok bool, err error) {
	statedb.muTxCache.RLock()
	if tx, found := statedb.txCache[txhash]; found {
		statedb.muTxCache.RUnlock()
		log.Trace("[statedb:getTxCache] FOUND")
		return tx, true, nil
	}
	statedb.muTxCache.RUnlock()

	txbytes, ok, err := statedb.Storage.GetChunk(ctx, txhash.Bytes())
	if err != nil {
		return tx, false, err
	} else if !ok {
		return tx, false, nil
	}
	tx, err = DecodeRLPTransaction(txbytes)
	if err != nil {
		return tx, false, err
	}
	return tx, true, nil
}

func (statedb *StateDB) GetName(ctx context.Context, name string, withProof bool) (owner common.Address, ok bool, proof *SMTProof, err error) {
	nameaddr := NameToAddress(name)
	log.Trace("[statedb_name:GetName]", "name", name, "nameaddr", nameaddr)
	var tx *Transaction

	txHashBytes, ok, _, nameproof, _, err := statedb.nameStorage.Get(ctx, nameaddr.Bytes(), true)
	if err != nil {
		return owner, ok, proof, fmt.Errorf("[statedb_name:GetName] %s", err)
	}
	if !ok {
		// statedb.nameStorage.Dump()
		log.Trace("[statedb_name:GetName] NOT OK", "name", name)
		return owner, false, proof, nil
	}
	txhash := common.BytesToHash(txHashBytes)
	tx, ok, err = statedb.getTxCache(ctx, txhash)
	if err != nil {
		return owner, ok, proof, err
	} else if !ok {
		return owner, ok, proof, nil
	}
	owner, err = tx.GetSignerAddress()
	if err != nil {
		return owner, ok, proof, fmt.Errorf("[statedb_name:GetName] GetSigner ERR %s", err)
	}
	if withProof {
		proof = new(SMTProof)
		// proof.BlockNumber =
		// proof.BlockHash =
		proof.ChunkHash = statedb.nameStorage.ChunkHash()
		proof.MerkleRoot = statedb.nameStorage.MerkleRoot(ctx)
		proof.Key = name
		proof.KeyHash = nameaddr // this is the 160bit
		proof.TxHash = txhash    // this is the value that is hashed to the merkle root
		proof.Tx = NewSerializedTransaction(tx)
		proof.TxSigner = owner
		proof.Proof = NewSerializedProof(nameproof)
		log.Info("[statedb_name:GetName] proof", "proof", proof.String())
		// statedb.nameStorage.Dump()
	}
	return owner, true, proof, nil
}
