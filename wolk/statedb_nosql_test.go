package wolk

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	//	"github.com/stretchr/testify/assert"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

// make a 'fake statedb' - never increment the block
func singleNodeStateDB(t *testing.T, cs ChunkStore, datadir string) (statedb *StateDB, owner string, privkey *ecdsa.PrivateKey) {
	var err error
	require := require.New(t)
	nAccounts := MAX_VALIDATORS
	genesisFileName := "genesis-1.json"
	privateKeys := make([]*crypto.PrivateKey, nAccounts)
	pubKeys := make([]string, nAccounts)
	privateKeysECDSA := make([]*ecdsa.PrivateKey, nAccounts)
	pubKeysECDSA := make([]string, nAccounts)
	accounts := make(map[common.Address]Account)
	addr := make([]common.Address, nAccounts)
	owners := make([]string, nAccounts)

	for i := 0; i < nAccounts; i++ {

		//r, _ := generateRandomBytes(32)
		owner := fmt.Sprintf("owner%d", i)
		owners[i] = owner
		keyString := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(owner)))
		privateKeys[i], err = crypto.HexToPrivateKey(keyString)
		require.NoError(err)
		// TODO:
		// this wont work for p2p peering, unless you make a change in the registry[i].PubKey line ( which is Edwards based ) --
		// privateKeysECDSA[i], err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		// this is necessary for p2p to work since the public keys in the genesis file match up with these determinstic keys --
		privateKeysECDSA[i], err = ethcrypto.HexToECDSA(keyString)
		require.NoError(err)
		pubKeys[i] = fmt.Sprintf("%x", privateKeys[i].PublicKey())
		pubKeysECDSA[i] = fmt.Sprintf("%x", ethcrypto.CompressPubkey(&privateKeysECDSA[i].PublicKey))
		//addr[i] = crypto.PubkeyToAddress(privateKeys[i].PublicKey())
		addr[i] = crypto.GetECDSAAddress(privateKeysECDSA[i]) //MUST FIX
		accounts[addr[i]] = Account{Balance: uint64(100000 + i)}

	}
	HTTPPort := uint16(81)
	registry := make([]SerializedNode, MAX_VALIDATORS)
	for i := 0; i < MAX_VALIDATORS; i++ {
		dns := fmt.Sprintf("c%d.wolk.com", i)
		registry[i] = SerializedNode{
			Address:     addr[i],
			PubKey:      pubKeys[i],
			ValueInt:    uint64(10000 + i),
			ValueExt:    uint64(0),
			StorageIP:   dns,
			ConsensusIP: dns,
			Region:      1,
			HTTPPort:    HTTPPort,
		}
	}
	networkID := int(1234)

	// for id, acct := range accounts {
	// 	tprint("id(%x) acct(%+v)", id, acct)
	// }
	err = CreateGenesisFile(networkID, genesisFileName, accounts, registry)
	require.NoError(err)
	genesisConfig, err := LoadGenesisFile(genesisFileName)
	require.NoError(err)
	// for id, acct := range genesisConfig.Accounts {
	// 	tprint("genesis config id(%x) acct(%+v)", id, acct)
	// }

	// create nodes
	// registeredNode := registry[i]
	// str := fmt.Sprintf("%d", int32(time.Now().Unix()))
	// datadir := fmt.Sprintf("/tmp/storage%s/datadir", str)
	cfg := &cloud.DefaultConfig
	cfg.GenesisFile = genesisFileName
	cfg.DataDir = datadir
	cfg.HTTPPort = int(HTTPPort)
	cfg.NodeType = "storage"
	cfg.ConsensusIdx = 3
	cfg.ConsensusAlgorithm = "poa"
	cfg.Preemptive = false
	cfg.Provider = "leveldb"
	//cfg.Address = address
	cfg.OperatorKey = privateKeys[0]
	cfg.OperatorECDSAKey = privateKeysECDSA[0]

	//_, genesisConfig = genConfigs(t) // (t, "testowner")
	_, statedb, err = CreateGenesis(cs, genesisConfig, true)
	require.NoError(err)

	return statedb, owners[0], privateKeysECDSA[0]
}

// make a txn for SetKey/GetKey singleNode testing. modeled after NewTransactionImplicit
// note: still need to use NewTransaction (official) to SetName, etc.
// func newTransaction(t *testing.T, r *http.Request, payload []byte) (tx *Transaction) {
// 	tx = new(Transaction)
// 	tx.Method = []byte(r.Method)
// 	tx.Path = []byte(r.URL.Path)
// 	//tx.Sig = common.FromHex(r.Header.Get("Sig"))
// 	//tx.Signer = []byte(r.Header.Get("Requester"))
// 	tx.Payload = payload
// 	tprint("new tx payload(%x)(%s)", payload, payload)
// 	return tx
// }

