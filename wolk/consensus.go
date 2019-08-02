package wolk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
)

var (
	errCancel         = errors.New("Cancel")
	errFinalizedRound = errors.New("finalized round")
	errForked         = errors.New("Forked")
	errHangForever    = errors.New("Hang forever")
)

const (
	useRealBalance = false
	useVRF         = false
)

// ConsensusEngine implements ConsensusEngine
type ConsensusEngine struct {
	round uint64
	step  uint64

	id      uint64
	privkey *crypto.PrivateKey
	pubkey  *crypto.PublicKey
	chain   *WolkStore

	proposer bool

	executorVRF      []byte
	executorProof    []byte
	executorSubusers int
	proposerVRF      []byte
	proposerProof    []byte
	proposerSubusers int

	// internal model
	votes map[uint64]*VoteMessage

	knownProposals   map[common.Hash]*Proposal
	knownProposalsMu sync.RWMutex

	// COUNT data, filled in from voteCh processing
	voters   map[uint64]map[string]struct{}
	votersMu sync.RWMutex

	// VOTE DATA
	voteCh       chan *VoteMessage
	votelist     map[common.Hash]*VoteMessage
	lastStepTime time.Time
	lastVoteTime time.Time
	lastVote     *VoteMessage
	vmu          sync.RWMutex // vote msg mutex for { knownMsgHashes, incomingVotes }

	certs   map[common.Hash]*VoteMessage
	certsMu sync.RWMutex

	// BLOCK DATA
	blockCh chan *Block
	blocks  map[common.Hash]*Block
	bmu     sync.RWMutex // block list mutex

	// LAST BLOCK [the highest vote known so far]
	lastBlock      *Block
	lastBlockVotes uint64
	ParentHash     common.Hash
	emptyBlk       *Block
	emptyHash      common.Hash

	// TREE
	allLastBlock      map[common.Hash]*Block
	allLastBlockVotes map[common.Hash]uint64 //votecounts
	allEmptyBlk       map[common.Hash]*Block
	allEmptyHash      map[common.Hash]common.Hash

	// Keep track of missing certs that have already been requested
	missingCerts map[common.Hash]common.Hash
}

// NewConsensusEngine creates a ConsensusEngine for the consensusEngine to use to manage protocol messages
func NewConsensusEngine(round uint64, wstore *WolkStore) *ConsensusEngine {
	isDeterministic := true
	id := wstore.consensusIdx
	rand.Seed(time.Now().UnixNano())
	var priv *crypto.PrivateKey
	var pub *crypto.PublicKey
	if isDeterministic {
		privateStr := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", id))))
		priv, _ = crypto.HexToEd25519(privateStr)
		pub = priv.PublicKey()
		addr := pub.Address()
		seed := priv.ToSeed()
		log.Info(fmt.Sprintf("[consensus:NewConsensusEngine] Node%d [%v] seed: [%v] deterministic: %v", id, addr.Hex(), seed, isDeterministic))
	} else {
		pub, priv, _ = crypto.NewKeyPair()
	}

	eng := &ConsensusEngine{
		chain: wstore,

		// incoming
		votelist:       make(map[common.Hash]*VoteMessage),
		voters:         make(map[uint64]map[string]struct{}),
		knownProposals: make(map[common.Hash]*Proposal),

		// outgoing
		votes: make(map[uint64]*VoteMessage),

		blocks: make(map[common.Hash]*Block),
		round:  round,

		voteCh:  make(chan *VoteMessage, 1000),
		blockCh: make(chan *Block, 500),

		allLastBlock:      make(map[common.Hash]*Block),
		allLastBlockVotes: make(map[common.Hash]uint64),
		allEmptyBlk:       make(map[common.Hash]*Block),
		allEmptyHash:      make(map[common.Hash]common.Hash),
		missingCerts:      make(map[common.Hash]common.Hash),

		certs:   make(map[common.Hash]*VoteMessage),
		id:      uint64(id),
		privkey: priv,
		pubkey:  pub,
	}

	certs, ok, err := wstore.Indexer.GetCertificate(eng.round)
	if err != nil {

	} else if ok {
		for _, cert := range certs {
			eng.certs[cert.BlockHash] = cert
		}
	}

	// set MULT
	setConsensusParameters(len(wstore.genesis.Registry))
	go func(r uint64) {
		//log.Info("[consensus:NewConsensusEngine] starting loop", "round", eng.round)
		// consensusLoop will come back either when
		// 1. consensus has been reached and a block has been verified
		// 2. when a fork happens
		// 3. when something cancels
		// 4. when it notices that the blockchain has moved on to a new block
		err := eng.consensusLoop(context.Background())
		if err == errCancel {
			log.Warn(fmt.Sprintf("[consensus:NewConsensusEngine] [Node%d] Cancelled ConsensusEngine", eng.id))
		} else if err == nil { // we finalized
			log.Info(fmt.Sprintf("[consensus:NewConsensusEngine] [Node%d] round #%d DONE", eng.id, round), "currentRound", wstore.currentRound())
		} else {
			log.Error(fmt.Sprintf("[consensus:NewConsensusEngine] [Node%d] ERR", eng.id), "err", err)
		}
	}(round)
	return eng
}

func (eng *ConsensusEngine) dumpChannelInfo() {
	if len(eng.voteCh)+len(eng.blockCh) > 500 {
		log.Info(fmt.Sprintf("[consensus:dumpChannelInfo] [Node%d] round #%d  CHANNEL OVERFLOW? (voteCH:%d|blockCH:%d)", eng.id, eng.round, len(eng.voteCh), len(eng.blockCh)))
	}

}

