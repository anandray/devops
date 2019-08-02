// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	//"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

type Person struct {
	Name   string  `json:"name,omitempty"`
	Age    int64   `json:"age,omitempty"`
	Weight float64 `json:"weight,omitempty"`
}

func TestAVLBytes(t *testing.T) {
	f := byteToFloat(common.FromHex("44174154fb210940"))
	if f != float64(3.1415926535) {
		t.Fatalf("small float fail -- %f", f)
	}
	g := byteToFloat(common.FromHex("cc06e525e198fa41"))
	if g != float64(7139627614.31415926535) {
		t.Fatalf("bigger float fail -- %f", g)
	}
	i := byteToInt64(common.FromHex("a2ed7156feffffff"))
	if i != int64(-7139627614) {
		t.Fatalf("negative int failure fail")
	}
	j := byteToInt(common.FromHex("ffffffffffffffff"))
	if j != uint64(18446744073709551615) {
		t.Fatalf("max int failure")
	}
	fmt.Printf("[avltree_test:TestAVLBytes] f: %f g: %f i: %d: j: %d\n", f, g, i, j)
	// TODO: what about null bytes
}

func dumpRecordsAndProof(recs [][]byte, valuesmap map[common.Hash]string, p interface{}) {
	for i, r := range recs {
		if _, ok := valuesmap[common.BytesToHash(r)]; !ok {
			tprint("valuesmap:")
			for k, v := range valuesmap {
				tprint("k(%x) v(%s)", k, v)
			}
			panic(fmt.Sprintf("hash(%x) not found in valuesmap", r))
			//tprint("rec(%d) (%s)", i, r)
			//tprint("rec(%x) %s", r, valuesmap[common.BytesToAddress(r)])
		}
		tprint("rec(%d) (%s)", i, valuesmap[common.BytesToHash(r)])
	}
	//tprint("AVL proof: %+v", p.(*AVLProof))
}

type TestTxBucket struct {
	Name       string             `json:"name,omitempty"`
	Schema     string             `json:"schemaURL,omitempty"`
	AVLIndexes []*TestBucketIndex `json:"avlindexes,omitempty"` // AVL only. TODO: consolidate with Indexes?
}

type TestBucketIndex struct {
	IndexName string      `json:"indexName"`
	IndexType string      `json:"indexType"` // Text, Number, Float, Date, Time, DateTime, ...
	Primary   bool        `json:"primary,omitempty"`
	RootHash  common.Hash `json:"rootHash"`
	t         *AVLTree
}

