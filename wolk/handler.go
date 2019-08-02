// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.\
package wolk

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	//	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"

	//set "gopkg.in/fatih/set.v0"
	mapset "github.com/deckarep/golang-set"
)

// Constants to match up protocol versions and messages
const (
	wolk67 = 67
)

const maxKnownTxs = 10000

// Official short name of the protocol used during capability negotiation.
var ProtocolName = "wolk"

// Supported versions of the eth protocol (first is primary).
var ProtocolVersions = []uint{wolk67}

// Number of implemented message corresponding to different protocol versions.
var ProtocolLengths = []uint64{9}

const ProtocolMaxMsgSize = 10 * 1024 * 1024 // Maximum cap on the size of a protocol message

// eth protocol message codes
const (
	// Wolk msg
	StatusMsg            = 0x01
	TxMsg                = 0x02
	NewWolkBlockMsg      = 0x03
	CertMsg              = 0x04
	CertRequestMsg       = 0x05
	VoteMsg              = 0x06
	ResvStatusMsg        = 0x07
	CertTargetRequestMsg = 0x08
)

func getMsgType(code uint64) string {
	switch code {
	case 0x01:
		return "StatusMsg"
	case 0x02:
		return "TxMsg"
	case 0x03:
		return "NewWolkBlockMsg"
	case 0x04:
		return "CertMsg"
	case 0x05:
		return "CertRequestMsg"
	case 0x06:
		return "VoteMsg"
	case 0x07:
		return "ResvStatusMsg"
	case 0x08:
		return "CertTargetRequestMsg"
	default:
		return "invalid Msg type"
	}
}

type errCode int

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrProtocolVersionMismatch
	ErrNetworkIdMismatch
	ErrGenesisBlockMismatch
	ErrNoStatusMsg
	ErrExtraStatusMsg
	ErrSuspendedPeer
)

// txChanSize is the size of channel listening to TxPreEvent.
// The number is referenced from the size of tx pool.
const (
	txChanSize       = 4096
	proposalChanSize = 409600
	voteChanSize     = 409600
)

func (e errCode) String() string {
	return errorToString[int(e)]
}

// TxPreEvent is posted when an transaction enters the transaction pool.
type TxPreEvent struct {
	Tx *Transaction
}

// ProposalPreEvent is posted when a blockproposal or forkproposal is broadcasted
type ProposalPreEvent struct {
	Proposal *Proposal
	Forked   bool
}

// VotePreEvent is posted when a Vote is broadcasted
type VotePreEvent struct {
	Vote *VoteMessage
}

// CertPreEvent is posted when a block Certificate is broadcasted
type CertPreEvent struct {
}

var errorToString = map[int]string{
	ErrMsgTooLarge:             "Message too long",
	ErrDecode:                  "Invalid message",
	ErrInvalidMsgCode:          "Invalid message code zzz",
	ErrProtocolVersionMismatch: "Protocol version mismatch",
	ErrNetworkIdMismatch:       "NetworkId mismatch",
	ErrGenesisBlockMismatch:    "Genesis block mismatch",
	ErrNoStatusMsg:             "No status message",
	ErrExtraStatusMsg:          "Extra status message",
	ErrSuspendedPeer:           "Suspended peer",
}

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type ProtocolManager struct {
	networkId uint64
	wolkChain *WolkStore
	P2PServer *p2p.Server

	SubProtocols []p2p.Protocol
	scope        event.SubscriptionScope
	newPeerCh    chan *peer // peer-related
	maxPeers     int
	peers        *peerSet
	noMorePeers  chan struct{}

	txn_Ch   chan TxPreEvent // tx-related
	txn_Sub  event.Subscription
	TxFeed   event.Feed
	txpool   *TxPool
	txsyncCh chan *txsync //not used?
	quitSync chan struct{}

	ConsensusEngines   map[uint64]*ConsensusEngine
	ConsensusEnginesMu sync.RWMutex
	//	ConsensusEngineCh chan uint64

	minedBlockSub     *event.TypeMuxSubscription // cert-related
	eventMux          *event.TypeMux             // channels for fetcher, syncer, txsyncLoop
	wg                sync.WaitGroup
	knownWolkBlocks   mapset.Set // Set of block hashes known to be known by this peer
	knownWolkTxs      mapset.Set // Set of tx hashes known to be known by this peer
	knownProposals    mapset.Set // Set of proposal hashes known to be known by this peer
	knownCertificates mapset.Set // Set of Certificates hashes known to be known by this peer
	knownVotes        mapset.Set
}

