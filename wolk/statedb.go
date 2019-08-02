// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

const (
	minBlocksForStorageClaims  = 10
	parallelApplyTransactions  = false
	parallelVerifyTransactions = false
	parallelSMTFlush           = false // stage 1 3 4
	parallelCommitNoSQL        = false // stage 1 3 4
)

type revision struct {
	id           int
	journalIndex int
}

// StateDB holds the block and all key objects and SMTs
type StateDB struct {
	Storage ChunkStore

	parentHash common.Hash
	blockHash  common.Hash
	// operatorKey *crypto.PrivateKey

	// Journal of state modifications
	journal        *journal
	validRevisions []revision
	nextRevisionID int
	mu             sync.RWMutex

	// Cloudstore account objects
	accountObjects    map[common.Address]*accountObject
	muAccountObjects  sync.RWMutex
	registryObjects   map[uint64]*registryObject
	muRegistryObjects sync.RWMutex
	checkObjects      map[common.Hash]*checkObject
	muCheckObjects    sync.RWMutex
	storageUsage      map[common.Address]uint64
	muStorageUsage    sync.RWMutex
	storageClaims     map[common.Address]uint64
	muStorageClaims   sync.RWMutex

	txCache   map[common.Hash]*Transaction
	muTxCache sync.RWMutex

	// NoSQL objects
	nosql *NoSQLStateDB

	// SQL objects
	sql *SQLStateDB

	// Name objects
	names *NamesStateDB

	accountStorage  *SparseMerkleTree
	registryStorage *SparseMerkleTree
	checkStorage    *SparseMerkleTree
	keyStorage      *SparseMerkleTree
	keyStorageAVL   *AVLTree
	productsStorage *SparseMerkleTree
	nameStorage     *SparseMerkleTree
	block           *Block

	// list of txs that succeeded in ApplyTransaction, which get packaged in a block
	committedTxes []*Transaction
	attemptedTxs  map[common.Hash]int8
	txSize        uint64
}

func New(cs ChunkStore) (statedb *StateDB) {
	statedb = new(StateDB)
	statedb.Storage = cs

	// initialize statedb with SMT structures
	statedb.accountStorage = NewSparseMerkleTree(NumBitsAddress, cs)
	statedb.registryStorage = NewSparseMerkleTree(NumBitsAddress, cs)
	statedb.checkStorage = NewSparseMerkleTree(NumBitsAddress, cs)
	statedb.keyStorage = NewSparseMerkleTree(NumBitsAddress, cs)
	statedb.productsStorage = NewSparseMerkleTree(NumBitsAddress, cs)
	statedb.nameStorage = NewSparseMerkleTree(NumBitsAddress, cs)

	//statedb.operatorKey = operatorKey
	statedb.accountObjects = make(map[common.Address]*accountObject)
	statedb.registryObjects = make(map[uint64]*registryObject)
	statedb.checkObjects = make(map[common.Hash]*checkObject)
	statedb.storageUsage = make(map[common.Address]uint64)
	statedb.storageClaims = make(map[common.Address]uint64)
	statedb.journal = newJournal()
	statedb.attemptedTxs = make(map[common.Hash]int8)

	statedb.txCache = make(map[common.Hash]*Transaction)

	// NoSQL
	statedb.nosql = new(NoSQLStateDB)
	statedb.nosql.systemProof = make(map[common.Address]*Proof)
	statedb.nosql.collectionProof = make(map[common.Address]*Proof)
	statedb.nosql.collections = make(map[common.Address]*SparseMerkleTree)
	statedb.nosql.indexedCollections = make(map[common.Address]*AVLTree)
	statedb.nosql.indexedCollectionProof = make(map[common.Address]*Proof)
	statedb.nosql.storageBytes = make(map[common.Address]uint64)
	statedb.nosql.collectionOwner = make(map[common.Address]string)

	// SQL
	statedb.sql = new(SQLStateDB)
	statedb.sql.stateObjects = make(map[common.Hash]*stateObject)
	statedb.sql.Owners = make(map[string]*SQLOwner)
	statedb.sql.Databases = make(map[string]*SQLDatabase)
	statedb.sql.Tables = make(map[string]*SQLTable)

	// Names
	statedb.names = new(NamesStateDB)
	statedb.block = new(Block)
	return statedb
}

func (statedb *StateDB) String() string {
	if statedb != nil {
		/*
			return fmt.Sprintf("{\"accountRoot\":\"%x\", \"registryRoot\":\"%x\", \"checkRoot\":\"%x\",\"keyRoot\":\"%x\", \"productsRoot\":\"%x\", \"nameRoot\":\"%x\", \"parentHash\":\"%x\"\"}",
				statedb.accountStorage.ChunkHash(), statedb.registryStorage.ChunkHash(), statedb.checkStorage.ChunkHash(), statedb.keyStorage.ChunkHash(), statedb.productsStorage.ChunkHash(), statedb.nameStorage.ChunkHash(), statedb.parentHash)
		*/
		return fmt.Sprintf("[accountRoot %x | keyRoot %x | nameRoot %x | parent %x]",
			statedb.accountStorage.ChunkHash(), statedb.keyStorage.ChunkHash(), statedb.nameStorage.ChunkHash(), statedb.parentHash)
	} else {
		return fmt.Sprint("{}")
	}
}

// add error
func newStateDB(ctx context.Context, cs ChunkStore, blockHash common.Hash) (statedb *StateDB, err error) {

	statedb = New(cs)
	block := new(Block)
	b := make([]byte, 32)
	zeroHash := common.BytesToHash(b)
	if bytes.Compare(blockHash.Bytes(), zeroHash.Bytes()) == 0 {
		// genesis block situation,
	} else {
		encodedBlock, ok, err := statedb.Storage.GetChunk(ctx, blockHash.Bytes())
		if err != nil {
			log.Error("[statedb:newStateDB] blockHash GetChunk ERR", "blockHash", blockHash.Hex(), "err", err)
			// MUST RETURN error here!!!  This has timeout
			return statedb, err
		} else if !ok {
			log.Error("[statedb:newStateDB] blockHash GetChunk NOT OK", "blockHash", blockHash.Hex())
			// MUST RETURN error here!!!  This has timeout
			return statedb, fmt.Errorf("[statedb:newStateDB] Chunk not found")
		}
		if err := rlp.Decode(bytes.NewReader(encodedBlock), block); err != nil {
			log.Error("[statedb:newStateDB] Decode Invalid block RLP", "v", blockHash.Hex(), "RLP", encodedBlock, "err", err)
		}
		statedb.block = block
		statedb.parentHash = block.ParentHash
		statedb.SetRootFromBlock(block)
	}
	return statedb, nil
}

// SetRootFromBlock initialize the SMT roots with the block state
func (statedb *StateDB) SetRootFromBlock(b *Block) {
	statedb.accountStorage.Init(b.AccountRoot)
	statedb.registryStorage.Init(b.RegistryRoot)
	statedb.keyStorage.Init(b.KeyRoot)
	statedb.nameStorage.Init(b.NameRoot)
}

