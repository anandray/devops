package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

const mailCollection = "mail"
const friendsCollection = "friends"

var dummyoptions = wolk.NewRequestOptions()

func init() {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
}

func getTestKeyVal() ([]byte, []byte) {
	s := wolkcommon.Computehash([]byte(time.Now().Format("2006-01-02 15:04:05")))
	t := wolkcommon.Computehash(s)
	return t, s
}

func generateRandomData(l int) (r io.Reader, slice []byte) {
	slice = make([]byte, l)
	rand.Seed(time.Now().Unix())
	if _, err := rand.Read(slice); err != nil {
		panic("rand error")
	}
	r = io.LimitReader(bytes.NewReader(slice), int64(l))
	return
}

func newWolkClient(t *testing.T) *WolkClient {
	rs, err := NewWolkClient("c0.wolk.com", 84, "")
	if err != nil {
		t.Fatalf("[client_test:newWolkClient] %s", err)
	}
	return rs
}

func TestClient(t *testing.T) {
	newWolkClient(t)
}

func TestAccountCreation(t *testing.T) {
	cl := newWolkClient(t)
	name := fmt.Sprintf("adam%d", time.Now().Unix())
	var options wolk.RequestOptions
	txhash, err := cl.CreateAccount(name, &options)
	if err != nil {
		t.Fatalf("CreateAccount %v", err)
	}
	cl.WaitForTransaction(txhash)
	cl.SetDefaultAccount(name)

	cl = newWolkClient(t)
	cl.LoadAccount(name)
	cl.DumpAccount()
	account, err := cl.GetAccount(name, nil)
	if err != nil {
		t.Fatalf("GetAccount %v", err)
	}
	fmt.Printf("ACCOUNT: %s\n", account.String())
}

func createUser(t *testing.T, owner string) (*WolkClient, error) {
	writers := make([]common.Address, 0)
	var emptyAddress common.Address
	var options wolk.RequestOptions
	options.ValidWriters = append(writers, emptyAddress)
	options.Schema = "wolk://wolk/schema/Mail"
	options.Encryption = "none"

	cl := newWolkClient(t)
	fmt.Printf("CreateAccount(%s)...", owner)
	txhash, err := cl.CreateAccount(owner, &options)
	if err != nil {
		return cl, fmt.Errorf("CreateAccount %v", err)
	}
	fmt.Printf(" txhash: %x\n", txhash)
	cl.WaitForTransaction(txhash)

	collection := fmt.Sprintf("wolk://%s/%s", owner, mailCollection)
	fmt.Printf("CreateCollection(%s)...", collection)
	txhash, err = cl.CreateBucket(wolk.BucketNoSQL, owner, collection, &options)

	if err != nil {
		return cl, fmt.Errorf("CreateCollection %v", err)
	}
	cl.WaitForTransaction(txhash)
	fmt.Printf(" txhash: %x\n", txhash)
	return cl, nil
}