// NewProtocolManager returns a new ethereum sub protocol manager. The Ethereum sub protocol manages peers capable
// with the ethereum network.
func NewProtocolManager(config *cloud.Config, mux *event.TypeMux, txpool *TxPool, chain *WolkStore) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		networkId:         config.NetworkID,
		eventMux:          mux,
		peers:             newPeerSet(),
		newPeerCh:         make(chan *peer),
		txsyncCh:          make(chan *txsync),
		quitSync:          make(chan struct{}),
		wolkChain:         chain,
		txn_Ch:            make(chan TxPreEvent, txChanSize),
		txpool:            txpool,
		knownWolkBlocks:   mapset.NewSet(),
		knownWolkTxs:      mapset.NewSet(),
		knownProposals:    mapset.NewSet(),
		knownCertificates: mapset.NewSet(),
		knownVotes:        mapset.NewSet(),
		ConsensusEngines:  make(map[uint64]*ConsensusEngine),
	}

	manager.txn_Sub = manager.SubscribeTxPreEvent(manager.txn_Ch)
	manager.minedBlockSub = mux.Subscribe(NewMinedBlockEvent{})

	// Initiate a sub-protocol for every implemented version we can handle
	manager.SubProtocols = make([]p2p.Protocol, 0, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		// Compatible; initialise the sub-protocol
		version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  ProtocolLengths[i],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := manager.newPeer(int(version), p, rw)
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					log.Trace("[handler:NewProtocolManager] NEW PEER", "Run", p, "version", version)
					return manager.handle(peer)
				case <-manager.quitSync:
					return p2p.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return manager.NodeInfo()
			},
			PeerInfo: func(id enode.ID) interface{} {
				if p := manager.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info()
				}
				return nil
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}
	return manager, nil
}

func (pm *ProtocolManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Info("[handler:removePeer] REMOVING Wolk peer", "peer", id)

	// Unregister the peer from the downloader and Ethereum peer set
	//pm.downloader.UnregisterPeer(id)
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("[handler:removePeer] Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (pm *ProtocolManager) maintainTrustedPeers() {
	for {
		/*
			if pm.P2PServer != nil {
				for _, info := range pm.P2PServer.PeersInfo() {
					//log.Trace("[handler:maintainTrustedPeers]", "enode", info.Enode, "id", info.ID, "name", info.Name, "trusted", info.Network.Trusted, "static", info.Network.Static)
				}
			} else {
				log.Debug("[handler:maintainTrustedPeers] node is nil")
			}
		*/
		time.Sleep(5 * time.Second)
	}
}

// SubscribeTxPreEvent registers a subscription of TxPreEvent and
// starts sending event to the given channel.
func (pm *ProtocolManager) SubscribeTxPreEvent(ch chan<- TxPreEvent) event.Subscription {
	return pm.scope.Track(pm.TxFeed.Subscribe(ch))
}

func (pm *ProtocolManager) Start(srvr *p2p.Server) {
	pm.maxPeers = srvr.MaxPeers
	log.Info("[handler:Start] START", "maxPeers", pm.maxPeers)
	pm.P2PServer = srvr
	//go pm.maintainTrustedPeers()
	go pm.txnBroadcastLoop()
	go pm.generatedWolkBlockBroadcastLoop()
	go pm.syncer()
	go pm.ConsensusEngineLoop()
}

func (pm *ProtocolManager) ConsensusEngineLoop() {

	checkInterval := time.NewTicker(100 * time.Millisecond)
	finalizeConsensusEngines := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-checkInterval.C:
			pm.getConsensusEngine(pm.wolkChain.currentRound())
		case <-finalizeConsensusEngines.C:
			currentRound := pm.wolkChain.currentRound()
			if currentRound > 10 {
				pm.finalizeConsensusEngines(pm.wolkChain.currentRound() - 10)
			}
		}
	}
}