// Copy generates a copy of the StateDB
func (s *StateDB) Copy() (statedb *StateDB) {
	s.mu.Lock()
	statedb = New(s.Storage)
	statedb.block = s.block.Copy()
	statedb.parentHash = s.parentHash
	statedb.accountStorage.Init(s.accountStorage.ChunkHash())
	statedb.registryStorage.Init(s.registryStorage.ChunkHash())
	statedb.checkStorage.Init(s.checkStorage.ChunkHash())
	statedb.keyStorage.Init(s.keyStorage.ChunkHash())
	statedb.productsStorage.Init(s.productsStorage.ChunkHash())
	statedb.nameStorage.Init(s.nameStorage.ChunkHash())
	log.Trace("[statedb:SetRootFromBlock]", fmt.Sprintf("SetRootFromBlock: accountStorage:%x registryStorage:%x keyStorage: %x nameStorage %x",
		s.accountStorage.ChunkHash(), s.registryStorage.ChunkHash(), s.keyStorage.ChunkHash(), s.nameStorage.ChunkHash()))

	s.mu.Unlock()
	return statedb
}

// NewStateDB creates a new StateDB object
func NewStateDB(ctx context.Context, cs ChunkStore, blockHash common.Hash) (statedb *StateDB, err error) {
	statedb, err = newStateDB(ctx, cs, blockHash)
	if err != nil {
		return statedb, err
	}
	return statedb, nil
}

// CreateGenesis generates the first StateDB object and genesis block
func CreateGenesis(storage ChunkStore, c *GenesisConfig, isBlockMinter bool) (b *Block, statedb *StateDB, err error) {
	statedb = New(storage)
	statedb.block = c.CreatePreGenesis()
	ctx := context.TODO()
	for addr, account := range c.Accounts {
		a, err := statedb.createAccount(ctx, addr)
		if err != nil {
			log.Error("[statedb:createGenesisBlock] createAccount", "err", err)
			return b, statedb, err
		}
		//dprint("creating acct! (%x)", addr)
		a.SetBalance(account.Balance)
		a.SetQuota(account.Quota)
		statedb.muAccountObjects.Lock()
		statedb.accountObjects[addr] = a
		statedb.muAccountObjects.Unlock()
	}

	for i, n := range c.Registry {
		a, err := statedb.createRegistryObject(ctx, uint64(i))
		if err != nil {
			log.Error("[backend:createGenesisBlock] createRegistryObject", "err", err)
			return b, statedb, err
		}
		a.SetAddress(n.Address)
		a.SetStorageIP([]byte(n.StorageIP))
		a.SetConsensusIP(ipstring_to_netip(n.ConsensusIP))
		a.SetValueInt(n.ValueInt)
		a.SetValueExt(n.ValueExt)
		a.SetRegion(n.Region)
		a.SetHTTPPort(n.HTTPPort)
		statedb.muRegistryObjects.Lock()
		statedb.registryObjects[uint64(i)] = a
		statedb.muRegistryObjects.Unlock()
	}

	var wg sync.WaitGroup
	_, commitToErr := statedb.CommitTo(ctx, &wg, isBlockMinter)
	if commitToErr != nil {
		log.Error("[backend:createGenesisBlock] CommitTo ERROR", "commitToErr", commitToErr)
	}

	flushError := statedb.Flush(ctx, &wg, isBlockMinter)
	if flushError != nil {
		log.Error("[backend:createGenesisBlock] Flush ERROR", "flushError", flushError)
	}
	//wg.Wait() // no wait here?

	log.Trace("[backend:createGenesisBlock] After Flushing!!!!", "statedb", statedb)

	b = NewBlock()
	b.NetworkID = statedb.block.NetworkID
	b.ParentHash = statedb.block.Hash()
	b.Seed = statedb.block.Seed
	b.StorageBeta = statedb.block.StorageBeta
	b.BandwidthBeta = statedb.block.BandwidthBeta
	b.AccountRoot = statedb.accountStorage.ChunkHash()
	b.RegistryRoot = statedb.registryStorage.ChunkHash()
	b.BlockNumber = 1
	statedb.block = b
	log.Trace("[statedb:createGenesis]", "HASH", b.Hash(), "statedb", statedb)
	return b, statedb, nil
}

// RevertToSnapshot
func (statedb *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(statedb.validRevisions), func(i int) bool {
		return statedb.validRevisions[i].id >= revid
	})
	if idx == len(statedb.validRevisions) || statedb.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := statedb.validRevisions[idx].journalIndex

	// Replay the journal to undo changes.
	for i := len(statedb.journal.entries) - 1; i >= snapshot; i-- {
		statedb.journal.entries[i].undo(statedb)
	}
	statedb.journal.entries = statedb.journal.entries[:snapshot]

	// Remove invalidated snapshots from the stack.
	statedb.validRevisions = statedb.validRevisions[:idx]
}

// Snapshot returns the integer revisionid
func (statedb *StateDB) Snapshot() (i int) {
	id := statedb.nextRevisionID
	statedb.nextRevisionID++
	statedb.validRevisions = append(statedb.validRevisions, revision{id, len(statedb.journal.entries)})
	return id
}

func (statedb *StateDB) clearJournal() {
	statedb.journal = newJournal()
	statedb.validRevisions = statedb.validRevisions[:0]
}

// VerifyAllTransactions verifies ALL the transactions supplied by some minter, ANY error and we error out
// With  parallelApplyTransactions we can parallelize operations but need to order dependent txes
//  Ordering: account creation > bucket creation > key inserts
func (statedb *StateDB) VerifyAllTransactions(ctx context.Context, txs []*Transaction, parallelVerify bool) (err error) {
	if len(txs) == 0 {
		return nil
	}
	if parallelVerify {
		st := time.Now()
		var errTotal error
		var errMu sync.RWMutex
		var wg sync.WaitGroup
		wg.Add(len(txs))
		for _, tx := range txs {
			go func(tx *Transaction) {
				err := statedb.ApplyTransaction(ctx, tx, "V")
				if err != nil {
					errMu.Lock()
					errTotal = err
					errMu.Unlock()
				}
				wg.Done()
			}(tx)
		}
		wg.Wait()
		log.Info("[backend:VerifyAllTransactions]", "len(wolktxns)", len(txs), "tm", time.Since(st))
		return errTotal
	}
	for _, tx := range txs {
		err := statedb.ApplyTransaction(ctx, tx, "V")
		if err != nil {
			// ABORT under ANY failure
			log.Error(fmt.Sprintf("[backend:VerifyAllTransactions] ApplyTransaction: %x", tx.Hash()))
			return err
		}
	}
	return nil
}