func TestFriends(t *testing.T) {
	AName := fmt.Sprintf("alice%d", time.Now().Unix())
	clA, err := createUser(t, AName)
	if err != nil {
		t.Fatalf("createUserA %v", err)
	}

	BName := fmt.Sprintf("bob%d", time.Now().Unix())
	clB, err := createUser(t, BName)
	if err != nil {
		t.Fatalf("createUserB %v", err)
	}

	// A looking up B's Public Key
	accountB, err := clA.GetAccount(BName, nil)
	if err != nil {
		t.Fatalf("GetAccount %v", err)
	}
	fmt.Printf("clA.GetAccount(%s) %s", BName, accountB.String())
	// B looking up A's Public Key
	accountA, err := clB.GetAccount(AName, nil)
	if err != nil {
		t.Fatalf("GetAccount %v", err)
	}
	fmt.Printf("clB.GetAccount(%s)=%s\n", AName, accountA.String())

	publicKeyB, err := accountB.GetRSAPublicKey()
	if err != nil {
		t.Fatalf("GetRSAPublicKey %v", err)
	}
	publicKeyA, err := accountA.GetRSAPublicKey()
	if err != nil {
		t.Fatalf("GetRSAPublicKey %v", err)
	}

	// B send friendsKey to A using A's publickey
	friendsKeyB := clB.friendsDecryptionKey
	fmt.Printf("friendsKeyB: %x\n", friendsKeyB)
	val, err := clB.AddRSABlob(friendsKeyB, publicKeyA)
	if err != nil {
		t.Fatalf("AddRSABlob %v", err)
	}
	txhashB, err := clB.SetKey(AName, friendsCollection, "msgB", val, dummyoptions)

	if err != nil {
		t.Fatalf("AddRSABlob %v", err)
	}
	fmt.Printf("txhashB: %x\n", txhashB)
	clB.WaitForTransaction(txhashB)

	// A send message to B using A's publickey
	friendsKeyA := clA.friendsDecryptionKey
	fmt.Printf("friendsKeyA: %x\n", friendsKeyA)
	val, err = clA.AddRSABlob(friendsKeyA, publicKeyB)
	if err != nil {
		t.Fatalf("AddRSABlob %v", err)
	}
	txhashA, err := clA.SetKey(BName, friendsCollection, "msgA", val, dummyoptions)
	if err != nil {
		t.Fatalf("AddRSABlob %v", err)
	}
	fmt.Printf("txhashA: %x\n", txhashA)
	clA.WaitForTransaction(txhashA)

	// B checks his mail
	_, itemsB, err := clB.GetCollection(BName, friendsCollection, nil)
	if err != nil {
		t.Fatalf("ScanCollection %v", err)
	}
	for i, item := range itemsB {
		friendsKey, ok, err := clB.GetRSABlob(item.ValHash.Bytes())
		if err != nil {
			t.Fatalf("GetRSABlob %v", err)
		} else if ok {
			fmt.Printf("itemsB[%d]: %x\n", i, string(friendsKey)) // show who
		}
	}

	// A checks her mail
	_, itemsA, err := clA.GetCollection(AName, friendsCollection, nil)
	if err != nil {
		t.Fatalf("ScanCollection %v", err)
	}
	for i, item := range itemsA {
		friendsKey, ok, err := clA.GetRSABlob(item.ValHash.Bytes())
		if err != nil {
			t.Fatalf("GetRSABlob %v", err)
		} else if ok {
			fmt.Printf("itemsA[%d]: %x\n", i, string(friendsKey)) // show who
		}
	}
}

func TestAccountMail(t *testing.T) {
	AName := fmt.Sprintf("alice%d", time.Now().Unix())
	clA, err := createUser(t, AName)
	if err != nil {
		t.Fatalf("createUserA %v", err)
	}

	BName := fmt.Sprintf("bob%d", time.Now().Unix())
	clB, err := createUser(t, BName)
	if err != nil {
		t.Fatalf("createUserB %v", err)
	}

	// A looking up B's Public Key
	accountB, err := clA.GetAccount(BName, nil)
	if err != nil {
		t.Fatalf("GetAccount %v", err)
	}
	fmt.Printf("clA.GetAccount(%s) %s", BName, accountB.String())
	// B looking up A's Public Key
	accountA, err := clB.GetAccount(AName, nil)
	if err != nil {
		t.Fatalf("GetAccount %v", err)
	}
	fmt.Printf("clB.GetAccount(%s)=%s\n", AName, accountA.String())

	publicKeyB, err := accountB.GetRSAPublicKey()
	if err != nil {
		t.Fatalf("GetRSAPublicKey %v", err)
	}
	publicKeyA, err := accountA.GetRSAPublicKey()
	if err != nil {
		t.Fatalf("GetRSAPublicKey %v", err)
	}

	// B send message to A using A's publickey
	val, err := clB.AddRSABlob([]byte("Hi Mom!  I love you!"), publicKeyA)
	if err != nil {
		t.Fatalf("AddRSABlob %v", err)
	}
	txhashB, err := clB.SetKey(AName, mailCollection, "msgB", val, dummyoptions)
	if err != nil {
		t.Fatalf("AddRSABlob %v", err)
	}
	fmt.Printf("txhashB: %x\n", txhashB)
	clB.WaitForTransaction(txhashB)

	// A send message to B using A's publickey
	val, err = clA.AddRSABlob([]byte("Hi Bobby! I love you too!"), publicKeyB)
	if err != nil {
		t.Fatalf("AddRSABlob %v", err)
	}
	txhashA, err := clA.SetKey(BName, mailCollection, "msgA", val, dummyoptions)
	if err != nil {
		t.Fatalf("AddRSABlob %v", err)
	}
	fmt.Printf("txhashA: %x\n", txhashA)
	clA.WaitForTransaction(txhashA)

	// B checks his mail
	_, itemsB, err := clB.GetCollection(BName, mailCollection, nil)
	if err != nil {
		t.Fatalf("GetCollection %v", err)
	}
	for i, item := range itemsB {
		chunk, ok, err := clB.GetRSABlob(item.ValHash.Bytes())
		if err != nil {
			t.Fatalf("GetRSABlob %v", err)
		} else if ok {
			fmt.Printf("itemsB[%d]: %s\n", i, string(chunk))
		}
	}

	// A checks her mail
	_, itemsA, err := clA.GetCollection(AName, mailCollection, nil)
	if err != nil {
		t.Fatalf("ScanCollection %v", err)
	}
	for i, item := range itemsA {
		chunk, ok, err := clA.GetRSABlob(item.ValHash.Bytes())
		if err != nil {
			t.Fatalf("GetRSABlob %v", err)
		} else if ok {
			fmt.Printf("itemsA[%d]: %s\n", i, string(chunk))
		}
	}
}

