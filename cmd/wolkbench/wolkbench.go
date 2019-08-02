package main

import (
	"bytes"
	"time"

	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/client"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk"
)

// keybench calls SetKey for a specific owner and collection with a random key-value pair
// if check is true, checks for correct value with a GetKey
func keyBench(cl *client.WolkClient, owner string, collection string, check int) (err error) {

	// make a random key
	keytoken := make([]byte, 6)
	rand.Read(keytoken)
	key := fmt.Sprintf("file-%x", keytoken)

	// make a random value
	input := make([]byte, 6)
	rand.Read(input)
	val := []byte(fmt.Sprintf("%s-%x", key, input))

	// write the key-value
	var txhash common.Hash
	var options wolk.RequestOptions
	options.WaitForTx = "1"
	//txhash, err = cl.SetKey(owner, collection, key, val, &options)
	txhash, err = cl.SetKey(owner, collection, key, val)
	if err != nil {
		log.Error("SetKey", "owner", owner, "collection", collection, "err", err)
		return fmt.Errorf("[filebench:KeyBench] SetKey %v", err)
	}
	if check < 0 {
		log.Info("SUCCESS")
		return nil
	}
	// wait for txhash to be included
	var tx *wolk.SerializedTransaction
	tx, err = cl.WaitForTransaction(txhash)
	if err != nil {
		log.Error("WaitForTransaction", "err", err)
		return err
	}
	// if no need to check, exit function without err
	if check <= 0 {
		log.Info("[filebench:KeyBench] SetKey tx included", "txhash", txhash, "owner", owner, "collection", collection, "key", key, "tx", tx)
		return nil
	}
	var resp []byte
	var sz uint64
	resp, sz, err = cl.GetKey(owner, collection, key, &options)
	if err != nil {
		return fmt.Errorf("[filebench:KeyBench] GetKey", "err", err)
	}
	if bytes.Compare(resp, val) != 0 {
		return fmt.Errorf("[filebench:KeyBench] GetKey incorrect value!", "expected", val, "resp", string(resp))
	}
	log.Info("[filebench:KeyBench] GetKey found", "owner", owner, "collection", collection, "key", key, "sz", sz, "resp", string(resp))
	return nil
}

var mem runtime.MemStats