func (pm *ProtocolManager) Stop() {
	log.Info("[handler:Stop] Stopping Wolk protocol")

	pm.txn_Sub.Unsubscribe()       // quits txBroadcastLoop
	pm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	pm.wg.Wait()

	log.Info("[handler:Stop] Wolk protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, p, rw)
}

// handle is the callback invoked to manage the life cycle of an eth peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers {
		log.Trace("[handler:handle] Too Many Peers", "numPeers", pm.peers.Len())
		return p2p.DiscTooManyPeers
	}
	log.Trace("[handler:handle] WOLK peer connecting", "name", p.Name(), "len(peers)", pm.peers.Len())

	lastbn := pm.wolkChain.round()
	lastfbn := pm.wolkChain.LastFinalized()

	if err := p.Handshake(pm.networkId, lastbn, lastfbn); err != nil {
		log.Debug("[handler:handle] WOLK handshake failed", "err", err)
		return err
	}

	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		log.Error("[handler:handle] WOLK peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)
	p.lastMessageReceived = time.Now()

	sendStatus := time.NewTicker(5000 * time.Millisecond)
	for {
		select {
		case <-sendStatus.C:
			if time.Since(p.lastMessageReceived) > 60*time.Second && time.Since(p.lastMessageReceived) < 24*time.Hour {
				return fmt.Errorf("[handler:handle] statusData Timeout")
			} else if time.Since(p.lastMessageReceived) > 2*time.Second {
				log.Trace("[p2p:handle] sendStatus", " time.Since(p.lastMessageReceived)", time.Since(p.lastMessageReceived), "LastFinalizedBlockNumber", pm.wolkChain.LastFinalized())
				p2p.Send(p.rw, StatusMsg, &statusData{
					ProtocolVersion:          uint32(p.version),
					NetworkId:                pm.networkId,
					LastBlockNumber:          pm.wolkChain.round(),
					LastFinalizedBlockNumber: pm.wolkChain.LastFinalized(),
				})
			}
		default:
			err := pm.handleMsg(p)
			if err != nil {
				return err
			}
		}
	}
}

func (pm *ProtocolManager) getConsensusEngine(round uint64) (eng *ConsensusEngine) {
	if round <= pm.wolkChain.round() {
		// we already processed this round
		return nil
	}
	var ok bool
	pm.ConsensusEnginesMu.Lock()
	defer pm.ConsensusEnginesMu.Unlock()
	eng, ok = pm.ConsensusEngines[round]
	if !ok {
		eng = NewConsensusEngine(round, pm.wolkChain)
		pm.ConsensusEngines[round] = eng
		log.Warn("[handler:getConsensusEngine] ******** Adding NewConsensusEngine ********", "round", round)
	}
	return eng
}