func TestCTR(t *testing.T) {
	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4097, 8191, 8192}

	key, _ := hex.DecodeString("6368616e6765207468697320706173736368616e676520746869732070617373")
	client := newWolkClient(t)
	for _, sz := range sizes {
		_, chunk := generateRandomData(sz)
		//chunk := []byte("Sourabh Niyogi.")
		ciphertext, err := client.encrypt(chunk, key)
		if err != nil {
			t.Fatalf("encrypt %v", err)
		}
		chunk2, err := client.decrypt(ciphertext, key)
		if err != nil {
			t.Fatalf("decrypt %v", err)
		}

		if bytes.Compare(chunk2, chunk) != 0 {
			t.Fatalf("fail compare %d/%d bytes\n", len(chunk2), len(chunk))
		}
		fmt.Printf("TestCTR %d bytes\n", sz)
	}
}

func TestChunk(t *testing.T) {
	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096}

	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	client := newWolkClient(t)

	// test EncryptionNone, EncryptionPersonal, EncryptionFriends
	for i := 0; i < 3; i++ {
		encryption := wolk.EncryptionNone
		if i == 1 {
			encryption = wolk.EncryptionPersonal
		} else if i == 2 {
			encryption = wolk.EncryptionFriends
		}
		options := new(wolk.RequestOptions)
		options.Encryption = encryption
		for _, sz := range sizes {
			_, chunk := generateRandomData(sz)

			// setChunk
			st := time.Now()
			chunkKey, err := client.SetChunk(ctx, chunk, options)
			if err != nil {
				t.Fatalf("[client_test:TestChunk] SetChunk ERR %v", err)
			}
			log.Info("[client_test:TestChunk] SetChunk SUCC", "encryption", options.Encryption, "sz", sz, "tm", time.Since(st), "chunkID", fmt.Sprintf("%x", chunkKey))

			chunk0, ok, err := client.GetChunk(ctx, chunkKey.Bytes(), options)
			if err != nil {
				t.Fatalf("[client_test:TestChunk] GetChunk ERR %v", err)
			}
			if !ok {
				log.Error("[client_test:TestChunk] GetChunk-ERR", "ok", ok)
			} else if bytes.Compare(chunk0, chunk) != 0 {
				log.Error("[client_test:TestChunk] GetChunk Fail", "resp fail", string(chunk0))
			} else {
				log.Info("[client_test:TestChunk] GetChunk SUCC", "encryption", options.Encryption, "sz", sz, "tm", time.Since(st))
				q := sz
				if q > 64 {
					q = 64
				}
				log.Debug("[client_test:TestChunk] GetChunk %x...(%d bytes) vs %x...(%d bytes)\n", chunk0[0:q], len(chunk0), chunk[0:q], len(chunk))
			}
		}
	}
}

