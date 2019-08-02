// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	chunker "github.com/ethereum/go-ethereum/swarm/storage"
	metrics "github.com/rcrowley/go-metrics"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
	"github.com/wolkdb/cloudstore/wolk/merkletree"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	maxChunksInChunkCache = 1000
	minCheckSize          = 10000
	chunkSize             = 4096
	numStoredNode         = 8
	degree                = 2
	maxFileSize           = 125000000
)

type BestBlockHash struct {
	vcnt        int
	blockHash   common.Hash
	isFinalized bool
}

type ChunkStore interface {
	QueueChunk(ctx context.Context, chunk []byte) (chunkID common.Hash, err error)
	FlushQueue() (err error)
	GetChunk(ctx context.Context, k []byte) ([]byte, bool, error)
	GetFile(context.Context, []byte) ([]byte, error)
	PutChunk(ctx context.Context, chunk *cloud.RawChunk, wg *sync.WaitGroup) (r chunker.Reference, err error)
}

type Storage struct {
	config *cloud.Config

	wolkChain *WolkStore //storage needs pointer back to preemptive at WolkStore

	edgeList [][]int

	// keyed by blockHash
	stateDBCacheMu   sync.RWMutex
	stateDBCache     map[common.Hash]*StateDB
	blockHashes      map[uint64]common.Hash
	maxVote          map[uint64]common.Hash
	blockHashesVotes map[uint64]map[common.Hash]int
	lastFlush        uint64
	operatorKey      *ecdsa.PrivateKey
	conn             net.PacketConn

	// chunk Cache
	chunkCache     map[common.Hash][]byte
	chunkCacheTime map[common.Hash]time.Time
	chunkCacheMu   sync.RWMutex

	cl cloud.Cloudstore

	// holds registry of nodes, currently fixed but should be made dynamic
	Registry []RegisteredNode

	muStorageReady sync.RWMutex
	storageReady   bool

	// ECDSA Key in raw and JWK ready form
	operatorECDSAKey     *ecdsa.PrivateKey
	operatorECDSAAddress common.Address
	jwkPrivateKey        *jose.JSONWebKey
	jwkPublicKeyString   string

	// Storage Accounting (SWAP) - A node maintains an individual balance with every other node
	// Only messages which have a price will be accounted for
	lock     sync.RWMutex             //lock the balances
	balances map[common.Address]int64 //map of balances for each peer

	// this chunk queue holds chunks generated in CommitTo, then Flush clears it out with a call to SetChunkBatch
	closed        chan struct{}
	wg            *sync.WaitGroup
	httpclient    []*http.Client
	chunkQueue    *PutChunkData
	getChunkQueue map[string][]chan *cloud.RawChunk
	qcount        int
	queuemu       sync.Mutex
	qstatus       int64
	flushCh       chan int
	queueCh       chan *cloud.RawChunk

	latestCacheBN   uint64 //Latest cached {tentative, finalzied} block number
	latestCacheBNMu sync.RWMutex

	devParam int
}

func NewStorage(config *cloud.Config, genesisConfig *GenesisConfig, chain *WolkStore) (self *Storage, err error) {
	// setup
	jwkPrivateKey, jwkPublicKeyString, operatorECDSAAddress, err := crypto.JWKSetupECDSA(config.OperatorECDSAKey)
	if err != nil {
		return self, err
	}
	self = &Storage{
		operatorECDSAKey:     config.OperatorECDSAKey,
		operatorECDSAAddress: operatorECDSAAddress,
		jwkPrivateKey:        jwkPrivateKey,
		jwkPublicKeyString:   jwkPublicKeyString,
		closed:               make(chan struct{}),
		wg:                   &sync.WaitGroup{},
		config:               config,
		stateDBCache:         make(map[common.Hash]*StateDB),
		blockHashes:          make(map[uint64]common.Hash),
		blockHashesVotes:     make(map[uint64]map[common.Hash]int),
		lastFlush:            0,
		wolkChain:            chain,
		balances:             make(map[common.Address]int64),
		chunkCache:           make(map[common.Hash][]byte),
		chunkCacheTime:       make(map[common.Hash]time.Time),
		queueCh:              make(chan *cloud.RawChunk, 10000),
		flushCh:              make(chan int),
		edgeList:             buildDRG(degree, maxFileSize/32+degree, 0),
	}

	cloudstore, err := cloud.NewCloudstore(config)
	if err != nil {
		return self, err
	}
	self.cl = cloudstore

	self.Registry = make([]RegisteredNode, len(genesisConfig.Registry))
	for i := 0; i < len(genesisConfig.Registry); i++ {
		self.Registry[i] = convertSerializedNode(genesisConfig.Registry[i])
	}

	self.httpclient = make([]*http.Client, len(self.Registry))

	for i := 0; i < len(self.Registry); i++ {
		self.httpclient[i] = &http.Client{Timeout: time.Second * 5, Transport: DefaultTransport}
	}

	self.chunkQueue = new(PutChunkData)
	self.chunkQueue.chunkData = make([][]*cloud.RawChunk, len(self.Registry))
	self.getChunkQueue = make(map[string][]chan *cloud.RawChunk)

	//TODO
	//self.devParam = len(self.Registry)/8 - 1

	////////// Temporary
	self.devParam = 3

	go self.QueueLoop()

	return self, nil
}

// flush out stateDBCache (old Blocks) + chunkCacheTime (chunks that have not been read/written for 60)
func (storage *Storage) flushStateDBCache(minBlockNumber uint64) {
	storage.stateDBCacheMu.Lock()
	for hash, statedb := range storage.stateDBCache {
		if statedb.block != nil {
			if statedb.block.BlockNumber < minBlockNumber {
				log.Trace("[storage:FlushCache] deleted statedb", "minBlockNumber", minBlockNumber, "bn", statedb.block.BlockNumber)
				delete(storage.stateDBCache, hash)
			}
		}
	}
	storage.lastFlush = minBlockNumber //TODO: flush only completed paths
	storage.stateDBCacheMu.Unlock()
}

func (storage *Storage) flushChunkCache() {
	storage.chunkCacheMu.RLock()
	if len(storage.chunkCacheTime) < maxChunksInChunkCache { // TODO: increase this
		storage.chunkCacheMu.RUnlock()
		return
	}
	storage.chunkCacheMu.RUnlock()

	storage.chunkCacheMu.Lock()
	for chunkID, tm := range storage.chunkCacheTime {
		if time.Since(tm) > 600*time.Second {
			delete(storage.chunkCacheTime, chunkID)
			delete(storage.chunkCache, chunkID)
			log.Trace("[storage:FlushCache] deleted chunkID", "chunkID", chunkID.Hex(), "tm", time.Since(tm))
		}
	}
	storage.chunkCacheMu.Unlock()
}

func (storage *Storage) StateDBCacheSize() (sz int) {
	storage.stateDBCacheMu.RLock()
	sz = len(storage.stateDBCache)
	storage.stateDBCacheMu.RUnlock()
	return sz
}

func (storage *Storage) ChunkCacheSize() (sz int) {
	storage.chunkCacheMu.RLock()
	sz = len(storage.chunkCache)
	storage.chunkCacheMu.RUnlock()
	return sz
}

type PutChunkData struct {
	chunkData [][]*cloud.RawChunk
	mu        sync.RWMutex
}

type GetChunkData struct {
	chunkData *cloud.RawChunk
	chunkCh   chan *cloud.RawChunk
}

type SetShareResponse struct {
	MerkleRoot string `json:"mr,omitempty"`
	Hash       string `json:"h,omitempty"`
	Len        int    `json:"len,omitempty"`
}

type Share struct {
	idx      uint64
	response []byte
	ok       bool
	success  bool
}

func ValidKey(k []byte) bool {
	if bytes.Compare(k, []byte("0000000000000000000000000000000000000000000000000000000000000000")) == 0 {
		return false
	}
	return true
}

func (wstore *Storage) SignRequest(req *http.Request, body []byte) (err error) {
	msg := PayloadBytes(req, body)
	signature, err := crypto.JWKSignECDSA(wstore.jwkPrivateKey.Key.(*ecdsa.PrivateKey), msg)
	if err != nil {
		log.Error("[storage:SignRequest] JWKSignECDSA", "err", err)
		return err
	}
	req.Header.Set("Sig", fmt.Sprintf("%x", signature))
	req.Header.Set("Requester", wstore.jwkPublicKeyString)
	//req.Header.Set("Msg", common.Bytes2Hex([]byte(msg)))
	//log.Trace("[storage:SignRequest]", "sig", signature, "requester", wstore.jwkPublicKeyString, "msg", common.Bytes2Hex([]byte(msg)))
	return nil
}

func (wstore *Storage) GetOperatorECDSAAddress() common.Address {
	return wstore.operatorECDSAAddress
}

// return MAX_VALIDATORS nodes in the neighborhood
func (wstore *Storage) GetRegisteredNodes(chunkID []byte) []*RegisteredNode {
	out := make([]*RegisteredNode, 0)
	for i := 0; i < len(wstore.Registry); i++ {
		out = append(out, &(wstore.Registry[i]))
		if len(out) >= MAX_VALIDATORS {
			return out
		}
	}
	return out
}

func (wstore *Storage) Close() {
	wstore.cl.Close()
}

/* TODO: Bring this back with the registry
func (wstore *Storage) getNodeChunkID(chunkID []byte) []byte {
	idxb := wolkcommon.UIntToByte(uint64(wstore.consensusIdx))
	return append(chunkID, idxb...)
}
*/
func (wstore *Storage) SetShare(octx context.Context, chunk []byte) (chunkID []byte, chunkKey []byte, chunklen int, err error) {
	chunkID = wolkcommon.Computehash(chunk)
	log.Chunk(chunkID, "SetShare")
	log.Trace("[storage:SetShare] in", "chunkid", fmt.Sprintf("%x", chunkID))
	wstore.chunkCacheMu.Lock()
	cID := common.BytesToHash(chunkID)
	wstore.chunkCache[cID] = chunk
	wstore.chunkCacheTime[cID] = time.Now()
	wstore.chunkCacheMu.Unlock()

	start := time.Now()

	metrics.GetOrRegisterCounter("setShare", nil).Inc(1)
	bytes := ldb_set(len(chunk), chunk)
	err = wstore.cl.SetChunk(octx, chunkID, bytes)
	if err != nil {
		log.Error("[storage:SetShare] cloud SetChunk Error", "chunkID", fmt.Sprintf("%x", chunkID), "err", err)
		return chunkID, chunkID, 0, err
	}
	log.Trace("[storage:SetShare]", "chunkid", fmt.Sprintf("%x", chunkID), "time", time.Since(start))
	return chunkID, chunkID, len(chunk), nil
}