func (b *TestTxBucket) Flush(ctx context.Context, wg *sync.WaitGroup, writeToCloudstore bool) (err error) {
	for _, idx := range b.AVLIndexes {
		_, err = idx.t.Flush(ctx, wg, writeToCloudstore) //note if the tree is empty, this will just return with no error and nothing done
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *TestTxBucket) Insert(key []byte, data []byte, index *TestBucketIndex) error {
	// assumes that index is correct.
	err := index.t.Insert(context.TODO(), key, data, 0, false)
	if err != nil {
		return fmt.Errorf("[txpayload:Insert] %s", err)
	}
	index.RootHash = index.t.ChunkHash()
	return nil
}

func (b *TestTxBucket) getIndex(indexName string) (idx *TestBucketIndex, err error) {
	//dprint("[txpayload:getIndex] indexName(%s)", indexName)
	for _, idx := range b.AVLIndexes {
		if strings.Compare(idx.IndexName, indexName) == 0 {
			return idx, nil
		}
	}
	return idx, fmt.Errorf("No index found")
}

func (index *TestBucketIndex) GetField(row map[string]interface{}) (field []byte, ok bool) {
	val, ok := row[index.IndexName]
	if !ok {
		return field, false
	}
	switch index.IndexType {
	case indexTypeText:
		field = []byte(val.(string))
	case indexTypeNumber:
		field = floatToByte(val.(float64)) // json always comes out in float64
	case indexTypeFloat:
		field = floatToByte(val.(float64))
	default:
		panic(fmt.Sprintf("unknown index type (%s)", index.IndexType))
	}

	return field, true
}

func (b *TestTxBucket) InsertRow(rec []byte) (err error) {
	//dprint("[txpayload:Insert] rec(%s)", rec)
	var row map[string]interface{}
	err = json.Unmarshal(rec, &row)
	if err != nil {
		return fmt.Errorf("[txpayload:Insert] %s", err)
	}
	//dprint("[txpayload:Insert] row(%+v)", row)

	// TODO: right now, non-primary indexes get their rows overwritten if there are duplicate keys
	for _, idx := range b.AVLIndexes {
		key, ok := idx.GetField(row)
		if !ok {
			return fmt.Errorf("[txpayload:Insert] index field(%s) not found in row(%+v)", idx.IndexName, row)
		}
		//dprint("inserting on key(%v)", idx.Translate(key))
		err = idx.t.Insert(context.TODO(), key, rec, 0, false) // TODO: storage bytes not 0
		if err != nil {
			return fmt.Errorf("[txpayload:Insert] %s", err)
		}
		idx.RootHash = idx.t.ChunkHash()
	}
	//find primary key index, insert in there with replacement
	// primaryIndex, err := b.getPrimaryIndex()
	// if err != nil {
	// 	return err
	// }
	// dprint("primaryindex(%+v)", primaryIndex)
	//
	// // TODO: get primary key field
	// primaryKey, ok := primaryIndex.getField(row)
	// if !ok {
	// 	return fmt.Errorf("Primary key not found %s", primaryIndex.IndexName)
	// }
	// dprint("primarykey(%s)", primaryKey)
	//
	// err = primaryIndex.t.Insert(primaryKey, rec, 0, false)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (b *TestTxBucket) Scan(fld string, start []byte, end []byte, asc bool) (recs [][]byte, p interface{}, err error) {
	//dprint("start(%v), end(%v)", start, end)
	index, err := b.getIndex(fld)
	if err != nil {
		return recs, p, fmt.Errorf("[txpayload:Scan] %s", err)
	}

	// TODO: return ScanProof
	//dprint("Scan:index: %+v\n", index)
	//dprint("start(%v), end(%v)", index.Translate(start), index.Translate(end))
	proof, _, values, err := index.t.Scan(start, end, 0) // 0 means no limit
	if err != nil {
		return recs, p, fmt.Errorf("[txpayload:Scan] %s", err)
	}
	return values, proof, nil
}

func (b *TestTxBucket) Get(fld string, key []byte) (val interface{}, found bool, deleted bool, p interface{}, err error) {
	index, err := b.getIndex(fld)
	if err != nil {
		return val, found, deleted, p, fmt.Errorf("[txpayload:Get] %s", err)
	}
	valbytes, found, deleted, proof, _, err := index.t.Get(key)
	if err != nil {
		return val, found, deleted, p, fmt.Errorf("[txpayload:Get] %s", err)
	}
	val = index.Translate(valbytes)
	return val, found, deleted, proof, nil
}

func (index *TestBucketIndex) Translate(val []byte) (field interface{}) {
	switch index.IndexType {
	case indexTypeText:
		field = string(val)
	case indexTypeNumber:
		//field = byteToInt64(val)
		ffield := byteToFloat(val)
		field = int64(ffield)
	case indexTypeFloat:
		field = byteToFloat(val)
	default:
		panic(fmt.Sprintf("unknown index type (%s)", index.IndexType))
	}
	return field
}

func (b *TestTxBucket) VerifyRange(fld string, proof interface{}) (bool, error) {
	// if AVLProof
	p := proof.(*AVLProof)
	index, err := b.getIndex(fld)
	if err != nil {
		return false, fmt.Errorf("[txpayload:Verify] %s", err)
	}
	return index.t.VerifyRangeProof(p), nil
}

func (b *TestTxBucket) Verify(fld string, key []byte, value []byte, proof interface{}) (bool, error) {
	// if AVLProof
	p := proof.(*AVLProof)
	index, err := b.getIndex(fld)
	if err != nil {
		return false, fmt.Errorf("[txpayload:Verify] %s", err)
	}
	return index.t.VerifyProof(key, value, p), nil
}

func TestAVLScanProof(t *testing.T) {
	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	var wg sync.WaitGroup

	var b TestTxBucket
	b.Name = "friends"
	b.Schema = "https://schema.org/Person.jsonld"

	indexName := new(TestBucketIndex)
	indexName.IndexName = "name"
	indexName.IndexType = indexTypeText
	indexName.Primary = true
	indexName.t = NewAVLTree(cs)

	indexAge := new(TestBucketIndex)
	indexAge.IndexName = "age"
	indexAge.IndexType = indexTypeNumber
	indexAge.t = NewAVLTree(cs)

	indexWeight := new(TestBucketIndex)
	indexWeight.IndexName = "weight"
	indexWeight.IndexType = indexTypeFloat
	indexWeight.t = NewAVLTree(cs)

	b.AVLIndexes = make([]*TestBucketIndex, 0)
	b.AVLIndexes = append(b.AVLIndexes, indexName)
	b.AVLIndexes = append(b.AVLIndexes, indexAge)
	b.AVLIndexes = append(b.AVLIndexes, indexWeight)

	tprint("bucket(%+v)", b)
	for _, bi := range b.AVLIndexes {
		tprint("index(%+v)", bi)
	}

	valuesmap := make(map[common.Hash]string) // chunkID, jsonstring of actual value. simulates

	numblocks := 10
	for block := 2; block < numblocks+2; block++ {
		nrecs := 100
		for r := 1; r <= nrecs; r++ {
			f := new(Person)
			f.Name = fmt.Sprintf("name%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d%d", block, r))[0:4]))
			//f.Name = fmt.Sprintf("name%d%d", block, r)
			f.Age = int64(rand.Intn(50)) + 1
			f.Weight = 100.1234 + float64(rand.Intn(10000))/100.0
			rec, err := json.Marshal(f)
			if err != nil {
				t.Fatal(err)
			}

			//err = b.InsertRow(jsonBytes)

			valhash := wolkcommon.Computehash(rec) // simulating a txhash
			tprint("blk(%v) rec(%d) (%s) (%x)", block, r, string(rec), valhash)
			valuesmap[common.BytesToHash(valhash)] = string(rec) // simulating a chunkstore

			for _, idx := range b.AVLIndexes {
				var row map[string]interface{}
				err = json.Unmarshal(rec, &row)
				require.NoError(t, err)
				key, ok := idx.GetField(row)
				require.True(t, ok)
				err := idx.t.Insert(context.TODO(), key, valhash, 0, false)
				require.NoError(t, err)
				idx.RootHash = idx.t.ChunkHash()
			}
			//tprint("blk(%v) rec(%d) (%s)", block, r, string(rec))
		}
		b.Flush(context.TODO(), &wg, true)
	}

	// tprint("test gets - no proof")
	// actualv, f, _, p, err := b.Get("name", []byte("name22"))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if !f {
	// 	tprint("why not found? k(name3)")
	// }
	// tprint("actualv(%v)", actualv)

	// scan name
	tprint("scan name: name3 - namea")
	recs, p, err := b.Scan("name", []byte("name3"), []byte("namea"), true)
	if err != nil {
		t.Fatalf("scan name err %v", err)
	}
	dumpRecordsAndProof(recs, valuesmap, p)
	verified, err := b.VerifyRange("name", p)
	if !verified {
		t.Fatal("not verified")
	}
	if err != nil {
		t.Fatal(err)
	}

	// scan age
	tprint("scan age: 30 - 50")
	recs, p, err = b.Scan("age", floatToByte(30), floatToByte(50), false)
	if err != nil {
		t.Fatalf("scan age err %v", err)
	}
	dumpRecordsAndProof(recs, valuesmap, p)
	verified = false
	verified, err = b.VerifyRange("age", p)
	if !verified {
		t.Fatal("not verified")
	}
	if err != nil {
		t.Fatal(err)
	}

	// scan weight
	tprint("scan wt: 110.0 - 140.5")
	recs, p, err = b.Scan("weight", floatToByte(110.0), floatToByte(130.5), true)
	if err != nil {
		t.Fatalf("scan weight err %v", err)
	}
	dumpRecordsAndProof(recs, valuesmap, p)
	verified = false
	verified, err = b.VerifyRange("weight", p)
	if !verified {
		t.Fatal("not verified")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestAVLScan(t *testing.T) {
	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	tree := NewAVLTree(cs)
	seed := rand.NewSource(42)
	r := rand.New(seed)
	var keys [][]byte
	var vals [][]byte
	numkeys := 10
	wg := new(sync.WaitGroup)

	// ints
	tprint("INTS")
	for j := 0; j < numkeys; j++ {
		i := r.Intn(1000)
		v := intToByte(uint64(i))
		k := intToByte(uint64(i))

		sb := randUInt64(5, 32)
		tprint("\n")
		tprint("(%v)---insert k(%v) v(%v)(%x)(%v) sb(%v)", j, byteToInt(k), byteToInt(v), v, v, sb)

		err := tree.Insert(context.TODO(), k, v, sb, false)
		if err != nil {
			//tprint("[NewKVChain] Insert ERR %v", err)
			t.Fatal(err)
		}
		keys = append(keys, k)
		vals = append(vals, v)
	}

	_, err := tree.Flush(context.TODO(), wg, true)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	hash := tree.ChunkHash()
	tree = NewAVLTree(cs)
	tree.Init(hash)
	tprint("hash: %x", hash)

	tprint("---get k(50-200)")
	proof, actualkeys, actualvals, err := tree.Scan(intToByte(50), intToByte(200), 0)
	if err != nil {
		t.Fatal(err)
	}
	jsonproof, _ := json.Marshal(proof)
	tprint("proof(%s)", jsonproof)
	for i, ak := range actualkeys {
		tprint("key(%v) val(%v)", byteToInt(ak), byteToInt(actualvals[i]))
	}
	//tree.PrintTree(nil, false, false)
	if !tree.VerifyRangeProof(proof.(*AVLProof)) {
		t.Fatal("proof not verified")
	}

	// strings
	tprint("\nSTRINGS")
	tree = NewAVLTree(cs)
	keys = nil
	vals = nil
	actualkeys = nil
	actualvals = nil
	var st []string
	st = append(st, "anteater")
	st = append(st, "bear")
	st = append(st, "beast")
	st = append(st, "cat")
	st = append(st, "dog")
	st = append(st, "elephant")

	for j := 0; j < 6; j++ {
		v := []byte(st[j])
		k := []byte(st[j])
		sb := randUInt64(5, 32)
		tprint("\n")
		tprint("(%v)---insert k(%v) v(%v)(%x)(%v) sb(%v)", j, string(k), string(v), v, v, sb)

		err := tree.Insert(context.TODO(), k, v, sb, false)
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, k)
		vals = append(vals, v)
	}
	_, err = tree.Flush(context.TODO(), wg, true)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	hash = tree.ChunkHash()
	tree = NewAVLTree(cs)
	tree.Init(hash)

	tprint("---get k(bear-dog)")
	proof = nil
	proof, actualkeys, actualvals, err = tree.Scan([]byte("bear"), []byte("dog"), 0)
	if err != nil {
		t.Fatal(err)
	}
	jsonproof, _ = json.Marshal(proof)
	tprint("proof(%s)", jsonproof)
	for i, ak := range actualkeys {
		tprint("key(%v)(%v) val(%v)(%v)", string(ak), ak, string(actualvals[i]), actualvals[i])
	}
	if !tree.VerifyRangeProof(proof.(*AVLProof)) {
		t.Fatal("proof not verified")
	}

	// floats
	tprint("\nFLOATS")
	tree = NewAVLTree(cs)
	keys = nil
	vals = nil
	actualkeys = nil
	actualvals = nil
	numkeys = 10
	for j := 0; j < numkeys; j++ {
		//i := rand.Float64() * float64(numkeys)
		v := floatToByte(float64(j) + 0.14)
		k := floatToByte(float64(j) + 0.14)

		sb := randUInt64(5, 32)
		tprint("\n")
		tprint("(%v)---insert k(%v) v(%v)(%x)(%v) sb(%v)", j, byteToFloat(k), byteToFloat(v), v, v, sb)

		err := tree.Insert(context.TODO(), k, v, sb, false)
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, k)
		vals = append(vals, v)
	}
	_, err = tree.Flush(context.TODO(), wg, true)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	hash = tree.ChunkHash()
	tree = NewAVLTree(cs)
	tree.Init(hash)

	tprint("---get k(0.14-7.14)")
	proof = nil
	proof, actualkeys, actualvals, err = tree.Scan(floatToByte(float64(0.14)), floatToByte(float64(7.14)), 0)
	if err != nil {
		t.Fatal(err)
	}
	jsonproof, _ = json.Marshal(proof)
	tprint("proof(%s)", jsonproof)
	for i, ak := range actualkeys {
		tprint("key(%v)(%v) val(%v)(%v)", byteToFloat(ak), ak, byteToFloat(actualvals[i]), actualvals[i])
	}
	if !tree.VerifyRangeProof(proof.(*AVLProof)) {
		t.Fatal("proof not verified")
	}

}

