package chunker

import (
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/wolk/cloud"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	//	"os"
	//	"io/ioutil"
)

type ClientPutter struct {
	url        string
	queue      putterQueue
	mu         sync.Mutex
	encryption string
}

type putterQueue struct {
	chunk []cloud.RawChunk
	wg    []*sync.WaitGroup
}

func (p *ClientPutter) Put(ctx context.Context, chunk ChunkData, wg *sync.WaitGroup) (Reference, error) {
	p.PutChunk(chunk, wg)
	chunkKey := wolkcommon.Computehash(chunk)
	return Reference(chunkKey), nil
}

func (p *ClientPutter) RefSize() int64 {
	return 32
}

func (p *ClientPutter) Close() {
}

func (p *ClientPutter) Wait(ctx context.Context) error {
	return nil
}

func NewPutter(url string, encryption string) *ClientPutter {
	putter := ClientPutter{url: url, encryption: encryption}
	go func() {
		ticker := time.NewTicker(time.Millisecond * 20)
		for {
			select {
			case _ = <-ticker.C:
				if len(putter.queue.chunk) > 0 {
					putter.Flush()
				}
			}
		}
	}()
	return &putter
}

func (p *ClientPutter) PutChunk(chunk ChunkData, wg *sync.WaitGroup) {
	p.mu.Lock()
	// TODO: use encryption
	p.queue.chunk = append(p.queue.chunk, cloud.RawChunk{Value: []byte(chunk)})
	p.queue.wg = append(p.queue.wg, wg)
	if len(p.queue.chunk) > 100 {
		chunks, err := json.Marshal(p.queue.chunk)
		if err != nil {
			fmt.Println("PutChunk json error", err)
		}
		url := p.url + "/setbatch"
		body := bytes.NewReader(chunks)
		req, err := http.NewRequest(http.MethodPost, url, body)
		httpclient := &http.Client{Timeout: time.Second * 1000, Transport: DefaultTransport}
		req.Header.Add("cache-control", "no-cache")
		_, err = httpclient.Do(req)
		if err != nil {
			fmt.Println("PutChunk json error", err)
		}

		for _, wg := range p.queue.wg {
			wg.Done()
		}
		p.queue.chunk = nil
		p.queue.wg = nil
	}
	p.mu.Unlock()
}

func (p *ClientPutter) Flush() {
	p.mu.Lock()
	chunks, err := json.Marshal(p.queue.chunk)
	url := p.url + "/setbatch"
	body := bytes.NewReader(chunks)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		fmt.Println("Flush req error", err)
	}
	httpclient := &http.Client{Timeout: time.Second * 1000, Transport: DefaultTransport}
	req.Header.Add("cache-control", "no-cache")
	_, err = httpclient.Do(req)
	if err != nil {
		fmt.Println("Flush res error", err)
	}

	for _, wg := range p.queue.wg {
		wg.Done()
	}

	p.queue.chunk = nil
	p.queue.wg = nil
	p.mu.Unlock()

}