func (wstore *Storage) GetShare(k []byte) (val []byte, len_chunk int, err error) {
	log.Chunk(k, "GetShare")
	log.Trace("[storage:GetShare]", "k", fmt.Sprintf("%x", k))
	metrics.GetOrRegisterCounter("getShare", nil).Inc(1)
	wstore.chunkCacheMu.RLock()
	chunkID := common.BytesToHash(k)
	if chunk, ok := wstore.chunkCache[chunkID]; ok {
		wstore.chunkCacheMu.RUnlock()
		return chunk, len(chunk), nil
	}
	wstore.chunkCacheMu.RUnlock()

	lcraw, ok, err := wstore.cl.GetChunk(k)
	if err != nil {
		log.Error("[storage:GetShare]", "k", k, "err", err)
		log.Chunk(k, fmt.Sprintf("GetShare|err %s", err))
		return val, len_chunk, err
	}
	if ok == false {
		// TODO: TREAT THIS PROPERLY
		// ERRORCHECK: log.Error("GetShare-ok false")
		log.Trace("[storage:GetShare] no result", "k", fmt.Sprintf("%x", k))
		log.Chunk(k, fmt.Sprintf("GetShare|not found"))
		return val, 0, fmt.Errorf("[storage:GetShare] not found")
	}
	len_chunk, val, err = ldb_get(lcraw)
	if err != nil {
		log.Error("[storage:GetShare] ldb_get", "err", err)
		log.Chunk(k, fmt.Sprintf("GetShare|err ldb_get %s", err))
		return val, len_chunk, err
	}
	h := wolkcommon.Computehash(val)
	if bytes.Compare(h, k) != 0 {
		log.Chunk(k, fmt.Sprintf("GetShare|MISMATCH HASH %x", h))
		log.Error("[storage:GetShare] MISMATCH HASH", "chunkID", fmt.Sprintf("%x", k), "hash(chunk)", fmt.Sprintf("%x", wolkcommon.Computehash(val)), "len(val)", len(val))
	}
	log.Chunk(k, fmt.Sprintf("GetShare|success %d", len_chunk))
	log.Trace("[storage:GetShare] done", "k", fmt.Sprintf("%x", k), "len", len_chunk)
	return val, len_chunk, nil
}

func (wstore *Storage) ChunkSearch(k []byte) (v []byte, okLocal bool, okRemote bool, okMemory bool, err error) {
	/*
		log.Chunk(k, "ChunkSearch")
		v, okLocal, okRemote, err = wstore.cl.ChunkSearch(k)
		if err != nil {
			return v, okLocal, okRemote, okMemory, err
		}
		chunkID := common.BytesToHash(k)
		wstore.chunkCacheMu.RLock()
		if _, ok := wstore.chunkCache[chunkID]; ok {
			okMemory = true
		}*/
	return v, okLocal, okRemote, okMemory, nil
}

func (wstore *Storage) GetChunk(ctx context.Context, k []byte) ([]byte, bool, error) {
	log.Chunk(k, "GetChunk")
	log.Trace("[storage:GetChunk] in", "k", fmt.Sprintf("%x", k))
	// store in chunkCache
	chunkID := common.BytesToHash(k)

	if wstore.config.ConsensusIdx%8 > 0 { // could be: wstore.skipMemory so that minter has a memorycache
		wstore.chunkCacheMu.RLock()
		if chunk, ok := wstore.chunkCache[chunkID]; ok {
			wstore.chunkCacheMu.RUnlock()
			return chunk, true, nil
		}
		wstore.chunkCacheMu.RUnlock()
	}

	//if !wstore.IsReady {
	//	return []byte{}, false, fmt.Errorf("Storage not ready yet.")
	//}
	start := time.Now()
	metrics.GetOrRegisterCounter("getChunk", nil).Inc(1)

	if !ValidKey(k) {
		return nil, false, fmt.Errorf("invalid key %x", k)
	}

	var val []byte

	var mu sync.Mutex
	decodech := make(chan *cloud.RawChunk, 1)
	closed := false
	not_oks := 0
	var quit chan interface{}
	quit = make(chan interface{})
	ctx, cancel := context.WithTimeout(context.Background(), interregionMaxGetShareTime)

	for i, r := range wstore.Registry {
		if wstore.Registry[i].GetStorageNode() == "" {
			continue
		}
		if (int(k[0])^i)&wstore.devParam != 0 {
			continue
		}
		go func(i uint64, url string) {
			log.Trace("[storage:GetChunk] Init", "i", i, "url", url)
			var share Share
			share.idx = i
			share.success = false
			var timeoutrate time.Duration
			timeoutrate = interregionMaxGetShareTime
			st := time.Now()
			tmr := metrics.GetOrRegisterTimer(fmt.Sprintf("%d.getShare", i), nil)

			for cnt := 0; cnt < retrySetShare; cnt++ {
				nstart := time.Now()
				req, err := http.NewRequest(http.MethodGet, url, nil)
				wstore.SignRequest(req, []byte{})
				signtime := time.Since(nstart)
				req.Cancel = ctx.Done()
				resp, err := wstore.httpclient[i].Do(req)
				if err != nil {
					mu.Lock()
					if !closed {
						log.Error("[storage:GetChunk] Do-closed", "i", i, "chunkID", chunkID, "tm", time.Since(st), "cnt", cnt, "err", err)
						not_oks++
						if not_oks >= len(wstore.Registry) && !closed && cnt == retrySetShare-1 {
							closed = true
							close(quit)
						}
					}
					mu.Unlock()
				} else if err == nil {
					len_chunk_str := resp.Header.Get("Chunk-Len")
					tmr.Update(time.Since(st))
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						//ERRORCHECK: log.Error("getchunk-err0", "i", i, "chunkID", chunkID, "tm", time.Since(st), "cnt", cnt, "err", err)
					}
					resp.Body.Close()
					statuscode := resp.StatusCode
					log.Trace("[storage:GetChunk] Success", "i", i, "chunkID", chunkID, "tm", time.Since(st), "statuscode", statuscode, "time", time.Since(nstart), "stime", signtime)
					if statuscode == 404 {
						// ERRORCHECK log.Error("getchunk-err404", "i", i, "chunkID", chunkID, "tm", time.Since(st), "status", statuscode, "cnt", cnt)
						mu.Lock()
						not_oks++
						//if not_oks > len(wstore.Registry)*2/3 && !closed {
						if not_oks > numStoredNode*2/3 && !closed {
							closed = true
							close(quit)
						}
						mu.Unlock()
						break
						// don't return inside of goroutine
						//return
					} else if statuscode == 503 {
						mu.Lock()
						not_oks++
						if not_oks >= len(wstore.Registry) && !closed && cnt == retrySetShare-1 {
							closed = true
							close(quit)
						}
						mu.Unlock()
						// TODO: handle 503
						// don't return inside of goroutine
						//return
					} else if statuscode == 200 {
						len_chunk2, err := strconv.Atoi(len_chunk_str)
						if err == nil {
							if len_chunk2 > 0 {

							}
						}
					}
					if err != nil {
						// ERRORCHECK: log.Error("getchunk-err1", "i", i, "chunkID", chunkID, "tm", time.Since(st), "cnt", cnt, "err", err)
					} else if statuscode >= 200 && statuscode <= 299 {
						share.success = true
						share.response = body
						if time.Since(st) > timeoutrate {
							log.Error("[storage:GetChunk] GetShare SLOW", "i", i, "chunkID", chunkID, "tm", time.Since(st), "cnt", cnt)
						}
						mu.Lock()
						if !closed {
							sd := time.Now()
							decoder := new(cloud.RawChunk)
							decoder.ChunkID = k
							decoder.Value = share.response
							decoder.OK = true
							closed = true
							decodech <- decoder
							log.Trace("[storage:GetChunk] determined", "i", i, "chunkID", chunkID, "tm", time.Since(st), "checktime", time.Since(sd))
						}
						mu.Unlock()
						break
					} else {
						log.Trace("[storage:GetChunk]", "i", i, "chunkID", chunkID, "tm", time.Since(st), "status", statuscode, "cnt", cnt, "status", resp.Status, "body", fmt.Sprintf("%s", body))
					}
				}
			}
			//ch <- share
		}(uint64(i), fmt.Sprintf("%s://%s:%d/wolk/share/%x", r.GetScheme(), r.GetStorageNode(), r.GetPort(), k))
	}

	//log.Trace("[storage:GetChunk] waiting", "chunkID", chunkID)
	ticker := time.NewTicker(interregionMaxGetShareTime)
	select {
	case decoder := <-decodech:
		dstart := time.Now()
		decoded := decoder.Value
		cancel()
		log.Trace("[storage:GetChunk] RESULT", "chunkID", chunkID, "hash(return)", fmt.Sprintf("%x", wolkcommon.Computehash(decoded)), "len(decoded)", len(decoded), "time", time.Since(start), "decode time", time.Since(dstart))
		log.Chunk(k, fmt.Sprintf("GetChunk|Got result len = %d", len(decoded)))
		wstore.chunkCacheMu.Lock()
		wstore.chunkCache[chunkID] = decoded
		wstore.chunkCacheTime[chunkID] = time.Now()
		wstore.chunkCacheMu.Unlock()
		return decoded, true, nil
	case <-quit:
		cancel()
		log.Trace("[storage:GetChunk] quit", "chunkID", chunkID, "k", fmt.Sprintf("%x", k), "time", time.Since(start))
		log.Chunk(k, "GetChunk|quit")
		return val, false, nil
	case <-ticker.C:
		mu.Lock()
		closed = true
		mu.Unlock()
		cancel()
		log.Trace("[storage:GetChunk] timeout", "chunkID", chunkID, "k", fmt.Sprintf("%x", k), "time", time.Since(start))
		log.Chunk(k, "GetChunk|timeout")
		return val, false, fmt.Errorf("[Storage:GetChunk] Inactive cloudstore node")
	}
}

