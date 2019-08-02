package wolk

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/syndtr/goleveldb/leveldb"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
	//"github.com/wolkdb/cloudstore/wolk/log/opentrace"
)

const (
	consensusSingleNode = "singlenode"
	consensusAlgorand   = "algorand"
)

const (
	useBlockchain = true
	isPreemptive  = false
	useFileHash   = false

	maxConcurrentChains = 10 //TODO: make a flag/param?

	MAX_VALIDATORS   = 8
	Q                = 2
	DefaultWolkIndex = 1
	MIN_SET_SUCCESS  = 4

	intraregionMaxSetShareTime = 4000 * time.Millisecond
	interregionMaxSetShareTime = 4000 * time.Millisecond
	intraregionMaxGetShareTime = 8000 * time.Millisecond
	interregionMaxGetShareTime = 8000 * time.Millisecond
	useFileLog                 = false
	retrySetShare              = 1
	sendCancel                 = true
)

// TransactionStatus options
const (
	transactionStatusNotFound = "Not Found"
	transactionStatusApplied  = "Applied"
	transactionStatusProposed = "Proposed"
)

// Consensus Node Status indicators
const (
	// setup from NewWolk constructor to node being ready
	consensusStatusStart        = "Init:Start"
	consensusStatusMakeGenesis  = "Init:MakeGenesis"
	consensusStatusGettingReady = "Init:GettingReady"
	consensusStatusSynchronise  = "Init:Synchronise"

	consensusStatusSynchroniseCert = "SynchroniseCert"

	// 1st step in processConsensus
	consensusStatusBlockProposal       = "BlockProposal"
	consensusStatusSilentObserver      = "BlockProposal:SilentObserver"
	consensusStatusProposingBlock      = "BlockProposal:ProposingBlock"
	consensusStatusWriteEmptyBlock     = "BlockProposal:WriteEmptyBlock"
	consensusStatusWaitingForBlock     = "BlockProposal:WaitingForBlock"
	consensusStatusBroadcastProposal   = "BlockProposal:BroadcastProposal"
	consensusStatusWriteCandidateBlock = "BlockProposal:WriteCandidateBlock"

	// next steps in processConsensus
	consensusStatusBA                = "BA"
	consensusStatusBAReduction       = "BAReduction"
	consensusStatusVerifyConsensus   = "VerifyConsensus"
	consensusStatusApplyTransactions = "ApplyTransactions"
	consensusStatusProposeFork       = "ProposeFork"
	consensusStatusWriteBlock        = "WriteBlock"
)

// WolkStore is the core object that holds all key objects to run the blockchain storage engine
type WolkStore struct {
	statusCh     chan struct{}
	hangForever  chan struct{}
	blocks       map[uint64]map[common.Hash]*Block
	last         *Block
	current      *Block
	genesisBlock *Block

	blockIndexLdb *leveldb.DB
	consensusIdx  int

	blockchainReady   bool
	muBlockchainReady sync.RWMutex
	muCurrentBlock    sync.RWMutex
	muLastBlock       sync.RWMutex

	// this is where incoming transactions are Applied
	PreemptiveStateDB   *StateDB
	muPreemptiveStateDB sync.RWMutex
	CurrentStateDB      *StateDB

	// Edwards, used in Consensus
	operatorKey *crypto.PrivateKey

	Region byte

	isPreemptive bool

	mu sync.RWMutex
	// status variables for testnet display
	mintedTS                 int64
	consensusStatusMu        sync.RWMutex
	consensusStatus          string
	consensusStatusUpdatedTS int64

	// blockchain
	wolktxpool *TxPool

	policy *Policy

	Indexer *Indexer
	Storage *Storage

	config  *cloud.Config
	genesis *GenesisConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down wolk

	// Handlers
	protocolManager *ProtocolManager
	eventMux        *event.TypeMux

	fetching          bool
	subChainHead      *Block
	subChainBlocks    []*Block
	fetchStatus       uint64
	lastFinalized     uint64
	lastFinalizedHash common.Hash

	insertChainLengthMu sync.RWMutex
	insertChainLength   int

	dispatcher *Dispatcher
}

