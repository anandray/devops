// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// registryObject represents a wallet address which is being modified.
type registryObject struct {
	idx            uint64
	db             *StateDB
	registeredNode RegisteredNode
	deleted        bool
}

//NewAccountObject creates an account object.
func NewRegistryObject(db *StateDB, idx uint64, a RegisteredNode) *registryObject {
	return &registryObject{
		db:             db,
		idx:            idx,
		registeredNode: a,
		deleted:        false,
	}
}

// EncodeRLP implements rlp.Encoder.
func (self *registryObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, self.registeredNode)
}

func (self *registryObject) SetAddress(addr common.Address) {
	self.db.journal.entries = append(self.db.journal.entries, addressChange{
		index: self.idx,
		prev:  self.registeredNode.address,
	})
	self.registeredNode.address = addr
}

func (self *registryObject) SetStorageIP(storageip []byte) {
	self.db.journal.entries = append(self.db.journal.entries, storageIPChange{
		index: self.idx,
		prev:  self.registeredNode.storageip,
	})
	self.registeredNode.storageip = string(storageip)
}

func (self *registryObject) SetConsensusIP(consensusip []byte) {
	self.db.journal.entries = append(self.db.journal.entries, consensusIPChange{
		index: self.idx,
		prev:  self.registeredNode.consensusip,
	})
	self.registeredNode.consensusip = string(consensusip)
}

func (self *registryObject) SetValueInt(valueInt uint64) {
	self.db.journal.entries = append(self.db.journal.entries, valueIntChange{
		index: self.idx,
		prev:  self.registeredNode.valueInt,
	})
	self.registeredNode.valueInt = valueInt
}

func (self *registryObject) SetValueExt(valueExt uint64) {
	self.db.journal.entries = append(self.db.journal.entries, valueExtChange{
		index: self.idx,
		prev:  self.registeredNode.valueExt,
	})
	self.registeredNode.valueExt = valueExt
}

func (self *registryObject) SetRegion(region byte) {
	self.db.journal.entries = append(self.db.journal.entries, regionChange{
		index: self.idx,
		prev:  self.registeredNode.region,
	})
	self.registeredNode.region = region
}

func (self *registryObject) SetHTTPPort(httpPort uint16) {
	self.db.journal.entries = append(self.db.journal.entries, httpPortChange{
		index: self.idx,
		prev:  self.registeredNode.httpPort,
	})
	self.registeredNode.httpPort = httpPort
}

//
// Attribute accessors
//
func (self *registryObject) Index() uint64 {
	return self.idx
}

func (self *registryObject) Address() common.Address {
	return self.registeredNode.address
}
func (self *registryObject) ValueInt() uint64 {
	return self.registeredNode.valueInt
}

func (self *registryObject) ValueExt() uint64 {
	return self.registeredNode.valueExt
}

func (self *registryObject) String() string {
	s := fmt.Sprintf("Obj: %v, account: %v, MarkDel: %v", self.idx, self.registeredNode, self.deleted)
	return s
}