func (wstore *Storage) SetChunk(octx context.Context, chunk []byte) (chunkID common.Hash, err error) {
	start := time.Now()
	chunkKey := wolkcommon.Computehash(chunk)
	log.Chunk(chunkKey, "SetChunk")
	log.Trace("[storage:SetChunk]", "chunkKey", fmt.Sprintf("%x", chunkKey))

	// store in chunkCache
	chunkID = common.BytesToHash(chunkKey)
	/*
		wstore.chunkCacheMu.Lock()
		wstore.chunkCache[chunkID] = chunk
		wstore.chunkCacheTime[chunkID] = time.Now()
		wstore.chunkCacheMu.Unlock()
	*/

	log.Trace("[storage:SetChunk]", "k", fmt.Sprintf("%x", chunkKey), "len(chunk)", len(chunk))
	metrics.GetOrRegisterCounter("setChunk", nil).Inc(1)

	var mu sync.Mutex
	var svrCount int
	var quit chan struct{}
	closed := false
	success := 0
	svrCount = MAX_VALIDATORS
	quit = make(chan struct{})
	ctxintra, cancelintra := context.WithDeadline(context.Background(), time.Now().Add(intraregionMaxSetShareTime))
	ctxinter, cancelinter := context.WithDeadline(context.Background(), time.Now().Add(interregionMaxSetShareTime))

	for i, _ := range wstore.Registry {
		if wstore.Registry[i].GetStorageNode() == "" {
			continue
		}
		if (int(chunkKey[0])^i)&wstore.devParam != 0 {
			continue
		}
		go func(i uint64, url string) {
			log.Trace("[storage:SetChunk]", "i", i, "url", url, "k", fmt.Sprintf("%x", chunkKey))
			nstart := time.Now()
			var err error
			var ctx context.Context

			var timeoutrate time.Duration
			if true { // wstore.config.Region == wstore.Registry[i].region {
				timeoutrate = intraregionMaxSetShareTime
				ctx = ctxintra
			} else {
				timeoutrate = interregionMaxSetShareTime
				ctx = ctxinter
			}
			for cnt := 0; cnt < retrySetShare; cnt++ {
				var share Share
				share.idx = i
				share.success = false
				st := time.Now()
				tmr := metrics.GetOrRegisterTimer(fmt.Sprintf("%d.setShare", i), nil)
				//ctx, _ := context.WithCancel(context.TODO())
				body := bytes.NewReader(chunk)
				req, reserr := http.NewRequest(http.MethodPost, url, body)
				if reserr != nil {
					// handle err
					log.Error("[storage:SetChunk] NewRequest", "i", i, "err1", reserr)
					if cnt == 2 {
						mu.Lock()
						err = reserr
						mu.Unlock()
					}
				} else {
					sigerr := wstore.SignRequest(req, chunk)
					stime := time.Since(nstart)
					if sigerr != nil {
						log.Error("SignRequest", "err", err)
					}
					req.Header.Set("Content-Type", "application/json")
					//ctx, cancel := context.WithTimeout(context.Background(), timeoutrate)
					//defer cancel()
					req.Cancel = ctx.Done()

					resptime := time.Now()
					resp, reserr := wstore.httpclient[i].Do(req)
					//serverlist = serverlist + ":" + fmt.Sprintf("%d", i)

					if reserr != nil {
						//statuscode := resp.StatusCode
						//if statuscode == 503 {
						// TODO: handle 503
						// }
						mu.Lock()
						if !closed {
							log.Error("[storage:SetChunk] Do", "i", i, "k", fmt.Sprintf("%x", chunkKey), "err", reserr)
						}
						mu.Unlock()
						if cnt == retrySetShare-1 {
							mu.Lock()
							err = reserr
							mu.Unlock()
						}
					} else {
						body2, reserr := ioutil.ReadAll(resp.Body)
						resp.Body.Close()
						if reserr != nil {
							log.Error("[storage:SetChunk] ReadAll", "i", i, "err", reserr)
							if cnt == retrySetShare-1 {
								mu.Lock()
								err = reserr
								mu.Unlock()
							}
						} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
							mu.Lock()
							err = fmt.Errorf("SetShare res status = %v", resp.StatusCode)
							log.Error("[storage:SetChunk] StatusCode Error", "i", i, "url", url, "body2", fmt.Sprintf("%s", body2), "StatusCode", resp.StatusCode, "time", time.Since(resptime))
							mu.Unlock()
						} else {
							var c SetShareResponse
							jerr := json.Unmarshal(body2, &c)
							if jerr != nil {
								log.Error("[storage:SetChunk] Unmarshal SetShareResponse", "i", i, "err", jerr, "body2", fmt.Sprintf("%s", body2), "StatusCode", resp.StatusCode, "time", time.Since(resptime))
							} else {
								share.success = true
								share.response = body2
								log.Debug("[storage:SetChunk] SetShareResponse", "i", i, "chunkKey", fmt.Sprintf("%x", chunkKey), "lenshare", c.Len, "time", time.Since(resptime), "stime", stime)
								mu.Lock()
								if share.success {
									success++
								}
								mu.Unlock()
							}
							tmr.Update(time.Since(st))
						}
					}
				}
			}
			log.Debug("[storage:SetChunk]", "i", i, "tm", time.Since(start))
			mu.Lock()
			svrCount = svrCount - 1
			if time.Since(start) > timeoutrate {
				log.Debug("[storage:SetChunk] SetShare SLOW", "i", i, "closed", closed, "tm", time.Since(start))
			}
			if svrCount <= MAX_VALIDATORS-MIN_SET_SUCCESS && closed == false && success >= Q {
				closed = true
				log.Trace("[storage:SetChunk] DONE", "Closed", closed)
				close(quit)
			}
			mu.Unlock()
		}(uint64(i), fmt.Sprintf("%s://%s:%d/wolk/share", wstore.Registry[i].GetScheme(), wstore.Registry[i].GetStorageNode(), wstore.Registry[i].GetPort()))
	}

	ticker := time.NewTicker(time.Second * 5)
	select {
	case <-quit:
		mu.Lock()
		closed = true
		log.Trace("[storage:SetChunk] quit", "Closed = ", closed, "time", time.Since(start))
		mu.Unlock()
		return chunkID, nil
	case <-ticker.C:
		mu.Lock()
		log.Trace("[storage:SetChunk] ticker", "Closed", closed, "time", time.Since(start))
		closed = true
		if err != nil && success > Q {
			err = nil
		}
		mu.Unlock()
		cancelinter()
		cancelintra()
		return chunkID, err
	}
}

func (wstore *Storage) SetChunkBatchToSetShareBatch(octx context.Context, chunklist [][]byte) (err error) {
	log.Trace("storage:SetChunkBatchToSetShareBatch] in", "len", len(chunklist))
	start := time.Now()

	var mu sync.Mutex
	var svrCount int
	var quit chan struct{}
	closed := false
	success := 0

	quit = make(chan struct{})
	ctxintra, cancelintra := context.WithDeadline(context.Background(), time.Now().Add(intraregionMaxSetShareTime))
	ctxinter, cancelinter := context.WithDeadline(context.Background(), time.Now().Add(interregionMaxSetShareTime))

	for i := 0; i < len(wstore.Registry); i++ {
		if wstore.Registry[i].GetStorageNode() == "" {
			continue
		}
		go func(i uint64, url string) {
			//var statuscode int
			var ctx context.Context

			chunk := chunklist[int(i)&wstore.devParam]
			if len(chunk) == 0 {
				return
			}

			var timeoutrate time.Duration
			if true {
				timeoutrate = intraregionMaxSetShareTime
				ctx = ctxintra
			} else {
				timeoutrate = interregionMaxSetShareTime
				ctx = ctxinter
			}
			for cnt := 0; cnt < retrySetShare; cnt++ {
				var share Share
				share.idx = i
				share.success = false
				st := time.Now()
				tmr := metrics.GetOrRegisterTimer(fmt.Sprintf("%d.setShare", i), nil)
				//ctx, _ := context.WithCancel(context.TODO())
				body := bytes.NewReader(chunk)
				req, reserr := http.NewRequest(http.MethodPost, url, body)
				if reserr != nil {
					// handle err
					log.Error("[storage:SetChunkBatchToSetShareBatch] NewRequest", "i", i, "err1", reserr)
					if cnt == 2 {
						mu.Lock()
						err = reserr
						mu.Unlock()
					}
				} else {
					sigerr := wstore.SignRequest(req, chunk)
					if sigerr != nil {
						log.Error("SignRequest", "err", err)
					}
					req.Header.Set("Content-Type", "application/json")
					//ctx, cancel := context.WithTimeout(context.Background(), timeoutrate)
					//defer cancel()
					req.Cancel = ctx.Done()
					log.Debug("[storage:SetChunk] Sending chunk data", "url", url, "len", len(chunk))
					//log.Debug(fmt.Sprintf("[storage:SetChunk] Req is: %+v", req))
					resp, reserr := wstore.httpclient[i].Do(req)

					if reserr != nil {
						//statuscode := resp.StatusCode
						//if statuscode == 503 {
						// TODO: handle 503
						// }
						mu.Lock()
						if !closed {
							log.Error("[storage:SetChunkBatchToSetShareBatch] Do", "i", i, "err", reserr)
						}
						mu.Unlock()
						if cnt == retrySetShare-1 {
							mu.Lock()
							err = reserr
							mu.Unlock()
						}
					} else {
						body2, reserr := ioutil.ReadAll(resp.Body)
						resp.Body.Close()
						if reserr != nil {
							log.Error("[storage:SetChunkBatchToSetShareBatch] ReadAll", "i", i, "err", reserr)
							if cnt == retrySetShare-1 {
								mu.Lock()
								err = reserr
								mu.Unlock()
							}
						} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
							mu.Lock()
							err = fmt.Errorf("SetShare res status = %v", resp.StatusCode)
							log.Error("[storage:SetChunkBatchToSetShareBatch] StatusCode Error", "i", i, "body2", fmt.Sprintf("%s", body2), "StatusCode", resp.StatusCode, "url", url)
							mu.Unlock()
						} else {
							var c SetShareResponse
							jerr := json.Unmarshal(body2, &c)
							if jerr != nil {
								log.Error("[storage:SetChunkBatchToSetShareBatch] Unmarshal SetShareResponse", "i", i, "url", url, "err", jerr, "body2", fmt.Sprintf("%s", body2), "StatusCode", resp.StatusCode)
							} else {
								share.success = true
								share.response = body2
								log.Debug("[storage:SetChunkBatchToSetShareBatch] SetShareResponse", "i", i, "lenshare", c.Len)
								mu.Lock()
								if share.success {
									success++
								}
								mu.Unlock()
							}
							tmr.Update(time.Since(st))
						}
					}
				}
			}
			log.Debug("[storage:SetChunkBatchToSetShareBatch]", "i", i, "tm", time.Since(start))
			mu.Lock()
			svrCount = svrCount - 1
			if time.Since(start) > timeoutrate {
				log.Debug("[storage:SetChunkBatchToSetShareBatch] SetShare SLOW", "i", i, "closed", closed, "tm", time.Since(start))
			}
			if svrCount <= MAX_VALIDATORS-MIN_SET_SUCCESS && closed == false && success >= Q {
				closed = true
				log.Trace("[storage:SetChunkBatchToSetShareBatch] DONE", "Closed", closed)
				close(quit)
			}
			mu.Unlock()
		}(uint64(i), fmt.Sprintf("%s://%s:%d/wolk/sbatch", wstore.Registry[i].GetScheme(), wstore.Registry[i].GetStorageNode(), wstore.Registry[i].GetPort()))
	}
	ticker := time.NewTicker(time.Second * 5)
	select {
	case <-quit:
		mu.Lock()
		closed = true // MUTEX NEEDED
		log.Trace("[storage:SetChunkBatchToSetShareBatch] quit", "Closed = ", closed, "time", time.Since(start))
		mu.Unlock()
		return nil
	case <-ticker.C:
		mu.Lock()
		log.Trace("[storage:SetChunkBatchToSetShareBatch] ticker", "Closed", closed, "time", time.Since(start))
		closed = true // MUTEX NEEDED
		if err != nil && success > Q {
			err = nil
		}
		mu.Unlock()
		cancelinter()
		cancelintra()
		return err
	}
} // MUTEX on err needed

