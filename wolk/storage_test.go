package wolk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	chunker "github.com/ethereum/go-ethereum/swarm/storage"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

func TestChunkAsyncMultinode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	wolk, _, _, _, _ := newWolkPeer(t, defaultConsensus)
	//defer ReleaseNodes(t, nodelist)

	ws := wolk[0].Storage
	//ws.Start()

	for wolk[0].isChainReady() == false {
		log.Info("[storage_test:TestChunkAsyncMultinode] Waiting for storage to be ready")
		time.Sleep(1 * time.Second)
	}
	ctx := context.Background()
	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289}

	values := make([][]byte, len(sizes))
	keys := make([]chunker.Reference, len(sizes))

	// set up Put operations for different sizes
	for i, sz := range sizes {
		_, val := generateRandomData(sz)
		values[i] = val
		key, err := ws.Put(ctx, values[i], nil)
		if err != nil {
			t.Fatalf("[storage_test:TestChunkAsyncMultinode] Put ERR %v\n", err)
		}
		keys[i] = key
	}

	for k := 0; k < len(sizes); k++ {
		st := time.Now()
		ws.GetChunkAsync(ctx, keys[k], func(v []byte, ok bool, err error) bool {
			if k < len(sizes) {
				if !bytes.Equal(v, values[k]) {
					t.Fatalf("input and output mismatch\n IN: %v\nOUT: %v\n", values[k], v)
				}
			}
			log.Info("[storage_test:TestChunkAsyncMultinode] GetChunkAsync", "tm", time.Since(st))
			return true
		})
	}

	var chunkIDs [][]byte
	inputs := make(map[string][]byte)
	for i, chunkid := range keys {
		chunkIDs = append(chunkIDs, []byte(keys[i]))
		inputs[fmt.Sprintf("%x", chunkid)] = values[i]
	}

	st := time.Now()
	ws.GetChunksAsync(ctx, chunkIDs, func(chunk *cloud.RawChunk) bool {
		if !bytes.Equal(chunk.Value, inputs[fmt.Sprintf("%x", chunk.ChunkID)]) {
			t.Fatalf("input and output mismatch\n IN: %v\nOUT: %v\n", inputs[fmt.Sprintf("%x", chunk.ChunkID)], chunk.Value)
		}
		log.Info("[storage_test:TestChunkAsyncMultinode] GetChunksAsync", "tm", time.Since(st))
		return true
	})
}