// DefaultTransport is used for all storage http calls
var DefaultTransport http.RoundTripper = &http.Transport{
	Dial: (&net.Dialer{
		// limits the time spent establishing a TCP connection (if a new one is needed)
		Timeout:   10 * time.Second,
		KeepAlive: 60 * time.Second, // 60 * time.Second,
	}).Dial,
	//MaxIdleConns: 5,
	MaxIdleConnsPerHost: 25, // changed from 100 -> 25

	// limits the time spent reading the headers of the response.
	ResponseHeaderTimeout: 10 * time.Second,
	IdleConnTimeout:       4 * time.Second, // 90 * time.Second,

	// limits the time the client will wait between sending the request headers when including an Expect: 100-continue and receiving the go-ahead to send the body.
	ExpectContinueTimeout: 1 * time.Second,

	// limits the time spent performing the TLS handshake.
	TLSHandshakeTimeout: 30 * time.Second,

	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// NewWolk returns the WolkStore object
func NewWolk(ctx *node.ServiceContext, config *cloud.Config) (cs *WolkStore, err error) {
	log.Info("[backend:NewWolk] ***** INIT ****", "consensusIdx", config.ConsensusIdx)
	chunkID := wolkcommon.Computehash([]byte("sourabh"))
	log.Chunk(chunkID, "PutChunk")

	genesisConfig, err := LoadGenesisFile(config.GenesisFile)
	log.Info("[backend:NewWolk]", "genesisConfig", genesisConfig)
	if err != nil {
		log.Error("FAILURE", "err", err)
		return cs, err
	}
	log.Info("[backend:NewWolk] LoadGenesisFile")

	//	genesisConfig.Dump()
	policy := &Policy{}
	policyFile := DefaultPolicyWolkFile
	LoadPolicy(policyFile, policy)

	indexer, err := NewIndexer(config.DataDir)
	if err != nil {
		log.Info("[backend:NewWolk] NewIndexer err", "err", err)
		return cs, err
	}
	log.Info("[backend:NewWolk] NewIndexer", "err", err)

	wstore := WolkStore{
		consensusIdx: config.ConsensusIdx,
		wolktxpool:   NewTxPool(),

		eventMux:    new(event.TypeMux), // SHOULD BE: ctx.EventMux
		Indexer:     indexer,
		genesis:     genesisConfig,
		operatorKey: config.OperatorKey,

		blocks:       make(map[uint64]map[common.Hash]*Block), // ADDED
		isPreemptive: config.Preemptive,
	}
	log.Info("[backend:NewWolk] WolkStore")

	storage, err := NewStorage(config, genesisConfig, &wstore)

	if err != nil {
		return cs, err
	}
	log.Info("[backend:NewWolk] NewStorage", "nodeID", wstore.consensusIdx)
	wstore.Storage = storage

	wstore.config = config

	wstore.wolktxpool.Start(&wstore)

	//TODO: need to check. it might be better think about mem size too
	handlerlimit := runtime.NumCPU() * 1000
	if handlerlimit > 1500 {
		handlerlimit = 1500
	}

	log.Info("[backend:NewWolk] set up ProtocolManager", "nodeID", wstore.consensusIdx)
	if wstore.protocolManager, err = NewProtocolManager(config, wstore.eventMux, wstore.wolktxpool, &wstore); err != nil {
		return nil, fmt.Errorf("[backend:NewWolk] NewProtocolManager: %v", err)
	}

	//handlerlimit = int(handlerlimit / maxConcurrentChains)
	server := NewHttpServer(&wstore, config.HTTPPort, handlerlimit)
	server.Start()
	wstore.setConsensusStatus(consensusStatusStart)

	wstore.policy = policy
	wstore.shutdownChan = make(chan bool)

	genesisBlock, _, err := wstore.Indexer.GetBlockByNumber(1)
	if genesisBlock != nil {
		log.Warn("[backend:NewWolk] genesisBlock found in indexer", "nodeID", wstore.consensusIdx)
		wstore.genesisBlock = genesisBlock
	} else {
		genesisBlock, _, err = CreateGenesis(wstore.Storage, wstore.genesis, false)
		if err != nil {
			log.Error("[backend:NewWolk] genesisBlock generation error", "nodeID", wstore.consensusIdx, "err", err)
		} else {
			wstore.genesisBlock = genesisBlock
			log.Trace("[backend:NewWolk] set genesis block ok", "nodeID", wstore.consensusIdx)
		}
	}

	// returns a finalized block -- at the beginning, this is 0
	lastbn, err := wstore.Indexer.GetLastKnownBlockNumber()
	if err != nil {
		return cs, nil
	}
	if lastbn > 0 {
		last, lastok, errLast := wstore.Indexer.GetBlockByNumber(lastbn)
		if errLast != nil {
			log.Error(fmt.Sprintf("[backend:NewWolk] indexer.GetBlockByNumber - unable to get block #%d from indexer", lastbn), "err", errLast)
			return nil, errLast
		} else if !lastok {
			log.Error(fmt.Sprintf("[backend:NewWolk] indexer.GetBlockByNumber - unable to get block #%d from indexer NOT OK", lastbn))
			return nil, errLast
		}
		wstore.setLastBlock(last)
		wstore.setCurrentBlock(last)
	} else {
		wstore.setLastBlock(genesisBlock)
		wstore.setCurrentBlock(genesisBlock)
	}

	dctx, cancel := context.WithCancel(context.Background())
	wstore.dispatcher = NewDispatcher(1000, cancel)
	if err := wstore.StartDispatcher(dctx); err != nil {
		panic(err)
	}

	return &wstore, nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (wstore *WolkStore) Protocols() (p []p2p.Protocol) {
	return wstore.protocolManager.SubProtocols
}

// APIs implements node.Service
func (wstore *WolkStore) APIs() (apis []rpc.API) {
	return []rpc.API{}
}

// Start kicks off the Consensus/Storage Node proceses
func (wstore *WolkStore) Start(srvr *p2p.Server) error {
	// Figure out a max peers count based on the server limits

	wstore.hangForever = make(chan struct{})
	wstore.Storage.Start()
	log.Info("[backend:Start]")
	go wstore.ChainStatusLoop()
	wstore.protocolManager.Start(srvr)
	lastbn, err := wstore.Indexer.GetLastKnownBlockNumber()
	if err != nil {
		return err
	} else if lastbn == 0 {
		err := wstore.MakeGenesis()
		if err != nil {
			log.Error("[backend:Start] MakeGenesis, err", err)
		}
		log.Info("[backend:Start]", "PreemptiveStateDB", wstore.PreemptiveStateDB)
	} else {
		log.Info("[backend:Start]", "LastKnownBlockNumber", lastbn)
	}
	return nil
}

// ChainStatusLoop check blockchain availability status in infinite loop.
func (wstore *WolkStore) ChainStatusLoop() {
	// synchronise requires (1) peers to be available (2) blocks are readable from cloudstore
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			wstore.chainStatus(context.Background())
		}
	}
}

// Stop implements node.Service, terminating all internal goroutines used by the protocol.
func (wstore *WolkStore) Stop() error {
	log.Info("[backend:Stop] Wolk Blockchain manager stopped")
	wstore.protocolManager.Stop()
	if wstore.consensusIdx < 0 {
		return nil
	}
	//	close(s.shutdownChan)

	close(wstore.hangForever)
	return nil
}

// Close halts all Storage operations
func (wstore *WolkStore) Close() {
	wstore.Storage.Close()
}

func (wstore *WolkStore) chainStatus(ctx context.Context) (err error) {
	// TODO: check if storage is available?
	if wstore.protocolManager.NumPeers() < minDesiredPeerCount {
		if wstore.isChainReady() {
			wstore.setChainReady(false)
		}
		return
	} else if wstore.isChainReady() {
		return
	}

	var update bool
	// This is blocking call -= to be run by both storage + consensus
	wstore.setConsensusStatus(consensusStatusSynchroniseCert)
	update, err = wstore.protocolManager.synchroniseCert(wstore.protocolManager.BestPeer())
	if err != nil {
		log.Error(fmt.Sprintf("[backend:chainStatus] synchroniseCert ERR [Node%v] %v", wstore.consensusIdx, err))
		return err
	}

	wstore.setConsensusStatus(consensusStatusGettingReady)
	log.Info(fmt.Sprintf("*** [backend:chainStatus] synchronise [Node%v] Update: %t", wstore.consensusIdx, update))
	lastbn := wstore.round()
	if lastbn > 1 { // CHECK: Consensus
		latestBlockHash, newStateDBError := wstore.Indexer.GetBlockHash(lastbn)
		if newStateDBError != nil {
			log.Error(fmt.Sprintf("*** [backend:chainStatus] NewStateDB - unable to get block #%d from indexer [Node%v] update", lastbn, wstore.consensusIdx))
			return newStateDBError
		}
		var preemptiveStateDB *StateDB
		preemptiveStateDB, newStateDBError = NewStateDB(ctx, wstore.Storage, latestBlockHash)
		if newStateDBError != nil {
			log.Error(fmt.Sprintf("*** [backend:chainStatus] NewStateDB - unable to generate StateDB #%d from indexer [Node%v] update", lastbn, wstore.consensusIdx))
			return newStateDBError
		}
		wstore.muPreemptiveStateDB.Lock()
		wstore.PreemptiveStateDB = preemptiveStateDB
		wstore.muPreemptiveStateDB.Unlock()
	}

	wstore.setChainReady(true)

	return nil
}

func (wstore *WolkStore) setConsensusStatus(status string) time.Time {
	wstore.consensusStatusMu.Lock()
	defer wstore.consensusStatusMu.Unlock()
	wstore.consensusStatus = status
	wstore.consensusStatusUpdatedTS = time.Now().Unix()
	return time.Now()
}