func TestAVLTypes(t *testing.T) {
	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	tree := NewAVLTree(cs)
	//seed := rand.NewSource(42)
	//r := rand.New(seed)
	testfloat := true
	testint := true
	teststring := true
	numkeys := 5

	if testfloat {
		//insert floats
		var fkeys []float64
		var fvals []float64
		for j := 0; j < numkeys; j++ {
			v := rand.Float64() * float64(100) //r := min + rand.Float64() * (max - min)
			//k := wolkcommon.Computehash(floatToByte(v))
			k := rand.Float64() * float64(100)
			sb := randUInt64(5, 32)
			tprint("\n")
			tprint("(%v)---insert k(%v) v(%v) sb(%v)", j, k, v, sb)

			err := tree.Insert(context.TODO(), floatToByte(k), floatToByte(v), sb, false)
			if err != nil {
				//tprint("[NewKVChain] Insert ERR %v", err)
				t.Fatal(err)
			}
			fkeys = append(fkeys, k)
			fvals = append(fvals, v)
		}
		//get with proof: floats
		for j := 0; j < len(fkeys); j++ {
			k := fkeys[j]
			expectedv := fvals[j]
			tprint("(%v)---get k(%v), expectedval(%v)", j, k, expectedv)
			actualv, found, _, proof, storageBytes, err := tree.GetWithProof(context.TODO(), floatToByte(k))
			if err != nil {
				t.Fatal(err)
			}
			if !found {
				t.Fatal("not found!")
			}
			if bytes.Compare(floatToByte(expectedv), actualv) == 0 {
				tprint("val matches(%v), sb(%v)", byteToFloat(actualv), storageBytes)
			} else {
				t.Fatalf("actualval(%v) does not match expectedval(%v)", actualv, expectedv)
			}
			// verify proof
			if tree.VerifyProof(floatToByte(k), actualv, proof.(*AVLProof)) {
				tprint("proof verified.")
			} else {
				t.Fatal("verify proof failed")
			}
			p, _ := json.Marshal(proof.(*AVLProof))
			tprint("float proof(%s)", string(p))
		}
	}

	if testint {
		// insert ints
		var keys []uint64
		var vals []uint64
		for j := 0; j < numkeys; j++ {
			v := randUInt64(-100, 100)
			k := randUInt64(-100, 100)
			sb := randUInt64(5, 32)
			tprint("\n")
			tprint("(%v)---insert k(%v) v(%v) sb(%v)", j, k, v, sb)

			err := tree.Insert(context.TODO(), intToByte(k), intToByte(v), sb, false)
			if err != nil {
				//tprint("[NewKVChain] Insert ERR %v", err)
				t.Fatal(err)
			}
			keys = append(keys, k)
			vals = append(vals, v)
		}
		//get with proof: ints
		for j := 0; j < len(keys); j++ {
			k := keys[j]
			expectedv := vals[j]
			tprint("(%v)---get k(%v), expectedval(%v)", j, k, expectedv)
			actualv, found, _, proof, storageBytes, err := tree.GetWithProof(context.TODO(), intToByte(k))
			if err != nil {
				t.Fatal(err)
			}
			if !found {
				t.Fatal("not found!")
			}
			if bytes.Compare(intToByte(expectedv), actualv) == 0 {
				tprint("val matches(%v), sb(%v)", byteToInt(actualv), storageBytes)
			} else {
				t.Fatalf("actualval(%v) does not match expectedval(%v)", actualv, expectedv)
			}
			// verify proof
			if tree.VerifyProof(intToByte(k), actualv, proof.(*AVLProof)) {
				tprint("proof verified.")
			} else {
				t.Fatal("verify proof failed")
			}
			p, _ := json.Marshal(proof.(*AVLProof))
			tprint("int proof(%s)", string(p))

		}

	}

	if teststring {
		// insert strings
		var keys []string
		var vals []string
		for j := 0; j < numkeys; j++ {
			v := randString(32)
			k := randString(32)
			sb := randUInt64(5, 32)
			tprint("\n")
			tprint("(%v)---insert k(%v) v(%v) sb(%v)", j, k, v, sb)

			err := tree.Insert(context.TODO(), []byte(k), []byte(v), sb, false)
			if err != nil {
				//tprint("[NewKVChain] Insert ERR %v", err)
				t.Fatal(err)
			}
			keys = append(keys, k)
			vals = append(vals, v)
		}
		//get with proof: strings
		for j := 0; j < len(keys); j++ {
			k := keys[j]
			expectedv := vals[j]
			tprint("(%v)---get k(%v), expectedval(%v)", j, k, expectedv)
			actualv, found, _, proof, storageBytes, err := tree.GetWithProof(context.TODO(), []byte(k))
			if err != nil {
				t.Fatal(err)
			}
			if !found {
				t.Fatal("not found!")
			}
			if bytes.Compare([]byte(expectedv), actualv) == 0 {
				tprint("val matches(%v), sb(%v)", string(actualv), storageBytes)
			} else {
				t.Fatalf("actualval(%v) does not match expectedval(%v)", actualv, expectedv)
			}
			// verify proof
			if tree.VerifyProof([]byte(k), actualv, proof.(*AVLProof)) {
				tprint("proof verified.")
			} else {
				t.Fatal("verify proof failed")
			}
			p, _ := json.Marshal(proof.(*AVLProof))
			tprint("string proof(%s)", string(p))

		}

	}

}