func TestChunkAsyncLocal(t *testing.T) {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte("test")))

	privateKeys, err := crypto.HexToPrivateKey(k_str)
	if err != nil {
		t.Fatalf("HexToEdwardsPrivateKey %v", err)
	}

	privateKeysECDSA, err := ethcrypto.HexToECDSA(k_str)
	if err != nil {
		t.Fatalf("HexToEdwardsPrivateKey %v", err)
	}

	datadir := "/tmp/storagetest" + fmt.Sprintf("%d", int32(time.Now().Unix()))
	fmt.Println("datadir", datadir)
	cfg := cloud.DefaultConfig
	cfg.DataDir = datadir
	cfg.OperatorKey = privateKeys
	cfg.OperatorECDSAKey = privateKeysECDSA
	cfg.GenesisFile = "/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis2.json"
	cfg.HTTPPort = 82

	fmt.Println("cfg", cfg)
	wc, err := NewWolk(nil, &cfg)
	wcs := wc.Storage
	wcs.Start()
	time.Sleep(time.Second * 3)

	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289}

	// Put test: the Puts are instanaeous, the putwg.Add(1) calls are matched with a single putwg.Wait()
	var keys [][]byte
	var vals [][]byte
	var putwg sync.WaitGroup
	putstart := time.Now()
	for _, size := range sizes {
		_, val := generateRandomData(size)
		vals = append(vals, val)
		ctx := context.Background()
		putwg.Add(1)
		// this is not blocking! BUT Issue: key is NOT the fileHash! key is hash(val)  Don't we want the fileHash
		key, err := wcs.Put(ctx, val, &putwg)
		if err != nil {
			log.Error("[storage_test:TestGetChunkAsyncLocal] Put", "err", err)
		}
		keys = append(keys, key)
	}
	log.Info("[storage_test:TestGetChunkAsyncLocal] Putwg.Add(1) submitted", "len(keys)", len(keys), "tm", time.Since(putstart))

	// waiting Put response
	putwg.Wait()
	log.Info("[storage_test:TestGetChunkAsyncLocal] Putwg.Wait() completed", "tm", time.Since(putstart))

	// TODO: Need to do a Get and check result against vals

	// PutChunk test: here there is one waitgroup per chunk
	var rawchunks []*cloud.RawChunk
	var rawkeys [][]byte
	putstart = time.Now()
	for _, size := range sizes {
		var wg sync.WaitGroup
		wg.Add(1)
		ctx := context.Background()
		_, val := generateRandomData(size)

		rawchunk := &cloud.RawChunk{Value: val}
		chunkKey := wolkcommon.Computehash(val)
		_, err := wcs.PutChunk(ctx, rawchunk, &wg)
		if err != nil {
			t.Fatalf("PutChunk ERR %v\n", err)
		}
		rawchunks = append(rawchunks, rawchunk)
		rawkeys = append(rawkeys, chunkKey)
	}
	log.Info("[storage_test:TestGetChunkAsyncLocal] Putwg.Add(1) submitted", "len(rawkeys)", len(rawkeys), "tm", time.Since(putstart))

	// Here, we are waiting for ALL the waitgroups
	for _, rawchunk := range rawchunks {
		rawchunk.Wg.Wait()
		if rawchunk.Error != nil {
			t.Fatalf("PutChunk : chunk has error %v", rawchunk.Error)
		}
	}
	log.Info("[storage_test:TestGetChunkAsyncLocal] rawchunk.Wg.Wait() done", "len(rawkeys)", len(rawkeys), "tm", time.Since(putstart))

	//PutChunkWithChannel test
	ch := make(chan *cloud.RawChunk, len(sizes))
	putstart = time.Now()
	for _, size := range sizes {
		ctx := context.Background()
		_, chunk := generateRandomData(size)
		//chunkKey := wolkcommon.Computehash(chunk)
		wcs.PutChunkWithChannel(ctx, chunk, ch)
	}
	log.Info("[storage_test:TestGetChunkAsyncLocal] PutChunkWithChannel", "len(sizes)", len(sizes), "tm", time.Since(putstart))

	// waiting PutChunkWithChannel response
	var channelChunk []*cloud.RawChunk
	var testChunk [][]byte
	var chunkMap map[string]*cloud.RawChunk
	chunkMap = make(map[string]*cloud.RawChunk)
	for i := 0; i < len(sizes); i++ {
		res := <-ch
		channelChunk = append(channelChunk, res)
		testChunk = append(testChunk, res.ChunkID)
		chunkMap[fmt.Sprintf("%x", res.ChunkID)] = res
	}
	log.Info("[storage_test:TestGetChunkAsyncLocal] PutChunkAsync completed", "len(sizes)", len(sizes), "tm", time.Since(putstart))

	//PutChunkAsync test
	putstart = time.Now()
	for _, size := range sizes {
		ctx := context.Background()
		_, chunk := generateRandomData(size)
		st := time.Now()
		// this is not a blocking call =)
		wcs.PutChunkAsync(ctx, chunk, func(key []byte, err error) bool {
			if err != nil {
				return false
			}
			log.Info("[storage_test:TestGetChunkAsyncLocal] PutChunkAsync complete", "tm", time.Since(st))
			return true
		})
	}
	log.Info("[storage_test:TestGetChunkAsyncLocal] PutChunkAsync ALL submitted", "len(sizes)", len(sizes), "tm", time.Since(putstart))

	// GetChunkAsync test
	putstart = time.Now()
	for i, key := range keys {
		ctx := context.Background()
		pos := i
		k := key
		st := time.Now()
		// this is not a blocking call =)
		wcs.GetChunkAsync(ctx, k, func(v []byte, ok bool, err error) bool {
			if !bytes.Equal(v, vals[pos]) {
				t.Fatalf("input and output mismatch: %x\n", k)
				return false
			}
			log.Info("[storage_test:TestGetChunkAsyncLocal] GetChunkAsync complete", "sz", len(v), "tm", time.Since(st))
			return true
		})
	}
	log.Info("[storage_test:TestGetChunkAsyncLocal] GetChunkAsync ALL submitted", "len(sizes)", len(sizes), "tm", time.Since(putstart))

	// GetChunksAsync test
	ctx := context.Background()
	putstart = time.Now()
	st := time.Now()
	wcs.GetChunksAsync(ctx, testChunk, func(chunkData *cloud.RawChunk) bool {
		if !bytes.Equal(chunkMap[fmt.Sprintf("%x", chunkData.ChunkID)].Value, chunkData.Value) {
			t.Fatalf("input and output mismatch: %x\n", chunkData.ChunkID)
			return false
		}
		log.Info("[storage_test:TestGetChunkAsyncLocal] GetChunksAsync complete", "sz", len(chunkData.Value), "tm", time.Since(st))
		return true
	})
	log.Info("[storage_test:TestGetChunkAsyncLocal] GetChunkAsync ALL done", "len(sizes)", len(sizes), "tm", time.Since(putstart))
}

func TestChunkFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte("test")))

	privateKeys, err := crypto.HexToPrivateKey(k_str)
	if err != nil {
		t.Fatalf("HexToEdwardsPrivateKey %v", err)
	}

	privateKeysECDSA, err := ethcrypto.HexToECDSA(k_str)
	if err != nil {
		t.Fatalf("HexToEdwardsPrivateKey %v", err)
	}

	datadir := "/tmp/storagetest" + fmt.Sprintf("%d", int32(time.Now().Unix()))
	cfg := cloud.DefaultConfig
	cfg.DataDir = datadir
	cfg.OperatorKey = privateKeys
	cfg.OperatorECDSAKey = privateKeysECDSA
	cfg.GenesisFile = "/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis2.json"
	cfg.HTTPPort = 82

	wc, err := NewWolk(nil, &cfg)
	wcs := wc.Storage
	wcs.Start()
	time.Sleep(time.Second * 3)

	//sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289, 524288, 524288 + 1, 524288 + 4097, 1000000, 7 * 524288, 7*524288 + 1, 7*524288 + 4097}
	//sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289, 524288, 524288 + 1, 524288 + 4097, 1000000}
	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289}

	var wg sync.WaitGroup
	var keys [][]byte
	var vals [][]byte
	for _, sz := range sizes {
		ctx := context.Background()
		_, val := generateRandomData(sz)
		vals = append(vals, val)
		wg.Add(1)
		key, err := wcs.PutFile(ctx, val, &wg)
		log.Info("PutFile return", fmt.Sprintf("%x", key), "size", sz, err)
		if err != nil {
		}
		keys = append(keys, key)
	}

	log.Info("wg.Wait start")
	wg.Wait()
	log.Info("wg.Wait Done")

	for i, key := range keys {
		ctx := context.Background()
		res, err := wcs.GetFile(ctx, key)
		log.Info("GetFile ", "key", fmt.Sprintf("%x", key), "len", len(res), "err", err)

		if err != nil && err != io.EOF {
			t.Fatalf("err %v", err)
		}

		if !bytes.Equal(vals[i], res) {
			t.Fatalf("GetFile error")
		}
	}

	ch := make(chan []byte, len(sizes))
	for i, key := range keys {
		ctx := context.Background()

		pos := i
		k := key
		log.Info("GetFileAsync ", "key", fmt.Sprintf("%x", key))
		var wg sync.WaitGroup
		wg.Add(1)
		wcs.GetFileAsync(ctx, k, func(ctx context.Context, res []byte, exist bool, err error) bool {
			defer wg.Done()
			if len(res) == 1 {
			}

			ch <- res

			if !bytes.Equal(vals[pos], res) {
				t.Fatalf("GetFileAsync1 error")
				return false
			}
			return true
		})
	}
	for i := 0; i < len(keys); i++ {
		<-ch
	}

	for i, key := range keys {
		ctx := context.Background()

		k := key
		var wg sync.WaitGroup
		var buf []byte
		pos := i

		wg.Add(1)
		wcs.GetFileAsync(ctx, k, func(ctx context.Context, res []byte, exist bool, err error) bool {
			defer wg.Done()
			if (err != nil && err != io.EOF) || !exist {
				return false
			}
			buf = res
			return true
		})
		wg.Wait()
		if !bytes.Equal(vals[pos], buf) {
			t.Fatalf("GetFileAsync2 error %d %d", len(buf), pos)
		}
	}
}