func (wstore *WolkStore) getConsensusStatus(round uint64) (status string, secondsAgo int64, step uint64,
	isVoting uint64, isMalicious uint64, expectedTentative int, expectedFinal int, tokenWeight uint64,
	hasBlock bool, reserved string) {
	wstore.consensusStatusMu.RLock()
	defer wstore.consensusStatusMu.RUnlock()
	isVoting = IsVoting
	isMalicious = Malicious
	hasBlock = false

	alg := wstore.protocolManager.getConsensusEngine(round)
	if alg != nil {
		var address common.Address
		step = alg.step
		_, maxProposal, _ := alg.getVoteLastStep()
		expectedTentative, expectedFinal = alg.expectedThresholds()

		tokenWeight = alg.weight(address, round)
		if maxProposal != nil {
			b := alg.getBlock(maxProposal.BlockHash)
			if b != nil {
				hasBlock = true
			}
		}
	}
	return wstore.consensusStatus, time.Now().Unix() - wstore.consensusStatusUpdatedTS, step,
		isVoting, isMalicious, expectedTentative, expectedFinal, tokenWeight,
		hasBlock, reserved
}

// GenesisFile returns the actual local genesis.json filename
func (wstore *WolkStore) GenesisFile() (genesisFile string) {
	return wstore.config.GenesisFile
}

func (wstore *WolkStore) setChainReady(ready bool) {
	wstore.muBlockchainReady.Lock()
	wstore.blockchainReady = ready
	wstore.muBlockchainReady.Unlock()
}

func (wstore *WolkStore) isChainReady() (ready bool) {
	wstore.muBlockchainReady.RLock()
	ready = wstore.blockchainReady
	wstore.muBlockchainReady.RUnlock()
	return ready
}

//called consensusprocess
func (wstore *WolkStore) setCandidateBlock(candidate *Block) {
	if candidate.Number() >= wstore.currentRound() {
		wstore.muCurrentBlock.Lock()
		wstore.current = candidate
		wstore.muCurrentBlock.Unlock()
	}
}

//called by insertchain
func (wstore *WolkStore) setCurrentBlock(curr *Block) {
	wstore.muCurrentBlock.Lock()
	wstore.current = curr
	wstore.muCurrentBlock.Unlock()
}

func (wstore *WolkStore) getCandidateBlock() (candidate *Block) {
	wstore.muCurrentBlock.RLock()
	candidate = wstore.current
	wstore.muCurrentBlock.RUnlock()
	return candidate
}

//filtering cert with less than Tentative vote
func (wstore *WolkStore) validCertificate(ctx context.Context, cert *VoteMessage) (h common.Hash, finalized bool, votecnt int, err error) {
	h, cv := cert.CountVotes(voteNext, cert.Seed, cert.BlockNumber)
	finalized = cv >= expectedTokensFinal
	if cv >= expectedTokensTentative {
		return cert.BlockHash, finalized, cv, nil
	}
	return h, finalized, cv, fmt.Errorf("Invalid cert %d < %d (bn %d step %d)", cv, expectedTokensTentative, cert.BlockNumber, voteCert)
}

func (wstore *WolkStore) processCerts(ctx context.Context, certs []*VoteMessage) (err error) {
	log.Info("[backend:processCerts] START", "len(certs)", len(certs))

	var statedb *StateDB
	var beststateDB *StateDB
	var preemptiveDB *StateDB
	for _, cert := range certs {
		if cert == nil {
			return nil
		}
		blockHash, finalized, voteCnt, err := wstore.validCertificate(ctx, cert)
		if err != nil {
			log.Error("[backend:processCerts] invalid certificate", "cert.BlockNumber", cert.BlockNumber, "err", err)
		} else {

			if finalized && cert.BlockNumber > wstore.LastFinalized() {
				err = wstore.setFinalizedPath(cert.BlockNumber, cert.BlockHash, cert.ParentHash)
				if err != nil {
					//TODO
				}

				missingCerts, err := wstore.Indexer.processFinalized(cert)
				if err != nil {

				} else if len(missingCerts) > 0 {
					// finalized path is missing
					wstore.protocolManager.requestMissingCerts(missingCerts)
				} else {
					wstore.setLastFinalized(cert.BlockNumber, cert.BlockHash)
				}
			}

			if cert.BlockNumber >= wstore.currentRound() {
				//catching up
				cert.dump(fmt.Sprintf("processCerts CERT %d %x", wstore.currentRound(), blockHash))
				statedb, err = NewStateDB(ctx, wstore.Storage, blockHash)
				if err != nil {
					log.Error("[backend:processCerts] NewStateDB", "err", err)
					// return err
				} else if statedb != nil {
					log.Info("[backend:processCerts]", "bn", cert.BlockNumber, "blockHash", blockHash) // , "block", statedb.block

					wstore.Storage.StoreVerifiedBlock(statedb.block, statedb, voteCnt)
					bestHash, vcnt, found := wstore.Storage.getBestHash(cert.BlockNumber)
					if !found || bytes.Equal(bestHash.Bytes(), blockHash.Bytes()) || voteCnt >= vcnt {
						//only update if bestblock
						log.Info("[backend:processCerts] record", "bn", cert.BlockNumber, "blockHash", blockHash) // , "block", statedb.block
						wstore.recordBlock(statedb.block)
						beststateDB = statedb
						if finalized {
							preemptiveDB = statedb
						}
					}
				}
			} else {
				//received a past cert, if the cert is from a missing descendant block of a finaled round, record the block to indexer!
				expectedCertHash, err := wstore.Indexer.getFinalizedPath(cert.BlockNumber)
				if err == nil {
					if bytes.Equal(cert.BlockHash.Bytes(), expectedCertHash.Bytes()) {
						ctx := context.TODO()
						prevBlk, err := wstore.GetBlockByHash(ctx, expectedCertHash, BlockReadOptions{ReadFromCloudstore: true})
						if err != nil || prevBlk == nil {
							log.Error(fmt.Sprintf("GetBlockByHash looking %x but err", expectedCertHash), "err", err)
						} else {
							log.Info("[backend:processCerts] record corrected block", "BN", cert.BlockNumber, "Bhash", expectedCertHash)
							wstore.recordBlock(prevBlk)
						}
					}
				}
			}

			wstore.Indexer.StoreCertificate(cert)

			eng := wstore.protocolManager.getConsensusEngine(cert.BlockNumber)
			if eng != nil {
				eng.addCertificate(cert)
			}
		}
	}

	if beststateDB != nil {
		wstore.muPreemptiveStateDB.Lock()
		wstore.CurrentStateDB = beststateDB.Copy()
		if preemptiveDB != nil {
			wstore.PreemptiveStateDB = preemptiveDB.Copy()
		}
		wstore.muPreemptiveStateDB.Unlock()
	}

	return nil
}

// TxPool returns the transaction pool live in memory
func (wstore *WolkStore) TxPool() *TxPool {
	return wstore.wolktxpool
}

func (wstore *WolkStore) validateTx(rtx *Transaction) (err error) {
	// TODO: compare how shallow signature validation relates to deep state transaction validation
	return nil
}

// assumes this has already been validated
func (wstore *WolkStore) addTransactionToPool(tx *Transaction) (err error) {
	wstore.wolktxpool.addTransactionToPool(tx)
	return nil
}