// TODO: treat errors, optimize this
func (wstore *Storage) SetChunkBatch(ctx context.Context, chunklist [][]*cloud.RawChunk) (err error) {
	log.Trace("[storage:setChunkBatch] in", "len", len(chunklist))
	/*
		var chunkbatch [][]byte
		for _, chunk := range chunks{
			chunkbatch = append(chunkbatch, chunk.Value)
		}
	*/
	start := time.Now()
	var batchchunk [][]byte
	for _, chunks := range chunklist {
		chs := chunks
		b, err := json.Marshal(chs)
		if err != nil {
			log.Error("[storage:SetChunkBatch] json Marshal", "err", err)
			return err
		}
		batchchunk = append(batchchunk, b)
		//TODO : batch length

		/*
			for {
				if len(chs) < 200 {
					b, err := json.Marshal(chs)
					if err != nil {
						log.Error("[storage:SetChunkBatch] json Marshal", "err", err)
						return err
					}
					log.Trace("[storage:SetChunkBatch] finish a SetChunkBatchToSetShareBatch", "len", len(chs), "time", time.Since(start))
					go func(b []byte) {
						err = wstore.SetChunkBatchToSetShareBatch(context.TODO(), b)
						if err != nil {
							//TODO
							log.Error("[storage:SetChunkBatch] SetChunkBatchToSetShareBatch", "err", err)
							//		return err
						}
					}(b)
					break
				} else {
					b, err := json.Marshal(chs[:200])
					if err != nil {
						log.Error("[storage:SetChunkBatch] json Marshal", "err", err)
						return err
					}
					err = wstore.SetChunkBatchToSetShareBatch(context.TODO(), b)
					log.Trace("[storage:SetChunkBatch] finish b SetChunkBatchToSetShareBatch", "len", len(chs), "time", time.Since(start))
					go func(b []byte) {
						if err != nil {
							//TODO
							log.Error("[storage:SetChunkBatch] SetChunkBatchToSetShareBatch", "err", err)
							//return err
						}
					}(b)
					chs = chs[200:]
					if len(chs) == 0 {
						break
					}
					//time.Sleep(100 * time.Millisecond)
				}
			}
		*/
	}
	if err != nil {
		log.Error("[Storage:SetChunkBatch]", "err", err)
	}
	go func() {
		err := wstore.SetChunkBatchToSetShareBatch(context.TODO(), batchchunk)
		if err != nil {
			log.Error("[storage:SetChunkBatch] SetChunkBatchToSetShareBatch error", "err", err)
		}
	}()
	for _, chunks := range chunklist {
		for _, chunk := range chunks {
			log.Trace("[storage:SetChunkBatch] done", "chunkID", fmt.Sprintf("%x", chunk.ChunkID))
			if chunk.Wg != nil {
				log.Trace("[storage:SetChunkBatch] done", "chunkID", fmt.Sprintf("%x", chunk.ChunkID), "wg", chunk.Wg, "wgp", fmt.Sprintf("%p", chunk.Wg), "time", time.Since(start))
				if chunk.Error != nil {
					// WE NEED A PLAN
					log.Error("Chunk Err", "err", chunk.Error)
				}
				chunk.Wg.Done()
				log.Chunk(chunk.ChunkID, "SetChunkBatch:done")
			}
		}
	}
	return err
}

func (wstore *Storage) SetShareBatch(ctx context.Context, shares []*cloud.RawChunk) (err error) {
	log.Trace("[storage:SetShareBatch] in", "number", len(shares))
	if len(shares) == 0 {
		return nil
	}
	start := time.Now()
	/*
		for _, share := range shares {
			if bytes.Compare(share.ChunkID, wolkcommon.Computehash(share.Value)) == 0 {
				log.Error("[Storage:SetShareBatch]", "ChunkID", fmt.Sprintf("%x", share.ChunkID), "len", len(share.Value))
			}
		}
		var chunks []*cloud.RawChunk
		for _, share := range shares {
			chunks = append(chunks, &cloud.RawChunk{Value: share, ChunkID:wolkcommon.Computehash(share)})
		}
		for _, share := range shares {
			chunkID, _, _, err := wstore.SetShare(ctx, share)
			if err != nil{
				log.Error("[Storage:SetShareBatch]", "chunkID", chunkID)
				// TODO
			}
		}
	*/
	err = wstore.cl.SetChunkBatch(ctx, shares)
	log.Trace("[storage:SetShareBatch]", "number", len(shares), "time", time.Since(start), "err", err)
	return nil
}

// TODO: treat errors, optimize this
func (wstore *Storage) GetChunkBatch(chunks []*cloud.RawChunk) (resp []*cloud.RawChunk, err error) {
	log.Trace("[storage:GetChunkBatch] in", "nchunks", len(chunks))
	start := time.Now()
	for i, ch := range chunks {
		v, ok, errc := wstore.GetChunk(context.TODO(), ch.ChunkID)
		if errc == nil {
			chunks[i].OK = ok
			if ok {
				chunks[i].Value = v
			}
		} else {
			return resp, err
		}
	}
	log.Trace("[storage:GetChunkBatch] in", "nchunks", len(chunks), "time", time.Since(start))
	return chunks, nil
}

func ldb_set(len_chunk int, d []byte) []byte {
	lc := make([]byte, 8)
	binary.BigEndian.PutUint64(lc, uint64(len_chunk))
	return append(lc, d...)
}

func ldb_get(raw []byte) (len_chunk int, b []byte, err error) {
	if len(raw) >= 8 {
		return int(binary.BigEndian.Uint64(raw[0:8])), raw[8:], nil
	}
	return 0, raw, fmt.Errorf("Insufficient length %d", len(raw))
}

func (storage *Storage) getLatestCacheBN() (bn uint64) {
	storage.latestCacheBNMu.RLock()
	defer storage.latestCacheBNMu.RUnlock()
	return storage.latestCacheBN
}

func (storage *Storage) setLatestCacheBN(bn uint64) {
	storage.latestCacheBNMu.Lock()
	defer storage.latestCacheBNMu.Unlock()
	storage.latestCacheBN = bn
}

// StoreVerifiedBlock actually records the block + statedb in the statedbcache
func (self *Storage) StoreVerifiedBlock(block *Block, stateDB *StateDB, votesCnt int) {
	self.stateDBCacheMu.Lock()
	self.stateDBCache[block.Hash()] = stateDB
	self.blockHashes[block.BlockNumber] = block.Hash() // should be removed
	_, roundOK := self.blockHashesVotes[block.BlockNumber]
	if roundOK {
		//blockround already initiated
		oldCnt, certFound := self.blockHashesVotes[block.BlockNumber][block.Hash()]
		if certFound {
			// update votesCnt with higher votes
			if votesCnt > oldCnt {
				self.blockHashesVotes[block.BlockNumber][block.Hash()] = votesCnt
			}
		} else {
			//store votesCnt for new block
			self.blockHashesVotes[block.BlockNumber][block.Hash()] = votesCnt
		}

	} else {
		//blockround not initiated for the cache yet
		self.blockHashesVotes[block.BlockNumber] = make(map[common.Hash]int)
		self.blockHashesVotes[block.BlockNumber][block.Hash()] = votesCnt
	}
	self.stateDBCacheMu.Unlock()
	//update best hash

	if votesCnt >= expectedTokensFinal && block.BlockNumber-self.lastFlush > 10 {
		//flush every 10 blocks After reaching final threshold
		self.flushStateDBCache(block.BlockNumber - 10)
		self.flushChunkCache()
		//TODO: gc on votecnt
	} else if block.BlockNumber-self.lastFlush >= 50 {
		//cap stateDBCache at max of 50 rounds (rolling)
		self.flushStateDBCache(block.BlockNumber - 5)
		self.flushChunkCache()
	}
	if block.BlockNumber > self.getLatestCacheBN() {
		self.setLatestCacheBN(block.BlockNumber)
	}
	log.Trace("[storage:StoreVerifiedBlock]", "LatestCacheBN", self.getLatestCacheBN(), "bn", block.BlockNumber, "hash", block.Hash().Hex())
}