func TestFull(t *testing.T) {
	rs := newWolkClient(t)

	k, v := getTestKeyVal()
	prefix := fmt.Sprintf("some random value %s", time.Now())
	missingKey := wolkcommon.Computehash([]byte(prefix))
	emptyVal := []byte("")

	// (1) NORMAL CASE
	// SetChunk
	st := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	_, err := rs.SetChunk(ctx, v, nil)
	if err != nil {
		log.Error("[client_test:TestFull] SetChunk", "err", err)
	} else {
		log.Info("[client_test:TestFull] SetChunk PASS", "w", time.Since(st))
	}

	// GetChunk
	st = time.Now()
	v0, _, err := rs.GetChunk(ctx, k, nil)
	if err != nil {
		log.Error("[client_test:TestFull] GetChunk", "err", err)
	} else if bytes.Compare(v0, v) != 0 {
		log.Error("[client_test:TestFull] Compare - bytes don't match")
	} else {
		log.Info("[client_test:TestFull] GetChunk PASS", "r", time.Since(st))
	}

	// (2) ZERO-BYTE CASE
	// SetChunk( k, []byte("") )
	k = wolkcommon.Computehash(emptyVal)
	_, err = rs.SetChunk(ctx, emptyVal, nil)
	if err != nil {
		log.Error("[client_test:TestFull] SetChunk", "err", err)
	} else {
		log.Info("[client_test:TestFull] SetChunk PASS", "r", time.Since(st))
	}
	// GetChunk(k) [with EMPTY value] should just work normally
	st = time.Now()
	v0, ok, err := rs.GetChunk(ctx, k, nil)
	if err != nil {
		log.Error("[client_test:TestFull] GetChunk", "err", err)
	} else if ok == false {
		log.Error("[client_test:TestFull] GetChunk 0-byte value has ok = false")
	} else if len(v0) != 0 {
		log.Error("[client_test:TestFull] GetChunk 0-byte value has non-zero byte value")
	} else {
		// GOOD! ok is true, 0-byte value returned
		log.Info("[client_test:TestFull] GetChunk 0-byte value PASS", "r", time.Since(st))
	}

	// (3) KEY NOT FOUND case: should NOT return an err -- instead, ok should be FALSE
	st = time.Now()
	v0, ok, err = rs.GetChunk(ctx, missingKey, nil)
	if err != nil {
		log.Error("[client_test:TestFull] FAILURE GetChunk MISSING key returning error", "err", err)
	} else if ok == true {
		log.Error("[client_test:TestFull] FAILURE GetChunk on MISSING key ok = true", "err", err)
	} else if len(v0) != 0 {
		log.Error("[client_test:TestFull] FAILURE GetChunk on MISSING key returning value", "err", err)
	} else {
		// GOOD! ok is false, err is nil
		log.Info("[client_test:TestFull] GetChunk 0-byte value PASS ", "r", time.Since(st))
	}

	// SetChunkBatch
	st = time.Now()
	NCHUNKS := 5
	chunks := make([]*cloud.RawChunk, 0)
	for i := 0; i < NCHUNKS; i++ {
		v := wolkcommon.Computehash([]byte(fmt.Sprintf("123456789j%d", i)))
		k := wolkcommon.Computehash(v)
		chunks = append(chunks, &cloud.RawChunk{ChunkID: k, Value: v})
	}
	//	st = time.Now()
	err = rs.SetChunkBatch(ctx, chunks)
	if err != nil {
		log.Error("[client_test:TestFull] SetChunkBatch", "err", err)
	} else {
		log.Info("[client_test:TestFull] SetChunkBatch", "w", time.Since(st))
	}

	// GetChunkBatch
	st = time.Now()
	chunksR := make([]*cloud.RawChunk, 0)
	for i := 0; i < NCHUNKS; i++ {
		v := wolkcommon.Computehash([]byte(fmt.Sprintf("123456789j%d", i)))
		k := wolkcommon.Computehash(v)
		chunksR = append(chunksR, &cloud.RawChunk{ChunkID: k})
	}
	respChunks, err := rs.GetChunkBatch(ctx, chunksR)
	if err != nil {
		log.Error("[client_test:TestFull] GetChunkBatch", "err", err)
	}
	nerrors := 0
	for i := 0; i < NCHUNKS; i++ {
		if bytes.Compare(chunks[i].Value, respChunks[i].Value) != 0 {
			log.Error("[client_test:TestFull] GetChunkBatch Compare mismatch", "i", i)
			nerrors++
		}
	}
	if nerrors == 0 {
		log.Info("[client_test:TestFull] GetChunkBatch", "w", time.Since(st))
	}
}

