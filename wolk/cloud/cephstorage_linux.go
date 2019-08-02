// +build linux

package cloud

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ceph/go-ceph/rados"
	"github.com/wolkdb/cloudstore/log"
	"sync"
	"sync/atomic"
	"time"
)

const radosNotFound = "rados: No such file or directory"

type CephStorage struct {
	ConfigFile string
	pool       string
	conn       *rados.Conn
	queue      []*RawChunk
	queuemu    sync.Mutex
	update     int64
	emptyBytes []byte
}

func NewCephStorage(configFile string, pool string) (cephif *CephStorage, err error) {
	cephif = &CephStorage{ConfigFile: configFile, pool: pool}
	conn, err := rados.NewConnWithClusterAndUser("ceph", "client.admin")

	if err != nil {
		return nil, err
	}
	cephif.conn = conn
	if cephif.ConfigFile != "" {
		err = conn.ReadConfigFile(configFile)
		if err != nil {
			log.Error("[cephstorage:NewCephStorage] Ceph Connection not established", "err", err)
			return cephif, err
		}
	}

	conn.Connect()
	log.Trace("[cephstorage:NewCephStorage] Ceph Connection established", "pool", pool)
	//defer conn.Shutdown()
	atomic.StoreInt64(&cephif.update, 0)
	go cephif.StoreLoop()

	cephif.emptyBytes = []byte("%eMptyByt3s@")
	return cephif, err
}

func (c *CephStorage) StoreLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if c.getQueueLen() > 0 {
				c.storeChunks()
			}
		}
	}
}

func (c *CephStorage) getQueueLen() int {
	c.queuemu.Lock()
	defer c.queuemu.Unlock()
	return len(c.queue)
}

func (c *CephStorage) addQueue(chunk *RawChunk) {
	c.queuemu.Lock()
	c.queue = append(c.queue, chunk)
	c.queuemu.Unlock()
}

func (c *CephStorage) addQueueBatch(chunks []*RawChunk) {
	c.queuemu.Lock()
	c.queue = append(c.queue, chunks...)
	log.Trace("[cephstorage:addQueueBatch]", "addlen", len(chunks), "totallen", len(c.queue))
	c.queuemu.Unlock()
}

func (c *CephStorage) storeChunks() error {
	if atomic.LoadInt64(&c.update) != 0 {
		log.Info("[CephStorage] storeChunks no update")
		return nil
	}
	atomic.StoreInt64(&c.update, 1)
	defer atomic.StoreInt64(&c.update, 0)
	c.queuemu.Lock()
	buf := c.queue
	c.queue = make([]*RawChunk, 0)
	c.queuemu.Unlock()
	ioctx, err := c.conn.OpenIOContext(c.pool)
	if err != nil {
	   log.Error("[cephstorage:storeChunks] OpenIOContext", "pool", c.pool, "err", err)
	  return err
	}
	var wglist []*sync.WaitGroup
	defwg := new(sync.WaitGroup)
	emptyBytes := c.emptyBytes
	var complist []*rados.Completion
	log.Trace("[cephstorage:storeChunks]", "len", len(buf))
	for _, c := range buf {
		comp, err := rados.NewCompletion()
		if err != nil {
		} else {
			v := c.Value
			if len(v) == 0 {
				v = emptyBytes
			}
			log.Info("[cephstorage:storeChunks] AsyncWrite", "chunkID", fmt.Sprintf("%x", c.ChunkID), "len(v)", len(c.Value), "len(v)", len(v))
			ioctx.AsyncWrite(fmt.Sprintf("%x", c.ChunkID), v, 0, comp)
			complist = append(complist, comp)
			if c.Wg != nil {
				wglist = append(wglist, c.Wg)
			} else {
				defwg.Add(1)
				wglist = append(wglist, defwg)
			}
		}
	}
	for i, cmp := range complist {
		wg := wglist[i]
		go func(cmp *rados.Completion) {
			cmp.WaitForComplete()
			cmp.WaitForSafe()
			wg.Done()
			cmp.Release()
		}(cmp)
	}
	defwg.Wait()
	//log.Info("[ceph] store done", "wg", wg)
	ioctx.Destroy()
	return err
}