func (eng *ConsensusEngine) addCertificate(cert *VoteMessage) (err error) {
	eng.certsMu.Lock()
	defer eng.certsMu.Unlock()
	if _, ok := eng.certs[cert.BlockHash]; ok {
		return fmt.Errorf("Already have cert")
	}
	eng.certs[cert.BlockHash] = cert
	return nil
}

func (eng *ConsensusEngine) getSize() (sz int) {
	eng.vmu.RLock()
	defer eng.vmu.RUnlock()
	return len(eng.blocks)
}

// VOTE DATA

// BLOCK DATA
func (eng *ConsensusEngine) setBlock(wolkBlock *Block) bool {
	hash := wolkBlock.Hash()
	eng.bmu.RLock()
	blk, ok := eng.blocks[hash]
	eng.bmu.RUnlock()

	if !ok || blk == nil {
		eng.bmu.Lock()
		eng.blocks[hash] = wolkBlock
		eng.bmu.Unlock()
		//log.Info("[consensus:setBlock] STORED", "hash", hash)
		return true
	}
	return false
}

func (eng *ConsensusEngine) getBlock(hash common.Hash) *Block {
	eng.bmu.RLock()
	defer eng.bmu.RUnlock()
	return eng.blocks[hash]
}

// weight returns the weight of the given address.
func (eng *ConsensusEngine) weight(address common.Address, round uint64) (w uint64) {
	return TokenPerUser
}

// tokenOwn returns the token amount (weight) owned by self node.
func (eng *ConsensusEngine) tokenOwn(round uint64) (balance uint64) {
	return eng.weight(eng.Address(), round)
}

func (eng *ConsensusEngine) getByRound(round uint64) *Block {
	if round == 0 {
		return eng.chain.genesisBlock
	}
	last := eng.lastBlock
	startblock := uint64(1)
	if round > startblock {
		if last.Round() == round {
			return last
		}
		block, ok, err := eng.chain.GetBlockByNumber(round)
		if err != nil {
			log.Error(fmt.Sprintf("[consensus:getByRound] [Node%d] GetBlockByNumber ConsensusEngine Sortition Seed Error", eng.id), "round", round, "err", err)
		} else if !ok {
			log.Error(fmt.Sprintf("[consensus:getByRound] [Node%d] GetBlockByNumber ConsensusEngine Sortition Seed Block not found", eng.id), "round", round)
			//Consensus can not proceed because past seed in last epoch is not available (fetal case)
		}
		return block
	}
	return last
}

// sortitionSeed returns the selection seed with a refresh interval R.
func (eng *ConsensusEngine) sortitionSeed(round uint64) (b []byte) {
	//At round r, seed = r-1-(r mod R)
	// r_minus_1 := round - 1
	// r_mod_r := round % R
	// r := r_minus_1 - r_mod_r
	// log.Info("[consensus:sortitionSeed] sortitionSeed", "r_minus_1", r_minus_1, "r_mod_r", r_mod_r, "r", r)

	realR := round - 1
	mod := round % R
	if realR < mod {
		realR = 0
	} else {
		realR -= mod
	}

	// log.Info(fmt.Sprintf("[consensus:sortitionSeed] Node%d", eng.id), "r_minus_1", mod, "realR", realR)
	rblk := eng.getByRound(realR)
	if rblk == nil {
		log.Error(fmt.Sprintf("[consensus:sortitionSeed] [Node%d] NIL eng.getByRound", eng.id), "realR", realR)
		return b
	}
	return rblk.Seed
}

// Address returns the consensus Node address
func (eng *ConsensusEngine) Address() common.Address {
	return common.BytesToAddress(eng.pubkey.Bytes())
}

// seed returns the vrf-based seed of block r.
func (eng *ConsensusEngine) vrfSeed(lastBlock *Block) (seed, proof []byte, err error) {
	return eng.privkey.Evaluate(bytes.Join([][]byte{lastBlock.Seed, wolkcommon.UIntToByte(eng.round + 1)}, nil))
}

func (eng *ConsensusEngine) isProposer() (proposer bool) {
	return eng.proposerSubusers > 0
}

func (eng *ConsensusEngine) getNetworkID() (networkID uint64) {
	return eng.chain.genesis.NetworkID
}

// Algorithm 4: committeeVote votes for `hash` by broadcasting to all peers
func (eng *ConsensusEngine) expectedThresholds() (expectedTentative int, expectedFinal int) {
	return int(float64(expectedSubusers) * thresholdTentative), int(float64(expectedSubusers) * thresholdFinal)
}

func (eng *ConsensusEngine) recordInternalVote(voteMsg *VoteMessage, votedOnSteps []uint64) {
	eng.vmu.Lock()
	eng.lastVote = voteMsg
	eng.lastVoteTime = time.Now()
	eng.vmu.Unlock()
	c := voteMsg.Copy()
	for _, s := range votedOnSteps {
		log.Info(fmt.Sprintf("[consensus:recordInternalVote] [Node%d] added to eng.votes[s]", eng.id), "s", s)
		eng.votes[s] = c
	}
}

func (eng *ConsensusEngine) getBestProposal(step uint64) (bp *Proposal, cnt int, total int, voted bool) {
	bp = nil
	cnt = 0
	total = 0
	voted = false
	eng.knownProposalsMu.RLock()
	defer eng.knownProposalsMu.RUnlock()

	// take highest priority proposal with voteSoft
	for _, p := range eng.knownProposals {
		// pull seed	b := p.ParentHash
		qcnt := p.countVotes(step, p.Seed, eng.round)
		total += qcnt
		if bp == nil || (step > 0 && qcnt > cnt) || (step == 0 && bytes.Compare(p.Prior, bp.Prior) > 0) {
			cnt = qcnt
			bp = p
		}
		if p.hasVote(eng.id, step) {
			voted = true
		}
	}
	return bp, cnt, total, voted
}