func TestStateDB_NoSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	require := require.New(t)
	var err error

	cs, closecs := getWolkTestDB(t) // local leveldb
	defer closecs()
	datadir := fmt.Sprintf("/tmp/testleveldb/") // same as local leveldb
	sdb, owner, privKey := singleNodeStateDB(t, cs, datadir)
	//tprint("sdb(%+v)", sdb)

	ctx := context.TODO()
	var wg sync.WaitGroup

	// make owner/name
	txbucket := NewTxBucket(owner, BucketNoSQL, nil)
	tx, err := NewTransaction(privKey, http.MethodPost, owner, txbucket)
	require.NoError(err)
	tprint("owner/name txn (%+v)", tx)
	err = sdb.ApplyTransaction(ctx, tx, "")
	require.NoError(err)
	err = sdb.Storage.FlushQueue()
	require.NoError(err)
	tprint("owner/name set")

	owneraddr, ok, _, err := sdb.GetName(ctx, owner, false)
	require.NoError(err)
	require.True(ok)
	tprint("owneraddr(%x) gotten", owneraddr)

	// set Bucket with AVLIndexes
	b := new(TxBucket)
	b.Name = "friends"
	b.Schema = "https://schema.org/Person.jsonld"

	indexName := new(BucketIndex)
	indexName.IndexName = "name"
	indexName.IndexType = indexTypeText
	indexName.Primary = true

	indexAge := new(BucketIndex)
	indexAge.IndexName = "age"
	indexAge.IndexType = indexTypeNumber

	indexWeight := new(BucketIndex)
	indexWeight.IndexName = "weight"
	indexWeight.IndexType = indexTypeFloat

	options := new(RequestOptions)
	options.Indexes = make([]*BucketIndex, 0)
	options.Indexes = append(options.Indexes, indexName)
	options.Indexes = append(options.Indexes, indexAge)
	options.Indexes = append(options.Indexes, indexWeight)
	options.PrimaryKey = indexName.IndexName
	tprint("options(%+v)", options)

	txbucket = NewTxBucket(b.Name, BucketNoSQL, options)
	tprint("tx bucket is (%+v)", txbucket)
	tx, err = NewTransaction(privKey, http.MethodPost, path.Join(owner, b.Name), txbucket)
	require.NoError(err)
	tprint("tx payload is (%s)", tx.Payload)
	err = sdb.ApplyTransaction(ctx, tx, "")
	require.NoError(err)

	tprint("random commit after buckets created w/ empty trees")
	sdb.mu.Lock()
	err = sdb.CommitNoSQL(ctx, &wg, true) //- negative test
	require.NoError(err)
	err = sdb.Flush(ctx, &wg, true)
	require.NoError(err)
	sdb.mu.Unlock()

	// double checking getting the bucket tx
	txHash := wolkcommon.Computehash(tx.Bytes())
	txActual, ok, err := sdb.Storage.GetChunk(ctx, txHash)
	require.NoError(err)
	require.True(ok)
	txbActual, err := DecodeRLPTransaction(txActual)
	require.NoError(err)
	require.Equal(txbActual, tx, "should be the same")

	// remove cached trees
	for i, _ := range sdb.nosql.indexedCollections {
		delete(sdb.nosql.indexedCollections, i)
	}

	// double check getting a txhash of an empty tree
	// icHash := IndexedCollectionHash("owner0", "friends", "name")
	// v, found, _, _, _, err := sdb.keyStorage.Get(ctx, icHash.Bytes(), false)
	// require.NoError(err)
	// require.True(found)
	// require.Equal(v, getEmptySMTChunkHash(), "should be the same")

	// insert keys and values
	tprint("insert keys and values")
	nrecs := 5
	for rec := 1; rec <= nrecs; rec++ {
		block := 2

		// make a new row
		person := new(Person)
		person.Name = fmt.Sprintf("name%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d%d", block, rec))[0:4]))
		person.Age = int64(rand.Intn(50)) + 1
		person.Weight = 100.1234 + float64(rand.Intn(10000))/100.0
		jsonBytes, errJ := json.Marshal(person)
		require.NoError(errJ)
		tprint("blk(%v) rec(%d) (%s)", block, rec, string(jsonBytes))

		// make the individual key transations for the row
		txns, errT := MakeIndexedTransactions(privKey, http.MethodPost, owner, txbucket, jsonBytes)
		require.NoError(errT)

		// apply each insert transaction for each index
		for _, tx := range txns {
			err = sdb.ApplyTransaction(ctx, tx, "")
			require.NoError(err)
		}

	}

	sdb.mu.Lock()
	err = sdb.CommitNoSQL(ctx, &wg, true)
	require.NoError(err)
	err = sdb.Flush(ctx, &wg, true)
	require.NoError(err)
	sdb.mu.Unlock()

	// do a simple get indexed key
	idx, err := txbucket.GetIndex("name")
	require.NoError(err)
	keyTxHash, ok, _, _, err := sdb.GetIndexedKey(ctx, owner, "friends", "name", []byte("namee561fd638f2b9ca3b112290bfcf06cde0b90766d665773435fce24df0d357da3"), false)
	require.NoError(err)
	require.True(ok)
	txhashToRecord(ctx, t, keyTxHash, idx, sdb)

	// scan primary index "name"
	tprint("scan name: name3 - namea")
	idx, err = txbucket.GetIndex("name")
	require.NoError(err)
	txhashes, ok, _, err := sdb.ScanIndexedCollection(ctx, owner, "friends", "name", []byte("name3"), []byte("namea"), 0, false)
	require.NoError(err)
	require.True(ok)
	require.NotEqual(len(txhashes), 0)
	for _, txh := range txhashes {
		txhashToRecord(ctx, t, txh, idx, sdb)
	}

	// scan age
	tprint("scan age: 30 - 50")
	idx, err = txbucket.GetIndex("age")
	require.NoError(err)
	txhashes, ok, _, err = sdb.ScanIndexedCollection(ctx, owner, "friends", "age", floatToByte(30), floatToByte(50), 0, false)
	require.NoError(err)
	require.True(ok)
	require.NotEqual(len(txhashes), 0)
	for _, txh := range txhashes {
		txhashToRecord(ctx, t, txh, idx, sdb)
	}

	// scan weight
	tprint("scan wt: 110.0 - 140.5")
	idx, err = txbucket.GetIndex("weight")
	require.NoError(err)
	txhashes, ok, _, err = sdb.ScanIndexedCollection(ctx, owner, "friends", "weight", floatToByte(110.0), floatToByte(130.5), 0, false)
	require.NoError(err)
	require.True(ok)
	require.NotEqual(len(txhashes), 0)
	for _, txh := range txhashes {
		txhashToRecord(ctx, t, txh, idx, sdb)
	}

	// TODO: proof & verify
	//dumpRecordsAndProof(recs, nil, p)
	//verified, err := b.VerifyRange("name", p)

}