func TestFile(t *testing.T) {

	rs := newWolkClient(t)
	ts := int32(time.Now().Unix())
	name := fmt.Sprintf("randomname%d", ts)

	// LatestBlockNumber
	latestBlockNumber, err := rs.LatestBlockNumber()
	if err != nil {
		t.Fatalf("LatestBlockNumber Err %v", err)
	}
	fmt.Printf("LatestBlockNumber()=%d\n\n", latestBlockNumber)
	options := new(wolk.RequestOptions)
	txhash, txhasherr := rs.CreateAccount(name, options)
	if txhasherr != nil {
		t.Fatalf("SetName %v", txhasherr)
	}
	fmt.Printf("SetName(%s): %x\n", name, txhash)
	_, err = rs.WaitForTransaction(txhash)

	options.Proof = "1"
	addr, err := rs.CreateAccount(name, options)
	if err != nil {
		t.Fatalf("GetName %v", err)
	}
	fmt.Printf("GetName(%s): %x\n", name, addr)

	// GetTransaction
	tx, err := rs.GetTransaction(txhash)
	if err != nil {
		t.Fatalf("GetTransaction error %v", err)
	}
	fmt.Printf("GetTransaction(txhash %x)=%s\n", txhash, tx.String())

	// GetBlock
	b, err := rs.GetBlock(latestBlockNumber)
	if err != nil {
		t.Fatalf("GetBlock error %v", err)
	}
	fmt.Printf("GetBlock(%d)=%s\n", latestBlockNumber, b.String())

	// GetBalance
	bucket := fmt.Sprintf("randombucket%d", ts)
	bucketurl := fmt.Sprintf("wolk://%s/%s", name, bucket)

	bucketDef := new(wolk.RequestOptions)
	bucketDef.RequesterPays = 1
	txhash, err = rs.CreateBucket(wolk.BucketNoSQL, name, bucket, bucketDef)
	fmt.Printf("MakeBucket(%s) => %x\n", bucketurl, txhash)
	_, err = rs.WaitForTransaction(txhash)

	sizes := []int{8193, 12287, 524288} // , 524288 + 1, 524288 + 4097, 1000000, 7 * 524288, 7*524288 + 1, 7*524288 + 4097}
	for i, sz := range sizes {
		_, v := generateRandomData(sz)
		localfile := fmt.Sprintf("fn%d", i)
		ioutil.WriteFile(localfile, v, 0644)
		wolkurl := fmt.Sprintf("wolk://%s/%s/fn%d", name, bucket, i)
		txhash, err := rs.PutFile(localfile, wolkurl, nil)
		if err != nil {
			t.Fatalf("PutFile %v", err)
		}
		fmt.Printf("PutFile(%s, %s) => %x\n", localfile, wolkurl, txhash)
		rs.WaitForTransaction(txhash)

		v2, err := rs.GetFile(wolkurl, options)
		if err != nil {
			t.Fatalf("GetFile %v", err)
		}
		if len(v) != len(v2) {
			t.Fatalf("Failure in size match %d != %d", len(v), len(v2))
		} else if bytes.Compare(v, v2) != 0 {
			t.Fatalf("Failure in byte match")
		}
	}

}

func TestBatch(t *testing.T) {
	rs := newWolkClient(t)
	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192}
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	// SetChunkBatch
	chunks := make([]*cloud.RawChunk, 0)
	for _, sz := range sizes {
		_, v := generateRandomData(sz)
		k := wolkcommon.Computehash(v)
		chunks = append(chunks, &cloud.RawChunk{ChunkID: k, Value: v})
	}
	start := time.Now()
	err := rs.SetChunkBatch(ctx, chunks)
	if err != nil {
		log.Error("[client_test:TestRemoteBatch] SetChunkBatch", "err", err)
	} else {
		log.Info("[client_test:TestRemoteBatch] SetChunkBatch", "w", time.Since(start))
	}

	// GetChunkBatch
	st := time.Now()
	chunksR := make([]*cloud.RawChunk, 0)
	for i, _ := range sizes {
		chunksR = append(chunksR, &cloud.RawChunk{ChunkID: chunks[i].ChunkID})
	}
	respChunks, err := rs.GetChunkBatch(ctx, chunksR)
	if err != nil {
		log.Error("[client_test:TestRemoteBatch] GetChunkBatch", "err", err)
	} else {
		successes := 0
		for i, _ := range sizes {
			if bytes.Compare(chunks[i].Value, respChunks[i].Value) != 0 {
				log.Error("[client_test:TestRemoteBatch] GetChunkBatch Compare mismatch", "w bytes", len(chunks[i].Value), "r bytes", len(chunksR[i].Value))
			} else {
				successes++
			}
		}
		log.Info("[client_test:TestRemoteBatch] GetChunkBatch", "r", time.Since(st), "rw", time.Since(start))
	}

}

