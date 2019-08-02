// +build darwin

package cloud

import (
	"context"
	"fmt"
)

type CephStorage struct {
}

func NewCephStorage(configFile string, pool string) (cephif *CephStorage, err error) {
	return cephif, fmt.Errorf("Not SUPPORTED outside linux currently")
}

func (c *CephStorage) SetChunk(ctx context.Context, key []byte, data []byte) (err error) {
	return err
}

func (c *CephStorage) GetChunk(key []byte) (data []byte, b bool, err error) {
	return data, true, err
}

func (c *CephStorage) GetChunkWithRange(key []byte, start int, end int) (data []byte, b bool, err error) {
	return data, true, err
}

func (c *CephStorage) SetChunkBatch(ctx context.Context, chunk []*RawChunk) (err error) {
	return err
}

func (c *CephStorage) GetChunkBatch(chunks []*RawChunk) (err error) {
	return nil
}

func (c *CephStorage) Close() {
}