func (c *CephStorage) SetChunk(ctx context.Context, key []byte, data []byte) (err error) {
	// which one is better? using addQueue or call ceph directly
	ioctx, err := c.conn.OpenIOContext(c.pool)
	if err != nil {
		log.Error("[CephStorage]SetChunk", "pool", c.pool)
		return err
	}

	name := fmt.Sprintf("%x", key)
	if len(data) == 0 {
		data = c.emptyBytes
	}
	err = ioctx.Write(name, data, 0)
	return err
}

func (c *CephStorage) GetChunk(key []byte) (data []byte, b bool, err error) {

	ioctx, err := c.conn.OpenIOContext(c.pool)
	if err != nil {
		log.Error("[cephstorage:GetChunk] OpenIOContext", "pool", c.pool, "err", err)
		return nil, false, err
	}
	name := fmt.Sprintf("%x", key)
	stat, err := ioctx.Stat(name)
	if err != nil {
		if err.Error() == radosNotFound {
			return nil, false, nil
		}
		log.Error("[cephstorage:GetChunk] Stat", "err", err)
		return data, false, err
	}
	data = make([]byte, stat.Size)
	buflen, err := ioctx.Read(name, data, 0)
	if err != nil {
		if err.Error() == radosNotFound {
			return nil, false, nil
		}
		log.Error("[cephstorage:GetChunk] Read", "key", fmt.Sprintf("%x", key), "buflen", buflen)
		return nil, false, err
	}
	if bytes.Compare(data, c.emptyBytes) == 0 {
		data = []byte("")
	}
	log.Trace("[cephstorage:GetChunk]", "key", fmt.Sprintf("%x", key), "buflen", buflen)
	return data, true, err
}

func (c *CephStorage) GetChunkWithRange(key []byte, start int, end int) (data []byte, b bool, err error) {
	ioctx, err := c.conn.OpenIOContext(c.pool)
	name := fmt.Sprintf("%x", key)
	if end == 0 {
		stat, err := ioctx.Stat(name)
		if err != nil {
			log.Error("[CephStorage:GetChunkWithRange] Stat error", "key", fmt.Sprintf("%x", key), "err", err)
			return nil, false, err
		}
		end = int(stat.Size)
	}
	if start > end {
		return nil, b, fmt.Errorf("[cloud:CephStorage] start(%d) > end(%d)", start, end)
	}
	data = make([]byte, end-start)
	_, err = ioctx.Read(name, data, uint64(start))
	if err != nil {
		if err.Error() == radosNotFound {
			return nil, false, nil
		}
		log.Error("Error at GetChunk", "key", fmt.Sprintf("%x", key), "err", err)
		return nil, false, err
	}
	if bytes.Compare(data, c.emptyBytes) == 0 {
		data = []byte("")
	}

	return data, true, err
}

func (c *CephStorage) SetChunkBatch(ctx context.Context, chunk []*RawChunk) (err error) {
	log.Trace("[cephstorage:SetChunkBatch]", "len", len(chunk))
	c.addQueueBatch(chunk)

	return nil
}

func (c *CephStorage) GetChunkBatch(chunks []*RawChunk) (err error) {
	var err_ldb error
	for _, chunk := range chunks {
		v, ok, err := c.GetChunk(chunk.ChunkID)
		if err != nil {
			log.Error("GetChunk", "k", chunk.ChunkID, "err", err)
			err_ldb = err
		} else {
			log.Trace("GetChunkBatch", "k", fmt.Sprintf("%x", chunk.ChunkID), "len(v)", len(v), "ok", ok)
		}
		chunk.OK = ok
		chunk.Value = v
		chunk.Error = err
	}
	return err_ldb // return the last error

}

func (c *CephStorage) Close() {
	c.conn.Shutdown()
}