func waitForTransactions(t *testing.T, rs *WolkClient, txhashes []common.Hash) error {
	for _, txhash := range txhashes {
		_, err := rs.WaitForTransaction(txhash)
		if err != nil {
			return fmt.Errorf("[client_test:WaitForTransactions] erred on tx(%x), err:%s", txhash, err)
		}
	}
	time.Sleep(5 * time.Second)
	return nil
}

func TestSQLClient(t *testing.T) {

	wclient := newWolkClient(t)
	//wclient.SetVerbose(true)
	mutatereqs, readreqs, readexpected, _, ownername, _ := getRequests_sqlclient(t)
	tprint("ownername(%s)", ownername)
	options := new(wolk.RequestOptions)
	txhash, err := wclient.CreateAccount(ownername, options)
	if err != nil {
		if strings.Contains(err.Error(), "Account exists already on blockchain") {
			// use existing account
		} else {
			t.Fatal(err)
		}
	} else {
		tprint("CreateAccount(%s), hash(%x)", ownername, txhash)
		_, err = wclient.WaitForTransaction(txhash)
		if err != nil {
			t.Fatal(err)
		}
	}

	//submit SQL mutate requests

	txhashes := make([]common.Hash, 0)
	for i := 0; i < len(mutatereqs); i++ {
		req := mutatereqs[i]
		tprint("submitting MUTATE req(%+v)", req)
		txhash, err := wclient.MutateSQL(req.Owner, req.Database, &req, options)
		if err != nil {
			t.Fatalf("[client_test:TestSQL] MutateSQL ERR %v", err)
		}
		//tprint("submitted mutate req(%+v)", req)
		txhashes = append(txhashes, txhash)
	}
	tprint("waiting for mutate txns...")
	err = waitForTransactions(t, wclient, txhashes)
	if err != nil {
		t.Fatalf("[client_test:TestSQL] waitForTransactions ERR %v", err)
	}
	tprint("finished waiting for mutate txns")

	//this one is an update
	// req := mutatereqs[len(txhashes)-1]
	// txhash, err = wclient.MutateSQL(req.Owner, req.Database, &req)
	// if err != nil {
	// 	t.Fatalf("[client_test:TestSQL] MutateSQL ERR %v", err)
	// }
	// txhashes = append(txhahses, txhash)
	// tprint("submitted mutate req(%+v)", req)
	// _, err = wclient.WaitForTransaction(txhashes[len(txhashes)-1])
	// if err != nil {
	// 	t.Fatalf("[client_test:TestSQL] waitForTransaction ERR %v", err)
	// }
	// time.Sleep(5 * time.Second)

	//sql read requests and check
	for i := 0; i <= 7; i++ {
		if _, ok := readreqs[i]; !ok {
			continue
		}
		req := readreqs[i]
		options := wolk.NewRequestOptions()
		actual, err := wclient.ReadSQL(req.Owner, req.Database, &req, options)
		if err != nil {
			t.Fatal(err)
		}
		tprint("submitted read req(%+v)", req)
		tprint("(%d) result: %+v", i, actual)
		if _, ok := readexpected[i]; !ok { //skip tests w/o expected to match
			continue
		}
		//tprint("(%d) expected: %+v", i, readexpected[i])
		if !reflect.DeepEqual(actual, readexpected[i]) {
			if wolk.MatchRowData(actual.Data, readexpected[i].Data) && actual.AffectedRowCount == readexpected[i].AffectedRowCount && actual.MatchedRowCount == readexpected[i].MatchedRowCount {
				//passed
			} else {
				fmt.Printf("expected: %+v", readexpected[i])
				t.Fatalf("test %d failed", i)
			}
		}
	}
}

