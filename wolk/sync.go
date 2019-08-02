// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/log"
)

var synchroniseInProgress bool
var synchroniseInProgressMu sync.RWMutex

const (
	forceSyncCycle      = 5 * time.Second // Time interval to force syncs, even if few peers are available
	minDesiredPeerCount = 2               // Amount of peers desired to start syncing

	// This is the target size for the packs of transactions sent by txsyncLoop.
	// A pack can get larger than this if a single transactions exceeds this size.
	txsyncPackSize = 100 * 1024
)

type txsync struct {
	p   *peer
	txs []*Transaction
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) syncTransactions(p *peer) {
	/*
		var txs Transactions
		pending, _ := pm.txpool.Pending()
		for _, batch := range pending {
			txs = append(txs, batch...)
		}
		if len(txs) == 0 {
			return
		}
		select {
		case pm.txsyncCh <- &txsync{p, txs}:
		case <-pm.quitSync:
		}
	*/
}

// txsyncLoop takes care of the initial transaction sync for each new
// connection. When a new peer appears, we relay all currently pending
// transactions. In order to minimise egress bandwidth usage, we send
// the transactions in small packs to one peer at a time.
func (pm *ProtocolManager) txsyncLoop() {
	/*
		var (
			pending = make(map[discover.NodeID]*txsync)
			sending = false               // whether a send is active
			pack    = new(txsync)         // the pack that is being sent
			done    = make(chan error, 1) // result of the send
		)

		// send starts a sending a pack of transactions from the sync.
		send := func(s *txsync) {
			// Fill pack with transactions up to the target size.
			size := common.StorageSize(0)
			pack.p = s.p
			pack.txs = pack.txs[:0]
			for i := 0; i < len(s.txs) && size < txsyncPackSize; i++ {
				pack.txs = append(pack.txs, s.txs[i])
				size += s.txs[i].Size()
			}
			// Remove the transactions that will be sent.
			s.txs = s.txs[:copy(s.txs, s.txs[len(pack.txs):])]
			if len(s.txs) == 0 {
				delete(pending, s.p.ID())
			}
			// Send the pack in the background.
			s.p.Log().Trace("Sending batch of transactions", "count", len(pack.txs), "bytes", size)
			sending = true
			go func() { done <- pack.p.SendTransactions(pack.txs) }()
		}

		// pick chooses the next pending sync.
		pick := func() *txsync {
			if len(pending) == 0 {
				return nil
			}
			n := rand.Intn(len(pending)) + 1
			for _, s := range pending {
				if n--; n == 0 {
					return s
				}
			}
			return nil
		}

		for {
			select {
			case s := <-pm.txsyncCh:
				pending[s.p.ID()] = s
				if !sending {
					send(s)
				}
			case err := <-done:
				sending = false
				// Stop tracking peers that cause send failures.
				if err != nil {
					pack.p.Log().Debug("Transaction send failed", "err", err)
					delete(pending, pack.p.ID())
				}
				// Schedule the next send.
				if s := pick(); s != nil {
					send(s)
				}
			case <-pm.quitSync:
				return
			}
		}
	*/
}

// syncer is responsible for periodically synchronising with the network, both
// downloading hashes and blocks as well as handling the announcement handler.
func (pm *ProtocolManager) syncer() {
	// Wait for different events to fire synchronisation operations
	forceSync := time.NewTicker(forceSyncCycle)
	defer forceSync.Stop()
	synchroniseInProgressMu.Lock()
	synchroniseInProgress = false
	synchroniseInProgressMu.Unlock()

	for {
		select {
		case <-pm.newPeerCh:
			// Make sure we have peers to select from, then sync
			if pm.peers.Len() < minDesiredPeerCount {
				log.Trace("[handler:syncer] new peer but", "pm.peers.Len()", pm.peers.Len(), "minDesiredPeerCount", minDesiredPeerCount)
				break
			}
			//go pm.synchroniseCert(pm.BestPeer())

		case <-forceSync.C:
			// Force a sync even if not enough peers are present
			bestpeer := pm.BestPeer()
			if bestpeer != nil {
				log.Trace("[handler:syncer] forceSync Start")
				pm.synchroniseCert(bestpeer)
			} else {
				log.Info("[handler:syncer] forceSync FAIL")
			}
		case <-pm.noMorePeers:
			return
		}
	}
}