// keybench calls PutFile for a specific owner and collection with a random file of filesize bytes
// if check is true, checks for correct value with GetFile
func fileBench(cl *client.WolkClient, owner string, collection string, filesize int, check int, bufpool *Pool) (err error) {
	filetoken := make([]byte, 8)
	rand.Read(filetoken)
	key := fmt.Sprintf("file-%x", filetoken)
	//input := make([]byte, filesize)
	obj := bufpool.GetBuf()
	obj.mu.Lock()
	//input := obj.buf
	//rand.Read(input)
	rand.Read(obj.buf)

	wolkpath := fmt.Sprintf("%s/%s/%s", owner, collection, key)
	tmpfile, err := ioutil.TempFile("", "test.*.txt")
	if err != nil {
		return fmt.Errorf("[filebench:fileBench] Tempfile", "err", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up
	//if _, err = tmpfile.Write(input); err != nil {
	if _, err = tmpfile.Write(obj.buf); err != nil {
		tmpfile.Close()
		return fmt.Errorf("[filebench:fileBench] Write", "err", err)
	}
	if err = tmpfile.Close(); err != nil {
		return fmt.Errorf("[filebench:fileBench] Close", "err", err)
	}
	//runtime.ReadMemStats(&mem)
	//log.Info("[http:setShareBatchHandler] mem", "alloc", mem.Alloc, "total", mem.TotalAlloc, "heap", mem.HeapAlloc, "heaps", mem.HeapSys)

	if check <= 0 {
		obj.mu.Unlock()
		bufpool.ReleaseBuf(obj) /// check order
	}
	//runtime.ReadMemStats(&mem)
	//log.Info("[http:setShareBatchHandler] mem", "alloc", mem.Alloc, "total", mem.TotalAlloc, "heap", mem.HeapAlloc, "heaps", mem.HeapSys)

	var txhash common.Hash
	var options wolk.RequestOptions
	options.WaitForTx = "1"
	//txhash, err = cl.PutFile(tmpfile.Name(), wolkpath, &options)
	txhash, err = cl.PutFile3(obj.buf, wolkpath, &options)
	//time.Sleep(time.Second * 10)
	log.Info("PutFile", "filename", tmpfile.Name(), "wolkpath", wolkpath, "txhash", fmt.Sprintf("%x", txhash))
	if err != nil {
		fmt.Printf("\n[filebench:fileBench] ERR: %s", err)
		os.Remove(tmpfile.Name())
		return fmt.Errorf("[filebench:fileBench] PutFile ERR %v")
	}
	//log.Info(fmt.Sprintf("[filebench:fileBench] put file with hash: %x", txhash))
	if check < 0 {
		os.Remove(tmpfile.Name())
		return nil
	}

	//cl.WaitForTransaction(txhash)
	if check <= 0 {
		//	log.Info("[filebench:PutFile] submitted tx", "file", tmpfile.Name(), "wolkpath", wolkpath, "len", len(input), "txhash", txhash)
		os.Remove(tmpfile.Name())
		return nil
	}

	tryGet := true
	getAttemptCount := 0
	var output []byte
	for tryGet {
		time.Sleep(10 * time.Second)
		getAttemptCount++
		if getAttemptCount > 15 {
			tryGet = false
		}
		log.Info("GetFile", "wolkpath", wolkpath, "txhash", fmt.Sprintf("%x", txhash))
		output, err = cl.GetFile(wolkpath, &options)
		if err != nil {
			os.Remove(tmpfile.Name())
			log.Error("[filebench:fileBench] GetFile FAILURE", "err", err)
			if tryGet == true {
				log.Info("GetFile | RETRY", "wolkpath", wolkpath, "txhash", fmt.Sprintf("%x", txhash))
			} else {
				return fmt.Errorf("[filebench:fileBench] GetFile | %s", err)
			}
		} else {
			tryGet = false
		}
	}

	//if bytes.Compare(output, input) != 0 {
	if bytes.Compare(output, obj.buf) != 0 {
		log.Error("[filebench:fileBench] GetFile FAILURE", "err", err)
		os.Remove(tmpfile.Name())
		return fmt.Errorf("filebench:fileBench] GetFile match")
	}
	obj.mu.Unlock()
	bufpool.ReleaseBuf(obj)
	log.Info("[filebench:filebench] GetFile SUCC", "wolkpath", wolkpath, "len", len(output))
	os.Remove(tmpfile.Name())
	return nil
}

// calls a benchmark operation (fileBench, keyBench) for files files perpetually!
func WolkBench(cl *client.WolkClient, run string, owner string, collection string, files int, filesize int, check int) (err error) {
	limit := make(chan int, files)
	var buf *Pool
	if strings.Compare(run, "file") == 0 {
		buf, err = NewBufPool(filesize, 10)
		if err != nil {
		}
	}
	i := 0
	for {
		limit <- 1
		i++
		go func() (err error) {
			if strings.Compare(run, "file") == 0 {
				err = fileBench(cl, owner, collection, filesize, check, buf)
			} else {
				err = keyBench(cl, owner, collection, check)
			}
			if err != nil {
				// log.Error("[filebench:KeyBench]", "err", err)
			}
			<-limit
			return nil
		}()
	}
}

func main() {
	server := flag.String("server", "c0.wolk.com", "address to submit transactions")
	httpport := flag.Uint("httpport", 443, "port cloudstore is listening on")
	users := flag.Uint("users", 5, "number of users uploading in parallel")
	files := flag.Uint("files", 5, "number of files/keys put/get in parallel")
	filesize := flag.Uint("filesize", 4096, "filesize")
	check := flag.Int("check", 0, "check write operations (-1 - do not wait for tx, 0 - wait for tx, 1 - wait and check values)")
	run := flag.String("run", "key", "test (key,nosql)")
	flag.Parse()
	fmt.Printf("filebench [server=%s:%d, users=%d, files=%d, filesize=%d bytes]\n", *server, *httpport, *users, *files, *filesize)
	log.New(log.LvlTrace, "client", "wolk-debug")

	rand.Seed(time.Now().Unix())
	var wg sync.WaitGroup
	for i := 0; i < int(*users); i++ {
		wg.Add(1)
		go func(i int) {
			cl, err := client.NewWolkClient(*server, int(*httpport), "")
			if err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(0)
			}

			usertoken := make([]byte, 4)
			rand.Read(usertoken)

			// Create Account
			owner := fmt.Sprintf("user%d-%x", i, usertoken)
			txhash, err := cl.CreateAccount(owner)
			if err != nil {
				log.Error("[filebench:main] CreateAccount", "err", err)
				return
			}
			time.Sleep(time.Second * 10)
			tx, err := cl.WaitForTransaction(txhash)
			if err != nil {
				log.Error("[filebench:main] CreateAccount WaitForTransaction took too long ... exiting", "err", err)
				os.Exit(0)
			}
			fmt.Printf("SetName(%s): %s\n", owner, tx.String())

			collection := "test"
			wolkurl := fmt.Sprintf("wolk://%s/%s", owner, collection)
			b := new(wolk.TxBucket)
			b.RequesterPays = 1
			var options wolk.RequestOptions
			txhash, err = cl.CreateBucket(wolk.BucketNoSQL, owner, collection, &options)
			if err != nil {
				log.Error("[filebench:main] CreateBucket", "err", err)
				os.Exit(0)
			}
			time.Sleep(time.Second * 10)
			tx, err = cl.WaitForTransaction(txhash)
			if err != nil {
				log.Error("[filebench:main] WaitForTransaction took too long... exiting", "err", err)
				os.Exit(0)
			}
			fmt.Printf("CreateBucket(%s) Complete %s\n", wolkurl, tx.String())
			err = WolkBench(cl, *run, owner, collection, int(*files), int(*filesize), *check)
			if err != nil {
				fmt.Printf("%+v\n", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

type Pool struct {
	obj     []*Obj
	bufsize int
	num     int
}

type Obj struct {
	pos       int
	buf       []byte
	available bool
	mu        sync.RWMutex
}

func NewBufPool(bufsize int, sz int) (*Pool, error) {
	pool := new(Pool)
	pool.obj = make([]*Obj, sz)
	pool.bufsize = bufsize
	pool.num = sz
	for i := 0; i < sz; i++ {
		pool.obj[i] = new(Obj)
		pool.obj[i].pos = i
		pool.obj[i].buf = make([]byte, bufsize)
		pool.obj[i].available = true
	}
	return pool, nil
}

func (p *Pool) GetBuf() *Obj {
	for {
		for _, b := range p.obj {
			b.mu.RLock()
			if b.available {
				b.mu.RUnlock()
				return b
			}
			b.mu.RUnlock()
		}
	}
}

func (p *Pool) ReleaseBuf(obj *Obj) {
	obj.mu.Lock()
	obj.available = true
	//obj.buf = make([]byte, p.bufsize)
	obj.mu.Unlock()
}