// ApplyTransactions updates the statedb object with the applytransaction
func (statedb *StateDB) ApplyTransactions(ctx context.Context, txs []*Transaction, parallelApplyTransactions bool) (err error) {
	if len(txs) == 0 {
		return nil
	}
	start := time.Now()
	if parallelApplyTransactions {
		var wg sync.WaitGroup
		wg.Add(len(txs))
		for _, tx := range txs {
			go func(tx *Transaction) {
				statedb.ApplyTransaction(ctx, tx, "Batch")
				wg.Done()
			}(tx)
		}
		wg.Wait()
		log.Info("[backend:ApplyTransactions] PARALLEL Complete", "tm", time.Since(start), "len(wolktxns)", len(txs))
		return nil
	}
	for i, tx := range txs {
		st := time.Now()
		err := statedb.ApplyTransaction(ctx, tx, "Batch")
		if err != nil {
			log.Error("[backend:ApplyTransactions]", "i", i, "tm", time.Since(st))
		} else {
			if time.Since(st) > warningTxTime {
				log.Error("[backend:ApplyTransactions]", "i", i, "tm", time.Since(st))
			}
		}
	}
	log.Info("[backend:ApplyTransactions] SERIAL Complete", "tm", time.Since(start), "len(wolktxns)", len(txs))
	return nil
}

func (statedb *StateDB) setTx(txhash common.Hash, status int8) {
	statedb.mu.Lock()
	defer statedb.mu.Unlock()
	statedb.attemptedTxs[txhash] = status
}

func (statedb *StateDB) getTx(txhash common.Hash) int8 {
	statedb.mu.Lock()
	defer statedb.mu.Unlock()
	if status, ok := statedb.attemptedTxs[txhash]; ok {
		return status
	}
	return -1
}

// ApplyTransaction actually applies the tx to the statedb
// important: need to check that the tx isn't being done over and over again!
func (statedb *StateDB) ApplyTransaction(ctx context.Context, tx *Transaction, code string) (err error) {
	incomingTxsize := uint64(len(tx.Bytes()))
	if statedb.txSize+incomingTxsize > 500000 {
		return nil
	}

	txhash := tx.Hash()
	status := statedb.getTx(txhash)
	if status < 0 {
		statedb.setTx(txhash, 1)
	} else if status == 1 {
		//log.Warn("[backend:ApplyTransaction] bop out", "len(attemptedTxs)", len(stateDB.attemptedTxs), "len(committedTxs)", len(stateDB.attemptedTxs))
		return nil
	}

	sender, _ := tx.GetSignerAddress()
	payload, err := tx.GetTxPayload()
	if err != nil {
		// skip this one! consider deleting it from the pool
		log.Error("[backend:ApplyTransactions] txn erred, will NOT be executed", "tx", tx, "err", err)
		return err
	}
	switch txp := payload.(type) {
	case TxTransfer:
		err = statedb.Transfer(ctx, sender, txp.Recipient, uint64(txp.Amount))
	case TxNode:
		log.Trace("[statedb:ApplyTransaction] TxNode", "sender", sender, "txp", txp)
		node, _ := strconv.Atoi(tx.Key())
		err = statedb.RegisterNode(ctx, sender, uint64(node), txp.StorageIP, txp.ConsensusIP, txp.Region, txp.Amount)
	case TxBucket:
		st := time.Now()
		pathpieces := strings.Split(strings.Trim(string(tx.Path), "/"), "/")
		//log.Info("[statedb:ApplyTransaction] checking rsapubkey for setname or setbucket", "rsapubkey", string(txp.RSAPublicKey))
		setKeyAllowed := len(txp.RSAPublicKey) == 0
		if len(pathpieces) == 1 {
			//dprint("going into ApplyTransaction: SETNAME")
			err = statedb.SetName(ctx, tx, &txp)
			if time.Since(st) > warningTxTime {
				log.Error("[statedb:ApplyTransaction] SetName", "tm", time.Since(st))
			}
		} else if setKeyAllowed {
			//log.Info("[statedb:ApplyTransaction] setKeyAllowed so entering SetBucket")
			err = statedb.SetBucket(ctx, tx)
			if time.Since(st) > warningTxTime {
				log.Error("[statedb:ApplyTransaction] SetBucket", "tm", time.Since(st))
			}
		}
	case TxKey:
		st := time.Now()
		txkey, err := tx.GetTxKey()
		if err != nil {
			return fmt.Errorf("[statedb:ApplyTransaction] %s", err)
		}
		if txkey.BucketIndexName != "" {
			err = statedb.SetIndexedKey(ctx, tx)
			if err != nil {
				return fmt.Errorf("[statedb:ApplyTransaction] %s", err)
			}
		} else {
			err = statedb.SetKey(ctx, tx)
			if err != nil {
				return fmt.Errorf("[statedb:ApplyTransaction] %s", err)
			}
		}
		if time.Since(st) > warningTxTime {
			log.Error("[statedb:ApplyTransaction] TxKey", "tm", time.Since(st))
		}
	case SQLRequest:
		log.Trace("[statedb:ApplyTransaction] SQL", "txp", txp, "tx", tx)
		err = statedb.ApplySQLTransaction(&txp)
	default:
		log.Error("[statedb:ApplyTransactions] UNKNOWN TYPE", "payload type", txp)
		err = fmt.Errorf("Unknown Payload Type %v %s", txp, tx.String())
	}
	if err != nil {
		log.Warn("[backend:ApplyTransaction]", "err", err)
		return err
	}
	// if we didnt get an error, then write it into Cloudstore, package into statedb committedTxes, which is used in proposing a block
	_, err = statedb.Storage.QueueChunk(context.TODO(), tx.Bytes())
	if err != nil {
		return err
	}
	//dprint("[statedb:ApplyTransaction] queued this id(%x), chunk(%+v)", chunkID, tx)

	// TODO: since the above statedb applications are NOT instantaoes and MAY be dependent, we need to be careful about when we add to this array!!!
	statedb.mu.Lock()
	statedb.committedTxes = append(statedb.committedTxes, tx)
	log.Trace("[statedb:ApplyTransaction] DONE", "len(stateDB.committedTxes)", len(statedb.committedTxes))

	statedb.txSize = statedb.txSize + incomingTxsize
	statedb.mu.Unlock()
	return nil
}

// Existed returns if the txhash has been committed already
func (statedb *StateDB) Existed(txhash common.Hash) bool {
	// check if the transaction is already commited in stateDB
	//stateDB.mu.RLock()
	//defer stateDB.mu.RUnlock()
	for _, tx := range statedb.committedTxes {
		if bytes.Compare(tx.Hash().Bytes(), txhash.Bytes()) == 0 {
			return true
		}
	}
	return false
}