// GetTransaction retrieves a transaction from Indexer and if not present, from cloudstore
func (wstore *WolkStore) GetTransaction(ctx context.Context, txhash common.Hash) (tx *Transaction, blockNumber uint64, status string, ok bool, err error) {
	//status := "Not Found"
	tx, blockNumber, ok, err = wstore.Indexer.GetTransaction(txhash)
	if err != nil {
		return tx, blockNumber, transactionStatusNotFound, ok, err
	} else if ok {
		return tx, blockNumber, transactionStatusApplied, true, nil
	}
	//TODO: Check somewhere for errored TX case
	txbytes, ok, err := wstore.Storage.GetChunk(ctx, txhash.Bytes())
	if err != nil {
		return tx, blockNumber, transactionStatusNotFound, ok, err
	} else if !ok {
		//Check TXPool
		if wstore.wolktxpool.HasTransaction(txhash) {
			return wstore.wolktxpool.GetTransaction(txhash), blockNumber, "Pending", true, nil
		}
		//if not found in TXPool
		return tx, blockNumber, transactionStatusNotFound, ok, err
	}
	tx, err = DecodeRLPTransaction(txbytes)
	if err != nil {
		return tx, 0, transactionStatusNotFound, false, err
	}
	return tx, 0, transactionStatusProposed, true, nil
}

// SendRawTransactions adds a list of transactions to the transaction pool
func (wstore *WolkStore) SendRawTransactions(txs []*Transaction) (err error) {
	for _, tx := range txs {
		_, err := wstore.SendRawTransaction(tx)
		if err != nil {
			log.Warn("[backend:SendRawTransactions]", "err", err)
			return err
		}
	}
	return nil
}

// SendRawTransactions2 adds a list of transactions to the transaction pool
func (wstore *WolkStore) SendRawTransactions2(txlist []*Transaction) map[common.Hash]error {
	txsresult := make(map[common.Hash]error)
	for _, tx := range txlist {
		txhash, err := wstore.SendRawTransaction(tx)
		txsresult[txhash] = err
	}
	return txsresult
}

// SendRawTransaction adds a single transaction to the transaction pool
func (wstore *WolkStore) SendRawTransaction(tx *Transaction) (hash common.Hash, err error) {
	log.Info("[backend:SendRawTransaction] skipping tx.ValidateTx, TODO: re-enable!!")
	/*
		ok, err := tx.ValidateTx()
		if err != nil {
			log.Error("[backend:SendRawTransaction] ValidateTx", "err", err)
			return hash, fmt.Errorf("[backend:SendRawTransaction] %s", err)
		}
		if !ok {
			log.Error("[backend:SendRawTransaction] NOT OK", "ok", false)
			return hash, fmt.Errorf("[backend:SendRawTransaction] transaction(%x) not validated", tx.Hash())
		}
	*/
	log.Info("[backend:SendRawTransaction]", "tx", fmt.Sprintf("%v", tx), "tx", tx, "NUM PEERS", wstore.protocolManager.NumPeers())

	// ***** try a StateDB change! ****

	if wstore.GetIsPremptive() {
		wstore.muPreemptiveStateDB.Lock()
		defer wstore.muPreemptiveStateDB.Unlock()
		if wstore.PreemptiveStateDB != nil {
			err = wstore.PreemptiveStateDB.ApplyTransaction(context.Background(), tx, "R")
			if err != nil {
				log.Error("[backend:SendRawTransaction] Preemptive ApplyTransaction ERR", "err", err)
				return hash, err
			}
			log.Info(" ****** [backend:SendRawTransaction] PreemptiveDB: APPLIED TRANSACTION", "TID", len(wstore.PreemptiveStateDB.committedTxes), "txhash", tx.Hash())
		}
	}

	// write to cloudstore!
	_, err = wstore.Storage.SetChunk(context.TODO(), tx.Bytes())
	if err != nil {
		return hash, fmt.Errorf("[backend:SendRawTransaction] %s", err)
	}

	err = wstore.addTransactionToPool(tx)
	if err != nil {
		return hash, fmt.Errorf("[backend:SendRawTransaction] %s", err)
	}

	hash = tx.Hash()

	//log.Trace("[backend:SendRawTransaction] BroadcastTx", "txhash", hash)
	wstore.protocolManager.BroadcastTx(hash, tx)
	return hash, nil
}

// GetLock locks the internal state of WolkStore
func (wstore *WolkStore) GetLock() {
	wstore.mu.Lock()
}

// ReleaseLock unlocks the internal state of WolkStore
func (wstore *WolkStore) ReleaseLock() {
	wstore.mu.Unlock()
}

// RLock read locks the internal state of WolkStore
func (wstore *WolkStore) RLock() {
	wstore.mu.RLock()
}

// RUnlock read unlocks the internal state of WolkStore
func (wstore *WolkStore) RUnlock() {
	wstore.mu.RUnlock()
}

// ReceiveTransaction adds to StateDB in progress with state.ApplyTransaction
func (wstore *WolkStore) ReceiveTransaction(tx *Transaction) (err error) {
	if wstore.wolktxpool.HasTransaction(tx.Hash()) {
		// we already handled it!
		//TODO: should we broadcast here?
		log.Trace("[backend:ReceiveTransaction] - already processed", "txhash", tx.Hash())
		return nil
	}
	if wstore.isPreemptive {
		wstore.muPreemptiveStateDB.Lock()
		defer wstore.muPreemptiveStateDB.Unlock()
		if wstore.PreemptiveStateDB != nil {
			err = wstore.PreemptiveStateDB.ApplyTransaction(context.Background(), tx, "P")
			if err != nil {
				log.Warn("[backend:ReceiveTransaction] ApplyTransaction", "idx", wstore.consensusIdx, "err", err)
				return err
			}
			if wstore.consensusIdx == 1 {
				log.Debug("[backend:ReceiveTransaction] -- APPLIED***********", "idx", wstore.consensusIdx, "tx.Hash", tx.Hash().Hex(), "err", err)
			}
		}

	} else {
		// log.Trace("[backend:ReceiveTransaction] -- DID NOT apply", "idx", wstore.consensusIdx, "tx.Hash", tx.Hash())
	}

	err = wstore.addTransactionToPool(tx)
	if err != nil {
		log.Error("[backend:ReceiveTransaction] ERROR Adding tx to pool", "tx", tx.Hash())
		return err
	}

	if wstore.consensusIdx == 1 {
		log.Trace("[backend:SendRawTransaction] BroadcastTx", "txhash", tx.Hash().Hex())
	}
	wstore.protocolManager.BroadcastTx(tx.Hash(), tx)
	return nil
}

// ValidateBlockSig validates ParentHash + ValidateTx Sig (Potentially different for POA/Consensus case)
func (wstore *WolkStore) ValidateBlockSig(parentBlock *Block, newBlock *Block) bool {
	//TODO: check block signiture is valid - Differnet for POA/Consensus case
	if bytes.Compare(parentBlock.Hash().Bytes(), newBlock.ParentHash.Bytes()) != 0 {
		log.Error("[backend:ValidateBlock] false 1", "parentBlock", parentBlock, "newBlock", newBlock)
		return false
	}

	// for all the transactions in the block, do basic validation
	for _, tx := range newBlock.Transactions {
		validated, err := tx.ValidateTx()
		if err != nil {
			log.Error("[backend:ValidateBlock] false 2", "parentBlock", parentBlock, "newBlock", newBlock)
			return false
		}
		if !validated {
			log.Error("[backend:ValidateBlock] false 3", "parentBlock", parentBlock, "newBlock", newBlock)
			return false
		}
	}
	log.Trace("[backend:ValidateBlock] VALIDATED", "parentBlock", parentBlock, "newBlock", newBlock)
	return true
}