func getRequests_sqlclient2(t *testing.T) (mutaterequests map[int]wolk.SQLRequest, readrequests map[int]wolk.SQLRequest, readexpected map[int]wolk.SQLResponse, cleanuprequests map[int]wolk.SQLRequest, owner string, dbname string) {
	owner = wolk.MakeName("testowner-debug.eth")
	dbname = wolk.MakeName("testdb")
	tablename := wolk.MakeName("newtable")
	encrypted := 0

	columns := []wolk.Column{
		wolk.Column{
			ColumnName: "person_id",
			ColumnType: wolk.CT_INTEGER,
			IndexType:  wolk.IT_BPLUSTREE,
			Primary:    0,
		},
		wolk.Column{
			ColumnName: "name",
			ColumnType: wolk.CT_STRING,
			IndexType:  wolk.IT_BPLUSTREE,
			Primary:    1,
		},
	}

	mutaterequests = map[int]wolk.SQLRequest{
		0: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_CREATE_DATABASE},
		1: wolk.SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, Columns: columns, RequestType: wolk.RT_CREATE_TABLE},
		2: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (6, 'dinozzo')"},
		3: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (23, 'ziva')"},
		4: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (2, 'mcgee')"},
		5: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (101, 'gibbs')"},
		6: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "delete from " + tablename + " where name = 'ziva'"},
		//6: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "update " + tablename + " set age = 23 where email = 'test07@wolk.com'"},
	}

	readrequests = map[int]wolk.SQLRequest{
		//0: wolk.SQLRequest{Owner: owner, Encrypted: encrypted, RequestType: wolk.RT_LIST_DATABASES},
		1: wolk.SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: wolk.RT_DESCRIBE_TABLE},
		2: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename},
		3: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where person_id = 6"},
		4: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where person_id > 6"},
		5: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where person_id >= 6"},
		6: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where person_id > 7"},
		7: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where person_id >= 7"},
	}

	// readexpected = map[int]wolk.SQLResponse{
	// 	//0: {Data: [] Row{ Row{"database": dbname}}},
	// 	1: {Data: []wolk.Row{wolk.Row{"ColumnName": "age", "IndexType": "BPLUS", "Primary": 0, "ColumnType": "INTEGER"}, wolk.Row{"ColumnName": "email", "IndexType": "BPLUS", "Primary": 1, "ColumnType": "STRING"}}, AffectedRowCount: 0, MatchedRowCount: 0},
	// 	2: {Data: []wolk.Row{wolk.Row{"age": 2, "email": "test06@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
	// 	3: {Data: []wolk.Row{
	// 		wolk.Row{"age": 1, "email": "test05@wolk.com"},
	// 		wolk.Row{"age": 2, "email": "test06@wolk.com"},
	// 		wolk.Row{"age": 23, "email": "test07@wolk.com"},
	// 		wolk.Row{"age": 4, "email": "test08@wolk.com"},
	// 		wolk.Row{"age": 5, "email": "test09@wolk.com"},
	// 		wolk.Row{"age": 6, "email": "test10@wolk.com"}},
	// 		AffectedRowCount: 0,
	// 		MatchedRowCount:  6},
	// 	4: {Data: []wolk.Row{wolk.Row{"age": 6, "email": "test10@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
	// 	5: {Data: []wolk.Row{
	// 		wolk.Row{"age": 5,"email": "test09@wolk.com"},
	// 		wolk.Row{"age": 6, "email": "test10@wolk.com"},
	// 		wolk.Row{"age": 23,"email": "test07@wolk.com"}},
	// 		AffectedRowCount: 0,
	// 		MatchedRowCount: 3},
	// 	6: {Data: []wolk.Row{
	// 		wolk.Row{"age": 5,"email": "test09@wolk.com"},
	// 		wolk.Row{"age": 6, "email": "test10@wolk.com"},
	// 		wolk.Row{"age": 23,"email": "test07@wolk.com"}},
	// 		AffectedRowCount: 0,
	// 		MatchedRowCount: 3},
	// 	7: {Data: []wolk.Row{
	// 		wolk.Row{"age":4,"email":"test08@wolk.com"},
	// 		wolk.Row{"age": 5,"email": "test09@wolk.com"},
	// 		wolk.Row{"age": 6, "email": "test10@wolk.com"},
	// 		wolk.Row{"age": 23,"email": "test07@wolk.com"}},
	// 		AffectedRowCount: 0,
	// 		MatchedRowCount: 4},
	//}

	cleanuprequests = map[int]wolk.SQLRequest{
		0: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_DROP_DATABASE},
	}

	return mutaterequests, readrequests, readexpected, cleanuprequests, owner, dbname
}