// MakeBlock returns a new block with the txs supplied
// In this model, the stateDB has undergone ApplyTransactions already, but
//   (1) the State Objects have not been commited
//   (2) the SMTs have not been finalized
// isBlockMinter is true when proposing, and then the chunks are flushed to cloudstore
// isBlockMinter is false when verifying block, and the block is computed, but nothing is flushed to cloudstore
func (statedb *StateDB) MakeBlock(ctx context.Context, txs []*Transaction, policy *Policy, isBlockMinter bool) (bl *Block, err error) {
	var wg *sync.WaitGroup
	if !isBlockMinter {
		// This goes through and verifies ALL transactions, serially, in the order supplied
		err = statedb.VerifyAllTransactions(ctx, txs, true)
		if err != nil {
			log.Error("[statedb:MakeBlock] FAILED VerifyAllTransactions", "len(txs)", len(txs))
			return bl, err
		}
	} else {
		log.Trace("[statedb:MakeBlock] SUCCESS ***", "len(txs)", len(txs), "committedTxes", len(statedb.committedTxes), "isBlockMinter", isBlockMinter)
		wg = new(sync.WaitGroup)
	}

	_, err = statedb.CommitTo(ctx, wg, isBlockMinter)
	if err != nil {
		log.Error("[statedb:MakeBlock] CommitTo", "err", err, "blockNumber", statedb.block.Number()+1)
	}

	// these flush operations compute new root chunkhashs / merkleroots
	statedb.mu.Lock()
	statedb.Flush(ctx, wg, isBlockMinter)
	statedb.mu.Unlock()

	b := NewBlock()
	//	b.Transactions = sortedTransactions(txs)
	sort.Sort(ByAge(txs))
	b.Transactions = txs

	log.Trace("[statedb:MakeBlock] Parent ***", "prev statedb.block", statedb.block)
	b.BlockNumber = statedb.block.Number() + 1
	b.NetworkID = statedb.block.NetworkID
	b.ParentHash = statedb.block.Hash()
	b.AccountRoot = statedb.accountStorage.ChunkHash()
	b.RegistryRoot = statedb.registryStorage.ChunkHash()
	b.KeyRoot = statedb.keyStorage.ChunkHash()
	b.NameRoot = statedb.nameStorage.ChunkHash()
	b.Seed = statedb.GetSeed()

	b.StorageBeta = policy.AdjustUint64("StorageBeta", statedb.GetStorageBeta())
	b.BandwidthBeta = policy.AdjustUint64("BandwidthBeta", statedb.GetBandwidthBeta())
	return b, nil
}

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByAge []*Transaction

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return bytes.Compare(a[i].Hash().Bytes(), a[j].Hash().Bytes()) < 0 }

// Flush calls each SMTs flush operation
func (statedb *StateDB) Flush(ctx context.Context, wg *sync.WaitGroup, isBlockMinter bool) error {
	if parallelSMTFlush {
		var wgf sync.WaitGroup
		wgf.Add(6)
		go func() {
			statedb.accountStorage.Flush(ctx, wg, isBlockMinter)
			wgf.Done()
		}()
		go func() {
			statedb.registryStorage.Flush(ctx, wg, isBlockMinter)
			wgf.Done()
		}()
		go func() {
			statedb.checkStorage.Flush(ctx, wg, isBlockMinter)
			wgf.Done()
		}()
		go func() {
			statedb.keyStorage.Flush(ctx, wg, isBlockMinter)
			wgf.Done()
		}()
		go func() {
			statedb.productsStorage.Flush(ctx, wg, isBlockMinter)
			wgf.Done()
		}()
		go func() {
			statedb.nameStorage.Flush(ctx, wg, isBlockMinter)
			wgf.Done()
		}()
		wgf.Wait()
	} else {
		statedb.accountStorage.Flush(ctx, wg, isBlockMinter)
		statedb.registryStorage.Flush(ctx, wg, isBlockMinter)
		statedb.checkStorage.Flush(ctx, wg, isBlockMinter)
		statedb.keyStorage.Flush(ctx, wg, isBlockMinter)
		statedb.productsStorage.Flush(ctx, wg, isBlockMinter)
		statedb.nameStorage.Flush(ctx, wg, isBlockMinter)
	}
	// this actually writes chunks to cloud
	if isBlockMinter {
		wg.Wait()
		log.Trace("[statedb:MakeBlock] FlushQueue")
	}
	return nil
}

// CommitTo flushes out all the updated state objects and SMTs out to cloudstore ( if minter )
// or just computes the top level roots of the SMTs
func (statedb *StateDB) CommitTo(ctx context.Context, wg *sync.WaitGroup, writeToCloudstore bool) (merkleroot common.Hash, err error) {
	defer statedb.clearJournal()
	log.Trace("[statedb:CommitTo] start")

	// Commit registryObjects
	for i, registryObject := range statedb.registryObjects {
		err = statedb.updateRegistryObject(ctx, wg, writeToCloudstore, registryObject)
		if err != nil {
			return merkleroot, fmt.Errorf("[statedb:Commit] %s", err)
		}
		log.Trace("[statedb:CommitTo]", "len(registryObjects)", len(statedb.registryObjects), "i", i)
		delete(statedb.registryObjects, i)
	}

	// Commit checkObjects
	for i, checkObject := range statedb.checkObjects {
		err = statedb.updateCheckObject(ctx, checkObject)
		if err != nil {
			return merkleroot, fmt.Errorf("[statedb:CommitTo] updateCheckObject ERROR %s", err)
		}
		log.Trace("[statedb:CommitTo] checkObjects", "len(checkObjects)", len(statedb.checkObjects), "i", i, "err", err)
		delete(statedb.checkObjects, i)
	}

	// Commit sqlObjects
	err = statedb.CommitSQL(ctx, wg, writeToCloudstore)
	if err != nil {
		log.Error("[statedb:CommitTo] CommitSQL", "err", err)
		return merkleroot, fmt.Errorf("[statedb:CommitTo] updateCheckObject ERROR %s", err)
	}

	// Commit nosql objects
	err = statedb.CommitNoSQL(ctx, wg, writeToCloudstore)
	if err != nil {
		log.Error("[statedb:CommitTo] CommitNoSQL", "err", err)
		return merkleroot, fmt.Errorf("[statedb:CommitTo] CommitNoSQL ERROR %s", err)
	}

	err = statedb.CommitStorageUsed(ctx, writeToCloudstore)
	if err != nil {
		log.Error("[statedb:CommitTo] CommitStorageUsed", "err", err)
		return merkleroot, fmt.Errorf("[statedb:CommitTo] CommitStorageUsed ERROR %s", err)
	}

	// account - balances, quotas, storage usage
	statedb.muAccountObjects.Lock()
	defer statedb.muAccountObjects.Unlock()
	for addr, accountObject := range statedb.accountObjects {
		err = statedb.updateAccountObject(ctx, wg, writeToCloudstore, accountObject)
		if err != nil {
			return merkleroot, fmt.Errorf("[statedb:CommitTo] updateAccountObject %s", err)
		}
		delete(statedb.accountObjects, addr)
	}

	return merkleroot, err
}