func (eng *ConsensusEngine) getBlockByHash(ctx context.Context, blockHash common.Hash, options BlockReadOptions) (b *Block, err error) {
	b = eng.getBlock(blockHash)
	if b != nil {
		return b, nil
	}
	if options.ReadFromCloudstore {
		b, err = eng.chain.GetBlockByHash(ctx, blockHash, options)
		if err != nil {
			log.Warn(fmt.Sprintf("[Node%d] getBlockByHash looking %x but err", eng.id, blockHash), "err", err)
		} else if b != nil {
			eng.setBlock(b)
		}
		return b, err
	}
	log.Info(fmt.Sprintf("[Node%d] getBlockByHash not looking %x", eng.id, blockHash))
	return nil, nil
}

func (eng *ConsensusEngine) getVoteLastStep() (vm *VoteMessage, p *Proposal, countVotes uint64) {
	//log.Info(fmt.Sprintf("[Node%d] getVoteLastStep", eng.id, blockHash))
	// eng.vmu.RLock()
	// defer eng.vmu.RUnlock()
	if eng.lastVote != nil { // step > 0 {
		vm = eng.lastVote //  eng.votes[eng.step-1]
	}
	p, _, _, _ = eng.getBestProposal(eng.step - 1)
	if p != nil {
		countVotes = uint64(p.countVotes(eng.step-1, []byte(SkipSeed), eng.round))
	}
	return
}

func displayHash(h common.Hash) string {
	return fmt.Sprintf("%x", h.Bytes()[0:4])
}

func (eng *ConsensusEngine) hasExternalVote(votehash common.Hash) bool {
	eng.vmu.RLock()
	defer eng.vmu.RUnlock()
	if _, exist := eng.votelist[votehash]; exist {
		return true
	}
	return false
}

func (eng *ConsensusEngine) recordExternalVote(v *VoteMessage) {
	eng.vmu.Lock()
	defer eng.vmu.Unlock()
	eng.votelist[v.Hash()] = v
}

func (eng *ConsensusEngine) validCertHash(blockHash common.Hash) bool {
	eng.certsMu.RLock()
	defer eng.certsMu.RUnlock()
	if _, ok := eng.certs[blockHash]; ok {
		return true
	}
	return false
}

func (eng *ConsensusEngine) validParentHash(prevHash common.Hash) bool {
	prevEngine := eng.chain.protocolManager.getConsensusEngine(eng.round - 1)
	if prevEngine != nil {
		if prevEngine.validCertHash(prevHash) {
			return true
		}
	}
	if _, ok := eng.allLastBlock[prevHash]; !ok {
		log.Error(fmt.Sprintf("[consensus:processVote] [Node%d] Prevhash problem!", eng.id), "prevHash", prevHash, "eng.ParentHash", eng.ParentHash)
		if _, requestedAlready := eng.missingCerts[prevHash]; !requestedAlready {
			eng.missingCerts[prevHash] = prevHash
			missed := make(map[uint64]common.Hash)
			missed[eng.round-1] = prevHash
			log.Error(fmt.Sprintf("[consensus:processVote] [Node%d] Prevhash missing - issue CertTargetRequest", eng.id), "prevHash", prevHash, "eng.ParentHash", eng.ParentHash)

			eng.chain.protocolManager.requestMissingCerts(missed)
		}

		return false
	}
	return true
}

func (eng *ConsensusEngine) processVote(ctx context.Context, voteMsg *VoteMessage) bool {
	round := eng.round
	if voteMsg.BlockNumber != round {
		log.Error(fmt.Sprintf("[consensus:processVote] [Node%d] round mismatch", eng.id), "voteMsg", voteMsg.BlockNumber, "round", round)
		return false
	}

	// discard messages that do not extend this chain
	if !eng.validParentHash(voteMsg.ParentHash) {
		return false
	}

	// check if we processed this vote already
	voteHash := voteMsg.Hash()
	if eng.hasExternalVote(voteHash) {
		// log.Error("[processVote] hasExternalVote EMPTY problem!", "voter", voteMsg.Voter, "bn", voteMsg.BlockNumber, "step", voteMsg.Step, "seed", fmt.Sprintf("%x", voteMsg.Seed))
		return false
	}
	if err := voteMsg.VerifySignature(); err != nil {
		//log.Warn("[consensus:processVote] VerifySignature", "err", err)
		//		return false
	}

	vstep := voteMsg.Step
	pubkey, err := voteMsg.RecoverPubkey()
	if err != nil {
		log.Error(fmt.Sprintf("[consensus:processVote] [Node%d] EMPTY VOTE PUBKEY err: %v", eng.id, err), "err", err)
		return false
	}
	votes := verifySort(pubkey, voteMsg.VRF, voteMsg.Proof, eng.sortitionSeed(round), role(roleVoter, round), round, expectedTokensVoter, TokenPerUser, TotalTokenAmount)
	if uint64(votes) != voteMsg.Sub {
		log.Error(fmt.Sprintf("[consensus:processMsg] [Node%v] Malicious Vote (sub:%d|claimed:%d) ", eng.id, votes, voteMsg.Sub))
		return false
	}
	eng.recordExternalVote(voteMsg)

	eng.votersMu.Lock()
	if _, ok := eng.voters[vstep]; !ok {
		eng.voters[vstep] = make(map[string]struct{})
	}

	// check if the voter has voted yet -- only one vote per step!
	if _, exist := eng.voters[vstep][string(pubkey.Bytes())]; exist || votes == 0 {
		eng.votersMu.Unlock()
		return false
	}
	eng.voters[vstep][string(pubkey.Bytes())] = struct{}{} // above enforces one vote per pub key!
	eng.votersMu.Unlock()

	// for all the proposals in the vote, evaluate them.
	for _, bp := range voteMsg.Proposals {
		eng.processProposal(ctx, bp)
	}
	return true
}

