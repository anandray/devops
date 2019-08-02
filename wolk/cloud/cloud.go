// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.
package cloud

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/wolkdb/cloudstore/log"
)

const (
	maxTimeout             = 5000 * time.Millisecond
	DefaultGenesisFileName = "/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json"
)

type RawChunk struct {
	ChunkID []byte          `json:"chunkID,omitempty"`
	Value   []byte          `json:"val,omitempty"`
	OK      bool            `json:"ok,omitempty"`
	Error   error           `json:"error,omitempty"`
	Sig     []byte          `json:"sig,omitempty"`
	Wg      *sync.WaitGroup `json:"-"`
}

type Cloudstore interface {
	GetChunk(k []byte) (v []byte, ok bool, err error)
	GetChunkBatch(chunks []*RawChunk) (err error)
	GetChunkWithRange(k []byte, start int, end int) (v []byte, ok bool, err error)
	SetChunk(ctx context.Context, k []byte, v []byte) (err error)
	SetChunkBatch(ctx context.Context, chunk []*RawChunk) (err error)
	Close()
}

var ErrWriteLimit = errors.New("Write Capacity exceeded")
var ErrReadLimit = errors.New("Read Capacity exceeded")

func NewCloudstore(config *Config) (cl Cloudstore, err error) {
	if config == nil {
		config = &Config{}
		LoadConfig(DefaultConfigWolkFile, config)
		log.Debug("NO CONFIG Using Default File")
	}

	switch config.Provider {
	case "leveldb":
		cl, err = NewLevelDB(config.DataDir, nil)
		log.Info("New Cloudstore: Leveldb", "dd", config.DataDir)
	default:
		log.Info("New Cloudstore: ceph", "POOL", fmt.Sprintf("%s%d", config.CephCluster, config.ConsensusIdx))
		cl, err = NewCephStorage(config.CephConfig, fmt.Sprintf("%s%d", config.CephCluster, config.ConsensusIdx))
	}
	if err != nil {
		return cl, err
	}
	return cl, err
}