// ComputeStorageClaims can be outsourced too much work for a miner
func (statedb *StateDB) ComputeStorageClaims(ctx context.Context, currentBlock uint64) (err error) {
	// TODO: scan accounts
	addresses, _, err := statedb.accountStorage.ScanAll(ctx, false)
	if err != nil {
		return err
	}
	for addr := range addresses {
		a, ok, errA := statedb.getAccountObject(ctx, addr)
		if errA != nil {

		} else if ok {
			if a.account.LastClaim-currentBlock > minBlocksForStorageClaims {
				statedb.storageClaims[addr] = a.account.LastClaim
			}
		}
	}
	return nil
}

func (statedb *StateDB) CommitStorageClaims(ctx context.Context, blockNumber uint64, storageBeta uint64) (err error) {
	// nosql - insert new SMT roots for all the collections that were updated from the keyobjects into keyStorage
	reward := uint64(0)
	for addr, _ := range statedb.storageClaims {
		account, errQ := statedb.getOrCreateAccount(ctx, addr)
		if errQ != nil {
			log.Error("[statedb:CommitStorageClaims] getOrCreateAccount", "addr", addr, "err", errQ)
			return errQ
		}
		reward += account.ChargeForStorageClaims(blockNumber, storageBeta)
		statedb.muAccountObjects.Lock()
		statedb.accountObjects[addr] = account
		statedb.muAccountObjects.Unlock()
		delete(statedb.storageClaims, addr)
	}

	// reward goes to the miner!
	if reward == 0 {
		return nil
	}
	/*
		minerAddr := statedb.GetOperatorAddress()
		minerAccount, errM := statedb.getOrCreateAccount(minerAddr)
		if errM != nil {
			return errM
		}
		minerAccount.TallyReward(reward)
	*/
	return nil
}

// CommitStorageUsed updates the account state objects with the tallies
func (statedb *StateDB) CommitStorageUsed(ctx context.Context, writeToCloudstore bool) (err error) {
	// nosql - insert new SMT roots for all the collections that were updated from the keyobjects into keyStorage
	statedb.muStorageUsage.Lock()
	defer statedb.muStorageUsage.Unlock()

	for addr, storageUsed := range statedb.storageUsage {
		account, err := statedb.getOrCreateAccount(ctx, addr)
		if err != nil {
			log.Error("[statedb:CommitStorageUsed] ERR", "addr", addr, "storageUsed", storageUsed, "err", err)
			return fmt.Errorf("[statedb:CommitStorageUsed] %s", err)
		}
		account.TallyStorageUsage(storageUsed)
		statedb.muAccountObjects.Lock()
		statedb.accountObjects[addr] = account
		statedb.muAccountObjects.Unlock()
		delete(statedb.storageUsage, addr)
	}
	return nil
}

// SetBalance updates the account balance of addr
func (statedb *StateDB) SetBalance(ctx context.Context, addr common.Address, amount uint64) {
	accountObject, err := statedb.getOrCreateAccount(ctx, addr)
	if err != nil {
	}
	if accountObject != nil {
		accountObject.SetBalance(amount)
	}
}

// SetQuota updates the quota of addr
func (statedb *StateDB) SetQuota(ctx context.Context, owner common.Address, quota uint64) error {
	account, ok, err := statedb.getAccountObject(ctx, owner)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("[statedb:SetQuota] Sender account not found")
	}
	usage := account.Usage()
	if quota < usage {
		return fmt.Errorf("[statedb:SetQuota] New quota %d is below usage of %d", quota, usage)
	}
	account.SetQuota(quota)
	log.Info("[statedb:SetQuota]", "owner", owner, "quota", quota, "usage", usage)
	statedb.muAccountObjects.Lock()
	statedb.accountObjects[owner] = account
	statedb.muAccountObjects.Unlock()
	return nil
}

func (statedb *StateDB) tallyStorageUsage(ctx context.Context, owner string, storageBytesUsed uint64) (err error) {
	var ownerAddr common.Address
	ownerAddr, ok, _, err := statedb.GetName(ctx, owner, false)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("Owner %s not found", owner)
	}
	statedb.muStorageUsage.Lock()
	defer statedb.muStorageUsage.Unlock()
	if _, ok = statedb.storageUsage[ownerAddr]; !ok {
		statedb.storageUsage[ownerAddr] = 0
	}
	currentUsage, quota, ok, err := statedb.getCurrentUsageAndQuota(ctx, ownerAddr)
	if ok {
		if currentUsage+storageBytesUsed > quota {
			//			return fmt.Errorf("[statedb:tallyStorageUsage] FAIL - Current Usage %d + storage Used %d brings address %x Over quota of %d", currentUsage, storageBytesUsed, owner, quota)
		}
	}
	statedb.storageUsage[ownerAddr] += storageBytesUsed
	log.Trace("[statedb:tallyStorageUsage] SUCC", "storageBytesUsed", storageBytesUsed, "new statedb.storageUsage[owner]", statedb.storageUsage[ownerAddr], "currentUsage", currentUsage, "quota", quota)
	return nil
}

func (statedb *StateDB) getCurrentUsageAndQuota(ctx context.Context, addr common.Address) (usage uint64, quota uint64, ok bool, err error) {
	a, ok, err := statedb.getAccountObject(ctx, addr)
	if err != nil {
		return 0, 0, false, err
	} else if !ok {
		return 0, 0, false, nil
	}
	return a.Usage(), a.Quota(), true, nil
}

func (statedb *StateDB) getAccountObject(ctx context.Context, addr common.Address) (a *accountObject, ok bool, err error) {
	log.Trace("[statedb:getAccountObject] getting account", "addr", addr)
	statedb.muAccountObjects.RLock()
	if a0, okAcc := statedb.accountObjects[addr]; okAcc {
		statedb.muAccountObjects.RUnlock()
		return a0, true, nil
	}
	statedb.muAccountObjects.RUnlock()

	shortAddr := addr[0 : NumBitsAddress/8]
	v, found, deleted, _, _, err := statedb.accountStorage.Get(ctx, shortAddr, false)
	log.Trace("[statedb:getAccountObject]", "addr", addr, "shortAddr", fmt.Sprintf("%x", shortAddr), "found", found, "deleted", deleted, "v", fmt.Sprintf("%x", v))
	if err != nil {
		return nil, false, fmt.Errorf("[statedb:getAccountObject] %s", err)
	}
	if found == false {
		return nil, false, nil
	}

	encoded, ok, err := statedb.Storage.GetChunk(ctx, v)
	if err != nil {
		log.Error("[statedb:getAccountObject] GetChunk", "v", v, "err", err)
		return nil, false, fmt.Errorf("[statedb:getAccountObject] %s", err)
	} else if !ok {
		log.Error("[statedb:getAccountObject] Not OK", "chunkID", fmt.Sprintf("%x", v))
		return nil, false, nil
	}
	log.Trace("[statedb:getAccountObject]", "chunkID", fmt.Sprintf("%x", v))
	var acct Account // []interface{}
	err = rlp.Decode(bytes.NewReader(encoded), &acct)
	if err != nil {
		log.Error("[statedb:getAccountObject] Decode", "account", addr, "shortAddr", shortAddr, "v", v, "ERR", err)
		return nil, false, fmt.Errorf("[statedb:getAccountObject] %s", err)
	}
	log.Trace("[statedb:getAccountObject]", "addr", addr, "account", acct)
	return NewAccountObject(statedb, addr, acct), true, nil
}