func TestSetChunkGetChunk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	storage, err := NewMockStorage()
	if err != nil {
		tprint("newleveldbstorage err (%v)", err)
		t.Fatal(err)
	}

	_, expected := generateRandomData(32768)
	chunkHash, err := storage.SetChunk(nil, expected)
	if err != nil {
		tprint("setchunk err (%v)", err)
		t.Fatal(err)
	}
	log.Info("Wrote chunk")
	ctx := context.TODO()
	actual, ok, err := storage.GetChunk(ctx, chunkHash.Bytes())
	if err != nil {
		tprint("getchunk err (%v)", err)
		t.Fatal(err)
	}
	if !ok {
		tprint("getchunk did not get (%v) chunk!", chunkHash)
		t.Fatalf("NO chunk gotten!")
	}
	//	tprint("chunk gotten (%s) (%x)", actual, actual)
	if string(actual) != string(expected) {
		//		tprint("actual (%s) != expected (%s), fail.", actual, expected)
		t.Fatalf("they didn't match!")
	}

}

func TestHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// need to fix newWolkPeer

	server := "127.0.0.1:9800"
	fmt.Printf("server %s\n", server)
	newWolkPeer(t, consensusSingleNode)

	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289, 524288, 524288 + 1, 524288 + 4097, 1000000, 7 * 524288, 7*524288 + 1, 7*524288 + 4097}
	sizes = sizes[0:15]
	var DefaultTransport http.RoundTripper = &http.Transport{
		Dial: (&net.Dialer{
			// limits the time spent establishing a TCP connection (if a new one is needed)

			Timeout:   5 * time.Second,
			KeepAlive: 3 * time.Second, // 60 * time.Second,
		}).Dial,
		//MaxIdleConns: 5,
		MaxIdleConnsPerHost: 100,

		// limits the time spent reading the headers of the response.
		ResponseHeaderTimeout: 5 * time.Second,
		IdleConnTimeout:       4 * time.Second, // 90 * time.Second,

		// limits the time the client will wait between sending the request headers when including an Expect: 100-continue and receiving the go-ahead to send the body.
		ExpectContinueTimeout: 1 * time.Second,

		// limits the time spent performing the TLS handshake.
		TLSHandshakeTimeout: 5 * time.Second,
	}

	for _, sz := range sizes {
		_, chunkdata := generateRandomData(sz)
		k := wolkcommon.Computehash(chunkdata)

		httpclient := &http.Client{Timeout: time.Second * 10, Transport: DefaultTransport}
		for phase := 0; phase < 2; phase++ {
			run := http.MethodPost
			if phase > 0 {
				run = http.MethodGet
			}
			baseurl := fmt.Sprintf("http://%s/wolk/chunk", server)
			st := time.Now()
			if run == http.MethodGet {
				url := baseurl + fmt.Sprintf("/%x", k)
				fmt.Printf("BASEURL %s\n", url)

				req, _ := http.NewRequest(http.MethodGet, url, nil)
				req.Header.Add("content-type", "application/json;")
				req.Header.Add("cache-control", "no-cache")
				req.Header.Add("Expect", "")
				res, err := httpclient.Do(req)
				if err != nil {
					fmt.Println(" error", err)
				} else {
					v, _ := ioutil.ReadAll(res.Body)
					res.Body.Close()
					fmt.Printf("%s\n", v)
					fmt.Printf("sz=%d len(%s) = %d = %x => %x\n", sz, run, len(v), k, wolkcommon.Computehash(v))
				}
			} else {
				body := bytes.NewReader(chunkdata)
				url := baseurl
				req, _ := http.NewRequest(http.MethodPost, url, body)
				fmt.Printf("BASEURL %s\n", url)
				req.Header.Add("content-type", "application/json;")
				req.Header.Add("cache-control", "no-cache")
				res, err := httpclient.Do(req)
				if err != nil {
					fmt.Printf(" ! post to %s (%d) error %v\n", url, len(chunkdata), err)
				} else {
					defer res.Body.Close()
					if res.StatusCode == 200 {
						body2, reserr := ioutil.ReadAll(res.Body)
						if reserr != nil {
							fmt.Printf("DONE %d %s\n", res.StatusCode, body2)
						} else {
							v := body2
							fmt.Printf("%s\n", v)
						}
					}
				}
			}
			log.Info("result", "run", run, "st", time.Since(st))
		}
	}
}