func (pm *ProtocolManager) finalizeConsensusEngines(finalized uint64) {
	pm.ConsensusEnginesMu.Lock()
	defer pm.ConsensusEnginesMu.Unlock()
	for round, _ := range pm.ConsensusEngines {
		if round <= finalized {
			// tell the round processor to finish up, which is necessary to not have leaky goroutines [created getConsensusEngine]
			delete(pm.ConsensusEngines, round)
		}
	}
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	//TODO: Fail out if Regional ID Missing peer.RegId
	// Read the next message from the remote peer, and ensure it's fully consume
	msg, err := p.rw.ReadMsg()
	if err != nil {
		log.Error(fmt.Sprintf("[handler:handleMsg] Node%d ReadMsg Error", pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name(), "Error", err)
		return err
	}
	//log.Trace(fmt.Sprintf("[handler:handleMsg] Node%d handleMsg received", pm.wolkChain.consensusIdx), "code", getMsgType(msg.Code), "peer", p.ID(), "p.Name", p.Name(), "msg", msg)
	/*WOLK review ProtocolMaxMsgSize */
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	// Handle the message depending on its contents

	msgType := getMsgType(msg.Code)
	switch {

	case msg.Code == NewWolkBlockMsg:
		//log.Info(fmt.Sprintf("[handler:handleMsg:%v] [Node%d]", msgType, pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name())
		var wolkBlock *Block
		if err := msg.Decode(&wolkBlock); err != nil {
			log.Error(fmt.Sprintf("[handler:handleMsg:%v] Node%d decode error", msgType, pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name(), "Error", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}

		eng := pm.getConsensusEngine(wolkBlock.BlockNumber)
		if eng != nil {
			//log.Info("[handler:handleMsg:block] received block", "bn", wolkBlock.BlockNumber, "h", wolkBlock.Hash())
			eng.blockCh <- wolkBlock
			p.MarkWolkBlock(wolkBlock.Hash())
			pm.BroadcastBlock(wolkBlock)
		} else {
			// log.Info("[handler:handleMsg:block] NOT received block", "wolkBlock.BlockNumber", wolkBlock.BlockNumber)
		}

	case msg.Code == TxMsg:
		//log.Trace(fmt.Sprintf("[handler:handleMsg:%v] [Node%d]", msgType, pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name())
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs WolkTransactions
		if err := msg.Decode(&txs); err != nil {
			log.Error(fmt.Sprintf("[handler:handleMsg:%v] Node%d decode TxMsg error", msgType, pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name(), "Error", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}
			if pm.checkTransaction(tx.Hash()) {
				continue
			}
			pm.markTransaction(tx.Hash())
			// don't send this back to the peer that just sent it to you!
			p.MarkWolkTransaction(tx.Hash())

			// add to stateDB
			//log.Trace(fmt.Sprintf("[handler:handleMsg:%v] Temp [Node%d]", msgType, pm.wolkChain.consensusIdx), "tx", tx)
			err = pm.wolkChain.ReceiveTransaction(tx)
			if err != nil {
				log.Warn(fmt.Sprintf("[handler:handleMsg:%v] Node%d ReceiveTransaction ERR", msgType, pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name(), "Error", err)
			} else {
				// put the received txs into the txn_Ch channel
			}
		}

	case msg.Code == CertRequestMsg:
		//log.Info("[handler:handleMsg:CertRequestMsg] RECEIVED CertRequestMsg", "p.Name", p.Name())
		var req certRequest
		if err := msg.Decode(&req); err != nil {
			p.muPeer.RLock()
			log.Error(fmt.Sprintf("[handler:handleMsg:CertRequestMsg]  [Node%d] decode err", pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name(), "err", err)
			p.muPeer.RUnlock()
		} else {
			log.Info("[handler:handleMsg:CertRequestMsg] RECEIVED CertRequestMsg", "p.Name", p.Name(), "st", req.BlockNumberStart, "end", req.BlockNumberEnd)
			certs, err := pm.wolkChain.Indexer.GetCertificates(req.BlockNumberStart, req.BlockNumberEnd)
			if err != nil {
				log.Error("[handler:handleMsg:CertRequestMsg] GetCertificates", "err", err)
			} else {
				for _, cert := range certs {
					log.Info("[handler:CertRequestMsg]", "BlockNumberStart", req.BlockNumberStart, "BlockNumberEnd", req.BlockNumberEnd, "cert", cert.BlockNumber, "step", cert.Step)
					cert.dump(" cert response ")
				}
				p.SendCerts(certs)
			}
		}

	case msg.Code == CertTargetRequestMsg:
		//log.Info("[handler:handleMsg:CertTargetRequestMsg] RECEIVED CertTargetRequestMsg", "p.Name", p.Name())
		var treq certTargetRequest
		if err := msg.Decode(&treq); err != nil {
			p.muPeer.RLock()
			log.Error(fmt.Sprintf("[handler:handleMsg:CertTargetRequestMsg]  [Node%d] decode err", pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name(), "err", err)
			p.muPeer.RUnlock()
		} else {
			log.Info("[handler:handleMsg:CertTargetRequestMsg] RECEIVED CertTargetRequestMsg", "p.Name", p.Name(), "tartget bn", treq.BlockNumber, "blockhash", treq.BlockHash)
			certs, err := pm.wolkChain.Indexer.GetTargetCertificate(treq.BlockNumber, treq.BlockHash)
			if err != nil {
				log.Error("[handler:handleMsg:CertTargetRequestMsg] GetCertificates", "err", err)
			} else if len(certs) > 0 {
				for _, cert := range certs {
					log.Info("[handler:CertTargetRequestMsg]", "tartget bn", cert.BlockNumber, "blockhash", cert.BlockHash)
					cert.dump(" cert response ")
				}
				p.SendCerts(certs)
			}
		}

	case msg.Code == CertMsg:
		//log.Info(fmt.Sprintf("[handler:handleMsg:%v] [Node%d]", msgType, pm.wolkChain.consensusIdx), "p.Name", p.Name())
		var certs Votes
		if err := msg.Decode(&certs); err != nil {
			log.Error(fmt.Sprintf("[handler:handleMsg:%v]  [Node%d] decode err", msgType, pm.wolkChain.consensusIdx), "peer", p.ID(), "p.Name", p.Name(), "err", err)
		}
		log.Info("[handler] CertMsg", "len(certs)", len(certs))
		if len(certs) > 0 {
			pm.wolkChain.processCerts(context.TODO(), certs)
		}

	case msg.Code == StatusMsg:
		// Decode the handshake and make sure everything matches
		var status statusData
		if err := msg.Decode(&status); err != nil {
			log.Debug("[peer:readStatus] msg err", "err", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		if status.NetworkId != pm.networkId {
			log.Debug("[peer:readStatus] ErrNetworkIdMismatch")
			return errResp(ErrNetworkIdMismatch, "%d (!= %d)", status.NetworkId, pm.networkId)
		}
		if int(status.ProtocolVersion) != p.version {
			log.Debug("[peer:readStatus] ErrProtocolVersionMismatch")
			return errResp(ErrProtocolVersionMismatch, "%d (!= %d)", status.ProtocolVersion, p.version)
		}

		log.Trace("[peer:readStatus] Received status", "id", p.id, "lastBlockNumber", p.lastBlockNumber)

		p.muPeer.Lock()
		p.lastBlockNumber = status.LastBlockNumber
		p.lastMessageReceived = time.Now()
		p.muPeer.Unlock()

		currentbn := pm.wolkChain.round()
		lastFinalized := pm.wolkChain.LastFinalized()
		p2p.Send(p.rw, ResvStatusMsg, &statusData{
			ProtocolVersion:          uint32(wolk67),
			NetworkId:                pm.networkId,
			LastBlockNumber:          currentbn,
			LastFinalizedBlockNumber: lastFinalized,
		})

		if currentbn > p.lastBlockNumber && currentbn-p.lastBlockNumber < 3 {
			certs, err := pm.wolkChain.Indexer.GetCertificates(p.lastBlockNumber+1, currentbn)
			if err != nil {
				log.Error("[peer:readStatus] GetCertificates", "err", err)
			} else {
				p.SendCerts(certs)
			}
		}
	case msg.Code == ResvStatusMsg:
		// Decode the handshake and make sure everything matches
		var status statusData
		if err := msg.Decode(&status); err != nil {
			log.Debug("[peer:readStatus] msg err", "err", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		if status.NetworkId != pm.networkId {
			log.Debug("[peer:readStatus] ErrNetworkIdMismatch")
			return errResp(ErrNetworkIdMismatch, "%d (!= %d)", status.NetworkId, pm.networkId)
		}
		if int(status.ProtocolVersion) != p.version {
			log.Debug("[peer:readStatus] ErrProtocolVersionMismatch")
			return errResp(ErrProtocolVersionMismatch, "%d (!= %d)", status.ProtocolVersion, p.version)
		}

		log.Trace("[peer:readStatus] Received status", "id", p.id, "lastBlockNumber", p.lastBlockNumber)

		p.muPeer.Lock()
		p.lastFinalizedBlockNumber = status.LastFinalizedBlockNumber
		p.lastBlockNumber = status.LastBlockNumber
		p.lastMessageReceived = time.Now()
		p.muPeer.Unlock()

		currentbn := pm.wolkChain.round()
		if currentbn > p.lastBlockNumber && currentbn-p.lastBlockNumber < 3 {
			certs, err := pm.wolkChain.Indexer.GetCertificates(p.lastBlockNumber+1, currentbn)
			if err != nil {
				log.Error("[peer:readStatus] GetCertificates", "err", err)
			} else {
				p.SendCerts(certs)
			}
		}

	default:
		log.Debug(fmt.Sprintf("[handler:handleMsg:default] [Node%d] Code%d - Unknown", pm.wolkChain.consensusIdx, msg.Code), "peer", p.ID(), "p.Name", p.Name())
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

func (pm *ProtocolManager) GetPeerIPs() []string {
	peersArr := make([]string, 0)
	for _, p := range pm.peers.GetPeers() {
		peersArr = append(peersArr, fmt.Sprintf("%s", p.RemoteAddr()))
	}
	return peersArr
}

func (pm *ProtocolManager) GetPeers() []string {
	peers := pm.peers.GetPeers()
	//peersLen := pm.peers.Len()
	peersArr := make([]string, 0)
	for _, p := range peers {
		id := p.ID()
		enodeID := fmt.Sprintf("%x", id[:])
		remoteAddr := p.RemoteAddr() // net.Addr
		addr := fmt.Sprintf("%s", remoteAddr)
		enode := enodeID + "@" + addr
		peersArr = append(peersArr, enode)
	}
	log.Debug(fmt.Sprintf("[handler:GetPeers] peersArr: %s ", peersArr))
	return peersArr
}

// BroadcastTx will propagate a transaction to all peers which are not known to already have the given transaction.
func (pm *ProtocolManager) BroadcastTx(hash common.Hash, tx *Transaction) {
	// Broadcast transaction to a batch of peers not knowing about it
	peers := pm.peers.PeersWithoutTx(hash)
	for _, peer := range peers {
		err := peer.SendTransactions(WolkTransactions{tx})
		if err != nil {
			log.Error("[handler:BroadcastTx] Error Encountered Sending Wolk Transaction", "Error", err, "p", peer)
		}
	}
	log.Trace("[handler:BroadcastTx] Broadcast wolk transaction", "hash", hash, "recipients", len(peers))
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(block *Block) {
	//time := time.Now()
	//timeStr := fmt.Sprintf("%s", time)
	hash := block.Hash()
	peers := pm.peers.PeersWithoutWolkBlock(hash)
	//	log.Info("[handler:BroadcastBlock] BroadcastBlock", "len(peerswithout)", len(peers), "len(peers)", pm.peers.Len())
	// Send the wolk block to all of our peers
	for _, peer := range peers {
		log.Debug("[handler:BroadcastBlock] Sending to Peer", "idx", pm.wolkChain.consensusIdx, "peer", peer.ID(), "p.Name", peer.Name())
		err := peer.SendNewWolkBlock(block)
		if err != nil {
			log.Error("[handler:BroadcastBlock] BroadcastBlock] Error Encountered Sending Wolk Block", "Error", err)
		}
	}
	if len(peers) > 0 {
		log.Debug(fmt.Sprintf("[handler:BroadcastBlock] Node%d BroadcastBlock", pm.wolkChain.consensusIdx), "blockhash", block.Hash().Hex(), "bn", block.BlockNumber, "peer", pm.peers.Len(), "recipients", len(peers))
	}
}

func (pm *ProtocolManager) ForcedBroadcastBlock(block *Block) {
	peerset := pm.peers
	for _, peer := range peerset.peers {
		log.Debug("[handler:ForcedBroadcastBlock] Sending to Peer", "idx", pm.wolkChain.consensusIdx, "peer", peer.ID(), "p.Name", peer.Name())
		err := peer.SendNewWolkBlock(block)
		if err != nil {
			log.Error("[handler:ForcedBroadcastBlock] BroadcastBlock] Error Encountered Sending Wolk Block", "Error", err)
		}
	}
}

func (pm *ProtocolManager) ForcedBroadcastCerts(certs Votes) {
	peerset := pm.peers
	for _, peer := range peerset.peers {
		log.Debug("[handler:ForcedBroadcastCerts] Sending to Peer", "idx", pm.wolkChain.consensusIdx, "peer", peer.ID(), "p.Name", peer.Name())
		err := peer.SendCerts(certs)
		if err != nil {
			log.Error("[handler:ForcedBroadcastCerts] ForcedBroadcastCerts] Error Encountered while Sending Certs", "Error", err)
		}
	}
}

func (self *ProtocolManager) generatedWolkBlockBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range self.minedBlockSub.Chan() {
		switch ev := obj.Data.(type) {
		case NewMinedBlockEvent:
			self.BroadcastBlock(ev.Block) // First propagate block to peers
		}
	}
	//log.Trace("[handler:BroadcastTx]", "hash", hash, "recipients", len(peers))
}

func (self *ProtocolManager) txnBroadcastLoop() {
	for {
		select {
		case event := <-self.txn_Ch:
			log.Trace("[handler:wolk_txBroadcastLoop] Encountered wolk_txCh event and will attempt to broadcast", "txpool", self.txpool)
			self.BroadcastTx(event.Tx.Hash(), event.Tx)
		case <-self.txn_Sub.Err():
			log.Error("[handler:txnBroadcastLoop] Encountered txn_Ch err", "err", self.txn_Sub.Err())
			log.Error("[handler:wolk_txBroadcastLoop]", "txpool", self.txpool)
			return
		}
	}
}

// NodeInfo represents a short summary of the Ethereum sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network uint64 `json:"network"` // Ethereum network ID
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (self *ProtocolManager) NodeInfo() *NodeInfo {
	return &NodeInfo{
		Network: self.networkId,
	}
}

func (self *ProtocolManager) checkTransaction(hash common.Hash) bool {
	return self.knownWolkTxs.Contains(hash)
}

func (self *ProtocolManager) markTransaction(hash common.Hash) {
	// If we reached the memory allowance, drop a previously known transaction hash
	for self.knownWolkTxs.Cardinality() >= maxKnownWolkTxs {
		self.knownWolkTxs.Pop()
	}
	self.knownWolkTxs.Add(hash)
}

func (pm *ProtocolManager) NumPeers() int {
	return pm.peers.Len()
}

func (pm *ProtocolManager) GetPeerList() []string {
	var list []string
	pm.peers.lock.Lock()
	defer pm.peers.lock.Unlock()
	for _, j := range pm.peers.peers {
		//log.Trace("[peer:GetPeerList]", "i", i, "j", j.RemoteAddr(), "name", j.Name())
		list = append(list, fmt.Sprintf("addr: %s, name: %s lbn: %d finalized: %d", j.RemoteAddr(), j.Name(), j.LastBlockNumber(), j.LastFinalizedBlockNumber()))
	}
	//log.Trace("[peer:GetPeerList]", "list", list)
	return list
}