func randString(bytelength int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, bytelength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// TODO: test proofs with concurrency
func TestAVLProof(t *testing.T) {
	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	tree := NewAVLTree(cs)
	seed := rand.NewSource(42)
	r := rand.New(seed)
	var keys [][]byte
	var vals [][]byte
	numkeys := 100
	for j := 0; j < numkeys; j++ {
		i := r.Intn(numkeys)
		v := []byte(fmt.Sprintf("%d", i))
		k := wolkcommon.Computehash([]byte(v))
		//sb := randUInt64(5, 32)
		sb := uint64(32)
		tprint("\n")
		tprint("(%v)---insert k(%x) v(%x) sb(%v) seed(%v)", j, k, v, sb, i)

		err := tree.Insert(context.TODO(), k, v, sb, false)
		if err != nil {
			//tprint("[NewKVChain] Insert ERR %v", err)
			t.Fatal(err)
		}
		keys = append(keys, k)
		vals = append(vals, v)
	}
	//tprint("root: %s", tree.root.String())
	//tprint("cache: %s", tree.ndb.String())

	for j := 0; j < len(keys); j++ {
		//expectedv := []byte(fmt.Sprintf("%d", j))
		//k := wolkcommon.Computehash([]byte(expectedv))
		k := keys[j]
		expectedv := vals[j]
		tprint("(%v)---get k(%x), expectedval(%x)", j, k, expectedv)
		actualv, found, _, proof, storageBytes, err := tree.GetWithProof(context.TODO(), k)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatal("not found!")
		}
		if bytes.Compare(expectedv, actualv) == 0 {
			tprint("val matches(%x), sb(%v)", actualv, storageBytes)
		} else {
			t.Fatalf("actualval(%x) does not match expectedval(%x)", actualv, expectedv)
		}

		// verify proof
		if tree.VerifyProof(k, actualv, proof.(*AVLProof)) {
			tprint("proof verified.")
		} else {
			t.Fatal("verify proof failed")
		}
	}
	//tree.PrintTree(nil, true, false)
}

