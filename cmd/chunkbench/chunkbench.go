package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/wolkdb/cloudstore/cmd/chunkbench/util"
	wolkcommon "github.com/wolkdb/cloudstore/common"
)

const (
	chunkSize = 4096
)

type rec struct {
	err  int
	t    util.Stat
	k    []byte
	reqv []byte
	resv []byte
}

func postRun(server string, run string, method string, c int, b int, chunks [][]byte) <-chan *rec {
	limit := make(chan int, c)
	receiver := make(chan *rec, 100)
	cnt := 0

	var DefaultTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   1000 * time.Second, // 30 * time.Second,
			KeepAlive: 1000 * time.Second, // 60 * time.Second,
		}).Dial,
		MaxIdleConns:          10000,
		IdleConnTimeout:       1000 * time.Second, // 90 * time.Second,
		TLSHandshakeTimeout:   1000 * time.Second,
		ResponseHeaderTimeout: 1000 * time.Second,
	}

	go func() {
		baseurl := fmt.Sprintf("http://%s/wolk/%s", server, run)
		httpclient := &http.Client{Timeout: time.Second * 1000, Transport: DefaultTransport}

		for _, chunkdata := range chunks {
			limit <- 1
			cnt++
			go func(chunkdata []byte) {
				st := time.Now()
				k := wolkcommon.Computehash(chunkdata)
				var v []byte
				if method == "GET" {
					url := baseurl + fmt.Sprintf("/%x", k)

					req, _ := http.NewRequest("GET", url, nil)
					req.Header.Add("cache-control", "no-cache")
					req.Header.Add("Expect", "")
					res, err := httpclient.Do(req)
					if err != nil {
						fmt.Printf("%s %s %v\n", run, url, err)
						var a util.Stat
						a.Tm = float64(time.Since(st) / time.Millisecond)
						receiver <- &rec{0, a, k, chunkdata, []byte("")}
					} else {
						defer res.Body.Close()
						var a util.Stat
						// fmt.Printf("%s %s %d\n", run, url, res.StatusCode)
						a.Tm = float64(time.Since(st) / time.Millisecond)
						if res.StatusCode == 200 {
							v, _ = ioutil.ReadAll(res.Body)
							if len(v) != chunkSize/2 {
								// fmt.Printf("ERR (%d)\n", len(v))
							}
						}
						receiver <- &rec{res.StatusCode, a, k, chunkdata, v}
					}
				} else {
					body := bytes.NewReader(chunkdata)
					url := baseurl
					req, _ := http.NewRequest("POST", url, body)
					req.Close = true
					req.Header.Add("cache-control", "no-cache")
					res, err := httpclient.Do(req)
					if err != nil {
						fmt.Printf("%s %s (%x) %v\n", run, url, k, err)
						var a util.Stat
						a.Tm = float64(time.Since(st) / time.Millisecond)
						receiver <- &rec{0, a, k, chunkdata, v}
					} else {
						// fmt.Printf("%s %s (%x) %v\n", run, url, k, res.StatusCode)
						//defer res.Body.Close()
						if res.StatusCode == 200 {
							body2, reserr := ioutil.ReadAll(res.Body)
							res.Body.Close()
							if reserr != nil {
								// fmt.Printf("DONE %d %s\n", res.StatusCode, body2)
							} else {
								v = body2
							}
						}
						var a util.Stat
						a.Tm = float64(time.Since(st) / time.Millisecond)
						receiver <- &rec{res.StatusCode, a, k, chunkdata, v}
					}
				}
				<-limit
			}(chunkdata)
		}
	}()
	return receiver
}

func main() {
	// Usage: wb -server c0.wolk.com -n=1000 -c=10 -run share
	var server = flag.String("server", "c0.wolk.com", "http server (port:80)")
	var n = flag.Int("n", 200, "number of chunks to do POST/GET operations")
	var c = flag.Int("c", 50, "number of concurrent processes")
	var b = flag.Int("b", 25, "number of chunks in a batch")
	var run = flag.String("run", "share", "share/chunk")
	var numbers []util.Stat
	flag.Parse()
	fmt.Printf("Wolk Benchmark Server %s [c=%d run=%s, n=%d Chunks, chunkSize=%d bytes]\n", *server, *c, *run, *n, chunkSize)

	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	chunks := wolkcommon.GenerateRandomChunks(*n, chunkSize)
	var op string
	for phase := 0; phase < 3; phase++ {
		success := 0
		numbers = make([]util.Stat, 0)
		fmt.Println(*run)
		if strings.Compare(*run, "chunk") == 0 {
			if phase == 0 {
				op = "POST"
			} else {
				op = "GET"
			}
		} else if strings.Compare(*run, "share") == 0 {
			if phase == 0 {
				op = "POST"
			} else {
				op = "GET"
			}
		}
		fmt.Printf("Phase %d - %s %s\n", phase, op, *run)
		start := time.Now()
		var errmap map[int]int
		errmap = make(map[int]int)
		receiver := postRun(*server, *run, op, *c, *b, chunks)
		for i := 0; i < *n; i++ {
			res := <-receiver
			if _, ok := errmap[res.err]; ok {
				errmap[res.err] = errmap[res.err] + 1
			} else {
				errmap[res.err] = 1
			}
			if phase > 0 {
				// check for data length
				if strings.Compare(*run, "chunk") == 0 {
					if len(res.resv) == chunkSize {
						// check for data hash correctness
						if bytes.Compare(res.k, wolkcommon.Computehash(res.resv)) == 0 {
							success++
						} else {
							fmt.Printf("chunk mismatch %x %x %x\n", res.k, res.reqv, res.resv)
						}
					} else {
						fmt.Printf("chunk mismatch len(chunk)=%d\n", len(res.resv))
					}
				} else if strings.Compare(*run, "share") == 0 {
					if len(res.resv) == chunkSize {
						success++
					} else if res.err == 200 {
						fmt.Printf("share mismatch len(share)=%d\n", len(res.resv))
					} else {
						fmt.Printf("share missing %x\n", res.k)
					}
				} else if strings.Compare(*run, "batch") == 0 {
					if len(res.resv) == chunkSize {
						// check for data hash correctness
						if bytes.Compare(res.k, wolkcommon.Computehash(res.resv)) == 0 {
							success++
						} else {
							fmt.Printf("chunk mismatch %x %x %x\n", res.k, res.reqv, res.resv)
						}
					} else {
						fmt.Printf("chunk mismatch len(chunk)=%d\n", len(res.resv))
					}
				}
			} else {
				// TODO: check for merkle root
				success++
			}
			numbers = append(numbers, res.t)
		}
		util.Print_percentiles(numbers)
		fmt.Printf("errmap: %v - elapsed: %vs Success:%d\n\n", errmap, time.Since(start).Seconds(), success)
		time.Sleep(100 * time.Millisecond)
	}
}