// deprecated
func (statedb *StateDB) createAccount(ctx context.Context, addr common.Address) (a *accountObject, err error) {
	shortAddr := addr[0 : NumBitsAddress/8]
	_, isFound, deleted, _, _, err := statedb.accountStorage.Get(ctx, shortAddr, false)
	if err != nil {
		log.Error("[statedb:createAccount]", "acct", addr, "shortAddr", shortAddr, "status", "ERROR")
		return a, err
	} else if isFound {
		log.Debug("[statedb:createAccount]", "acct", addr, "shortAddr", shortAddr, "status", "shortAddr collation", "deleted", deleted)
		return a, fmt.Errorf("shortAddr already exist")
	}

	log.Debug("[statedb:createAccount]", "acct", addr, "status", "canOpen")
	acct := Account{
		Balance: 0,
		Quota:   0,
	}
	return NewAccountObject(statedb, addr, acct), nil
}

func (statedb *StateDB) accountExists(ctx context.Context, addr common.Address) bool {
	statedb.muAccountObjects.RLock()
	if a0, isSet := statedb.accountObjects[addr]; isSet {
		statedb.muAccountObjects.RUnlock()
		log.Info("[statedb:accountExists] CACHE", "addr", addr, "a0", a0.String())
		return true
	}
	statedb.muAccountObjects.RUnlock()
	_, isFound, _, _, _, err := statedb.accountStorage.Get(ctx, addr.Bytes(), false)
	if err != nil {
		log.Error("[statedb:accountExists]", "addr", addr, "err", err)
		return false
	}
	return isFound
}

// A 3-tier lookup procedure
func (statedb *StateDB) getOrCreateAccount(ctx context.Context, addr common.Address) (a *accountObject, err error) {
	//Tier1 : Lookup from current state
	statedb.muAccountObjects.RLock()
	if a0, isSet := statedb.accountObjects[addr]; isSet {
		statedb.muAccountObjects.RUnlock()
		log.Trace("[statedb:getOrCreateAccount] CACHE", "addr", addr, "a0", a0.String())
		return a0, nil
	}
	statedb.muAccountObjects.RUnlock()
	v, isFound, _, _, _, err := statedb.accountStorage.Get(ctx, addr.Bytes(), false)
	if err != nil {
		log.Error("[statedb:getOrCreateAccount]", "addr", addr, "err", err)
		return a, err
	}

	//Tier2 : Lookup from rs
	if isFound {
		encoded, ok, err := statedb.Storage.GetChunk(ctx, v)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, fmt.Errorf("NOT FOUND CHUNK: %x", v)
		}
		var acct Account
		err = rlp.Decode(bytes.NewReader(encoded), &acct)
		if err != nil {
			return nil, err
		}
		log.Debug("[statedb:getOrCreateAccount] FOUND", "acct", addr, "account", acct.String())
		return NewAccountObject(statedb, addr, acct), nil
	}

	//Tier 3: createAccount if still not found
	acct := Account{
		Balance: 0,
		Quota:   0,
	}
	log.Debug("[statedb:getOrCreateAccount] NOT FOUND", "acct", addr, "account", acct.String)
	return NewAccountObject(statedb, addr, acct), nil
}

func (statedb *StateDB) updateAccountObject(ctx context.Context, wg *sync.WaitGroup, writeToCloudstore bool, a *accountObject) (err error) {
	encoded, err := rlp.EncodeToBytes(&(a.account))
	if err != nil {
		log.Error("[statedb:updateAccountObject] EncodeToBytes", "err", err)
		return err
	}
	v := wolkcommon.Computehash(encoded)
	if writeToCloudstore {
		if wg != nil {
			wg.Add(1)
		}
		// put chunk into storage queue so that final Flush on storage uses setchunkbatch call
		c := new(cloud.RawChunk)
		c.Value = encoded
		_, err = statedb.Storage.PutChunk(context.TODO(), c, wg)
		if err != nil {
			log.Error("[statedb:updateAccountObject] | Error doing StoreChunk", "error", err)
			//return err
		} else {
			log.Trace(fmt.Sprintf("statedb:updateAccountObject] PutChunk! %x", v))
		}
	}

	var a2 Account // []interface{}
	err = rlp.Decode(bytes.NewReader(encoded), &a2)
	if err != nil {
		log.Error("[statedb:updateAccountObject]", "err", err)
	}
	shortAddr := a.address[0 : NumBitsAddress/8]

	if a.deleted {
		err = statedb.accountStorage.Insert(ctx, shortAddr, v, 0, true)
	} else {
		err = statedb.accountStorage.Insert(ctx, shortAddr, v, 0, false)
	}
	log.Trace("[statedb:updateAccountObject] StoreChunk MAPPING", "addr", a.Address(), "shortAddr", shortAddr, "v", v)
	if err != nil {
		log.Error("[statedb:updateAccountObject]", "a", a, "ERROR", err)
		return err
	}
	return nil
}

func (statedb *StateDB) createRegistryObject(ctx context.Context, idx uint64) (a *registryObject, err error) {
	addr := IndexToAddress(idx)
	_, isFound, _, _, _, err := statedb.registryStorage.Get(ctx, addr.Bytes(), false)
	if err != nil {
		log.Error("[statedb:createRegistryObject]", "addr", fmt.Sprintf("%x", addr), "status", "ERROR")
		return a, err
	} else if isFound {
		log.Debug("[statedb:createRegistryObject]", "addr", fmt.Sprintf("%x", addr), "status", "addr collation")
		return a, fmt.Errorf("[statedb:createRegistryObject] registry object already exists")
	}

	log.Debug("[statedb:createRegistryObject]", "addr", addr, "status", "canOpen")
	n := RegisteredNode{}
	return NewRegistryObject(statedb, idx, n), nil
}