// for TestAVLConcurrency below
const (
	keysPerBlock_AVL = 100
	nblocks          = uint64(100)
	// storageType_AVL  = "leveldb"
)

func TestAVLConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	offset := 0
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace%d", offset))
	var pcs ChunkStore
	var err error
	pcs, closepcs := getWolkTestDB(t)
	defer closepcs()
	tree := NewAVLTree(pcs)

	// this gets recorded in blocks
	chunkHash := make(map[uint64]common.Hash)

	// make a bunch of keys
	//nblocks:= uint64(100)
	nkeys := nblocks * keysPerBlock_AVL
	key := make(map[uint64]common.Address)
	val := make(map[uint64]common.Hash)
	for i := uint64(0); i < nkeys; i++ {
		b := make([]byte, 20)
		rand.Read(b)
		key[i] = common.BytesToAddress(b)
		val[i] = common.BytesToHash(wolkcommon.Computehash(key[i].Bytes()))
	}
	maxkey := uint64(0)
	for block := uint64(0); block < nblocks; block++ {
		wg := new(sync.WaitGroup)
		st := time.Now()
		var failsMu sync.RWMutex
		fails := 0
		success := 0
		if block > 1 {
			for j := uint64(0); j < keysPerBlock_AVL; j++ {
				wg.Add(1)
			}
			// these reads are happening concurrently with the writes below
			go func(maxkey uint64) {
				for j := uint64(0); j < keysPerBlock_AVL; j++ {
					n := uint64(maxkey - keysPerBlock_AVL - j - 1)
					k := key[n]
					expectedv := val[n]
					v, found, _, _, _, err := tree.Get(k.Bytes())
					if err != nil {
						t.Fatalf("[avltree_test:TestAVLConcurrency] Get ERR %v\n", err)
					} else if !found {
						log.Error("[avltree_test:TestAVLConcurrency] fail", "n", n, "k", k)
						failsMu.Lock()
						fails++
						failsMu.Unlock()
						t.Fatalf("[avltree_test:TestAVLConcurrency] ABANDON n=%d k=%x\n", n, k)
					} else if bytes.Compare(v, expectedv.Bytes()) != 0 {
						t.Fatalf("[avltree_test:TestAVLConcurrency] Get(%x) incorrect value got %x but expected %x", k, v, expectedv)
					}
					failsMu.Lock()
					success++
					failsMu.Unlock()
					wg.Done()
				}
			}(maxkey)
		}
		// insert keysPerBlock keys
		for j := uint64(0); j < keysPerBlock_AVL; j++ {
			k := key[maxkey+j]
			v := val[maxkey+j]
			//tprint("blk(%v) keynum(%v) insert k(%x) v(%x)", block, j, k, v)
			err = tree.Insert(context.TODO(), k.Bytes(), v.Bytes(), uint64(32), false)
			if err != nil {
				log.Error(fmt.Sprintf("[avltree_test:TestAVLConcurrency] Insert ERR %v\n", err))
			}
		}
		//tprint("blk(%v) flushing after insert set", block)
		tree.Flush(context.TODO(), wg, true)
		wg.Wait()
		chunkHash[block] = tree.ChunkHash()
		failsMu.RLock()
		maxkey = maxkey + keysPerBlock_AVL
		log.Info("[avltree_test:TestAVLConcurrency] New block", "block", block, "chunkHash", tree.ChunkHash(), "success", success, "fails", fails, "maxkey", maxkey, "tm", time.Since(st))
		failsMu.RUnlock()
		tree = NewAVLTree(pcs)
		tree.Init(chunkHash[block])
	}
}