func (self *Storage) DumpCache() {
	self.stateDBCacheMu.RLock()
	log.Error("DumpCache", "yipes", len(self.blockHashes))
	for i, hash := range self.blockHashes {
		log.Info("DumpCache", "i", i, "hash", hash)
		fmt.Printf("%d => %x\n", i, hash)
	}
	self.stateDBCacheMu.RUnlock()
}

func (self *Storage) getBestHash(blockNumber uint64) (blockHash common.Hash, vcnt int, found bool) {
	self.stateDBCacheMu.RLock()
	defer self.stateDBCacheMu.RUnlock()
	blockHashes, roundOK := self.blockHashesVotes[uint64(blockNumber)]
	if !roundOK {
		return blockHash, vcnt, roundOK
	}
	cnt := 0
	for bHash, votecnt := range blockHashes {
		if votecnt > cnt {
			cnt = votecnt
			blockHash = bHash
		}
	}
	return blockHash, cnt, roundOK
}

func (self *Storage) setStateDBCache(blockHash common.Hash, statedb *StateDB) {
	self.stateDBCacheMu.Lock()
	defer self.stateDBCacheMu.Unlock()
	self.stateDBCache[blockHash] = statedb

}

func (self *Storage) getStateDBCache(blockHash common.Hash) (statedb *StateDB, ok bool) {
	self.stateDBCacheMu.RLock()
	defer self.stateDBCacheMu.RUnlock()
	statedb, ok = self.stateDBCache[blockHash]
	return statedb, ok
}

// getStateDB returns statedb given its block number
func (self *Storage) getStateDB(ctx context.Context, blockNumber int) (statedb *StateDB, ok bool, err error) {

	var requestedBN int
	var isPastBlk, forcedPreemptive bool

	// look in cache by blockNumber
	acceptPreemptive := self.wolkChain.GetIsPremptive()
	localBN := int(self.getLatestCacheBN())
	localFBN := int(self.wolkChain.LastFinalized())
	switch blockNumber {
	case LastConsensusState:
		//returns the last final/tentative state. Used by Consensus
		requestedBN = localBN
	case LocalBestState:
		//fallback to preemptive, potentially mixed by tentative states. Special Option. NO docs
		requestedBN = int(self.wolkChain.currentRound())
		log.Trace("[storage:getStateDB] preemptive request", "auto requestedBN", requestedBN)
	case PreemptiveState:
		//forcing preemptive. tentative are purposely ignored. Externally Accessible
		forcedPreemptive = true
		requestedBN = int(self.wolkChain.currentRound() + 10)
		log.Trace("[storage:getStateDB] forced preemptive request", "auto requestedBN", requestedBN)
	case LastFinalizedState:
		//currently returns the last known finalized block. Externally Accessible
		requestedBN = localFBN
	default:
		//User specified a targeted blockNumber. User should not resquest future stateDB without using special index
		if blockNumber > localBN {
			//This is the future/present block case. TODO: check for the newwolk case
			requestedBN = localBN
		} else {
			//past block
			requestedBN = blockNumber
			isPastBlk = true
		}
	}

	blockHash, vcnt, bhashFound := self.getBestHash(uint64(requestedBN))
	if bhashFound && !forcedPreemptive {
		log.Trace("[storage:getStateDB] getBestHash FOUND", "requestedBN", requestedBN, "bhash", blockHash, "vcnt", vcnt, "bhashFound", bhashFound)
		//stateDB-Hash available given the requestedBN
		statedb, ok = self.getStateDBCache(blockHash)
		if ok {
			log.Info("[storage:getStateDB] FOUND", "vcnt", vcnt)
			return statedb, true, nil
		}
		// make a new statedb on cache miss
		statedb, err = NewStateDB(ctx, self, blockHash)
		if err != nil {
			log.Error("[storage:getStateDB] ERR", "err", err)
			return statedb, false, err
		}
		self.setStateDBCache(blockHash, statedb)
		log.Warn("[storage:getStateDB] CREATED new statedb cache", "BN", requestedBN, "bHash", blockHash, "stateDB", statedb)
		return statedb, true, nil
	} else {
		log.Error("[storage:getStateDB] getBestHash NOT FOUND", "requestedBN", requestedBN, "bhash", blockHash, "vcnt", vcnt, "bhashFound", bhashFound)
		//stateDBCache unavailable
		//case1: past stateDB has been removed from cache. But can be re-initiated from indexer's FP-%d key
		if isPastBlk {
			debugcase := "FP"
			blkHash, fpErr := self.wolkChain.Indexer.getFinalizedPath(uint64(requestedBN))
			if fpErr != nil {
				self.DumpCache()
				log.Error("[storage:getStateDB:getFinalizedPath] FPKey not found", "requested BN", requestedBN)
				//case1a: Fall back to unverified stateDB
				blk, ok, err := self.wolkChain.Indexer.GetBlockByNumber(uint64(requestedBN))
				if err != nil || !ok {
					//case1a-I: pastblock missing. TODO: calling FP patch
					log.Error("[storage:getStateDB:GetBlockByNumber] Past block unavailable", "requested BN", requestedBN)
					return statedb, false, fmt.Errorf("Past Block Unavailable")
				} else {
					//case1a-II: reinitiate stateDB
					debugcase = "T.Local"
					blkHash = blk.Hash()
				}
			}
			//case1b: Fallback to FP block
			statedb, dbErr := NewStateDB(ctx, self, blkHash)
			if dbErr != nil {
				log.Error("[storage:getStateDB:NewStateDB] Init Error", "err", dbErr, "case", debugcase)
				return statedb, false, dbErr
			}
			//self.setStateDBCache(blkHash, statedb)
			log.Trace("[storage:getStateDB] stateDB load successfully", "requested BN", requestedBN, "hash", blkHash, "case", debugcase, "statedb", statedb)
			return statedb, true, nil
		} else {
			//case2: asking for future block. return preemptiveDB if preemptive is enabled
			if acceptPreemptive {
				preemptiveDB, err := self.getPreemptiveStateDB()
				if err != nil || preemptiveDB == nil {
					log.Error("[storage:getStateDB] PreemptiveStateDB NOT FOUND", "requested BN", requestedBN, "inputBN", blockNumber, "ERROR", err)
					return statedb, false, fmt.Errorf("PreemptiveStateDB NOT OK")
				}
				return preemptiveDB, true, nil
			} else {
				//TODO: verify this case
				log.Error("[storage:getStateDB] PreemptiveStateDB Not Allowed!", "requested BN", requestedBN, "inputBN", blockNumber, "ERROR", err)
				return statedb, false, fmt.Errorf("PreemptiveStateDB Not Allowed!")
			}
		}
	}
	// should not go here
}

//getPreemptiveStateDB return the preemptiveDB. preemptive state were available until next finalized round
func (wstore *Storage) getPreemptiveStateDB() (statedb *StateDB, err error) {
	wstore.wolkChain.muPreemptiveStateDB.RLock()
	statedb = wstore.wolkChain.PreemptiveStateDB
	wstore.wolkChain.muPreemptiveStateDB.RUnlock()
	return statedb, err
}

// Put stores the chunkData into Storage and returns the reference.
// Asynchronous function, the data will not necessarily be stored when it returns.
func (wstore *Storage) Put(ctx context.Context, chunkData chunker.ChunkData, wg *sync.WaitGroup) (r chunker.Reference, err error) {
	chunkKey := wolkcommon.Computehash(chunkData)
	//wstore.queuemu.Lock()
	//defer wstore.queuemu.Unlock()

	rawChunk := new(cloud.RawChunk)
	rawChunk.Value = chunkData
	rawChunk.ChunkID = wolkcommon.Computehash(chunkData)
	//wstore.chunkQueue.chunkData = append(wstore.chunkQueue.chunkData, &cloud.RawChunk{Value: chunkData, Wg: wg})
	wstore.AddQueue(&cloud.RawChunk{Value: chunkData, Wg: wg})
	//wstore.qcount++

	return chunker.Reference(chunkKey), nil
}

func (wstore *Storage) PutChunk(ctx context.Context, chunkData *cloud.RawChunk, wg *sync.WaitGroup) (r chunker.Reference, err error) {
	chunkData.ChunkID = wolkcommon.Computehash(chunkData.Value)
	log.Chunk(chunkData.ChunkID, "PutChunk")
	//wstore.queuemu.Lock()
	//defer wstore.queuemu.Unlock()
	chunkData.Wg = wg
	//chunkKey := wolkcommon.Computehash(chunkData.Value)
	//wstore.chunkQueue.chunkData = append(wstore.chunkQueue.chunkData, chunkData)
	wstore.AddQueue(chunkData)
	//wstore.qcount++

	return chunkData.ChunkID, nil
}

func (wstore *Storage) PutChunkWithChannel(ctx context.Context, chunkData []byte, ch chan *cloud.RawChunk) (r chunker.Reference, err error) {
	chunkID := wolkcommon.Computehash(chunkData)
	log.Chunk(chunkID, "PutChunkWithChannel")
	//wstore.queuemu.Lock()
	//defer wstore.queuemu.Unlock()
	var wg sync.WaitGroup
	wg.Add(1)
	chunkKey := wolkcommon.Computehash(chunkData)
	chunk := &cloud.RawChunk{ChunkID: chunkKey, Value: chunkData, Wg: &wg}
	//wstore.chunkQueue.chunkData = append(wstore.chunkQueue.chunkData, chunk)
	wstore.AddQueue(chunk)
	//wstore.qcount++
	go func(chunk *cloud.RawChunk, ch chan *cloud.RawChunk) {
		chunk.Wg.Wait()
		ch <- chunk
	}(chunk, ch)

	return chunkID, nil
}

func (wstore *Storage) PutChunkAsync(ctx context.Context, chunkData []byte, f func([]byte, error) bool) (r chunker.Reference, err error) {
	//wstore.queuemu.Lock()
	//defer wstore.queuemu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	chunkKey := wolkcommon.Computehash(chunkData)
	log.Chunk(chunkKey, "PutChunkAsync")

	wg.Add(1)
	chunk := &cloud.RawChunk{ChunkID: chunkKey, Value: chunkData, Wg: &wg}
	//wstore.chunkQueue.chunkData = append(wstore.chunkQueue.chunkData, chunk)
	wstore.AddQueue(chunk)
	//wstore.qcount++

	go func(chunk *cloud.RawChunk) {
		chunk.Wg.Wait()
		if !f(chunk.ChunkID, chunk.Error) {
			// TODO
			//cancel()
		}
	}(chunk)

	return chunk.ChunkID, nil
}