// VerifyTransition validates state transition
func (wstore *WolkStore) VerifyTransition(ctx context.Context, parentBlock *Block, newBlock *Block) (verified bool, statedb *StateDB, err error) {
	writeToCloudstore := false
	//log.Debug("[backend:VerifyTransition]", "parent", parentBlock.hash, "newBlock", newBlock.hash)
	start := time.Now()
	statedb, err = NewStateDB(ctx, wstore.Storage, parentBlock.Hash())
	if err != nil {
		log.Error("VerifyTransition NewStateDB", "err", err)
		return false, statedb, err
	}

	var expectedBlock *Block

	expectedBlock, err = statedb.MakeBlock(ctx, newBlock.Transactions, wstore.policy, writeToCloudstore)
	if err != nil {
		log.Error("VerifyTransition MakeBlock", "err", err)
		return false, statedb, err
	}
	log.Info("[backend:VerifyTransition] MakeBlock", "time", time.Since(start))

	//log.Trace("[backend:VerifyTransition]", "expectedBlock", expectedBlock)
	start = time.Now()
	verified = expectedBlock.IsIdentical(newBlock)
	log.Info("[backend:VerifyTransition] IsIdentical", "time", time.Since(start), "verified", verified)
	if !verified {
		log.Error("[backend:VerifyTransition] NOT VERIFIED", "parentBlock.Hash()", parentBlock.Hash(), "parentBlock", len(parentBlock.Transactions), "expectedBlock", expectedBlock, "newBlock", newBlock)
		return false, statedb, nil
	}
	statedb.block = newBlock //returning statedb
	return verified, statedb, nil
}

// VerifyBlockTransition validates state transition using newBlock
func (wstore *WolkStore) VerifyBlockTransition(ctx context.Context, blk *Block) (verified bool, statedb *StateDB, err error) {
	writeToCloudstore := false
	parentHash := blk.ParentHash
	log.Debug("[backend:VerifyBlockTransition]", "parentHash", parentHash.Hex(), "block", blk)
	statedb, err = NewStateDB(ctx, wstore.Storage, parentHash)
	if err != nil {
		log.Error("VerifyBlockTransition NewStateDB", "err", err)
		return false, statedb, err
	}

	var expectedBlock *Block
	expectedBlock, err = statedb.MakeBlock(ctx, blk.Transactions, wstore.policy, writeToCloudstore)
	if err != nil {
		log.Error("VerifyBlockTransition MakeBlock", "err", err)
		return false, statedb, err
	}

	log.Trace("[backend:VerifyBlockTransition]", "expectedBlock", expectedBlock)
	verified = expectedBlock.IsIdentical(blk)
	if !verified {
		log.Error("[backend:VerifyBlockTransition] NOT VERIFIED", "parentHash", parentHash.Hex(), "expectedBlock", len(expectedBlock.Transactions), "newBlock", len(blk.Transactions))
		if wstore.consensusIdx == 1 {
			//	statedb.keyStorage.Dump()
		}
		return false, statedb, fmt.Errorf("NOT VERIFIED %x", blk.Hash()) //shoudl return "NOT VERIFIED" error
	}

	statedb.block = blk //returning statedb
	return verified, statedb, nil
}

func (wstore *WolkStore) getPeerIPs() []string {
	return wstore.protocolManager.GetPeerIPs()
}

// WriteBlock stores block in Cloudstore.
// This does not write to the Indexer ( recordBlock does this )
func (wstore *WolkStore) WriteBlock(b *Block) (err error) {
	data := b.Bytes()

	if useFileHash {
		var putwg sync.WaitGroup
		putstart := time.Now()
		putwg.Add(1)
		// this is not blocking!
		ctx := context.Background()
		val := b.Bytes()
		var fileHash []byte
		fileHash, err = wstore.Storage.PutFile(ctx, val, &putwg)
		if err != nil {
			return err
		}

		putwg.Wait()
		log.Info("[backend:WriteBlock] SetChunk", "fileHash", fmt.Sprintf("%x", fileHash), "tm", time.Since(putstart))
	} else {
		var chunkID common.Hash
		chunkID, err = wstore.Storage.SetChunk(context.TODO(), data)
		if err != nil {
			log.Error("[backend:WriteBlock] SetChunk ERR", "err", err)
			return err
		}
		if false {
			log.Info("[backend:WriteBlock] SetChunk", "blockhash", fmt.Sprintf("%x", chunkID), "block", b.String())
		}
	}

	if useFileHash {
		// now, for all the txs, make them chunks!
		for _, tx := range b.Transactions {
			txhash, errtx := wstore.Storage.SetChunk(context.TODO(), tx.Bytes())
			if errtx != nil {
				return errtx
			}
			log.Trace("[backend:WriteBlock] WROTE TX CHUNK!!!! ", "txhash", txhash.Hex(), "tx.Hash()", tx.Hash())
		}
	} else {
		// now, for all the txs, make them chunks!
		var rawchunks []*cloud.RawChunk
		//putstart := time.Now()
		var wg sync.WaitGroup
		for _, tx := range b.Transactions {
			wg.Add(1)

			ctx := context.Background()
			val := tx.Bytes()
			rawchunk := &cloud.RawChunk{Value: val}
			txhash := wolkcommon.Computehash(val)
			_, err := wstore.Storage.PutChunk(ctx, rawchunk, &wg)
			if err != nil {
				return err
			}
			rawchunks = append(rawchunks, rawchunk)
			log.Debug("[backend:WriteBlock] WROTE TX CHUNK!!!! ", "txhash", fmt.Sprintf("%x", txhash), "tx.Hash()", tx.Hash())
		}
		//log.Debug("[backend:WriteBlock] tx writes sent", "len(b.Transactions)", len(b.Transactions), "tm", time.Since(putstart))
		wg.Wait()
		log.Info("[backend:WriteBlock] tx writes done", "len(b.Transactions)", len(b.Transactions))
	}

	return nil
}

// GetBlockByNumber uses Indexer to return a block
func (wstore *WolkStore) GetBlockByNumber(blocknum uint64) (*Block, bool, error) {
	if blocknum < 1 {
		blocknum, _ = wstore.LatestBlockNumber()
	}
	if blocknum == 1 {
		return wstore.genesisBlock, true, nil
	}
	block, ok, err := wstore.Indexer.GetBlockByNumber(uint64(blocknum))
	if err != nil {
		return block, ok, fmt.Errorf("[backend:GetBlockByNumber] GetBlockByNumber(%d) %s", blocknum, err)
	}
	return block, ok, nil
}

// BlockReadOptions is used as options parameter to GetBlockByHash -- we could consider reading from StateDB (memory) and Indexer (local) as well
type BlockReadOptions struct {
	ReadFromCloudstore bool
}