func TestAVLBlocks(t *testing.T) {
	// TODO: Flush & Load with maxSize < one tree

	prove := true // verify proofs, or not

	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	tree := NewAVLTree(cs)
	var hash common.Hash
	numkeys := 100
	numblks := 10
	for blk := 0; blk < numblks; blk++ {

		// inserts
		for j := blk * numkeys; j < blk*numkeys+numkeys; j++ {
			v := []byte(fmt.Sprintf("%d", j))
			k := wolkcommon.Computehash([]byte(v))
			//sb := randUInt64(5, 32)
			sb := uint64(32)
			tprint("blk(%v)itr(%v)-----insert k(%x) v(%x) sb(%v)", blk, j, k, v, sb)

			err := tree.Insert(context.TODO(), k, v, sb, false)
			if err != nil {
				//tprint("[NewKVChain] Insert ERR %v", err)
				t.Fatal(err)
			}

			// double check the key was inserted
			// tprint("getting key(%x) just inserted...", shortbytes(k))
			// actualv, found, _, _, _, err := tree.Get(k)
			// if err != nil {
			// 	t.Fatal(err)
			// }
			// if !found {
			// 	tree.PrintTree(nil, false, false)
			// 	t.Fatalf("not found key(%x)!", shortbytes(k))
			// }
			// if bytes.Compare(v, actualv) == 0 {
			// 	tprint("val matches(%x)", actualv)
			// } else {
			// 	t.Fatalf("actualval(%x) does not match expectedval(%x)", actualv, v)
			// }
		}

		// flush block
		tprint("blk(%v)---flushing", blk)
		var wg sync.WaitGroup
		tree.Flush(context.TODO(), &wg, true)
		wg.Wait()
		//tprint("\n")
		//tree.PrintTree(true, false)
		//tprint("\n")
		hash = tree.ChunkHash()
		tprint("tree hash(%x)", hash)

		// make a new tree, load the old one
		tree = NewAVLTree(cs)
		tprint("loading a new tree")
		tree.Init(hash)
	}

	// do random gets on flushed blocks
	for j := 0; j < numblks*numkeys; j++ {

		i := randUInt64(0, numblks*numkeys-1)
		expectedv := []byte(fmt.Sprintf("%d", i))
		k := wolkcommon.Computehash([]byte(expectedv))
		tprint("(%v)---get k(%x) made from (%v)", j, k, i)
		var actualv []byte
		var found bool
		var proof interface{}
		var storageBytes uint64
		var err error
		if prove {
			actualv, found, _, proof, storageBytes, err = tree.GetWithProof(context.TODO(), k)
		} else {
			actualv, found, _, _, storageBytes, err = tree.Get(k)
		}
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatal("not found!")
		}
		if bytes.Compare(expectedv, actualv) == 0 {
			tprint("val matches(%x), sb(%v)", actualv, storageBytes)
		} else {
			t.Fatalf("actualval(%x) does not match expectedval(%x)", actualv, expectedv)
		}
		if prove && !tree.VerifyProof(k, actualv, proof.(*AVLProof)) {
			t.Fatal("proof not verified")
		}
	}
	tprint("passed")

}