func (statedb *StateDB) getRegistryObject(ctx context.Context, idx uint64) (a *registryObject, err error) {
	if a0, ok := statedb.registryObjects[idx]; ok {
		return a0, nil
	}
	addr := IndexToAddress(idx)
	v, found, _, _, _, err := statedb.registryStorage.Get(ctx, addr.Bytes(), false)
	log.Trace("[statedb:getRegistryObject]", "addr", fmt.Sprintf("%x", addr), "found", found, "v", fmt.Sprintf("%x", v))
	if err != nil {
		return nil, err
	}
	if found == false {
		return nil, fmt.Errorf("[statedb:getRegistryObject] Not found")
	}

	encoded, ok, err := statedb.Storage.GetChunk(ctx, v)
	if err != nil {
		log.Error("[statedb:getRegistryObject] GetChunk", "v", v, "err", err)
		return nil, err
	} else if !ok {
		log.Error("[statedb:getRegistryObject] Not OK", "chunkID", fmt.Sprintf("%x", v))
		return nil, nil
	}
	log.Trace("[statedb:getRegistryObject]", "chunkID", fmt.Sprintf("%x", v))
	var n RegisteredNode
	err = rlp.Decode(bytes.NewReader(encoded), &n)
	if err != nil {
		log.Error("[statedb:getRegistryObject] Decode", "addr", fmt.Sprintf("%x", addr), "v", v, "ERR", err)
		return nil, err
	}
	log.Trace("[statedb:getRegistryObject]", "addr", fmt.Sprintf("%x", addr), "n", n.String())
	return NewRegistryObject(statedb, idx, n), nil
}

func (statedb *StateDB) updateRegistryObject(ctx context.Context, wg *sync.WaitGroup, writeToCloudstore bool, a *registryObject) (err error) {
	encoded, err := rlp.EncodeToBytes(&(a.registeredNode))
	if err != nil {
		log.Error("[statedb:updateRegistryObject] EncodeToBytes err", err)
		return err
	}
	v := wolkcommon.Computehash(encoded)

	if writeToCloudstore {
		if wg != nil {
			wg.Add(1)
		}
		// put chunk into storage queue so that final Flush on storage uses setchunkbatch call
		c := new(cloud.RawChunk)
		c.Value = encoded
		_, errC := statedb.Storage.PutChunk(context.TODO(), c, wg)
		if errC != nil {
			log.Error("[statedb:updateRegistryObject] | Error doing StoreChunk", "error", errC)
			//return err
		} else {
			log.Trace(fmt.Sprintf("statedb:updateRegistryObject] PutChunk! %x", v))
		}
	}

	var a2 RegisteredNode
	err = rlp.Decode(bytes.NewReader(encoded), &a2)
	if err != nil {
		log.Error("[statedb:updateRegistryObject]", "err", err)
	}
	addr := IndexToAddress(a.idx)
	err = statedb.registryStorage.Insert(ctx, addr.Bytes(), v, 0, a.deleted)
	if err != nil {
		log.Error("[statedb:updateRegistryObject]", "a", a, "ERROR", err)
		return err
	}
	return nil
}

// GetAccount returns account associated with addr
func (statedb *StateDB) GetAccount(ctx context.Context, addr common.Address) (a *Account, ok bool, err error) {
	accountObject, ok, err := statedb.getAccountObject(ctx, addr)
	if err != nil {
		return a, ok, err
	} else if !ok {
		return a, ok, nil
	}
	return &(accountObject.account), ok, nil
}

// GetRegisteredNode returns node associated with idx
func (statedb *StateDB) GetRegisteredNode(ctx context.Context, idx uint64) (a RegisteredNode, err error) {
	registryObject, err := statedb.getRegistryObject(ctx, idx)
	if err != nil {
		return a, err
	}
	return registryObject.registeredNode, nil
}

func (statedb *StateDB) RegisterNode(ctx context.Context, sender common.Address, node uint64, storageip []byte, consensusip []byte, region uint8, value uint64) error {
	// fetch the node
	n, err := statedb.getRegistryObject(ctx, node)
	if err != nil {
		log.Error("[statedb:RegisterNode] getRegistryObject", "err", err)
		return fmt.Errorf("[statedb:RegisterNode] %s", err)
	}
	if n == nil {
		log.Error("[statedb:RegisterNode]", "node not found")
		return fmt.Errorf("[statedb:RegisterNode] node not found")
	}
	valueMin := int64(n.ValueInt()) - int64(n.ValueExt())
	// fetch account
	senderAccount, ok, err := statedb.getAccountObject(ctx, sender)
	if err != nil {
		log.Error("[statedb:RegisterNode] getAccountObject", "err", err)
		return fmt.Errorf("[statedb:RegisterNode] %s", err)
	} else if !ok {
		log.Error("[statedb:RegisterNode] getAccountObject NOT OK")
		return fmt.Errorf("[statedb:RegisterNode] Sender account not found")
	}
	log.Trace("[statedb:RegisterNode] getAccountObject SUCC", "senderAccount", senderAccount.String())

	// check for amount in tx vs in the senders account
	if int64(value) < valueMin {
		log.Error("[statedb:RegisterNode] Insufficient Value", "value", value, "valueMin", valueMin, "valueInt", n.ValueInt(), "valueExt", n.ValueExt())
		return fmt.Errorf("[statedb:RegisterNode] Insufficient value")
	}
	vMin := uint64(valueMin)
	//	funds := senderAccount.Balance()
	if vMin < 0 {
		return fmt.Errorf("[statedb:RegisterNode] Insufficient funds")
	}
	recipient := n.registeredNode.address
	recipientAccount, ok, err := statedb.getAccountObject(ctx, recipient)
	if err != nil {
		return fmt.Errorf("[statedb:RegisterNode] %s", err)
	} else if !ok {
		// ??
		log.Error("[statedb:RegisterNode] getAccountObject was NOT OK. Continuing.")
	}

	senderAmount := senderAccount.Balance() - vMin
	recipientAmount := recipientAccount.Balance() + vMin
	senderAccount.SetBalance(senderAmount)
	recipientAccount.SetBalance(recipientAmount)
	log.Info("[statedb:RegisterNode] RegisterNode", "vmin", vMin, "sender after", senderAccount.String(), "recipient after", recipientAccount.String())

	n.SetAddress(sender)
	n.SetStorageIP(storageip)
	n.SetConsensusIP(consensusip)
	n.SetRegion(region)
	n.SetValueInt(value)
	statedb.muAccountObjects.Lock()
	statedb.accountObjects[sender] = senderAccount
	statedb.accountObjects[recipient] = recipientAccount
	statedb.muAccountObjects.Unlock()
	statedb.muRegistryObjects.Lock()
	statedb.registryObjects[node] = n
	statedb.muRegistryObjects.Unlock()
	return nil
}

func (statedb *StateDB) UpdateNode(ctx context.Context, sender common.Address, node uint64, storageip []byte, consensusip []byte, region uint8, valueInternal uint64) error {
	n, err := statedb.getRegistryObject(ctx, node)
	if err != nil {
		log.Error("[statedb:UpdateNode] getRegistryObject", "err", err)
		return err
	}
	if n == nil {
		log.Error("[statedb:UpdateNode] getRegistryObject", "err", "node not found")
		return fmt.Errorf("node not found")
	}
	if bytes.Compare(n.Address().Bytes(), sender.Bytes()) != 0 {
		log.Error("[statedb:UpdateNode] Not owner", "sender", sender, "n.Addr", n.Address())
		return fmt.Errorf("[statedb:UpdateNode] Not owner of node")
	}
	n.SetStorageIP(storageip)
	n.SetConsensusIP(consensusip)
	n.SetRegion(region)
	n.SetValueInt(valueInternal)
	statedb.registryObjects[node] = n
	return nil
}