func (eng *ConsensusEngine) getVote(node string, port uint16) (v *VoteMessage, err error) {
	url := fmt.Sprintf("https://%s:%d/wolk/vote/%d", node, port, eng.round)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error(fmt.Sprintf("[consensus:getVote] [Node%d] Get FAIL err: %v", eng.id, err))
		return v, err
	}

	var DefaultTransport http.RoundTripper = &http.Transport{
		Dial: (&net.Dialer{
			// limits the time spent establishing a TCP connection (if a new one is needed)
			Timeout:   255 * time.Second,
			KeepAlive: 3 * time.Second, // 60 * time.Second,
		}).Dial,
		//MaxIdleConns: 5,
		MaxIdleConnsPerHost: 25, // changed from 100 -> 25

		// limits the time spent reading the headers of the response.
		ResponseHeaderTimeout: 255 * time.Second,
		IdleConnTimeout:       4 * time.Second, // 90 * time.Second,

		// limits the time the client will wait between sending the request headers when including an Expect: 100-continue and receiving the go-ahead to send the body.
		ExpectContinueTimeout: 1 * time.Second,

		// limits the time spent performing the TLS handshake.
		TLSHandshakeTimeout: 255 * time.Second,
	}

	httpclient := &http.Client{Timeout: time.Millisecond * 500, Transport: DefaultTransport}
	// ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(intraregionMaxSetShareTime))
	// req.Cancel = ctx.Done()

	resp, err := httpclient.Do(req)
	if err != nil {
		//log.Error(fmt.Sprintf("[consensus:getVote] [Node%d] Get FAIL err: %v", eng.id, err))
		return v, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Trace(fmt.Sprintf("[consensus:getVote] [Node%d] ReadAll err: %v", eng.id, err))
		return v, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		//	log.Error("[consensus:getVote] VoteMessage FAIL", "body", string(body), "code", resp.StatusCode)
		return v, nil
	}
	if len(body) == 0 {
		log.Error(fmt.Sprintf("[consensus:getVote] [Node%d] ReadAll no body", eng.id), "url", url, "code", resp.StatusCode)
		return v, fmt.Errorf("No body")
	}
	v = new(VoteMessage)
	v, err = DecodeRLPVote(body)
	if err != nil {
		log.Error(fmt.Sprintf("[consensus:getVote] [Node%d] DecodeRLPVote", eng.id), "len(body2)", len(body), "err", err)
		return v, err
	}
	eng.voteCh <- v
	return v, nil
}

func (eng *ConsensusEngine) alreadyVoted(step uint64) (ok bool) {
	eng.vmu.Lock()
	defer eng.vmu.Unlock()
	_, ok = eng.votes[step]
	return ok
}

func (eng *ConsensusEngine) setKnownProposal(p *Proposal, subusers int) {
	eng.knownProposalsMu.Lock()
	defer eng.knownProposalsMu.Unlock()
	eng.knownProposals[p.RootHash] = p
	p.subusers = subusers
}

func (eng *ConsensusEngine) getKnownProposal(h common.Hash) (p *Proposal, ok bool) {
	eng.knownProposalsMu.RLock()
	defer eng.knownProposalsMu.RUnlock()
	p, ok = eng.knownProposals[h]
	return p, ok
}

func (eng *ConsensusEngine) verifyProposal(ctx context.Context, bp *Proposal) (err error) {
	if bp.verified {
		return nil
	}
	p, ok := eng.getKnownProposal(bp.RootHash)
	if !ok {
		return fmt.Errorf("[consensus:verifyProposal] Unknown proposal rootHash %x", bp.RootHash)
	}

	// TODO: permit looser constraint of being anything with tentative certs in previous round
	if bytes.Compare(eng.lastBlock.Hash().Bytes(), bp.ParentHash.Bytes()) != 0 {
		return fmt.Errorf("[consensus:verifyProposal] Parent hash mismatch to last block")
	}

	p.txs = make([]*Transaction, 0)
	for _, txhash := range bp.TxList {
		tx := eng.chain.wolktxpool.GetTransaction(txhash)
		if tx == nil {
			tx, _, _, ok, err = eng.chain.GetTransaction(ctx, txhash)
			if err != nil {
				return fmt.Errorf("[consensus:verifyProposal] GetTransaction ERR %v", err)
			} else if !ok {
				return fmt.Errorf("[consensus:verifyProposal] GetTransaction tx not found: %x", txhash)
			}
		}
		p.txs = append(p.txs, tx)
	}
	p.verified = true

	return nil
}

func (eng *ConsensusEngine) confirmProposal(ctx context.Context, bp *Proposal) (err error) {
	if bp.executed {
		// the certVotes being counted must contain the blockhash -- this is handled in execProposal
		bp.confirmed = true
	}
	return nil
}

