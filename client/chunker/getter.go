package chunker

import (
	//        wolkcommon "github.com/wolkdb/cloudstore/common"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	//        "net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

type ClientGetter struct {
	url        string
	queue      getQueue
	mu         sync.Mutex
	at         int64
	encryption string
}

type getQueue struct {
	ret    map[string](chan retData)
	chunks []*cloud.RawChunk
}

type retData struct {
	chunk ChunkData
	err   error
}

var counter int

func (g *ClientGetter) Get(ctx context.Context, key Reference) (ChunkData, error) {

	resch := make(chan retData)
	go g.Queue(key, resch)
	counter++
	ret := <-resch
	counter--

	return ret.chunk, ret.err
}

func (g *ClientGetter) Force() {
	atomic.StoreInt64(&g.at, 1)
	g.mu.Lock()
	chunks, err := json.Marshal(&g.queue.chunks)
	wglist := g.queue.ret
	g.queue.ret = make(map[string](chan retData))
	g.queue.chunks = nil
	atomic.StoreInt64(&g.at, 0)
	g.mu.Unlock()
	body := bytes.NewReader(chunks)
	url := fmt.Sprintf("%s/getbatch", g.url)
	req, err := http.NewRequest(http.MethodGet, url, body)
	httpclient := &http.Client{Timeout: time.Second * 1000, Transport: DefaultTransport}

	res, err := httpclient.Do(req)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
		os.Exit(0)
	} else {
		// fmt.Printf("%s %s (%x) %v\n", run, url, k, res.StatusCode)
		defer res.Body.Close()
		if res.StatusCode == 200 {
			body, err := ioutil.ReadAll(res.Body)
			var c []*cloud.RawChunk
			err = json.Unmarshal(body, &c)
			if err != nil {
				fmt.Printf("DONE %d %s\n", res.StatusCode, body)
			} else {
				// fmt.Printf("%s", body)
				//return body, nil
			}
			for _, cdata := range c {
				if _, ok := wglist[common.Bytes2Hex(cdata.ChunkID)]; ok {
				} else {
				}
				wglist[common.Bytes2Hex(cdata.ChunkID)] <- retData{chunk: cdata.Value, err: cdata.Error}
			}
		}
	}
}

func (g *ClientGetter) Queue(k Reference, chunkC chan retData) {
	g.mu.Lock()
	key := common.Bytes2Hex(k)
	g.queue.ret[key] = chunkC
	g.queue.chunks = append(g.queue.chunks, &cloud.RawChunk{ChunkID: k})
	lenchunks := len(g.queue.chunks)
	g.mu.Unlock()
	if atomic.LoadInt64(&g.at) == 0 && lenchunks > 23 {
		g.Force()
	}
}

func NewGetter(url string, encryption string) *ClientGetter {
	ret := make(map[string](chan retData))
	queue := getQueue{ret: ret}
	// TODO: use encryption here
	getter := ClientGetter{url: url, queue: queue, encryption: encryption}
	go func() {
		ticker := time.NewTicker(time.Millisecond * 50)
		for {
			select {
			case _ = <-ticker.C:
				if atomic.LoadInt64(&getter.at) == 0 {
					if len(getter.queue.chunks) > 0 {
						getter.Force()
					}
				}
			}
		}
	}()
	return &getter
}
