package cloud

import (
	"context"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type CloudLevelDB struct {
	ldb  *leveldb.DB
	iter iterator.Iterator
}

func NewLevelDB(path string, o *opt.Options) (c *CloudLevelDB, err error) {
	ldb, err := leveldb.OpenFile(path, o)
	if err != nil {
		return c, fmt.Errorf("[NewCloudLevelDB] OpenFile %s: %v\n", path, err)
	}

	var cl CloudLevelDB
	cl.ldb = ldb
	return &cl, nil
}

func (self *CloudLevelDB) setChunk(k []byte, v []byte) (err error) {
	err = self.ldb.Put(k, v, nil)
	if err != nil {
		return err
	}
	return nil
}

func (self *CloudLevelDB) getChunk(k []byte) (v []byte, ok bool, err error) {
	v, err = self.ldb.Get(k, nil)
	if err == leveldb.ErrNotFound {
		return v, false, nil
	} else if err != nil {
		return v, false, err
	}
	return v, true, nil
}

func (self *CloudLevelDB) setChunkBatch(chunks []*RawChunk) (err error) {
	batch := new(leveldb.Batch)
	for _, ch := range chunks {
		batch.Put(ch.ChunkID, ch.Value)
		if ch.Wg != nil {
			ch.Wg.Done()
		}
	}

	err = self.ldb.Write(batch, nil)
	return err
}

func (self *CloudLevelDB) getChunkBatch(chunks []*RawChunk) (err error) {
	//	ctx, cancel := context.WithTimeout(context.Background(), maxTimeout) // take about 0.201s
	//	defer cancel()
	var err_ldb error
	for _, chunk := range chunks {
		v, ok, err := self.GetChunk(chunk.ChunkID)
		if err != nil {
			err_ldb = err
		}
		chunk.OK = ok
		chunk.Value = v
		chunk.Error = err
	}
	return err_ldb // return the last error
}

func (self *CloudLevelDB) SetChunk(ctx context.Context, k []byte, v []byte) (err error) {

	return self.setChunk(k, v)
}

func (self *CloudLevelDB) GetChunk(k []byte) (v []byte, ok bool, err error) {
	return self.getChunk(k)
}

func (self *CloudLevelDB) GetChunkWithRange(k []byte, start int, end int) (v []byte, ok bool, err error) {
	v, ok, err = self.getChunk(k)
	if end == 0 {
		end = len(v)
	}
	if start > end {
		return nil, ok, fmt.Errorf("[cloud:CloudLevelDB:GetChunkWithRange] start(%d) > end(%d)", start, end)
	}
	buf := v[start:end]
	return buf, ok, err
}

func (self *CloudLevelDB) SetChunkBatch(ctx context.Context, chunks []*RawChunk) (err error) {
	return self.setChunkBatch(chunks)
}

func (self *CloudLevelDB) GetChunkBatch(chunks []*RawChunk) (err error) {
	return self.getChunkBatch(chunks)
}

func (self *CloudLevelDB) Close() {
	self.ldb.Close()
}

func (self *CloudLevelDB) Delete(k []byte) (err error) {
	err = self.ldb.Delete(k, nil)
	return err
}

func (self *CloudLevelDB) DeleteBatch(chunks []*RawChunk) (err error) {
	batch := new(leveldb.Batch)
	for _, ch := range chunks {
		batch.Delete(ch.ChunkID)
	}

	err = self.ldb.Write(batch, nil)
	return err
}

func (self *CloudLevelDB) NewIterator(slice *util.Range, ro *opt.ReadOptions) (err error) {
	self.iter = self.ldb.NewIterator(slice, ro)
	return err
}

func (self *CloudLevelDB) ReleaseIterator() (err error) {
	self.iter.Release()
	return self.iter.Error()
}

func (self *CloudLevelDB) IterNext() bool {
	if self.iter == nil {
		return false
	}
	return self.iter.Next()
}

func (self *CloudLevelDB) IterKey() []byte {
	return self.iter.Key()
}

func (self *CloudLevelDB) IterValue() []byte {
	return self.iter.Value()
}