func (wstore *Storage) AddQueue(chunk *cloud.RawChunk) {
	start := time.Now()
	wstore.queueCh <- chunk
	log.Trace("[storage:AddQueue]", "time", time.Since(start))
	/*
		wstore.queuemu.Lock()
		wstore.chunkQueue.chunkData = append(wstore.chunkQueue.chunkData, chunk)
		bytes := ldb_set(len(chunk.Value), chunk.Value)
		wstore.qcount++
		chunkID := wolkcommon.Computehash(chunk.Value)
		chunk.ChunkID = chunkID
		chunk.Value = bytes
		//qc := wstore.qcount
		log.Trace("AddQueue", "chunkID", fmt.Sprintf("%x", chunkID), "qcount", wstore.qcount, "len", len(wstore.chunkQueue.chunkData))
		wstore.queuemu.Unlock()
	*/
	/*
		if qc > 100 {
			wstore.FlushQueue()
		}
	*/
}

var mem runtime.MemStats

func (wstore *Storage) QueueLoop() {
	counter := 0
	for {
		select {
		case chunk := <-wstore.queueCh:
			chunk.ChunkID = wolkcommon.Computehash(chunk.Value)
			chunk.Value = ldb_set(len(chunk.Value), chunk.Value)
			pos := int(chunk.ChunkID[0]) & wstore.devParam
			log.Trace("[storage:QueueLoop]", "pos", pos, "k", fmt.Sprintf("%x", chunk.ChunkID), "len", len(wstore.chunkQueue.chunkData[pos]))
			wstore.chunkQueue.chunkData[pos] = append(wstore.chunkQueue.chunkData[pos], chunk)
			wstore.qcount++
		case <-wstore.flushCh:
			counter++
			if counter > 100 {
				runtime.ReadMemStats(&mem)
				log.Trace("[storage:QueueLoop] mem", "alloc", mem.Alloc, "total", mem.TotalAlloc, "Sys", mem.Sys, "lookup", mem.Lookups, "Malloc", mem.Mallocs, "free", mem.Frees)
				log.Trace("[storage:QueueLoop] mem Heap", "alloc", mem.HeapAlloc, "Sys", mem.HeapSys, "idle", mem.HeapIdle, "inuse", mem.HeapInuse, "released", mem.HeapReleased, "object", mem.HeapObjects)
				log.Trace("[storage:QueueLoop] GC", "next", mem.NextGC, "last", mem.LastGC, "num", mem.NumGC)
				counter = 0
			}

			atomic.AddInt64(&wstore.qstatus, 1)
			if wstore.qcount > 0 {
				chunkData := wstore.chunkQueue.chunkData
				wstore.chunkQueue.chunkData = make([][]*cloud.RawChunk, len(wstore.Registry))
				wstore.qcount = 0
				atomic.StoreInt64(&wstore.qstatus, 0)
				go func(chunkData [][]*cloud.RawChunk) {
					start := time.Now()
					err := wstore.SetChunkBatch(context.Background(), chunkData)
					log.Trace("[storage:FlushQueue] SetChunkBatch", "time", time.Since(start))
					if err != nil {
						log.Error("[storage:FlushQueue] SetChunkBatch", "err", err)
					}
				}(chunkData)
			}
			atomic.StoreInt64(&wstore.qstatus, 0)
		}
	}
}

func (storage *Storage) isReady() (res bool) {
	res = true
	// TODO: use Peers???
	return res
}

func (wstore *Storage) FlushQueue() error {
	if atomic.LoadInt64(&wstore.qstatus) == 1 {
		return nil
	}
	wstore.flushCh <- 1
	/*
		wstore.queuemu.Lock()
		if wstore.qcount == 0 {
			wstore.queuemu.Unlock()
			return nil
		}

		log.Trace("[storage:FlushQueue] start", "qcount", wstore.qcount, "len", len(wstore.chunkQueue.chunkData))
		atomic.AddInt64(&wstore.qstatus, 1)
		chunkData := wstore.chunkQueue.chunkData
		wstore.chunkQueue.chunkData = nil
		wstore.qcount = 0
		wstore.queuemu.Unlock()

		log.Trace("FlushQueue calling SetChunkBatch", "len", len(chunkData), "len", len(wstore.chunkQueue.chunkData))
		go func() {
			start := time.Now()
			err := wstore.SetChunkBatch(context.Background(), chunkData)
			log.Trace("[storage:FlushQueue] SetChunkBatch", "time", time.Since(start))
			if err != nil {
				log.Error("[storage:FlushQueue] SetChunkBatch", "err", err)
			}
		}()
		atomic.StoreInt64(&wstore.qstatus, 0)
		log.Trace("FlushQueue done SetChunkBatch", "len", len(chunkData), "len", len(wstore.chunkQueue.chunkData))
		//return err
	*/
	return nil
}

func (storage *Storage) setStorageReady(ready bool) {
	log.Debug("[storage:setStorageReady] start", "ready", ready)
	storage.muStorageReady.Lock()
	storage.storageReady = ready
	storage.muStorageReady.Unlock()
	log.Trace("[storage:setStorageReady]", "ready", ready)
}

func (storage *Storage) isStorageReady() (ready bool) {
	storage.muStorageReady.RLock()
	ready = storage.storageReady
	storage.muStorageReady.RUnlock()
	return ready
}

func (storage *Storage) Start() {
	go storage.storageStatusLoop()
}

// ChainStatusLoop check blockchain availability status in infinite loop.
func (storage *Storage) storageStatusLoop() {
	// synchronise requires (1) peers to be available (2) blocks are readable from cloudstore
	//ticker := time.NewTicker(1000 * time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	for {
		select {
		// case <-ticker.C:
		// 	storage.storageStatus()
		default:
			storage.storageStatus()
		}
		time.Sleep(100 * time.Millisecond)
	}

}

// NumStorageConnections asks all nodes for wolk/info, and if they respond, it is considered active and a valid storage connection
func (storage *Storage) NumStorageConnections() (numConnections int, err error) {
	log.Info("[storage:NumStorageConnections] start ", "numConnections", numConnections)
	var numValidators int
	if len(storage.Registry) > MAX_VALIDATORS {
		numValidators = MAX_VALIDATORS
	} else {
		numValidators = len(storage.Registry)
	}
	//for i := 0; i < MAX_VALIDATORS; i++ {
	for i := 0; i < numValidators; i++ {
		url := fmt.Sprintf("%s://%s:%d/wolk/info", storage.Registry[i].GetScheme(), storage.Registry[i].GetStorageNode(), storage.Registry[i].GetPort())
		log.Trace("[storage:NumStorageConnections]", "url", url)
		req, reserr := http.NewRequest(http.MethodGet, url, nil)
		if reserr != nil {
			// handle err
			log.Error("NewRequest", "err", reserr)
		} else {
			resp, reserr := storage.httpclient[i].Do(req)
			if reserr != nil {
			} else {
				_, reserr := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				if reserr != nil {
				} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					log.Warn("[storage:NumStorageConnections] ", "StatusCode", resp.StatusCode, "url", url)
				} else {
					numConnections++
					log.Trace("[storage:NumStorageConnections] ", "StatusCode", resp.StatusCode, "numConnections", numConnections)
				}
			}
		}
	}
	log.Trace("[storage:NumStorageConnections] RESULT ", "numConnections", numConnections)
	return numConnections, err
}

func (storage *Storage) storageStatus() (err error) {
	if storage.isStorageReady() {
		return
	}

	numStorageConnections, err := storage.NumStorageConnections()
	if err != nil {
		return err
	}
	if numStorageConnections >= 2 {
		log.Trace("[storage:storageStatus] NumStorageConnections ", "numStorageConnections", numStorageConnections)
		storage.setStorageReady(true)
		go storage.chunkQueueStart()

	} else {
		log.Error("[storage:storageStatus] NumStorageConnections ", "numStorageConnections", numStorageConnections)
	}
	return nil
}

func (storage *Storage) chunkQueueStart() {

	ticker := time.NewTicker(25 * time.Millisecond)
	for {
		select {
		case _ = <-ticker.C:
			storage.FlushQueue()
			storage.GetChunkQueueBatch()
		}
	}
}

// Get returns data of the chunk with the given reference.
func (storage *Storage) Get(ctx context.Context, ref chunker.Reference) (chunkData chunker.ChunkData, err error) {
	key := []byte(ref)
	log.Chunk(key, "Get")

	chunkData, _, err = storage.GetChunk(ctx, key)
	if err != nil {
		return nil, err
	}
	return chunkData, nil
}

// RefSize returns the number of bytes used by chunkIDs
func (storage *Storage) RefSize() int64 {
	return 32
}

// Wait returns when
//    1) the Close() function has been called and
//    2) all the chunks which has been Put has been stored
func (storage *Storage) Wait(ctx context.Context) error {
	<-storage.closed
	storage.wg.Wait()
	return nil
}

func (self *Storage) QueueChunk(ctx context.Context, chunkData []byte) (chunkID common.Hash, err error) {
	//self.queuemu.Lock()
	//defer self.queuemu.Unlock()
	rawChunk := new(cloud.RawChunk)
	rawChunk.Value = chunkData
	rawChunk.ChunkID = wolkcommon.Computehash(chunkData)
	log.Chunk(rawChunk.ChunkID, "QueueChunk")
	//self.chunkQueue.chunkData = append(self.chunkQueue.chunkData, rawChunk)
	self.AddQueue(rawChunk)

	/*
		// put in chunkCache
		chunkID = common.BytesToHash(rawChunk.ChunkID)
		self.chunkCacheMu.Lock()
		self.chunkCache[chunkID] = chunkData
		self.chunkCacheTime[chunkID] = time.Now()
		self.chunkCacheMu.Unlock()
	*/

	return chunkID, nil
}

// Storage tallying
func (b *Storage) AddBandwidth(amount int64, addr common.Address) (err error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.balances[addr] += amount
	return err
}

