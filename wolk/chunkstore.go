package wolk

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	chunker "github.com/ethereum/go-ethereum/swarm/storage"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

type MockStorage struct {
	cloudstore cloud.Cloudstore
}

func NewMockStorage() (c *MockStorage, err error) {
	config := &cloud.DefaultConfig
	config.ConsensusIdx = 0
	var cl cloud.Cloudstore
	if true {
		cl, err = cloud.NewLevelDB(config.DataDir, nil)
	} else {
		cl, err = cloud.NewCephStorage(config.CephConfig, fmt.Sprintf("%s%d", config.CephCluster, config.ConsensusIdx))
	}
	if err != nil {
		return c, err
	}

	c = new(MockStorage)
	c.cloudstore = cl
	return c, nil
}

func (cs *MockStorage) SetChunk(ctx context.Context, v []byte) (k common.Hash, err error) {
	k0 := wolkcommon.Computehash(v)
	err = cs.cloudstore.SetChunk(ctx, k0, v)
	return common.BytesToHash(k0), err
}

func (cs *MockStorage) QueueChunk(ctx context.Context, v []byte) (k common.Hash, err error) {
	return cs.SetChunk(ctx, v)
}

func (cs *MockStorage) PutChunk(ctx context.Context, chunkData *cloud.RawChunk, wg *sync.WaitGroup) (r chunker.Reference, err error) {
	defer wg.Done()
	chunkData.ChunkID = wolkcommon.Computehash(chunkData.Value)
	chunkData.Wg = wg
	err = cs.cloudstore.SetChunk(ctx, chunkData.ChunkID, chunkData.Value)
	if err != nil {
		return r, err
	}
	return chunkData.ChunkID, nil
}

func (cs *MockStorage) GetChunk(ctx context.Context, k []byte) (v []byte, ok bool, err error) {
	return cs.cloudstore.GetChunk(k)
}

func (cs *MockStorage) FlushQueue() error {
	return nil
}

func (cs *MockStorage) Close() {
	cs.cloudstore.Close()
}

func (ldb *MockStorage) GetFile(ctx context.Context, key []byte) ([]byte, error) {
	return nil, fmt.Errorf("[ChunkStore:GetFile] GetFile is not supported")
}