func (statedb *StateDB) getCheckObject(ctx context.Context, checkID common.Hash) (a *checkObject, err error) {
	if a0, ok := statedb.checkObjects[checkID]; ok {
		return a0, nil
	}
	addr := CheckIDToAddress(checkID)
	blockHashBytes, found, _, _, _, err := statedb.checkStorage.Get(ctx, addr.Bytes(), false)
	if err != nil {
		return a, err
	}
	if found {
		if len(blockHashBytes) != 32 {
			return a, fmt.Errorf("[statedb:getCheckObject] Incorrect blockhash for check %d", len(blockHashBytes))
		} else {
			return NewCheckObject(statedb, checkID, common.BytesToHash(blockHashBytes)), nil
		}
	}
	return nil, nil
}

func (statedb *StateDB) updateCheckObject(ctx context.Context, a *checkObject) (err error) {
	addr := CheckIDToAddress(a.CheckID())
	log.Trace("[statedb:updateCheckObject]", "addr", addr.Bytes(), "val", a.BlockHash().Bytes())
	err = statedb.checkStorage.Insert(ctx, addr.Bytes(), a.BlockHash().Bytes(), 0, false)
	if err != nil {
		log.Error("[statedb:updateCheckObject]", "a", a, "ERROR", err)
		return err
	}
	return nil
}

// the check recipient will be the one submitting a check by its checkID, and checkRecipient=check.recipient
func (statedb *StateDB) BandwidthCheck(ctx context.Context, checkRecipient common.Address, checkID common.Hash) error {
	checkBytes, _, err := statedb.Storage.GetChunk(ctx, checkID.Bytes())
	if err != nil {
		log.Error("[statedb:BandwidthCheck] RetrieveCheck", "err", err)
		return err
	}
	var check BandwidthCheck
	err = json.Unmarshal(checkBytes, &check)
	if err != nil {
		return err
	}
	// TODO: map checkBytes into BandwidthCheck
	log.Trace("[statedb:BandwidthCheck] RetrieveCheck", "check", check.String())
	amount := check.Amount
	recipient := check.Recipient

	if bytes.Compare(checkRecipient.Bytes(), recipient.Bytes()) != 0 {
		log.Error("[statedb:BandwidthCheck] Compare failure")
		return fmt.Errorf("[statedb:BandwidthCheck] Only the check recipient can cash check")
	}

	giver, err := check.GetSigner()
	if err != nil {
		log.Error("[statedb:BandwidthCheck] GetSigner", "err", err)
		return err
	}
	log.Trace("[statedb:BandwidthCheck] GetSigner", "giver", giver)
	giverAccount, ok, err := statedb.getAccountObject(ctx, giver)
	if err != nil {
		log.Error("[statedb:BandwidthCheck] getAccountObject", "err", err)
		return err
	} else if !ok {
		return fmt.Errorf("Cannot receive")
	}
	log.Trace("[statedb:BandwidthCheck] getAccountObject", "giverAccount", giverAccount)

	// check for amount
	giverBalance := giverAccount.Balance()
	if giverBalance < amount {
		log.Error("[statedb:BandwidthCheck] Insufficient funds")
		return fmt.Errorf("Insufficient funds")
	}
	log.Trace("[statedb:BandwidthCheck]", "giverBalance", giverBalance)

	recipientAccount, err := statedb.getOrCreateAccount(ctx, recipient)
	if err != nil {
		log.Error("[statedb:BandwidthCheck] getOrCreateAccount", "err", err)
		return err
	}
	log.Trace("[statedb:BandwidthCheck] getOrCreateAccount", "recipient", fmt.Sprintf("%x", recipient), "recipientAccount", recipientAccount)

	checkObject, err := statedb.getCheckObject(ctx, checkID)
	if err != nil {
		log.Error("[statedb:BandwidthCheck] getCheckObject", "err", err)
		return err
	}
	if checkObject != nil {
		log.Error("[statedb:BandwidthCheck] Check has been cashed already")
		return fmt.Errorf("[statedb:BandwidthCheck] Check has been cashed already")
	}

	checkObject = NewCheckObject(statedb, checkID, checkID)
	giverAmount := giverAccount.Balance() - uint64(amount)
	recipientAmount := recipientAccount.Balance() + uint64(amount)
	giverAccount.SetBalance(giverAmount)
	recipientAccount.SetBalance(recipientAmount)
	log.Trace("[statedb:BandwidthCheck] BALANCES", "giverAccount", giverAmount, "recipientAccount", recipientAmount)
	statedb.muAccountObjects.Lock()
	statedb.accountObjects[giver] = giverAccount
	statedb.accountObjects[recipient] = recipientAccount
	statedb.muAccountObjects.Unlock()
	statedb.muCheckObjects.Lock()
	statedb.checkObjects[checkID] = checkObject
	statedb.muCheckObjects.Unlock()
	return nil
}

// Transfer update account objects balances
func (statedb *StateDB) Transfer(ctx context.Context, sender common.Address, recipient string, amount uint64) error {
	senderAccount, ok, err := statedb.getAccountObject(ctx, sender)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("[statedb:Transfer] Sender account not found [%x]", sender)
	}
	// check for amount
	recipientAddr, ok, _, err := statedb.GetName(ctx, recipient, false)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("[statedb:Transfer] Recipient account not found")
	}

	recipientAccount, err := statedb.getOrCreateAccount(ctx, recipientAddr)
	if err != nil {
		return err
	}
	senderAmount := senderAccount.Balance() - amount
	recipientAmount := recipientAccount.Balance() + amount
	senderAccount.SetBalance(senderAmount)
	recipientAccount.SetBalance(recipientAmount)
	log.Trace("[statedb:Transfer]", "sender", senderAmount, "recipient", recipientAmount)
	statedb.muAccountObjects.Lock()
	statedb.accountObjects[sender] = senderAccount
	statedb.accountObjects[recipientAddr] = recipientAccount
	statedb.muAccountObjects.Unlock()
	return nil
}

// GetSeed returns block seed
func (statedb *StateDB) GetSeed() []byte {
	return statedb.block.Seed
}

// GetStorageBeta returns storage beta, used to compute storage costs
func (statedb *StateDB) GetStorageBeta() uint64 {
	return statedb.block.StorageBeta
}

// GetBandwidthBeta returns bandwidth beta, used to compute bandwidth costs
func (statedb *StateDB) GetBandwidthBeta() uint64 {
	return statedb.block.BandwidthBeta
}