func (b *Storage) GetBandwidthBalance(addr common.Address) (int64, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if p, ok := b.balances[addr]; ok {
		return p, nil
	}
	return 0, nil
}

func (b *Storage) WriteCheck(nodeID uint64, recipient common.Address, balance uint64) (checkID common.Hash, err error) {
	if balance < minCheckSize {
		log.Error("WriteCheck: Insufficient balance")
		return checkID, fmt.Errorf("Insufficient balance")
	}
	check := NewBandwidthCheck(nodeID, recipient, uint64(balance))
	//TODO: check.SignCheck(b.operatorKey)

	encoded, err := rlp.EncodeToBytes(&(check))
	if err != nil {
		log.Error("WriteCheck EncodeToBytes err", err)
		return checkID, err
	}
	checkIDh := wolkcommon.Computehash(encoded)
	_, err = b.QueueChunk(nil, encoded)
	if err != nil {
		log.Error("WriteCheck3: SetChunk ERR", "checkIDh", fmt.Sprintf("%x", checkIDh), "ERROR", err)
		return checkID, err
	}
	checkID = common.BytesToHash(checkIDh)

	log.Info("***** WriteCheck ****", "node", nodeID, "recipient", recipient)
	return checkID, err
}

func (b *Storage) RetrieveCheck(checkID common.Hash) (c *BandwidthCheck, err error) {
	encoded, ok, err := b.GetChunk(context.TODO(), checkID.Bytes())
	if err != nil {
		log.Error("RetrieveCheck: GetChunk ERR", "checkID", checkID, "ERROR", err)
		return c, err
	}
	log.Trace("RetrieveCheck: GetChunk SUCC", "chunkID", fmt.Sprintf("%x", checkID), "ok", ok, "len(encoded)", len(encoded), "encoded", fmt.Sprintf("%x", encoded))

	var check BandwidthCheck
	err = rlp.Decode(bytes.NewReader(encoded), &check)
	if err != nil {
		log.Error("RetrieveCheck", "err", err)
		return c, err
	}
	log.Trace("RetrieveCheck:  SUCC", "chunkID", fmt.Sprintf("%x", checkID), "check", check.String())
	return &check, nil
}

func (b *Storage) CashCheck(checkID common.Hash) (err error) {
	check, err := b.RetrieveCheck(checkID)
	if err != nil {
		return err
	}
	addr, err := check.GetSigner()
	if err != nil {
		return err
	}
	b.AddBandwidth(-1*int64(check.Amount), addr)
	return nil
}

func (wstore *Storage) GetChunkQueueBatch() {
	wstore.queuemu.Lock()
	chunkData := wstore.getChunkQueue
	wstore.getChunkQueue = make(map[string][]chan *cloud.RawChunk)
	wstore.queuemu.Unlock()
	if len(chunkData) == 0 {
		return
	}

	var chunks []*cloud.RawChunk

	for strkey, _ := range chunkData {
		chunkID := common.Hex2Bytes(strkey)
		chunks = append(chunks, &cloud.RawChunk{ChunkID: chunkID})
	}

	chunkRes, err := wstore.GetChunkBatch(chunks)

	if err != nil {
		log.Error("[Storage:GetChunkQueueBatch] GetChunkBatch error %v", err)
	}

	for _, res := range chunkRes {
		for _, ch := range chunkData[fmt.Sprintf("%x", res.ChunkID)] {
			ch <- res
		}
	}
}

func (wstore *Storage) GetChunkQueue(ctx context.Context, chunkID []byte, resch chan *cloud.RawChunk) {
	wstore.queuemu.Lock()
	strkey := fmt.Sprintf("%x", chunkID)
	wstore.getChunkQueue[strkey] = append(wstore.getChunkQueue[strkey], resch)
	wstore.queuemu.Unlock()
}

func (wstore *Storage) GetChunksQueue(ctx context.Context, chunkIDs [][]byte) *Stream {
	resch := make(chan *cloud.RawChunk)
	wstore.queuemu.Lock()
	for _, chunk := range chunkIDs {
		strkey := fmt.Sprintf("%x", chunk)
		wstore.getChunkQueue[strkey] = append(wstore.getChunkQueue[strkey], resch)
	}
	wstore.queuemu.Unlock()
	stream := new(Stream)
	stream.ch = make(chan struct{}, len(chunkIDs))
	stream.total = len(chunkIDs)
	go func(ctx context.Context, stream *Stream, resch chan *cloud.RawChunk) {
		_, cancel := context.WithCancel(ctx)
		defer cancel()
		for i := 0; i < stream.total; i++ {
			select {
			case <-ctx.Done():
				stream.done = true
				stream.err = errors.New("cancel called")
			default:
				res := <-resch
				stream.Put(res)
			}
		}
		stream.done = true
	}(ctx, stream, resch)
	return stream
}

type Stream struct {
	buf   []*cloud.RawChunk
	total int
	puts  int
	gets  int
	done  bool
	mu    sync.Mutex
	err   error
	ch    chan struct{}
}

func (stream *Stream) Put(chunkData *cloud.RawChunk) {
	stream.enqueue(chunkData)
	stream.ch <- struct{}{}
}

func (stream *Stream) Get() (*cloud.RawChunk, error) {
	log.Trace("[storage:Stream Get]", "stream.gets", stream.gets, "stream.total", stream.total, "stream.done", stream.done)
	if stream.total == stream.gets {
		return nil, io.EOF
	}
	<-stream.ch
	res := stream.dequeue()
	if res == nil && stream.done {
		return nil, io.EOF
	}
	return res, nil
}

func (stream *Stream) enqueue(chunkData *cloud.RawChunk) {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	stream.buf = append(stream.buf, chunkData)
	stream.puts++
}

func (stream *Stream) dequeue() *cloud.RawChunk {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if len(stream.buf) == 0 {
		return nil
	}

	chunkData := stream.buf[0]
	if len(stream.buf) > 1 {
		stream.buf = stream.buf[1:]
	} else {
		stream.buf = nil
	}
	stream.gets++
	return chunkData
}

func (wstore *Storage) GetChunkAsync(ctx context.Context, k []byte, f func([]byte, bool, error) bool) error {
	res := make(chan *cloud.RawChunk)
	wstore.GetChunkQueue(ctx, k, res)
	go func(chan *cloud.RawChunk) {
		chunk := <-res

		if !f(chunk.Value, chunk.OK, chunk.Error) {

		}
	}(res)
	return nil
}

func (wstore *Storage) GetChunksAsync(ctx context.Context, chunkIDs [][]byte, f func(*cloud.RawChunk) bool) error {

	stream := wstore.GetChunksQueue(ctx, chunkIDs)

	for {
		res, err := stream.Get()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("[Storage:GetChunksAsync] Stream.Get error %v", err)
			return err
		}
		if !f(res) {
			log.Error("[storage:GetChunksAsync] func return false")
			break
		}
	}
	return nil
}

// PutGetter is only for PutFile/GetFile related functions
type PutGetter struct {
	wstore      *Storage
	counter     int64
	total       int64
	wg          *sync.WaitGroup
	closed      chan struct{}
	done        chan struct{}
	checkClosed bool
	errC        chan error
	mu          sync.Mutex
}

func NewPutGetter(wstore *Storage) *PutGetter {
	return &PutGetter{
		wstore:      wstore,
		wg:          &sync.WaitGroup{},
		closed:      make(chan struct{}),
		done:        make(chan struct{}),
		checkClosed: false,
		errC:        make(chan error),
	}
}

func (putgetter *PutGetter) Put(ctx context.Context, chunkData chunker.ChunkData) (r chunker.Reference, err error) {
	//return wolkcommon.Computehash(chunkData), nil
	rawchunk := &cloud.RawChunk{Value: chunkData}
	putgetter.wg.Add(1)
	start := time.Now()
	r, err = putgetter.wstore.PutChunk(ctx, rawchunk, putgetter.wg)
	log.Trace("[storage:putgetter:PutChunk]", "k", fmt.Sprintf("%x", r), "wg", putgetter.wg, "wgp", fmt.Sprintf("%p", putgetter.wg), "time", time.Since(start), "len", len(chunkData))
	return r, err
}

func (putgetter *PutGetter) Get(ctx context.Context, ref chunker.Reference) (chunkData chunker.ChunkData, err error) {
	return putgetter.wstore.Get(ctx, ref)
}

func (putgetter *PutGetter) RefSize() int64 {
	return 32
}

func (putgetter *PutGetter) Close() {
	close(putgetter.closed)
}

func (putgetter *PutGetter) Wait(ctx context.Context) error {
	<-putgetter.closed
	putgetter.wg.Wait()
	return nil
}

func (wstore *Storage) PutFile(ctx context.Context, chunkData []byte, wg *sync.WaitGroup) ([]byte, error) {
	log.Trace("[Storage:PutFile]", "len", len(chunkData))
	start := time.Now()
	defer log.Trace("[Storage:PutFile] done", "len", len(chunkData), "time", time.Since(start))
	putter := NewPutGetter(wstore)
	//go putter.Wait()

	reader := bytes.NewReader(chunkData)
	hash, wait, err := chunker.TreeSplit(ctx, reader, int64(len(chunkData)), putter)
	log.Trace("[Storage:PutFile] done chunker", "hash", hash, "len", len(chunkData), "wgp", fmt.Sprintf("%p", putter.wg), "time", time.Since(start))
	if err != nil {
		log.Error("[Storage:PutFile] TreeSplit Error ", "err", err)
	}

	if wg != nil {
		go func() {
			st := time.Now()
			log.Trace("[Storage:PutFile] in go func", "hash", hash, "len", len(chunkData))
			err := wait(ctx)
			if err != nil {
				// TODO
			}
			log.Trace("[Storage:PutFile] out go func", "hash", hash, "len", len(chunkData), "time", time.Since(start), "time", time.Since(st), "wgp", fmt.Sprintf("%p", putter.wg))
			wg.Done()
		}()
	}
	return hash, err
}

func (wstore *Storage) PutFileAsync(ctx context.Context, chunkData []byte, f func([]byte, error) bool) ([]byte, error) {
	putter := NewPutGetter(wstore)
	//go putter.Wait()
	reader := bytes.NewReader(chunkData)
	hash, wait, err := chunker.TreeSplit(ctx, reader, int64(len(chunkData)), putter)
	if err != nil {
		log.Error("[Storage:PutFileAsync] TreeSplit Error ", "err", err)
	}

	go func(hash []byte) {
		wait(ctx)
		if !f(hash, err) {
		}
	}(hash)
	return hash, err
}

