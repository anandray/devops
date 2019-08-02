// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (

	// "encoding/json"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/wolkdb/cloudstore/log"

	wolkcommon "github.com/wolkdb/cloudstore/common"
)

// NewMinedBlockEvent is posted when a block has been imported.
type NewMinedBlockEvent struct{ Block *Block }

const (
	reservedBytes = 10
)

type Block struct {
	NetworkID     uint64         `json:"networkID"`
	ParentHash    common.Hash    `json:"parentHash"`
	BlockNumber   uint64         `json:"blockNumber"`
	Seed          []byte         `json:"seed"`         // vrf-based seed for next round
	AccountRoot   common.Hash    `json:"accountRoot"`  // SMT with keys being addresses, values being Account
	RegistryRoot  common.Hash    `json:"registryRoot"` // SMT with keys being node numbers, values being RegistryNode
	KeyRoot       common.Hash    `json:"keyRoot"`
	NameRoot      common.Hash    `json:"nameRoot"`
	Transactions  []*Transaction `json:"transactions"`
	StorageBeta   uint64         `json:"storageBeta"`
	BandwidthBeta uint64         `json:"bandwidthBeta"`
}

// blockHeaderData is the network packet for the blockHash message.
type blockHeaderData struct {
	NetworkId   uint64      `json:"networkID"`
	BlockNumber uint64      `json:"blockNumber"`
	BlockHash   common.Hash `json:"blockHash"`
	ParentHash  common.Hash `json:"parentHash"`
}

func (b *Block) Head() (bhd *blockHeaderData) {
	bhd = new(blockHeaderData)
	bhd.NetworkId = b.NetworkID
	bhd.BlockNumber = b.BlockNumber
	bhd.BlockHash = b.Hash()
	bhd.ParentHash = b.ParentHash
	return bhd
}

func NewBlock() *Block {
	var b Block
	return &b
}

func (b *Block) Copy() (c *Block) {
	c = new(Block)
	c.NetworkID = b.NetworkID
	c.ParentHash = b.ParentHash
	c.BlockNumber = b.BlockNumber
	c.Seed = b.Seed
	c.AccountRoot = b.AccountRoot
	c.RegistryRoot = b.RegistryRoot
	c.KeyRoot = b.KeyRoot
	c.NameRoot = b.NameRoot
	//c.Transactions = b.Transactions //txns needs to be removed!
	c.StorageBeta = b.StorageBeta
	c.BandwidthBeta = b.BandwidthBeta
	return c
}

func (block Block) Hash() (h common.Hash) {
	data, _ := rlp.EncodeToBytes(&block)
	return common.BytesToHash(wolkcommon.Computehash(data))
}

func (block *Block) BytesWithoutSig() []byte {
	enc, _ := rlp.EncodeToBytes(&block)
	return enc
}

func (b *Block) UnsignedHash() common.Hash {
	unsignedBytes := b.BytesWithoutSig()
	return common.BytesToHash(wolkcommon.Computehash(unsignedBytes))
}

func (b *Block) IsEmpty() bool {
	//block is empty if: (1) txlen == 0 and (2) emptyseed matches with current seed
	if b == nil {
		return false
	}
	if len(b.Transactions) > 0 {
		return false
	} else if bytes.Compare(b.Seed, wolkcommon.Computehash(b.ParentHash.Bytes())) != 0 {
		return false
	} else {
		return true
	}
}