func (eng *ConsensusEngine) executeProposal(ctx context.Context, bp *Proposal) (err error) {
	if bp.executed {
		return nil
	}
	p, ok := eng.getKnownProposal(bp.RootHash)
	if !ok {
		return fmt.Errorf("Unknown proposal")
	}
	if p.executing {
		return nil
	}
	p.executing = true
	txs := bp.txs

	var statedb *StateDB
	statedb, err = NewStateDB(ctx, eng.chain.Storage, bp.ParentHash)
	if err != nil {
		return err
	}
	// TODO: if we cant execute some of the transactions, that's ok,
	// because everyone ends up with the same result incommittedTxes
	statedb.ApplyTransactions(ctx, txs, false)
	var blk *Block
	blk, err = statedb.MakeBlock(ctx, statedb.committedTxes, eng.chain.policy, true)
	if err != nil {
		err = eng.chain.WriteBlock(eng.emptyBlk)
		log.Error(fmt.Sprintf("[consensus:executeProposal] [Node%d] MakeBlock ERR", eng.id), "ERR", err)
		return err
	}

	blk.Seed = bp.Seed
	// blk.Reserved = []byte(fmt.Sprintf("Mined-By-NODE%d", p.Proposer))

	// write to Cloudstore and broadcast
	err = eng.chain.WriteBlock(blk)
	if err != nil {
		log.Error(fmt.Sprintf("[consensus:executeProposal] [Node%d] block proposal writeblock ERR", eng.id), "ERR", err)
		return err
	}
	eng.setBlock(blk)
	go eng.chain.protocolManager.BroadcastBlock(blk)
	p.BlockHash = blk.Hash()

	log.Info(fmt.Sprintf("[consensus:executeProposal] [Node%d] executeProposal", eng.id), "blockHash", p.BlockHash)
	p.executed = true
	return nil
}

func (eng *ConsensusEngine) processProposal(ctx context.Context, bp *Proposal) bool {
	round := eng.round

	// check if we have already considered this proposal
	if p, ok := eng.getKnownProposal(bp.RootHash); ok {
		for _, v := range bp.Votes {
			// TODO: validate VoteVRFProof before recording
			// log.Trace("[consensus:processProposal] RECORDING VOTE for REAL PROPOSAL", "round", eng.round, "step", v.Step, "voter", v.Voter, "subusers", v.Subusers)
			p.recordVotes(v)
		}
	} else if bp.IsEmpty() {
		log.Info(fmt.Sprintf("[consensus:processProposal] [Node%d] PROCESS EMPTY PROPOSAL", eng.id))
		eng.setKnownProposal(bp, 0)
	} else {
		//TODO: not properly verified!!
		pubkey, err := bp.RecoverPubkey()
		if err != nil {
			return false
		}
		subusers := verifySort(pubkey, bp.VRF, bp.Proof, eng.sortitionSeed(round), role(roleProposer, round), round, expectedTokensProposer, TokenPerUser, TotalTokenAmount)
		if subusers <= 0 {
			log.Error(fmt.Sprintf("[consensus:processProposal] [Node%d] round #%d proposal verifySort error: %v", round, eng.id, errors.New("malicious proposal")))
			return false
		} else if !bytes.Equal(maxPriority(bp.VRF, subusers), bp.Prior) {
			log.Error(fmt.Sprintf("[consensus:processProposal] [Node%d] round #%d proposal verifySort error: %v", round, eng.id, errors.New("max priority mismatch")))
			return false
		}
		// log.Trace("[consensus:processProposal] setKnownProposal", "subusers", subusers, "bp", bp)
		eng.setKnownProposal(bp, subusers)
	}
	return true
}

func (eng *ConsensusEngine) sendVote0(ctx context.Context) (err error) {
	round := eng.round
	FPlen := eng.round - eng.chain.LastFinalized()
	eng.chain.setConsensusStatus(fmt.Sprintf("%d sendVote0", eng.round))
	eng.step = voteProposal

	lastBlock := eng.lastBlock
	parentHash := lastBlock.Hash()
	// check if user is in committee using Sortition with  votes > 0, because only committee members can originate a message
	eng.executorVRF, eng.executorProof, eng.executorSubusers = sortition(eng.privkey, eng.sortitionSeed(round), role(roleVoter, round), expectedTokensVoter, TokenPerUser, TotalTokenAmount)
	vote := &VoteMessage{
		Voter:       uint8(eng.id),
		BlockNumber: eng.round,
		ParentHash:  parentHash,
		Step:        eng.step,
		Sub:         uint64(eng.executorSubusers),
		VRF:         eng.executorVRF,
		Proof:       eng.executorProof,
		Proposals:   make([]*Proposal, 0),
	}

	eng.proposerVRF, eng.proposerProof, eng.proposerSubusers = sortition(eng.privkey, eng.sortitionSeed(round), role(roleProposer, round), expectedTokensProposer, TokenPerUser, TotalTokenAmount)
	if eng.proposerSubusers > 0 { // && eng.round%5 != 0
		txs, _ := eng.chain.TxPool().PendingTxlist() //get all pending txs.

		if FPlen >= maxTrailingBlocks && len(txs) > tentativeTxnCap {
			txs = txs[0 : tentativeTxnCap-1]
		} else if len(txs) > maxTransactionsProposal {
			txs = txs[0 : maxTransactionsProposal-1]
		}
		seed, seedproof, err := eng.vrfSeed(lastBlock)
		if err != nil {
			log.Error(fmt.Sprintf("[consensus:sendVote0] [Node%d] vrfSeed ERR", eng.id), "ERR", err)
			return err
		}
		txhashes := make([]common.Hash, len(txs))
		b := make([]byte, len(txs)*32+len(seed))
		for i, tx := range txs {
			txhashes[i] = tx.Hash()
			copy(b[i*32:i*32+32], tx.Hash().Bytes())
		}
		copy(b[len(txs)*32:], seed[0:])

		roothash := common.BytesToHash(wolkcommon.Computehash(b))
		proposal := &Proposal{
			Proposer:   eng.id,
			ParentHash: parentHash,
			Seed:       seed,
			SeedProof:  seedproof,
			RootHash:   roothash,
			TxList:     txhashes,
			Prior:      maxPriority(eng.proposerVRF, eng.proposerSubusers),
			VRF:        eng.proposerVRF,
			Proof:      eng.proposerProof,
			Signature:  make([]byte, 65),
		}

		if eng.executorSubusers > 0 {
			var prf VoteVRFProof
			prf.VRF = eng.executorVRF
			prf.Proof = eng.executorProof
			prf.Step = voteProposal
			prf.Voter = eng.id
			prf.Subusers = uint64(eng.executorSubusers)
			prf.Sign(eng.privkey)
			proposal.recordVotes(&prf)
		}

		// sign proposal and consider it the max proposal
		_, err = proposal.Sign(eng.privkey)
		if err != nil {
			log.Error(fmt.Sprintf("[consensus:sendVote0] [Node%d] sign ERR", eng.id), "ERR", err)
		}
		eng.setKnownProposal(proposal, eng.proposerSubusers)
		// the generated proposal is the only proposal
		vote.Proposals = append(vote.Proposals, proposal)
		vote.Seed = proposal.Seed
		vote.Step = voteProposal
		eng.step = voteProposal
		eng.chain.setConsensusStatus(fmt.Sprintf("%d/%d BroadcastVote", eng.round, eng.step))
		log.Info(fmt.Sprintf("[consensus:sendVote0] [Node%d] VOTED on step 0 with proposal", eng.id), "proposal", proposal)
	} else {
		log.Info(fmt.Sprintf("[consensus:sendVote0] [Node%d] VOTED on step 0", eng.id))
	}

	vote.Sign(eng.privkey)
	votedOnSteps := make([]uint64, 1)
	votedOnSteps[0] = 0
	eng.recordInternalVote(vote, votedOnSteps)
	eng.lastVote = vote
	eng.lastVoteTime = time.Now()
	eng.lastStepTime = time.Now()

	return nil
}