func txhashToRecord(ctx context.Context, t *testing.T, txh common.Hash, idx *BucketIndex, sdb *StateDB) string {

	txa, ok, err := sdb.Storage.GetChunk(ctx, txh.Bytes())
	require.NoError(t, err)
	require.True(t, ok)
	txaKey, err := DecodeRLPTransaction(txa)
	require.NoError(t, err)
	txp, err := txaKey.GetTxKey()
	require.NoError(t, err)
	rec, ok, err := sdb.Storage.GetChunk(ctx, txp.ValHash.Bytes())
	require.NoError(t, err)
	require.True(t, ok)
	tprint("key(%v) val(%s)", idx.Translate(txp.Key), rec)
	return string(rec)

}

func TestIndexedKeyToAddress(t *testing.T) {
	nrecs := 5
	for rec := 1; rec <= nrecs; rec++ {
		block := 2
		// make a new row
		person := new(Person)
		person.Name = fmt.Sprintf("name%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d%d", block, rec))[0:4]))
		person.Age = int64(rand.Intn(50)) + 1
		person.Weight = 100.1234 + float64(rand.Intn(10000))/100.0
		//jsonBytes, err := json.Marshal(person)
		//require.NoError(err)
		tprint("key(%s) ktoa(%x)", person.Name, IndexedKeyToAddress([]byte(person.Name)))
	}
	tprint("\n")
	tprint("key(%s) ktoa(%x)", "name3", IndexedKeyToAddress([]byte("name3")))
	tprint("key(%s) ktoa(%x)", "namea", IndexedKeyToAddress([]byte("namea")))

	tprint("\n")
	flt := float64(12)
	bte := floatToByte(flt)
	tprint("len of flt(%v)", len(floatToByte(flt)))
	tprint("orig(%v) key(%v) ktoa(%x)", flt, bte, IndexedKeyToAddress(bte))
	flt1 := float64(13)
	bte = floatToByte(flt1)
	tprint("len of flt(%v)", len(floatToByte(flt)))
	tprint("orig(%v) key(%v) ktoa(%x)", flt1, bte, IndexedKeyToAddress(bte))
	flt2 := float64(10000)
	bte = floatToByte(flt2)
	tprint("len of flt(%v)", len(floatToByte(flt)))
	tprint("orig(%v) key(%v) ktoa(%x)", flt2, bte, IndexedKeyToAddress(bte))
	if bytes.Compare(IndexedKeyToAddress(floatToByte(flt2)).Bytes(), IndexedKeyToAddress(floatToByte(flt1)).Bytes()) == 1 && bytes.Compare(IndexedKeyToAddress(floatToByte(flt1)).Bytes(), IndexedKeyToAddress(floatToByte(flt)).Bytes()) == 1 {
	} else {
		t.Fatal("nope")
	}

	tprint("\n")
	it := int64(12)
	itbte := int64ToByte(it)
	tprint("len of int(%v)", len(int64ToByte(it)))
	tprint("orig(%v) key(%v) ktoa(%x)", it, itbte, IndexedKeyToAddress(itbte))
	it1 := int64(13)
	itbte = int64ToByte(it1)
	tprint("len of int(%v)", len(int64ToByte(it1)))
	tprint("orig(%v) key(%v) ktoa(%x)", it1, itbte, IndexedKeyToAddress(itbte))
	it2 := int64(10000)
	itbte = int64ToByte(it2)
	tprint("len of int(%v)", len(int64ToByte(it2)))
	tprint("orig(%v) key(%v) ktoa(%x)", it2, itbte, IndexedKeyToAddress(itbte))
	if bytes.Compare(IndexedKeyToAddress(int64ToByte(it2)).Bytes(), IndexedKeyToAddress(int64ToByte(it1)).Bytes()) == 1 && bytes.Compare(IndexedKeyToAddress(int64ToByte(it1)).Bytes(), IndexedKeyToAddress(int64ToByte(it)).Bytes()) == 1 {
	} else {
		t.Fatal("nope")
	}
}