func (block *Block) IsIdentical(b *Block) (validated bool) {
	if block.BlockNumber != b.BlockNumber {
		log.Error("[block:IsIdentical] blockNumber mismatch", "derived", block.BlockNumber, "received", b.BlockNumber)
		return false
	}

	if bytes.Compare(block.ParentHash.Bytes(), b.ParentHash.Bytes()) != 0 {
		log.Error("[block:IsIdentical] ParentHash mismatch", "derived", block.ParentHash.Hex(), "received", b.ParentHash.Hex())
		return false
	}

	if bytes.Compare(block.AccountRoot.Bytes(), b.AccountRoot.Bytes()) != 0 {
		log.Error("[block:IsIdentical] AccountRoot mismatch", "derived", block.AccountRoot.Hex(), "received", b.AccountRoot.Hex())
		return false
	}

	if bytes.Compare(block.RegistryRoot.Bytes(), b.RegistryRoot.Bytes()) != 0 {
		log.Error("[block:IsIdentical] RegistryRoot mismatch", "derived", block.RegistryRoot.Hex(), "received", b.RegistryRoot.Hex())
		return false
	}

	if bytes.Compare(block.KeyRoot.Bytes(), b.KeyRoot.Bytes()) != 0 {
		log.Error("[block:IsIdentical] KeyRoot mismatch", "derived", block.KeyRoot.Hex(), "received", b.KeyRoot.Hex())
		return false
	}

	if bytes.Compare(block.NameRoot.Bytes(), b.NameRoot.Bytes()) != 0 {
		log.Error("[block:IsIdentical] NameRoot mismatch", "derived", block.NameRoot.Hex(), "received", b.NameRoot.Hex())
		return false
	}

	//TODO: check other parameters (if implemented)
	return true
}

func (blk *Block) Round() uint64 {
	return blk.BlockNumber
}

func (block Block) Root() (p common.Hash) {
	return block.UnsignedHash()
}

func (block Block) Number() (n uint64) {
	return block.BlockNumber
}

func FromChunk(in []byte) (b *Block) {
	var ob Block // []interface{}
	err := rlp.Decode(bytes.NewReader(in), &ob)
	if err != nil {
		return nil
	}
	return &ob
}

func (block Block) Encode() ([]byte, error) {
	return rlp.EncodeToBytes(&block)
}

type SerializedBlock struct {
	Hash          common.Hash              `json:"hash"`
	ParentHash    common.Hash              `json:"parentHash"`
	NetworkID     uint64                   `json:"networkID"`
	BlockNumber   uint64                   `json:"blockNumber"`
	Seed          string                   `json:"seed"`
	AccountRoot   common.Hash              `json:"accountRoot"`  // SMT with keys being addresses, values being Account
	RegistryRoot  common.Hash              `json:"registryRoot"` // SMT with keys being node numbers, values being RegistryNode
	KeyRoot       common.Hash              `json:"keyRoot"`
	NameRoot      common.Hash              `json:"nameRoot"`
	Transactions  []*SerializedTransaction `json:"transactions"`
	TxList        []*common.Hash           `json:"txlist"`
	StorageBeta   uint64                   `json:"storageBeta"`
	BandwidthBeta uint64                   `json:"bandwidthBeta"`
	EmptyBlock    bool                     `json:"emptyBlock"`
	Blocksize     uint64                   `json:"blocksize"`
}

func (b *Block) Bytes() (enc []byte) {
	enc, _ = rlp.EncodeToBytes(b)
	return enc
}

func (b *Block) String() string {
	blocksize := uint64(len(b.Bytes()))
	s := &SerializedBlock{
		Hash:         b.Hash(),
		NetworkID:    b.NetworkID,
		ParentHash:   b.ParentHash,
		BlockNumber:  b.BlockNumber,
		Seed:         fmt.Sprintf("%x", b.Seed),
		AccountRoot:  b.AccountRoot,
		RegistryRoot: b.RegistryRoot,
		KeyRoot:      b.KeyRoot,
		NameRoot:     b.NameRoot,
		//TxList:        make([]*common.Hash, 0),
		Transactions:  make([]*SerializedTransaction, 0),
		StorageBeta:   b.StorageBeta,
		BandwidthBeta: b.BandwidthBeta,
		EmptyBlock:    b.IsEmpty(),
		Blocksize:     blocksize,
	}

	for _, tx := range b.Transactions {
		//txhash := tx.Hash()
		//s.TxList = append(s.TxList, &txhash)
		stx := NewSerializedTransaction(tx)
		stx.BlockNumber = b.BlockNumber
		s.Transactions = append(s.Transactions, stx)
	}
	return s.String()
}

func (s *SerializedBlock) String() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}