func (eng *ConsensusEngine) sendVote(ctx context.Context) bool {
	if IsVoting == 0 {
		return false
	}

	for _, r := range eng.chain.Storage.Registry {
		go func(node string) {
			eng.getVote(node, r.GetPort())
		}(r.GetStorageNode())
	}

	voteMsg := eng.lastVote
	votedOnSteps := make([]uint64, 0)
	voteEmpty := false

	for step := uint64(0); step < 5; step++ {
		var tallyStep uint64
		var voteStep uint64
		// TODO: do analysis of how binomial sampling of X results in 5X nodes threshold for 8 nodes
		required := expectedTokensTentative
		switch step {
		case 0:
			// voteReduction has vote for maxPriority proposal from sendVote0 (a set of txlist)
			// requirement: none, but any transactions not in txpool could be fetched from cloudstore into txpool
			tallyStep = voteProposal // we desire that the proposer "sees" the certified blockhash vs empty
			voteStep = voteReduction
		case 1:
			// voteSoft has vote for highest priority proposal where:
			//  (a) all the chunks of TxList are readable from txpool or cloudstore and pass validateTx
			//  (b) all txs are *valid*
			// EMPTY is chosen if `voteReduction` did not succeed in getting enough valid proposals (totalcnt < required)
			tallyStep = voteReduction
			voteStep = voteSoft
		case 2:
			// voteCert has vote for highest voteSoft proposal where:
			//  (a) the previous block's state is read
			//  (b) ApplyTransactions are executed, culminating in a WriteBlock operation
			// EMPTY is chosen if for some reason the ApplyTransactions on the TxList (chosen from `voteSoft`) could not succeed or required votes from `voteSoft` not achieved in time.
			tallyStep = voteSoft
			voteStep = voteCert
		case 3:
			// voteNext has vote for highest voteCert proposal where:
			//  (a) the blockHash written in verified to be readable
			// EMPTY is chosen if for some reason the block is not readable or required votes from `voteNext` not achieved in time.
			tallyStep = voteCert // AND voteSoft?
			voteStep = voteNext
		case 4:
			// need to wait until 75% of cert votes are in before nextvotes start
			tallyStep = voteNext
			voteStep = voteDone
		}
		bestProposal, cnt, totalcnt, voted := eng.getBestProposal(tallyStep)
		if bestProposal != nil {

			if cnt >= required || (totalcnt > required && voteStep == voteReduction) {
				vote := false

				log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v) - - - getBestProposal (top:%d|tot:%d)", eng.id, eng.round, eng.step, getStepType(eng.step), cnt, totalcnt), "step", step)
				if voteStep == voteDone && voted {
					eng.vmu.RLock()
					certvote := eng.votes[tallyStep]
					eng.vmu.RUnlock()
					if certvote != nil {
						b, err := eng.getBlockByHash(ctx, certvote.BlockHash, BlockReadOptions{ReadFromCloudstore: true})
						if err == nil && b != nil {
							//log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v) - - - extendChain (sub:%d|top:%d|tot:%d)", eng.id, eng.round, step, getStepType(step), voteMsg.SubUser(), cnt, totalcnt), "eng.step", eng.step, "tallyStep", getStepType(tallyStep), "certvote", certvote)
							certvote.absorbProposalVotes(bestProposal, tallyStep)
							//log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v) - - - after absorbProposalVotes", eng.id, eng.round, step, getStepType(step)), "eng.step", eng.step, "tallyStep", getStepType(tallyStep), "certvote", certvote)
							cert, err := eng.chain.extendChain(ctx, b, certvote)
							if err == nil {
								eng.addCertificate(cert)
							}
						}
					}
				} else if eng.alreadyVoted(voteStep) {
					// do nothing
					log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v) - - - ALREADY VOTED (sub:%d|top:%d|tot:%d)", eng.id, eng.round, step, getStepType(step), voteMsg.SubUser(), cnt, totalcnt), "seed", fmt.Sprintf("%x", voteMsg.Seed), "h", fmt.Sprintf("%x", voteMsg.Seed))
				} else if (voteStep == voteSoft) && bestProposal.verified == false { // verifiedTxAvailble
					// voteSoft has vote for highest priority proposal where:
					//  (a) all the chunks of TxList are readable from txpool or cloudstore and pass validateTx
					//  (b) all txs are *valid*
					// EMPTY is chosen if `voteReduction` did not succeed in getting enough valid proposals (totalcnt < required)
					err := eng.verifyProposal(ctx, bestProposal)
					if err != nil {
						log.Error("[consensus:sendVote] verifyProposal", "err", err)
						// TODO: we should REJECT this proposal and choose the NEXT proposal in the countvotes process
					} else if bestProposal.verified {
						vote = true
					}
				} else if (voteStep == voteCert) && bestProposal.executed == false { // blockWritten
					// voteCert has vote for highest voteSoft proposal where:
					//  (a) the previous block's state is read
					//  (b) ApplyTransactions are executed, culminating in a WriteBlock operation
					// EMPTY is chosen if for some reason the ApplyTransactions on the TxList (chosen from `voteSoft`) could not succeed or required votes from `voteSoft` not achieved in time.
					err := eng.executeProposal(ctx, bestProposal)
					if err != nil {
						log.Error("[consensus:sendVote] executeProposal")
						// TODO: we should REJECT this proposal and choose the NEXT proposal in the countvotes process
					} else if bestProposal.executed {
						vote = true
					}
				} else if (voteStep == voteNext) && bestProposal.confirmed == false {
					// voteNext requires that the blockhash derived by the executeproposal is the same being confirmed
					eng.confirmProposal(ctx, bestProposal)
					if bestProposal.confirmed {
						vote = true
					}
				} else {
					vote = true
				}

				if vote {
					var prf VoteVRFProof
					prf.VRF = eng.executorVRF
					prf.Proof = eng.executorProof
					prf.Step = voteStep
					prf.Voter = eng.id
					prf.Subusers = uint64(eng.executorSubusers)
					prf.Sign(eng.privkey)
					if voteStep >= voteCert {
						prf.BlockHash = bestProposal.BlockHash
						voteMsg.BlockHash = bestProposal.BlockHash
					}

					voteMsg.Seed = bestProposal.Seed
					voteMsg.Step = voteStep
					voteMsg.addProposal(bestProposal)

					bestProposal.recordVotes(&prf)
					eng.step = voteStep
					eng.chain.setConsensusStatus(fmt.Sprintf("%d/%d: %s", eng.round, eng.step, displayHash(bestProposal.RootHash)))
					log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v) ---> ADDED VOTE (sub:%d|top:%d|tot:%d)", eng.id, eng.round, voteStep, getStepType(voteStep), voteMsg.Sub, cnt, totalcnt), "h", voteMsg.BlockHash)

					votedOnSteps = append(votedOnSteps, voteStep)
				}
			} else if totalcnt >= expectedTokensTentative || (time.Since(eng.lastStepTime) > lambdaStep) {
				// vote empty
				if totalcnt >= expectedTokensTentative {
					log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v)   >>>>>>>> RECORDING EMPTY: THRESH EXCEEDED (total %d >= %d tentative) (sub:%d|top:%d)", eng.id, eng.round, step, getStepType(step), totalcnt, expectedTokensTentative, voteMsg.SubUser(), cnt))
				} else {
					log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v)   >>>>>>>> RECORDING EMPTY: TIMEOUT (total:%d|tentative:%d) (sub:%d|top:%d)", eng.id, eng.round, step, getStepType(step), totalcnt, expectedTokensTentative, voteMsg.SubUser(), cnt), "tm", time.Since(eng.lastStepTime))
				}
				voteEmpty = true
			} else if cnt > 0 || totalcnt > 0 {
				log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v)   >>>>>>>> INSUFFICENT TOTAL COUNT (total:%d|required:%d) (sub:%d|top:%d)", eng.id, eng.round, step, getStepType(step), totalcnt, required, voteMsg.SubUser(), cnt), "tm", time.Since(eng.lastStepTime))
			}
		} else {
			if time.Since(eng.lastStepTime) > lambdaStep {
				voteEmpty = true
				log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v)   >>>>>>>> RECORDING EMPTY: no best proposal (total:%d|required:%d) (sub:%d|top:%d)", eng.id, eng.round, step, getStepType(step), totalcnt, required, voteMsg.SubUser(), cnt))
			}
		}
	}
	if voteEmpty {

		emptyBlk := eng.emptyBlk
		log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d   >>>>>>>> RECORDING EMPTY Vote", eng.id, eng.round), "empty", eng.emptyBlk)
		emptyProposal := NewProposal(emptyBlk)
		errE := eng.chain.WriteBlock(emptyBlk)
		if errE != nil {
			log.Error(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d  >>>>>>>> WriteBlock ERR for Empty Block", eng.id, eng.round))
			return false
		}
		step := uint64(voteNext)
		if _, ok := eng.votes[step]; !ok {
			var prf VoteVRFProof
			prf.VRF = eng.executorVRF
			prf.Proof = eng.executorProof
			prf.Step = step
			prf.Voter = eng.id
			prf.Subusers = uint64(eng.executorSubusers)
			prf.Sign(eng.privkey)
			voteMsg.BlockHash = emptyBlk.Hash()
			voteMsg.Seed = emptyBlk.Seed
			voteMsg.Step = step
			voteMsg.addProposal(emptyProposal)
			log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d Step%d (%v)   >>>>>>>> RECORDING EMPTY VOTE (sub:%d)", eng.id, eng.round, step, getStepType(step), voteMsg.SubUser()))
			emptyProposal.recordVotes(&prf)
			eng.step = voteEmptyBlock
			votedOnSteps = append(votedOnSteps, step)
		}
	}

	if len(votedOnSteps) > 0 {
		_, err := voteMsg.Sign(eng.privkey)
		if err != nil {
			log.Error(fmt.Sprintf("[consensus:sendVote] [Node%d] sign err", eng.id), "err", err)
			return false
		}

		// record vote
		eng.lastStepTime = time.Now()
		eng.recordInternalVote(voteMsg, votedOnSteps)
		log.Info(fmt.Sprintf("[consensus:sendVote] [Node%d] round #%d recordInternalVote", eng.id, eng.round), "steps", fmt.Sprintf("%v", votedOnSteps))
		return true
	}
	return false
}