/*
func TestStateDB_NoSQL(t *testing.T) {
	setName(t, owner, privateKey, wolkStore)

	rawKey := []byte("key_a")
	rawVal := "val_a"
	rawKeyHash := deep.Keccak256(rawKey)
	testKey := deep.BytesToUint64(rawKeyHash)
	testVal := common.BytesToHash(deep.Keccak256([]byte(rawVal)))

	//TODO: SetKeyVal using tx
	log.Info("Setting StateDB KeyVal", "testKey", testKey)
	testTxHash := common.BytesToHash(deep.Keccak256(testVal.Bytes()))
	sobj, err := stateDb.getOrCreateStateObject(testKey)
	if err != nil {
		t.Fatalf("getOrCreateStateObject error: +%v", err)
	}

	sobj.SetVal(testVal, testTxHash)
	stateDb.Append(testKey, sobj, testTxHash)

	_, _ = stateDb.Commit(ChunkStore, 1, TestHeader.Hash(), TestHeader.ParentAnchorHash)
	err = ChunkStore.Flush()
	if err != nil {
		t.Fatalf("Flusing error: +%v", err)
	}

	log.Info("Getting StateDB KeyVal", "testKey", testKey)
	sobj, retrievedProof, getKeyValErr := stateDb.GetKey(testKey)
	retrievedVal := sobj.Val()
	if getKeyValErr != nil {
		t.Fatalf("GetKeyVal Failure - Error [%+v] Proof [%+v]", getKeyValErr, retrievedProof)
	} else {
		log.Info("GetKeyVal", "testKey", testKey, "retrievedVal", retrievedVal, "Proof", retrievedProof)
	}

	if testVal != retrievedVal {
		t.Fatalf("INITIAL: Val mismatch between what was set and what is in stateobject - testVal [%+v] retrieved statedb val [%+v]", testVal, retrievedVal)
	}
	alteredRawVal := rawVal + "_altered"
	alteredTestVal := common.BytesToHash(deep.Keccak256([]byte(alteredRawVal)))
	alteredTestTxHash := alteredTestVal

	sobj, alterr := stateDb.getOrCreateStateObject(testKey)
	if alterr != nil {
		t.Fatalf("getOrCreateStateObject error: +%v", alterr)
	}
	sobj.SetVal(alteredTestVal, alteredTestTxHash)
	stateDb.Append(testKey, sobj, alteredTestTxHash)

	_, _ = stateDb.Commit(ChunkStore, 1, TestHeader.Hash(), TestHeader.ParentAnchorHash)
	err = ChunkStore.Flush()
	if err != nil {
		t.Fatalf("Flusing error: +%v", err)
	}

	sobj, retrievedProof, getKeyValErr = stateDb.GetKey(testKey)
	retrievedVal = sobj.Val()
	if getKeyValErr != nil {
		t.Fatalf("GetKeyVal Failure - Error [%+v] Proof [%+v]", getKeyValErr, retrievedProof)
	}

	if alteredTestVal != retrievedVal {
		t.Fatalf("ALTERED: Val mismatch between what was set and what is in stateobject - testVal [%+v] retrieved statedb val [%+v]", alteredTestVal, retrievedVal)
	}
}
*/