func (pm *ProtocolManager) BestPeer() *peer {
	pm.peers.lock.RLock()
	defer pm.peers.lock.RUnlock()
	ps := pm.peers
	var (
		bestPeer                     *peer
		bestBlockNumber              uint64
		bestLastFinalizedBlockNumber uint64
	)
	for _, p := range ps.peers {
		p.muPeer.RLock()
		if p.lastBlockNumber > bestBlockNumber || (p.lastBlockNumber == bestBlockNumber && p.lastFinalizedBlockNumber > bestLastFinalizedBlockNumber) {
			bestPeer, bestBlockNumber, bestLastFinalizedBlockNumber = p, p.lastBlockNumber, p.lastFinalizedBlockNumber
		}
		p.muPeer.RUnlock()
	}
	if bestPeer != nil {
		log.Info("[handler:BestPeer] return", "p.ID()", bestPeer.ID(), "bestPeer.name", bestPeer.Name(), "lastBlockNumber", bestPeer.lastBlockNumber, "lastFinalized", bestPeer.lastFinalizedBlockNumber)
	}
	return bestPeer
}

func (pm *ProtocolManager) requestMissingCerts(missingCerts map[uint64]common.Hash) (err error) {
	pm.peers.lock.RLock()
	defer pm.peers.lock.RUnlock()
	ps := pm.peers
	//TODO: optimize the current wasteful treq!
	for _, p := range ps.peers {
		p.muPeer.RLock()
		for bn, blockhash := range missingCerts {
			log.Info("[sync:requestMissingCerts]", "bn", bn, "blockhash", blockhash)
			if p.lastBlockNumber >= bn {
				log.Info("[sync:requestMissingCerts]", "bn", bn, "blockhash", blockhash, "pName", p.Name())
				p.SendCertTargetRequest(bn, blockhash)
			}
		}
		p.muPeer.RUnlock()
	}
	return nil
}

// synchroniseCert tries to sync up our local block chain with a remote peer.
func (pm *ProtocolManager) synchroniseCert(peer *peer) (update bool, err error) {
	if peer == nil {
		log.Info("[sync:synchroniseCert] SendCertRequest fail0")
		return false, err
	}

	batchLimit := uint64(50)
	lastBN, err := pm.wolkChain.Indexer.GetLastKnownBlockNumber()
	if err != nil {
		log.Error("[sync:synchroniseCert] SendCertRequest fail1", "err", err)
		return false, err
	}
	targetEnd := lastBN + batchLimit
	lastFinalBN := pm.wolkChain.LastFinalized()
	peer.muPeer.RLock()
	peerFinalBN := peer.lastFinalizedBlockNumber
	peer.muPeer.RUnlock()

	if lastBN >= lastFinalBN && lastBN-lastFinalBN < 10 {
		lastBN = lastFinalBN

		if peerFinalBN > lastFinalBN {
			targetEnd = peerFinalBN
			if peerFinalBN-lastFinalBN >= batchLimit {
				targetEnd = lastFinalBN + batchLimit
			}
		} else if lastFinalBN > peerFinalBN {
			//exclude the equal case; help peer catch up
			certs, err := pm.wolkChain.Indexer.GetCertificates(lastFinalBN, lastFinalBN)
			if err != nil {
				log.Error("[handler:handleMsg:synchroniseCert] local cert mismatch", "err", err)
			} else {
				log.Info("[sync:synchroniseCert] Samaritan", "lbn", lastFinalBN, "end", lastFinalBN)
				peer.SendCerts(certs)
			}
			return false, nil
		} else {
			//no exchanage if at same height
			targetEnd = lastFinalBN + batchLimit
			return false, nil
		}
	}
	// Make sure the peer's lastBlockNumber is higher than our own
	log.Info("[sync:synchroniseCert] SendCertRequest", "lbn", lastBN+1, "end", targetEnd)
	peer.SendCertRequest(lastBN+1, targetEnd)
	return false, nil
}