func TestAVLFlushLoad(t *testing.T) {
	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	tree := NewAVLTree(cs)
	numkeys := 30
	// inserts
	for j := 0; j < numkeys; j++ {
		v := []byte(fmt.Sprintf("%d", j))
		k := wolkcommon.Computehash([]byte(v))
		//sb := randUInt64(5, 32)
		sb := uint64(32)
		tprint("(%v)-----insert k(%x) v(%x) sb(%v)", j, k, v, sb)

		err := tree.Insert(context.TODO(), k, v, sb, false)
		if err != nil {
			//tprint("[NewKVChain] Insert ERR %v", err)
			t.Fatal(err)
		}
		//tprint("\n")
		//tprint("nodecache: %s", tree.ndb.String())
		//tree.PrintTree(false)
		//tprint("\n")
	}

	// flush block
	var wg sync.WaitGroup
	tree.Flush(context.TODO(), &wg, true)
	wg.Wait()
	//tprint("\n")
	//tree.PrintTree(true, false)
	//tprint("\n")
	hash := tree.ChunkHash()
	tprint("tree hash(%x)", hash)

	// make a new tree, load the old one
	newTree := NewAVLTree(cs)
	newTree.Init(hash)

	// do some gets on flushed block
	for j := 0; j < numkeys; j++ {
		expectedv := []byte(fmt.Sprintf("%d", j))
		k := wolkcommon.Computehash([]byte(expectedv))
		tprint("(%v)---get k(%x)", j, k)
		actualv, found, _, _, storageBytes, err := newTree.Get(k)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatal("not found!")
		}
		if bytes.Compare(expectedv, actualv) == 0 {
			tprint("val matches(%x), sb(%v)", actualv, storageBytes)
		} else {
			t.Fatalf("actualval(%x) does not match expectedval(%x)", actualv, expectedv)
		}
	}
}

func TestAVLInsert(t *testing.T) {
	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	tree := NewAVLTree(cs)
	seed := rand.NewSource(42)
	r := rand.New(seed)
	var keys [][]byte
	var vals [][]byte
	numkeys := 10000
	for j := 0; j < numkeys; j++ {
		i := r.Intn(numkeys)
		v := []byte(fmt.Sprintf("%d", i))
		k := wolkcommon.Computehash([]byte(v))
		//sb := randUInt64(5, 32)
		sb := uint64(32)
		tprint("(%v)---insert k(%x) v(%x) sb(%v) seed(%v)", j, k, v, sb, i)

		err := tree.Insert(context.TODO(), k, v, sb, false)
		if err != nil {
			//tprint("[NewKVChain] Insert ERR %v", err)
			t.Fatal(err)
		}
		//tprint("\n")
		//tree.PrintTree(false)
		//tprint("\n")
		tprint("(%v) getting k(%x) just inserted", j, k)
		actualv, found, _, _, storageBytes, err := tree.Get(k)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatal("not found!")
		}
		if bytes.Compare(v, actualv) == 0 {
			tprint("val matches(%x), sb(%v)", actualv, storageBytes)
		} else {
			t.Fatalf("actualval(%x) does not match expectedval(%x)", actualv, v)
		}
		keys = append(keys, k)
		vals = append(vals, v)
	}
	//tprint("root: %s", tree.root.String())
	//tprint("cache: %s", tree.ndb.String())
	for j := 0; j < len(keys); j++ {
		//expectedv := []byte(fmt.Sprintf("%d", j))
		//k := wolkcommon.Computehash([]byte(expectedv))
		k := keys[j]
		expectedv := vals[j]
		tprint("(%v)---get k(%x), expectedval(%x)", j, k, expectedv)
		actualv, found, _, _, storageBytes, err := tree.Get(k)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatal("not found!")
		}
		if bytes.Compare(expectedv, actualv) == 0 {
			tprint("val matches(%x), sb(%v)", actualv, storageBytes)
		} else {
			t.Fatalf("actualval(%x) does not match expectedval(%x)", actualv, expectedv)
		}
	}
	tree.PrintTree(nil, true, false)
}