// GetBlockByHash using { ConsensusEngine, Cloudstore } to return a block from memory as specified by options
// If no options are specified, a cascade of ConsensusEngine > Cloudstore are used
// if block is not found, nil is returned
// If err is returned from Cloudstore, it is returned, but no error possibility exists from ConsensusEngine
func (wstore *WolkStore) GetBlockByHash(ctx context.Context, blockHash common.Hash, options BlockReadOptions) (*Block, error) {
	// default assumption is ReadFromCloudstore: true
	if !(options.ReadFromCloudstore) {

		options.ReadFromCloudstore = true
	}

	if options.ReadFromCloudstore {
		block := new(Block)
		var blockRaw []byte
		var ok bool
		var err error
		if useFileHash {
			ctx := context.Background()
			blockRaw, err = wstore.Storage.GetFile(ctx, blockHash.Bytes())
			if len(blockRaw) > 0 {
				ok = true
			}
		} else {
			blockRaw, ok, err = wstore.Storage.GetChunk(ctx, blockHash.Bytes())
		}
		if err != nil {
			log.Error("[backend:GetBlockByHash] Error retrieving Block ", "error", err)
			return block, fmt.Errorf("[backend:GetBlockByHash] %s", err)
		} else if !ok {
			log.Debug("[backend:GetBlockByHash] Block not found", "blockhash", blockHash.Hex())
			return nil, nil
		} else if blockRaw == nil {
			log.Debug("[backend:GetBlockByHash] Block nil", "blockhash", blockHash.Hex())
			return block, nil
		}

		if err := rlp.Decode(bytes.NewReader(blockRaw), block); err != nil {
			log.Error("[backend:GetBlockByHash] Invalid block RLP", "hash", blockHash.Hex(), "err", err)
			return nil, fmt.Errorf("[backend:GetBlockByHash] %s", err)
		}
		return block, nil
	}

	return nil, nil
}

// getStateByNumber returns the lastest (Cached) stateDB if found
func (wstore *WolkStore) getStateByNumber(ctx context.Context, blockNumber int) (state *StateDB, ok bool, err error) {
	/*
		if blockNumber == 0 && isPreemptive {
			if wstore.PreemptiveStateDB != nil {
				return wstore.PreemptiveStateDB, true, nil
			}
		}
	*/
	state, ok, err = wstore.Storage.getStateDB(ctx, blockNumber)
	if err != nil {
		return state, ok, fmt.Errorf("[backend_names:getStateDBByBlockNumber] %s", err)
	}
	if !ok {
		return state, ok, fmt.Errorf("[backend_names:getStateDBByBlockNumber] no block")
	}
	return state, ok, nil
}

// GetOperatorAddress returns the consensus node's ECDSA address
func (wstore *WolkStore) GetOperatorAddress() common.Address {
	return crypto.PubkeyToAddress(wstore.operatorKey.PublicKey())
}

// GetNode returns node information (with a proof) from the blockchain
func (wstore *WolkStore) GetNode(node uint64, options *RequestOptions) (n *RegisteredNode, ok bool, proof *Proof, err error) {
	log.Info("[backend:GetNode]", "node", node)
	s, ok, err := wstore.getStateByNumber(context.Background(), options.BlockNumber)
	if err != nil {
		return n, false, proof, err
	} else if !ok {
		return n, false, proof, fmt.Errorf("[backend:GetNode] number not found")
	}

	n0, err := s.GetRegisteredNode(context.Background(), node)
	if err != nil {
		return n, false, proof, err
	}
	log.Info("[backend:GetNode]", "n0", n0)
	return &n0, true, proof, nil
}

// GetIndex returns the node's consensus index
func (wstore *WolkStore) GetIndex() (idx int) {
	return wstore.consensusIdx
}

// GetBalance returns the balance of the address from the blockchain
func (wstore *WolkStore) GetBalance(address common.Address, blockNumber int) (balance uint64, ok bool, err error) {
	account, ok, err := wstore.GetAccount(address, blockNumber)
	if err != nil {
		return balance, ok, err
	} else if !ok {
		return balance, false, nil
	}
	return account.Balance, true, nil
}

// GetAccount returns an account object from the blockchain
func (wstore *WolkStore) GetAccount(address common.Address, blockNumber int) (account *Account, ok bool, err error) {
	log.Trace("[backend:GetAccount]", "address", address)
	s, ok, err := wstore.getStateByNumber(context.Background(), blockNumber)
	if err != nil {
		return account, false, err
	} else if !ok {
		return account, false, fmt.Errorf("[backend:GetAccount] Block number not found")
	}
	account, ok, err = s.GetAccount(context.Background(), address)
	if err != nil {
		return account, false, err
	} else if !ok {
		return account, false, nil
	}
	return account, true, nil
}

// GetLastBlockMintedSecondsAgo returns the Unix timestamp of when the last block was minted
func (wstore *WolkStore) GetLastBlockMintedSecondsAgo() (lastBlockMintedSecondsAgo int64) {
	wstore.mu.RLock()
	lastBlockMintedSecondsAgo = time.Now().Unix() - wstore.mintedTS
	wstore.mu.RUnlock()
	return lastBlockMintedSecondsAgo
}

// GetPendingTxCount returns the total number of transactions held in the transaction pool
func (wstore *WolkStore) GetPendingTxCount() int64 {
	return wstore.TxPool().TxCount()
}

// GetIsPremptive returns whether consensus node operates with isPreemptive
func (wstore *WolkStore) GetIsPremptive() (isPreemptive bool) {
	wstore.mu.RLock()
	isPreemptive = wstore.isPreemptive
	wstore.mu.RUnlock()
	return isPreemptive
}

func (wstore *WolkStore) extendChain(ctx context.Context, block *Block, certvote *VoteMessage) (cert *VoteMessage, err error) {
	statedb, err := NewStateDB(ctx, wstore.Storage, block.Hash())
	if err != nil {
		log.Error("[backend:insertChain] NewStateDB", "err", err)
		return cert, err
	}
	log.Info("[backend:extendChain]", "bn", block.BlockNumber, "h", block.Hash())
	//both tentative/finalized block will clear CurrentStateDB
	wstore.muPreemptiveStateDB.Lock()
	wstore.CurrentStateDB = statedb.Copy()
	wstore.muPreemptiveStateDB.Unlock()
	wstore.recordBlock(block)
	certvote.dump(fmt.Sprintf("---- extendChain uncompact CERT %d", block.BlockNumber))
	cert = certvote.MakeCertificate(block.Hash())
	cert.dump(fmt.Sprintf("---- extendChain compact CERT %d", block.BlockNumber))
	err = wstore.Indexer.StoreCertificate(cert)
	if err != nil {
		log.Error("[backend:extendChain] StoreCertificate", "err", err)
		return cert, err
	}

	_, finalized, voteCnt, err := wstore.validCertificate(ctx, cert)
	if err == nil && finalized {
		//update preemptiveDB if finalzed
		wstore.muPreemptiveStateDB.Lock()
		wstore.PreemptiveStateDB = statedb.Copy()
		wstore.muPreemptiveStateDB.Unlock()
		wstore.setFinalizedPath(cert.BlockNumber, cert.BlockHash, cert.ParentHash)
		lastF := wstore.LastFinalized()
		if cert.BlockNumber-lastF >= 5 {
			certs := make([]*VoteMessage, 0)
			certs = append(certs, cert)
			go wstore.protocolManager.ForcedBroadcastCerts(certs)
		}
		wstore.setLastFinalized(cert.BlockNumber, cert.BlockHash)
	}
	wstore.Storage.StoreVerifiedBlock(block, statedb, voteCnt)
	return cert, nil
}