func (wstore *Storage) GetFile(ctx context.Context, key []byte) ([]byte, error) {
	getter := NewPutGetter(wstore)
	//go getter.Wait()
	reader := chunker.TreeJoin(ctx, key, getter, 0)
	ch := make(chan bool)
	size, err := reader.Size(ctx, ch)
	if err != nil && err != io.EOF {
		log.Error("[Storage:GetFile] TreeJoin Error ", "err", err)
	}
	output := make([]byte, size)
	_, err = reader.Read(output)
	getter.Close()
	return output, err
}

func (wstore *Storage) GetFileAsync(ctx context.Context, key []byte, f func(context.Context, []byte, bool, error) bool) error {
	getter := NewPutGetter(wstore)
	//go getter.Wait()
	go func() {
		reader := chunker.TreeJoin(ctx, key, getter, 0)
		ch := make(chan bool)
		size, err := reader.Size(ctx, ch)
		if err != nil {
			log.Error("[Storage:GetFileAsync] TreeJoin Reader Error ", "err", err)
		}
		output := make([]byte, size)
		_, err = reader.Read(output)
		exist := false
		if size > 0 {
			exist = true
		}
		if err == io.EOF {
			err = nil
		}
		if !f(ctx, output, exist, err) {
		}
		getter.Close()
	}()
	return nil
}

func (wstore *Storage) GetFileWithRange(ctx context.Context, key []byte, start int64, end int64) ([]byte, error) {
	if start > end && end != -1 {
		return nil, fmt.Errorf("[Storage:GetFileWithRange] start is greater than end start: %d end: %d", start, end)
	}
	getter := NewPutGetter(wstore)
	//go getter.Wait()
	reader := chunker.TreeJoin(ctx, key, getter, 0)
	ch := make(chan bool)
	size, err := reader.Size(ctx, ch)
	if end == -1 || end > size {
		end = size
	}
	if err != nil {
		return nil, fmt.Errorf("[Storage:GetFileWithRange] Size error %v", err)
	}
	if size < start {
		return nil, fmt.Errorf("[Storage:GetFileWithRange] start is greater than size start: %d size: %d", start, size)
	}
	output := make([]byte, end-start)
	log.Info(fmt.Sprintf("[Storage:GetFileWithRange] getting key [%s] with range %d to %d", key, start, end))
	_, err = reader.ReadAt(output, start)
	if err == io.EOF {
		log.Error(fmt.Sprintf("[Storage:GetFileWithRange] ReadAt EOF error %v", err))
		err = nil
	}
	if err != nil {
		log.Error(fmt.Sprintf("[Storage:GetFileWithRange] ReadAt error %v", err))
	}
	getter.Close()
	return output, err
}

func buildDRG(degree int, numNodes int, l int) (edgeList [][]int) {
	edgeList = make([][]int, numNodes)
	for i := 1; i < numNodes; i++ {
		edgeList[i] = make([]int, degree)
		for j := 0; j < degree; j++ {
			edgeList[i][j] = rand.Intn(i) // (j + l) % i
		}
	}
	return edgeList
}

func xorhash(a, b []byte) (r []byte) {
	r = make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		r[i] = a[i] ^ b[i]
	}
	return r
}

func numDataBlocks(d []byte) (lend int) {
	n := len(d)
	lend = n / 32
	if n%32 > 0 {
		lend++
	}
	return lend
}

func dataBlocks(input []byte) (d [][]byte) {
	numBlocks := numDataBlocks(input)
	d = make([][]byte, numBlocks)
	for i := 0; i < numBlocks; i++ {
		b0 := i * 32
		b1 := (i + 1) * 32
		if b1 > len(input) {
			b1 = len(input)
		}
		d[i] = input[b0:b1]
	}
	return d
}

func padInput(d []byte) (o []byte) {
	padding := 32 - (len(d) % 32)
	if padding > 0 && padding < 32 {
		return append(d, make([]byte, padding)...)
	}
	return d
}

func computeLabels(edgeList [][]int, inp []byte, rootHash []byte) (dlabels []byte) {
	input := padInput(inp)
	lend := numDataBlocks(input)
	labels := make([][]byte, lend+degree)

	// compute labels using DRG
	for start := 0; start < lend+degree; start++ {
		if start == 0 {
			labels[start] = rootHash
		} else {
			res := make([]byte, 0)
			for _, end := range edgeList[start] {
				res = append(res, labels[end]...)
			}
			l := wolkcommon.Computehash(res)
			if start >= degree {
				b := (start - degree) * 32
				b2 := b + 32
				if b2 > len(input) {
					b2 = len(input)
				}
				d := input[b:b2]
				if len(d) < 32 {
					extra := make([]byte, 32-len(d))
					d = append(d, extra...)
				}
				l = xorhash(l, d)
			}
			labels[start] = l
		}
	}
	dlabels = make([]byte, 0)
	for _, l := range labels {
		dlabels = append(dlabels, l...)
	}
	return dlabels
}

func (s *Storage) encodeReplica(edgeList [][]int, inp []byte) (labels []byte, storageRoot []byte, err error) {
	// pad the input
	input := padInput(inp)

	// Merkelize the data
	mr := merkletree.Merkelize(dataBlocks(input))

	// compute labels using DRG edgeList
	labels = computeLabels(edgeList, input, mr[1])

	// merkleize labels
	labels_mr := merkletree.Merkelize(dataBlocks(labels))

	storageRoot = labels_mr[1]

	return labels, storageRoot, nil
}

func (s *Storage) GetReplica(k []byte, startoffset int, endoffset int) (chunk []byte, storageRoot []byte, err error) {
	bytesOut, ok, err := s.cl.GetChunkWithRange(k, startoffset, endoffset)
	if err != nil || !ok {
		return chunk, storageRoot, err
	}
	if len(bytesOut) < 40 {
		return chunk, storageRoot, fmt.Errorf("Incorrect file read")
	}
	content := bytesOut
	chunkSize := wolkcommon.BytesToUint64(content[0:8])
	storageRoot = content[8:40]
	labels := content[40:]
	chunk = s.decodeReplica(labels, chunkSize)

	return chunk, storageRoot, nil
}

type ReplicaProof struct {
	StorageRoot string
	Challenge   uint
	Branch      []string
}

func (s *Storage) GetReplicaProof(k []byte, challenge uint) (p *ReplicaProof, err error) {
	bytesOut, ok, err := s.cl.GetChunk(k)
	if len(bytesOut) < 40 || err != nil || !ok {
		return p, fmt.Errorf("Incorrect file read")
	}
	content := bytesOut
	storageRoot := content[8:40]
	labels := content[40:]

	// for a requested challenge node, compute merkle branch from the labels
	// Yes, its inefficient to get the whole file when the branch refers to a specific node, but the intermediate nodes demand the whole file
	o := dataBlocks(labels)
	labelsMerkelized := merkletree.Merkelize(o)

	branch, err := merkletree.Mk_branch(labelsMerkelized, challenge)
	if err != nil {
		return p, fmt.Errorf("Mk_branch ERR %v", err)
	}

	p = new(ReplicaProof)
	p.StorageRoot = fmt.Sprintf("%x", storageRoot)
	for _, e := range branch {
		p.Branch = append(p.Branch, fmt.Sprintf("%x", e))
	}
	p.Challenge = challenge
	return p, nil
}

func (s *Storage) VerifyReplicaProof(storageRoot []byte, proof *ReplicaProof) (err error) {
	storageRootProvided, err := hex.DecodeString(proof.StorageRoot)
	if err != nil {
		return err
	}

	if bytes.Compare(storageRootProvided, storageRoot) != 0 {
		return fmt.Errorf("Storage Root Mismatch")
	}
	branch := make([][]byte, len(proof.Branch))
	for i, e := range proof.Branch {
		branch[i], err = hex.DecodeString(e)
		if err != nil {
			return err
		}
	}
	// verify the branch
	_, err = merkletree.Verify_branch(storageRoot, proof.Challenge, branch)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) SetReplica(ctx context.Context, chunk []byte) (k []byte, storageRoot []byte, err error) {
	if len(chunk) > maxFileSize {
		return k, storageRoot, fmt.Errorf("chunksize %d exceed maxFileSize %d", len(chunk), maxFileSize)
	}

	k = wolkcommon.Computehash(chunk)
	var labels []byte
	labels, storageRoot, err = s.encodeReplica(s.edgeList, chunk)

	// save labels and len(chunk) to disk keyed in by hash of chunk.
	content := make([]byte, 40+len(labels))
	copy(content[0:8], wolkcommon.UIntToByte(uint64(len(chunk))))
	copy(content[8:40], storageRoot)
	copy(content[40:], labels[:])

	err = s.cl.SetChunk(ctx, k, content)
	return k, storageRoot, err
}

func (s *Storage) decodeReplica(labels []byte, chunkSize uint64) (d []byte) {
	//fmt.Printf("recoverData len(labels)=%d len(edgeList)=%d\n", len(labels), len(edgeList))
	numBlocks := numDataBlocks(labels)
	d = make([]byte, numBlocks*32) // (labels)-32*degree)
	for start := degree; start < numBlocks; start++ {
		res := make([]byte, 0)
		for _, end := range s.edgeList[start] {
			e0 := (end) * 32
			e1 := (end + 1) * 32
			res = append(res, labels[e0:e1]...)
		}
		b0 := (start - degree) * 32
		b1 := (start - degree + 1) * 32
		c0 := start * 32
		c1 := (start + 1) * 32
		if c1 > len(labels) {
			c1 = len(labels)
		}
		l := labels[c0:c1]
		if len(l) < 32 {
			extra := make([]byte, 32-len(l))
			l = append(l, extra...)
		}
		n := xorhash(wolkcommon.Computehash(res), l)
		copy(d[b0:b1], n[:])
		//fmt.Printf("d[%d]=%x == xorhash(%x, labels[%d]=%x)\n", start-degree, d[start-degree], Computehash(res), start, labels[start])
	}
	if uint64(len(d)) > chunkSize {
		d = d[0:chunkSize]
	}
	return d
}