// under construction
func (eng *ConsensusEngine) setLastBlock(lastBlock *Block, votes uint64) {
	h := lastBlock.Hash()
	eng.allLastBlock[h] = lastBlock
	eng.allLastBlockVotes[h] = votes
	emptyBlk := eng.emptyBlock(eng.round, lastBlock)
	eng.allEmptyBlk[h] = emptyBlk
	eng.allEmptyHash[h] = emptyBlk.Hash()
	prev := uint64(0)
	if eng.lastBlock != nil {
		prev = eng.allLastBlockVotes[eng.lastBlock.Hash()]
	}
	if votes > prev {
		eng.lastBlock = lastBlock
		eng.lastBlockVotes = votes
		eng.ParentHash = h
		eng.emptyBlk = eng.emptyBlock(eng.round, lastBlock)
		eng.emptyHash = eng.emptyBlk.Hash()
	}
}

func (eng *ConsensusEngine) consensusLoop(ctx context.Context) (err error) {
	round := eng.round
	eng.step = 0

	dump := time.NewTicker(200 * time.Millisecond)
	waitForStep1 := time.NewTicker(250 * time.Millisecond)
	timeoutForStep := time.NewTimer(24 * time.Hour)

	for {
		select {
		case <-ctx.Done(): // if some other round has been finalized =)
			log.Error(fmt.Sprintf("[Node%d] ctxdone", eng.id))
			return errCancel
		case <-dump.C:
			eng.dumpChannelInfo()
		case <-waitForStep1.C: // valueProposal
			if eng.chain.isChainReady() == false {
				var preemptiveStateDB *StateDB
				eng.chain.muPreemptiveStateDB.RLock()
				preemptiveStateDB = eng.chain.PreemptiveStateDB
				eng.chain.muPreemptiveStateDB.RUnlock()

				log.Error(fmt.Sprintf("[consensus:consensusLoop] [Node%d] round #%d - waitForStep1 chain not ready", eng.id, round), "currentRound", eng.chain.currentRound(), "statedb", preemptiveStateDB)

			} else if eng.chain.currentRound() > eng.round {
				log.Trace(fmt.Sprintf("[consensus:consensusLoop] [Node%d] Moved onto new round #%d", eng.id, eng.chain.currentRound()))

				return errCancel
			} else if eng.chain.currentRound() == eng.round {
				eng.chain.setConsensusStatus(fmt.Sprintf("%d: proposing block", eng.round))

				// TODO: get this from previous round
				votes := uint64(1)
				eng.setLastBlock(eng.chain.lastBlock(), votes)

				waitForStep1.Stop()
				errVote := eng.sendVote0(ctx)
				if errVote == nil {
					eng.step = 1
					log.Info(fmt.Sprintf("[consensus:consensusLoop] [Node%d] sendVote0 DONE", eng.id), "eng.step", eng.step)
					timeoutForStep = time.NewTimer(lambdaHeartbeat)
				} else {
					log.Error(fmt.Sprintf("[consensus:consensusLoop] [Node%d] sendVote0 ERR", eng.id), "errVote", errVote)

				}
			}
		case <-timeoutForStep.C:
			if eng.chain.currentRound() > eng.round {
				log.Trace(fmt.Sprintf("[consensus:consensusLoop] [Node%d] Moved onto new round #%d", eng.id, eng.chain.currentRound()))
				return nil
			}
			lastblk := eng.chain.lastBlock()
			if bytes.Equal(lastblk.Bytes(), eng.ParentHash.Bytes()) {
				log.Info(fmt.Sprintf("[consensus:consensusLoop] [Node%d] Round #%d - Updating lastblock #%d %x)", eng.id, eng.chain.currentRound(), lastblk.Number(), lastblk.Hash()))
				votes := uint64(1) // not used yet
				eng.setLastBlock(lastblk, votes)
			}

			eng.sendVote(ctx)
			timeoutForStep = time.NewTimer(lambdaHeartbeat)
		case voteMsg := <-eng.voteCh:
			eng.processVote(ctx, voteMsg)
			break
		case b := <-eng.blockCh:
			eng.setBlock(b)
			break
		}
	}
}

//EmptyBlock: A block that doesn't involve state transition
// round not really used
func (eng *ConsensusEngine) emptyBlock(round uint64, p *Block) (e *Block) {
	e = p.Copy() //copying previous state without txns
	e.ParentHash = p.Hash()
	e.BlockNumber = eng.round
	e.Seed = wolkcommon.Computehash(e.ParentHash.Bytes())
	return e
}