func TestAVLSetGet(t *testing.T) {
	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	tree := NewAVLTree(cs)

	child1 := NewAVLNode(wolkcommon.Computehash([]byte("a")), []byte("a"), uint64(32))
	child2 := NewAVLNode(wolkcommon.Computehash([]byte("b")), []byte("b"), uint64(5))
	child3 := NewAVLNode(wolkcommon.Computehash([]byte("c")), []byte("c"), uint64(8))

	tprint("child1: %s", child1.String())
	tprint("child2: %s", child2.String())
	tprint("child3: %s", child3.String())

	tree.ndb.SetNode(child1)
	tree.ndb.SetNode(child2)
	child2.LeftID = &child1.id
	tree.ndb.SetNode(child3) // make maxSize = 2
	child2.RightID = &child3.id
	child2.calcHeightAndSize(tree.ndb)

	tprint("after child2: %s", child2.String())
	tprint("after child1: %s", child1.String())
	tprint("after child3: %s", child3.String())
	tprint("nodecache: %s", tree.ndb.String())

	tprint("root balance: %v", child2.calcBalance(tree.ndb))

	actual1 := tree.ndb.GetNode(child1.id, EMPTYBYTES)
	tprint("after get 1: %s", actual1.String())
	if actual1.Hash() != child1.Hash() {
		t.Fatal("actua1 != expected1")
	}

	actual3 := tree.ndb.GetNode(child3.id, EMPTYBYTES)
	tprint("after get 3: %s", actual3.String())
	if actual3.Hash() != child3.Hash() {
		t.Fatal("actual3 != expected3")
	}
	tree.printNode(child2, 0, false, false)
}

func TestAVLEmptyTree(t *testing.T) {
	cs, closecs := getWolkTestDB(t) // this is levelDB
	defer closecs()
	tree := NewAVLTree(cs)

	tprint("empty tree's chunkhash(%x)", tree.ChunkHash())
}

func TestAVLHashCheck(t *testing.T) {
	hleft, _ := hex.DecodeString("81ebc13d72b1d2f37170f1fc521e7835dddd3a2989722af89ce9f44a590990b6")
	hright, _ :=
		hex.DecodeString("84e491e2e849c7b5c7aa2e0eb4fedfdb43f817500ad07cb96a921794ff884f8d")

	tprint("hleft(%x) hright(%x)", hleft, hright)
	hash := wolkcommon.Computehash(hleft, hright)
	tprint("res(%x)", hash)

}

func TestAVLid(t *testing.T) {
	key := EMPTYBYTES
	tprint("key(%v)", key)
	tprint("leaf(%x)", makeID(true, key.Bytes()))
	tprint("inner(%x)", makeID(false, key.Bytes()))

	for j := 0; j < 5; j++ {
		v := []byte(fmt.Sprintf("%d", j))
		k := wolkcommon.Computehash([]byte(v))
		tprint("key(%v)", k)
		tprint("leaf(%x)", makeID(true, k))
		tprint("inner(%x)", makeID(false, k))
		tprint("\n")

		n := NewAVLNode(k, v, uint64(32))
		tprint("node(%s)", n.String())
		if n.RightID == nil {
			tprint("new hash is nil")
		}
	}

}

func TestShortBytes(t *testing.T) {
	// set frontBytes to false or true
	writesPerBlock := 5
	maxkeys := 20
	for i := 0; i < writesPerBlock; i++ {
		k := []byte(fmt.Sprintf("K%d", maxkeys+i))
		k = NameToAddress(string(k)).Bytes()
		v := wolkcommon.Computehash(k)
		fmt.Printf("actual: k(%x) v(%x)\n", k, v)
		fmt.Printf("padded: k(%x) v(%x)\n", padBytes(k, 32), padBytes(v, 32))
		fmt.Printf("short: k(%x) v(%x)\n", shortbytes(k), shortbytes(v))
	}
}