func (wstore *WolkStore) setFinalizedPath(blockNumber uint64, bhash common.Hash, prevH common.Hash) (err error) {
	if blockNumber < 1 {
		return nil
	}
	err = wstore.Indexer.setFinalizedRound(blockNumber, bhash)
	if err != nil {
		return err
	}
	log.Info("[indexer:setFinalizedHash]", "BN", blockNumber, "BHash", bhash, "PrevH", prevH)
	prevBN := blockNumber - 1
	prevHash, err := wstore.Indexer.getFinalizedPath(prevBN)
	if err == nil {
		if bytes.Equal(prevHash.Bytes(), prevH.Bytes()) {
			//prevHash found and match. Exit condition
			return nil
		}
	}
	// either (1) prevHash not found, (2) prevHash mismatch, overwite it
	ctx := context.TODO()
	prevBlk, err := wstore.GetBlockByHash(ctx, prevH, BlockReadOptions{ReadFromCloudstore: true})
	if err != nil || prevBlk == nil {
		log.Error(fmt.Sprintf("GetBlockByHash looking %x but err", prevH), "err", err)
		return fmt.Errorf("prevH unavailable")
	}
	prevprevH := prevBlk.ParentHash //grandparentH
	wstore.recordBlock(prevBlk)
	log.Info("[backend:setFinalizedPath] recordBlock - Path Patching! ", "prevBN", prevBN, "prevH", prevH)
	wstore.checkMissingCert(prevBlk)
	return wstore.setFinalizedPath(prevBN, prevH, prevprevH)
}

func (wstore *WolkStore) checkMissingCert(block *Block) {
	fpcerts, err := wstore.Indexer.GetTargetCertificate(block.Number(), block.Hash())
	if err != nil || len(fpcerts) == 0 {
		missingCerts := make(map[uint64]common.Hash)
		missingCerts[block.Number()] = block.Hash()
		log.Info("[backend:checkMissingCert] requestMissingCerts! ", "BN", "certHash", block.Hash())
		//TODO: (OPTIMIZE) batch request for missing cert
		wstore.protocolManager.requestMissingCerts(missingCerts)
	}
}

// this is called in insertchain -- this is only place it should be called  --- EXCEPT for MakeGenesis
func (wstore *WolkStore) recordBlock(block *Block) (err error) {
	if block.NetworkID != wstore.genesis.NetworkID {
		log.Error("[backend:recordBlock] Incorrect NetworkID", "network", block.NetworkID)
		return fmt.Errorf("Incorrect networkID")
	}
	wstore.wolktxpool.Reset(block)
	wstore.Indexer.IndexBlock(block)

	wstore.GetLock()
	if block.Round() >= wstore.currentRound() {
		wstore.mintedTS = time.Now().Unix()
		wstore.setLastBlock(block)    // updating lastblock for both Consensus and poa
		wstore.setCurrentBlock(block) // updating currentblock for both Consensus and poa
	}
	wstore.ReleaseLock()
	return nil
}

func (wstore *WolkStore) LastKnownBlockHash() (h common.Hash) {
	if wstore.lastBlock() == nil {
		return h
	}
	lbn, _ := wstore.LatestBlockNumber()
	b, ok, err := wstore.GetBlockByNumber(uint64(lbn))
	if err == nil && ok {
		return b.Hash()
	}
	return h
}

// LastKnownBlock returns the last known block
func (wstore *WolkStore) LastKnownBlock() (b *Block) {
	// fetch block from indexer
	b, ok, err := wstore.GetBlockByNumber(uint64(wstore.round()))
	if err != nil || !ok {
		return nil
	}
	return b
}

//TODO: clean this
func (wstore *WolkStore) LatestBlockNumber() (blocknum uint64, err error) {
	last := wstore.lastBlock()
	if last != nil {
		return wstore.last.Round(), nil
	}
	return 1, nil // wrong condition, should be 0
}

// returns the latest tentative/finalized block
func (wstore *WolkStore) lastBlock() (last *Block) {
	wstore.muLastBlock.RLock()
	defer wstore.muLastBlock.RUnlock()
	last = wstore.last
	return last
}

// set the latest tentative/finalized block
func (wstore *WolkStore) setLastBlock(last *Block) {
	wstore.muLastBlock.Lock()
	wstore.last = last
	wstore.muLastBlock.Unlock()
}

// Genesis returns the genesis block
func (wstore *WolkStore) Genesis() *Block {
	return wstore.genesisBlock
}

// returns the latest finalized round number.
func (wstore *WolkStore) LastFinalized() (lastFBN uint64) {
	wstore.muLastBlock.RLock()
	defer wstore.muLastBlock.RUnlock()
	return wstore.lastFinalized
}

// returns the latest finalized blockhash.
func (wstore *WolkStore) LastFinalizedHash() (finalizedHash common.Hash) {
	wstore.muLastBlock.RLock()
	defer wstore.muLastBlock.RUnlock()
	return wstore.lastFinalizedHash
}

// set the latest finalized round number.
func (wstore *WolkStore) setLastFinalized(finalizedBlock uint64, finalizedHash common.Hash) {
	wstore.muLastBlock.Lock()
	defer wstore.muLastBlock.Unlock()
	wstore.lastFinalized = finalizedBlock
	wstore.lastFinalizedHash = finalizedHash
}

// round returns the latest round number.
func (wstore *WolkStore) round() uint64 {
	b := wstore.lastBlock()
	if b != nil {
		return b.Round()
	}
	return 0
}

func (wstore *WolkStore) currentRound() uint64 {
	b := wstore.lastBlock()
	if b != nil {
		return b.Round() + 1
	}
	return 1 // TODO: Michael to review in backend test
}

// NumPeers returns the number of peers the node currently has
func (wstore *WolkStore) NumPeers() int {
	return wstore.protocolManager.NumPeers()
}

func (wstore *WolkStore) GetPeerList() []string {
	return wstore.protocolManager.GetPeerList()
}

func (wstore *WolkStore) resolveFork(fork *Block) {
	wstore.mu.Lock()
	defer wstore.mu.Unlock()
	wstore.setLastBlock(fork)
}