func getRequests_sqlclient(t *testing.T) (mutaterequests map[int]wolk.SQLRequest, readrequests map[int]wolk.SQLRequest, readexpected map[int]wolk.SQLResponse, cleanuprequests map[int]wolk.SQLRequest, owner string, dbname string) {
	owner = wolk.MakeName("testowner-debug.eth")
	dbname = wolk.MakeName("testdb")
	tablename := wolk.MakeName("newtable")
	encrypted := 0

	columns := []wolk.Column{
		wolk.Column{
			ColumnName: "age",
			ColumnType: wolk.CT_INTEGER,
			IndexType:  wolk.IT_BPLUSTREE,
			Primary:    0,
		},
		wolk.Column{
			ColumnName: "email",
			ColumnType: wolk.CT_STRING,
			IndexType:  wolk.IT_BPLUSTREE,
			Primary:    1,
		},
	}

	putrows := []wolk.Row{
		wolk.Row{"email": "test05@wolk.com", "age": 1},
		wolk.Row{"email": "test06@wolk.com", "age": 2},
		wolk.Row{"email": "test07@wolk.com", "age": 3},
		wolk.Row{"email": "test08@wolk.com", "age": 4},
		wolk.Row{"email": "test09@wolk.com", "age": 5},
	}

	mutaterequests = map[int]wolk.SQLRequest{
		0: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_CREATE_DATABASE},
		1: wolk.SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, Columns: columns, RequestType: wolk.RT_CREATE_TABLE},
		2: wolk.SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: wolk.RT_PUT, Rows: putrows},
		3: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "insert into " + tablename + " (email, age) values ('test10@wolk.com', 6)"},
		4: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "update " + tablename + " set age = 23 where email = 'test07@wolk.com'"},
	}

	readrequests = map[int]wolk.SQLRequest{
		//0: wolk.SQLRequest{Owner: owner, Encrypted: encrypted, RequestType: wolk.RT_LIST_DATABASES},
		1: wolk.SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: wolk.RT_DESCRIBE_TABLE},
		2: wolk.SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: wolk.RT_GET, Key: "test06@wolk.com"},
		3: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename},
		4: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where age = 6"},
		5: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where age > 4"},
		6: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where age >= 5"},
		7: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_QUERY, RawQuery: "select * from " + tablename + " where age > 3"},
	}

	readexpected = map[int]wolk.SQLResponse{
		//0: {Data: []wolk.Row{wolk.Row{"database": dbname}}},
		1: {Data: []wolk.Row{wolk.Row{"ColumnName": "age", "IndexType": "BPLUS", "Primary": 0, "ColumnType": "INTEGER"}, wolk.Row{"ColumnName": "email", "IndexType": "BPLUS", "Primary": 1, "ColumnType": "STRING"}}, AffectedRowCount: 0, MatchedRowCount: 0},
		2: {Data: []wolk.Row{wolk.Row{"age": 2, "email": "test06@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		3: {Data: []wolk.Row{
			wolk.Row{"age": 1, "email": "test05@wolk.com"},
			wolk.Row{"age": 2, "email": "test06@wolk.com"},
			wolk.Row{"age": 23, "email": "test07@wolk.com"},
			wolk.Row{"age": 4, "email": "test08@wolk.com"},
			wolk.Row{"age": 5, "email": "test09@wolk.com"},
			wolk.Row{"age": 6, "email": "test10@wolk.com"}},
			AffectedRowCount: 0,
			MatchedRowCount:  6},
		4: {Data: []wolk.Row{wolk.Row{"age": 6, "email": "test10@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		5: {Data: []wolk.Row{
			wolk.Row{"age": 5, "email": "test09@wolk.com"},
			wolk.Row{"age": 6, "email": "test10@wolk.com"},
			wolk.Row{"age": 23, "email": "test07@wolk.com"}},
			AffectedRowCount: 0,
			MatchedRowCount:  3},
		6: {Data: []wolk.Row{
			wolk.Row{"age": 5, "email": "test09@wolk.com"},
			wolk.Row{"age": 6, "email": "test10@wolk.com"},
			wolk.Row{"age": 23, "email": "test07@wolk.com"}},
			AffectedRowCount: 0,
			MatchedRowCount:  3},
		7: {Data: []wolk.Row{
			wolk.Row{"age": 4, "email": "test08@wolk.com"},
			wolk.Row{"age": 5, "email": "test09@wolk.com"},
			wolk.Row{"age": 6, "email": "test10@wolk.com"},
			wolk.Row{"age": 23, "email": "test07@wolk.com"}},
			AffectedRowCount: 0,
			MatchedRowCount:  4},
	}

	cleanuprequests = map[int]wolk.SQLRequest{
		0: wolk.SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: wolk.RT_DROP_DATABASE},
	}

	return mutaterequests, readrequests, readexpected, cleanuprequests, owner, dbname
}

func tprint(in string, args ...interface{}) {
	if in == "" {
		fmt.Println()
	} else {
		fmt.Printf("[test] "+in+"\n", args...)
	}
}
