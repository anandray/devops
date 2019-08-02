// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.
package cloud

import (
	"bytes"
	"context"
	"fmt"
	"github.com/wolkdb/cloudstore/log"
	"golang.org/x/crypto/blake2s"
	"os"
	"sync"
	"testing"
	"time"
)

func init() {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
}

func Blakeb(b []byte) []byte {
	h, _ := blake2s.New256(nil)
	h.Write(b)
	return h.Sum(nil)
}

func getTestKeyVal() ([]byte, []byte) {
	t := time.Now()
	s := t.Format("2006-01-02 15:04:05")
	return []byte(fmt.Sprintf("k%s", s)), []byte(s)
}

func TestLevelDB(t *testing.T) {
	testCloudstore("leveldb", t)
}

func TestCeph(t *testing.T) {
	testCloudstore("cephif", t)
}

func testCloudstore(provider string, t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Root().SetHandler(log.CallerFileHandler(log.LvlFilterHandler(log.LvlDebug, log.StreamHandler(os.Stderr, log.TerminalFormat(true)))))

	// Load config file.
	cfg := &DefaultConfig
	cfg.Provider = provider
	cfg.ConsensusIdx = 0
	c, err := NewCloudstore(cfg)
	if err != nil {
		t.Fatalf("NewCloudstore %v", err)
	}

	k, v := getTestKeyVal()
	prefix := fmt.Sprintf("some random value %s", time.Now())
	missingKey := Blakeb([]byte(prefix))
	emptyVal := []byte("")

	// (1) NORMAL CASE
	// SetChunk
	st := time.Now()
	var ctx context.Context
	err = c.SetChunk(ctx, k, v)
	if err != nil {
		log.Error("[cloud_test] SetChunk", "provider", provider, "err", err)
	} else {
		log.Info("[cloud_test] SetChunk PASS", "provider", provider, "w", time.Since(st))
	}

	// GetChunk
	st = time.Now()
	v0, _, err := c.GetChunk(k)

	if err != nil {
		log.Error("[cloud_test] GetChunk", "provider", provider, "err", err)
	} else if bytes.Compare(v0, v) != 0 {
		log.Error("[cloud_test] Compare - strings don't match", "k", k, "retrieved", string(v0), "should be:", string(v))
	} else {
		log.Info("[cloud_test] GetChunk PASS", "provider", provider, "r", time.Since(st))
	}

	// (2) ZERO-BYTE CASE
	// SetChunk( k, []byte("") )
	k = Blakeb(k)
	err = c.SetChunk(ctx, k, emptyVal)
	if err != nil {
		log.Error("[cloud_test] SetChunk", "provider", provider, "err", err)
	} else {
		log.Info("[cloud_test] SetChunk PASS", "provider", provider, "r", time.Since(st))
	}

	// GetChunk(k) [with EMPTY value] should just work normally
	st = time.Now()
	v0, ok, err := c.GetChunk(k)
	if err != nil {
		log.Error("[cloud_test] GetChunk", "provider", provider, "err", err)
	} else if ok == false {
		log.Error("[cloud_test] GetChunk 0-byte value has ok = false", "provider", provider)
	} else if len(v0) != 0 {
		log.Error("[cloud_test] GetChunk 0-byte value has non-zero byte value", "provider", provider)
	} else {
		// GOOD! ok is true, 0-byte value returned
		log.Info("[cloud_test] GetChunk 0-byte value PASS", "provider", provider, "r", time.Since(st))
	}

	// (3) KEY NOT FOUND case: should NOT return an err -- instead, ok should be FALSE
	st = time.Now()
	v0, ok, err = c.GetChunk(missingKey)
	if err != nil {
		log.Error("[cloud_test] FAILURE GetChunk MISSING key returning error", "provider", provider, "err", err)
	} else if ok == true {
		log.Error("[cloud_test] FAILURE GetChunk on MISSING key ok = true", "provider", provider, "err", err)
	} else if len(v0) != 0 {
		log.Error("[cloud_test] FAILURE GetChunk on MISSING key returning value", "provider", provider, "err", err)
	} else {
		// GOOD! ok is false, err is nil
		log.Info("[cloud_test] GetChunk 0-byte value PASS ", "provider", provider, "r", time.Since(st))
	}

	// SetChunkBatch - write 2 chunks out, i=0 is a zero-byte value, i=1 is a non-zero byte value, i=2 is a missingKey (that we won't write)

	chunks := make([]*RawChunk, 0)
	wg := new(sync.WaitGroup)
	for i := 0; i < 2; i++ {
		switch i {
		case 0:
			k = Blakeb(Blakeb(Blakeb([]byte(prefix))))
			// empty Value
			v = []byte("")
		case 1:
			k = Blakeb(Blakeb(Blakeb(Blakeb([]byte(prefix)))))
			v = Blakeb([]byte(prefix))
		}
		chunks = append(chunks, &RawChunk{ChunkID: k, Value: v, Wg: wg})
	}
	wg.Add(2)
	err = c.SetChunkBatch(context.TODO(), chunks)
	if err != nil {
		log.Error("[cloud_test] FAILURE SetChunkBatch", "provider", provider, "err", err)
	} else {
		log.Info("[cloud_test] SetChunkBatch of 0-byte and N-byte value PASS", "provider", provider, "w", time.Since(st))
	}
	wg.Wait()

	// GetChunkBatch
	st = time.Now()
	chunksR := make([]*RawChunk, 0)
	for i := 0; i < 2; i++ {
		chunksR = append(chunksR, &RawChunk{ChunkID: chunks[i].ChunkID})
	}

	prefix = fmt.Sprintf("some random value %s", time.Now())
	missingKey = Blakeb([]byte(prefix))

	chunksR = append(chunksR, &RawChunk{ChunkID: missingKey})
	err = c.GetChunkBatch(chunksR)
	if err != nil {
		log.Error("[cloud_test] GetChunkBatch 3 chunk batch returned err", "provider", provider, "err", err)
	} else {
		// check empty byte key
		if chunksR[0].OK == false {
			log.Error("[cloud_test] GetChunkBatch 0-byte value has ok = false", "provider", provider)
		} else if len(chunksR[0].Value) != 0 {
			log.Error("[cloud_test] GetChunkBatch 0-byte value has non-zero byte value", "provider", provider)
		} else {
			// GOOD! ok is true, 0-byte value returned
			log.Info("[cloud_test] GetChunkBatch 0-byte value PASS", "provider", provider)
		}

		// check NORMAL key
		if chunksR[1].OK == false {
			log.Error("[cloud_test] FAILURE GetChunkBatch on NORMAL key ok = true", "provider", provider)
		} else if bytes.Compare(chunksR[1].Value, chunks[1].Value) != 0 {
			log.Error("[cloud_test] FAILURE GetChunkBatch on NORMAL key mismatch", "provider", provider)
		} else {
			// GOOD! ok is true, matching value returned
			log.Info("[cloud_test] GetChunkBatch NORMAL key PASS", "provider", provider)
		}
		// check MISSING key
		if chunksR[2].OK == true {
			log.Info("[cloud_test] GetChunkBatch MISSING key PASS", "v", string(chunksR[2].Value))
			log.Error("[cloud_test] FAILURE GetChunkBatch on MISSING key ok = true\n", provider)
		} else if len(chunksR[2].Value) != 0 {
			log.Error("[cloud_test] FAILURE GetChunkBatch on MISSING key returning value\n", provider)
		} else {
			// GOOD! ok is false, err is nil
			log.Info("[cloud_test] GetChunkBatch MISSING key PASS", "provider", provider)
		}
	}

	// SetChunkBatch
	st = time.Now()
	NCHUNKS := 10 //5 //74 //100
	chunks = make([]*RawChunk, 0)
	wg = new(sync.WaitGroup)
	for i := 0; i < NCHUNKS; i++ {
		v := Blakeb([]byte(fmt.Sprintf("123456789j%d", i)))
		k := Blakeb(v)
		chunks = append(chunks, &RawChunk{ChunkID: k, Value: v, Wg: wg})
	}
	wg.Add(NCHUNKS)

	err = c.SetChunkBatch(context.TODO(), chunks)
	if err != nil {
		log.Error("[cloud_test] SetChunkBatch err", "provider", provider, err)
	} else {
		log.Info("[cloud_test] SetChunkBatch", "provider", provider, "w", time.Since(st))
	}
	wg.Wait()

	// GetChunkBatch
	st = time.Now()
	chunksR = make([]*RawChunk, 0)
	for i := 0; i < NCHUNKS; i++ {
		v := Blakeb([]byte(fmt.Sprintf("123456789j%d", i)))
		k := Blakeb(v)
		chunksR = append(chunksR, &RawChunk{ChunkID: k})
	}
	err = c.GetChunkBatch(chunksR)
	if err != nil {
		log.Error("[cloud_test] GetChunkBatch %v", provider, err)
	}
	nerrors := 0
	for i := 0; i < NCHUNKS; i++ {
		//fmt.Printf("[cloud_test:%s] GetChunkBatch Compare %d: %x %x\n", provider, i, chunks[i].Value, chunksR[i].Value)
		if bytes.Compare(chunks[i].Value, chunksR[i].Value) != 0 {
			log.Error("[cloud_test] GetChunkBatch Compare mismatch", "provider", provider, "i", i)
			nerrors++
		}
	}
	if nerrors == 0 {
		log.Info("[cloud_test] GetChunkBatch", "provider", provider, "w", time.Since(st))
	}

}