// MakeGenesis uses the genesis configuration loaded to write a genesis block locally and to cloudstore
func (wstore *WolkStore) MakeGenesis() (err error) {
	// mint the genesis block
	genesisBlock, genesisState, err := CreateGenesis(wstore.Storage, wstore.genesis, true) //true so that everyone is writing genesis smt chunk to cloudstore
	if err != nil {
		log.Error("[backend:CreateGenesis] CreateGenesis", "err", err)
		return err
	}

	log.Error("[backend:MakeGenesis] START")
	genesisState.block = genesisBlock
	wstore.recordBlock(genesisBlock)

	// try writing the genesisBlock over and over until you can retrieve it!
	done := false
	st := time.Now()
	for tries := 0; !done; tries++ {
		log.Error("[backend:MakeGenesis] attempt", "tries", tries)
		if !wstore.Storage.isStorageReady() {
			time.Sleep(1 * time.Second)
		} else {
			log.Error("[backend:MakeGenesis] WriteBlock", "tries", tries)
			wstore.WriteBlock(genesisBlock)
			blc, ok, err := wstore.GetBlockByNumber(1)
			if err != nil {
				log.Error("[backend:MakeGenesis] GetBlockByNumber", "err", err)
				tries++
			} else if !ok {
				log.Error("[backend:MakeGenesis] Genesis Block Not OK")
				tries++
			} else if bytes.Compare(blc.Hash().Bytes(), genesisBlock.Hash().Bytes()) == 0 {
				done = true
			} else {
				log.Error("[backend:MakeGenesis] GetBlockByNumber", "err", err)
			}
		}
		tries++
		if tries > 100 {
			log.Error("[backend:MakeGenesis] COULD NOT WRITE GENESIS BLOCK", "tm", time.Since(st))
			os.Exit(0)
		}
	}

	wstore.muPreemptiveStateDB.Lock()
	wstore.PreemptiveStateDB = genesisState
	wstore.CurrentStateDB = genesisState.Copy()
	wstore.recordBlock(genesisBlock)
	wstore.Storage.StoreVerifiedBlock(genesisBlock, genesisState, 150) //genesis is finalized by default
	log.Info("[backend:createGenesisBlock] STORED createGenesisState", "statedb", genesisState)
	wstore.PreemptiveStateDB = wstore.PreemptiveStateDB.Copy()
	wstore.muPreemptiveStateDB.Unlock()

	// write to localdb and chunkstore
	wstore.Storage.chunkCacheMu.Lock()
	wstore.Storage.chunkCache[genesisBlock.Hash()] = genesisBlock.Bytes()
	wstore.Storage.chunkCacheMu.Unlock()

	return nil
}

// Queue is a FIFO Block data strucutre
type Queue struct {
	blocks []*Block
}

// Enqueue adds a block to the queue
func (queue *Queue) Enqueue(block *Block) error {
	queue.blocks = append(queue.blocks, block)
	return nil
}

// Dequeue removes a block from the queue
func (queue *Queue) Dequeue() (block *Block) {
	if len(queue.blocks) == 0 {
		return nil
	}

	block = queue.blocks[0]
	if len(queue.blocks) > 1 {
		queue.blocks = queue.blocks[1:]
	} else {
		queue.blocks = nil
	}
	return block
}

// Clear removes all items from the queue
func (queue *Queue) Clear() {
	queue.blocks = nil
	return
}

type Dispatcher struct {
	Requestch  chan *Req
	workerPool chan chan *Req
	workerNum  int
	cancel     context.CancelFunc
}

type Ret struct {
	data [][]byte
	err  error
}

type Req struct {
	f      func(data chan *Ret, quit chan interface{}, errCh chan error, closed *bool, mu sync.RWMutex, ctx context.Context, flg *uint64, args ...[]byte)
	data   chan *Ret
	quit   chan interface{}
	errCh  chan error
	closed *bool
	mu     sync.RWMutex
	ctx    context.Context
	flg    *uint64
	args   [][]byte
}

func NewDispatcher(workerNum int, cancel context.CancelFunc) *Dispatcher {
	return &Dispatcher{
		Requestch:  make(chan *Req, workerNum*10),
		workerPool: make(chan chan *Req, workerNum),
		workerNum:  workerNum,
		cancel:     cancel,
	}
}

func (self *WolkStore) StartDispatcher(ctx context.Context) error {
	poolLength := len(self.dispatcher.workerPool)
	if poolLength != 0 {
		return fmt.Errorf("already started")
	}
	for i := 0; i < self.dispatcher.workerNum; i++ {
		self.startWorker(ctx, 1)
	}

	go self.dispatcher.dispatch(ctx)
	return nil
}

func (self *WolkStore) startWorker(ctx context.Context, num int) {
	requestch := make(chan *Req)

	for i := 0; i < num; i++ {
		go func() {
			for {
				// workerPoolにchanを入れる(終わったらまだ戻る)
				self.dispatcher.workerPool <- requestch
				select {
				case req := <-requestch:
					//request(req)
					log.Info("get workerPool", "args[0]", fmt.Sprintf("%x", req.args[0]), "len", len(self.dispatcher.workerPool), "num go", runtime.NumGoroutine(), "self.dispatcher", self.dispatcher, "self.dispatcher.workerPool", self.dispatcher.workerPool)
					req.f(req.data, req.quit, req.errCh, req.closed, req.mu, req.ctx, req.flg, req.args...)
					log.Info("done workerPool", "argv[0]", fmt.Sprintf("%x", req.args[0]), "len", len(self.dispatcher.workerPool), "num go", runtime.NumGoroutine(), "self.dispatcher", self.dispatcher, "self.dispatcher.workerPool", self.dispatcher.workerPool)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return
}

func (d *Dispatcher) dispatch(ctx context.Context) {
	log.Info("start dispatch")
	for {
		//log.Info("loop dispatch")
		select {
		case req := <-d.Requestch:
			//log.Info("dispatch")
			// workerPoolからchanを取り出しreqを入れる
			worker := <-d.workerPool
			//log.Info("dispatch get worker")
			worker <- req
		case <-ctx.Done():
			return
		}
	}
}

func (self *WolkStore) GetShare(ctx context.Context, k []byte) (val []byte, len_chunk int, err error) {
	data := make(chan *Ret)
	quit := make(chan interface{})
	errCh := make(chan error)
	closed := false
	var flg uint64
	var mu sync.RWMutex
	var args [][]byte
	args = append(args, k)
	self.dispatcher.Requestch <- &Req{self.GetShareWorker, data, quit, errCh, &closed, mu, ctx, &flg, args}
	select {
	case ret := <-data:
		return ret.data[0], len(ret.data), ret.err
	case err = <-errCh:
		return nil, 0, err
	}
	return nil, 0, err
}

func (self *WolkStore) GetChunk(ctx context.Context, k []byte) (val []byte, ok bool, err error) {
	data := make(chan *Ret)
	quit := make(chan interface{})
	errCh := make(chan error)
	closed := false
	var flg uint64
	var mu sync.RWMutex
	var args [][]byte
	args = append(args, k)
	self.dispatcher.Requestch <- &Req{self.GetChunkWorker, data, quit, errCh, &closed, mu, ctx, &flg, args}
	select {
	case ret := <-data:
		var ok bool
		if ret.err == nil {
			ok = true
		}
		return ret.data[0], ok, ret.err
	case err = <-errCh:
		return nil, false, err
	}
	return nil, false, err
}

func (self *WolkStore) GetShareWorker(data chan *Ret, quit chan interface{}, errCh chan error, closed *bool, mu sync.RWMutex, ctx context.Context, flg *uint64, args ...[]byte) {
	val, _, err := self.Storage.GetShare(args[0])
	var arg [][]byte
	arg = append(arg, val)
	data <- &Ret{arg, err}
}

func (self *WolkStore) GetChunkWorker(data chan *Ret, quit chan interface{}, errCh chan error, closed *bool, mu sync.RWMutex, ctx context.Context, flg *uint64, args ...[]byte) {
	val, ok, err := self.Storage.GetChunk(context.TODO(), args[0])
	if err == nil && !ok {
		err = fmt.Errorf("Not Found")
	}
	var arg [][]byte
	arg = append(arg, val)
	data <- &Ret{arg, err}
}

/* ########################################################################## */
