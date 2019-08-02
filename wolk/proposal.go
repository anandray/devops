package wolk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
)

type VoteVRFProof struct {
	Voter     uint64      `json:"voter"`    // who is voting
	Step      uint64      `json:"step"`     // what they are voting on
	Subusers  uint64      `json:"subusers"` // with how much stake
	VRF       []byte      `json:"vrf"`
	Proof     []byte      `json:"proof"`
	BlockHash common.Hash `json:"blockHash"`
	Signature []byte      `json:"signature"`
	validated int
	mu        sync.RWMutex
}

// short hash
func (p *VoteVRFProof) ShortHash() (hash common.Hash) {
	unsigned := &VoteVRFProof{
		Voter:     p.Voter,
		Step:      p.Step,
		Subusers:  p.Subusers,
		VRF:       p.VRF,
		Proof:     p.Proof,
		Signature: make([]byte, 0),
	}
	enc, _ := rlp.EncodeToBytes(&unsigned)
	return common.BytesToHash(wolkcommon.Computehash(enc))
}

func (p *VoteVRFProof) Sign(priv *crypto.PrivateKey) ([]byte, error) {
	proofhash := p.ShortHash()
	sig, err := priv.Sign(proofhash.Bytes())
	if err != nil {
		return nil, err
	}
	p.Signature = sig
	return sig, nil
}

func (p *VoteVRFProof) Copy() (n *VoteVRFProof) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	n = new(VoteVRFProof)
	n.Voter = p.Voter
	n.Step = p.Step
	n.Subusers = p.Subusers
	n.Proof = p.Proof
	n.Signature = p.Signature
	n.VRF = p.VRF
	n.validated = p.validated
	return n
}

func (p *VoteVRFProof) RecoverPubkey() (*crypto.PublicKey, error) {
	return crypto.RecoverPubkey(p.Signature)
}

// verify Vote VRF with m
func (p *VoteVRFProof) Verify(m []byte) error {
	pubkey, err := p.RecoverPubkey()
	if err != nil { // note that you can vote on a proposal for an empty block (which has no signature) but your vote still has a signature!
		return err
	}

	if err = pubkey.VerifyVRF(p.Proof, m); err != nil {
		return err
	}
	return nil
}

func (prf *VoteVRFProof) String() string {
	s, _ := json.Marshal(prf)
	return string(s)
}

// Proposal is used inside Vote
type Proposal struct {
	Proposer   uint64        `json:"proposer"`
	ParentHash common.Hash   `json:"parentHash"`
	Seed       []byte        `json:"seed"`
	SeedProof  []byte        `json:"seedProof"`
	RootHash   common.Hash   `json:"rootHash"` // hash of all the data of tx put together
	TxList     []common.Hash `json:"txlist"`   // Proposer must have validated that these txs are in cloudstore
	VRF        []byte        `json:"vrf"`      // vrf of user's sortition hash
	Proof      []byte        `json:"proof"`
	Signature  []byte        `json:"signature"`
	Prior      []byte        `json:"prior"`

	Votes []*VoteVRFProof `json:"softVotesVRFProof"`

	verified  bool
	executing bool
	executed  bool
	confirmed bool
	subusers  int
	txs       []*Transaction
	BlockHash common.Hash `json:"blockHash"`
	mu        sync.RWMutex
}

func (bp *Proposal) Copy() (n *Proposal) {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	n = new(Proposal)
	n.Proposer = bp.Proposer
	n.ParentHash = bp.ParentHash
	n.Seed = bp.Seed
	n.SeedProof = bp.SeedProof
	n.RootHash = bp.RootHash
	n.TxList = bp.TxList
	n.VRF = bp.VRF
	n.Proof = bp.Proof
	n.Prior = bp.Prior
	n.Signature = bp.Signature
	n.verified = bp.verified
	n.subusers = bp.subusers
	for _, v := range bp.Votes {
		n.Votes = append(n.Votes, v.Copy())
	}
	return n
}

func (bp *Proposal) MakeCertificate() (n *Proposal) {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	n = new(Proposal)
	n.Proposer = bp.Proposer
	n.ParentHash = bp.ParentHash
	n.Seed = bp.Seed
	n.SeedProof = bp.SeedProof
	n.RootHash = bp.RootHash
	n.TxList = bp.TxList
	n.VRF = bp.VRF
	n.Proof = bp.Proof
	n.Prior = bp.Prior
	n.Signature = bp.Signature
	n.verified = bp.verified
	n.subusers = bp.subusers
	for _, v := range bp.Votes {
		if v.Step == voteNext {
			n.Votes = append(n.Votes, v.Copy())
		}
	}
	return n
}

func (bp *Proposal) hasVote(id uint64, step uint64) bool {
	for _, v := range bp.Votes {
		if id == v.Voter && v.Step == step {
			return true
		}
	}
	return false
}

type Proposals []*Proposal

const (
	voteProposal   = 0
	voteReduction  = 1
	voteSoft       = 2
	voteCert       = 3
	voteNext       = 4
	voteDone       = 5
	voteEmptyBlock = 6
)

const (
	SkipSeed = "skip"
)

func (p *Proposal) countVotes(step uint64, seed []byte, bn uint64) (cnt int) {
	// tally from SoftVotesVRFProof -- assumption is VRF Proofs are validated in processProposal
	cnt = 0
	for _, prf := range p.Votes {
		if prf.validated == 0 && step > voteReduction {
			// check proof
			err := prf.Verify(constructSeed(seed, role(roleVoter, bn)))
			if err == nil {
				prf.validated = 1
				//	log.Info("[proposal:countVotes] valid proof", "bn", bn, "step", step)
			} else {
				prf.validated = -1
				log.Error("[proposal:countVotes] *invalid* proof", "bn", bn, "step", step)
			}
		}
		if step <= voteReduction || prf.validated == 1 {
			if step == voteNext {
				if prf.Step == step {
					if bytes.Compare(prf.BlockHash.Bytes(), p.BlockHash.Bytes()) == 0 {
						cnt += int(prf.Subusers)
					}
				}
			} else if prf.Step == step {
				cnt += int(prf.Subusers)
			}
		}
	}

	return cnt // len(b.SoftVotes)
}

func getStepType(code uint64) interface{} {
	switch code {
	case 0:
		return "PROPOSAL"
	case 1:
		return "REDUCTION"
	case 2:
		return "SOFT"
	case 3:
		return "CERT"
	case 4:
		return "NEXT"
	case 5:
		return "DONE"
	case 6:
		return "EMPTY"
	default:
		return code
	}
}

func NewProposal(b *Block) *Proposal {
	return &Proposal{
		RootHash:   b.ParentHash,
		Seed:       b.Seed,
		ParentHash: b.ParentHash,
	}
}

func (p *Proposal) IsEmpty() bool {
	//proposal is empty if: (1) len(TxList) == 0 and (2) ParentHash matches RootHash
	if p == nil {
		return false
	}
	return bytes.Compare(p.RootHash.Bytes(), p.ParentHash.Bytes()) == 0 && len(p.TxList) == 0
}

func (bp *Proposal) recordVotes(sv *VoteVRFProof) bool {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	for _, q := range bp.Votes {
		if q.Voter == sv.Voter && q.Step == sv.Step {
			return false
		}
	}
	bp.Votes = append(bp.Votes, sv)

	return true
}

func (bp *Proposal) absorbProposalVotes(kp *Proposal, step uint64) bool {
	for _, q := range kp.Votes {
		if q.Step == step {
			if bp.recordVotes(q) {
				//		fmt.Printf("  --- absorbProposalVotes from...: %s\n", q.String())
			}
		}
	}
	return true
}

func (p *Proposal) dump() {
	fmt.Printf("  Proposer %d Seed %x empty? %t", p.Proposer, p.Seed, p.IsEmpty())
	fmt.Printf("   Votes: [")
	for s := uint64(0); s <= uint64(6); s++ {
		var voters []uint64
		for _, v := range p.Votes {
			if v.Step == s {
				voters = append(voters, v.Voter)
			}
		}
		if len(voters) > 0 {
			fmt.Printf("(%d: %v)", s, voters)
		}
	}
	fmt.Printf("]\n")
}

func DecodeRLPProposal(bytes []byte) (p *Proposal, err error) {
	var po Proposal
	err = rlp.DecodeBytes(bytes, &po)
	if err != nil {
		return p, err
	}
	return &po, nil
}

func (p *Proposal) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Proposal) Deserialize(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *Proposal) RecoverPubkey() (pk *crypto.PublicKey, err error) {
	return crypto.RecoverPubkey(p.Signature)
}

func (p *Proposal) Address() (a common.Address) {
	pubkey, err := crypto.RecoverPubkey(p.Signature)
	if err != nil {
		return a
	}
	return pubkey.Address()
}

func (p *Proposal) Verify(weight uint64, m []byte) (err error) {
	// verify vrf
	pubkey, err := p.RecoverPubkey()
	if err != nil {
		return err
	}
	if err = pubkey.VerifyVRF(p.Proof, m); err != nil {
		return err
	}

	// verify priority
	subusers := subUsers(expectedTokensProposer, weight, p.VRF, 1)
	if !bytes.Equal(maxPriority(p.VRF, subusers), p.Prior) {
		return fmt.Errorf("max priority mismatch")
	}
	return nil
}

func (p *Proposal) Hash() (h common.Hash) {
	enc, _ := rlp.EncodeToBytes(&p)
	return common.BytesToHash(wolkcommon.Computehash(enc))
}

//# go:generate gencodec -type Proposal -field-override transactionMarshaling -out proposal_json.go

//marshalling store external type, if different
/*type proposalMarshaling struct {
}*/

func (p *Proposal) String() string {
	if p != nil {
		jsonb, _ := json.Marshal(p)
		return string(jsonb)
	} else {
		return fmt.Sprint("{}")
	}
}

func (p *Proposal) Hex() string {
	return fmt.Sprintf("%x", p.Bytes())
}

func (p *Proposal) Bytes() (enc []byte) {
	enc, _ = rlp.EncodeToBytes(&p)
	return enc
}

// short hash
func (p *Proposal) ShortHash() (hash common.Hash) {
	unsigned := &Proposal{
		Proposer:   p.Proposer,
		ParentHash: p.ParentHash,
		Seed:       p.Seed,
		SeedProof:  p.SeedProof,
		RootHash:   p.RootHash,
		TxList:     p.TxList,
		VRF:        p.VRF,
		Proof:      p.Proof,
		Prior:      p.Prior,
		Signature:  make([]byte, 0),
		//TODO: Votes validations
	}
	enc, _ := rlp.EncodeToBytes(&unsigned)
	return common.BytesToHash(wolkcommon.Computehash(enc))
}

// full hash
func (p *Proposal) Sign(priv *crypto.PrivateKey) ([]byte, error) {
	proposalhash := p.ShortHash()
	sig, err := priv.Sign(proposalhash.Bytes())
	if err != nil {
		return nil, err
	}
	p.Signature = sig
	return sig, nil
}

// sortition runs cryptographic selection procedure and returns vrf,proof and amount of selected sub-users.
func sortition(privkey *crypto.PrivateKey, seed, role []byte, expectedNum int, tokens uint64, totaltokenamount uint64) (vrf, proof []byte, subusers int) {
	vrf, proof, _ = privkey.Evaluate(constructSeed(seed, role))
	subusers = subUsers(expectedNum, tokens, vrf, totaltokenamount)
	return
}

// verifySort verifies the vrf and returns the amount of selected sub-users.
// NOTE: weight is not part of the input
func verifySort(pk *crypto.PublicKey, vrf, proof, seed, role []byte, round uint64, expectedNum int, tokens uint64, totaltokenamount uint64) (subusers int) {
	if err := pk.VerifyVRF(proof, constructSeed(seed, role)); err != nil {
		//		log.Info(fmt.Sprintf("[consensus:verifySort] VerifyVRF ERR: pk[%x] proof[%x] seed[%x] role[%x] seed||role [%x] ", eng.pubkey.Bytes(), vrf, seed, role, constructSeed(seed, role)))
		return 0
	}
	return subUsers(expectedNum, tokens, vrf, totaltokenamount)
}

// role returns the role bytes from current round
func role(iden string, round uint64) []byte {
	return bytes.Join([][]byte{
		[]byte(iden),
		wolkcommon.UIntToByte(round),
	}, nil)
}

func maxPriority(vrf []byte, users int) []byte {
	var maxPrior []byte
	for i := 0; i < users; i++ {
		prior := wolkcommon.Computehash(bytes.Join([][]byte{vrf, wolkcommon.UIntToByte(uint64(i))}, nil))
		if bytes.Compare(prior, maxPrior) > 0 {
			maxPrior = prior
		}
	}
	return maxPrior
}

// subUsers return the selected amount of sub-users determined from the mathematics protocol.
func subUsers(expectedNum int, weight uint64, vrf []byte, totaltokenamount uint64) int {
	binomial := NewBinomial(int64(weight), int64(expectedNum), int64(totaltokenamount))
	//binomial := NewApproxBinomial(int64(expectedNum), weight)
	//binomial := &distuv.Binomial{
	//	N: float64(weight),
	//	P: float64(expectedNum) / float64(TotalTokenAmount()),
	//}
	// hash / 2^hashlen ∉ [ ∑0,j B(k;w,p), ∑0,j+1 B(k;w,p))
	hashBig := new(big.Int).SetBytes(vrf)
	maxHash := new(big.Int).Exp(big.NewInt(2), big.NewInt(common.HashLength*8), nil)
	hash := new(big.Rat).SetFrac(hashBig, maxHash)
	var lower, upper *big.Rat
	j := 0
	for uint64(j) <= weight {
		if upper != nil {
			lower = upper
		} else {
			lower = binomial.CDF(int64(j))
		}
		upper = binomial.CDF(int64(j + 1))
		if hash.Cmp(lower) >= 0 && hash.Cmp(upper) < 0 {
			break
		}
		j++
	}
	if uint64(j) > weight {
		j = 0
	}
	return j
}

// constructSeed construct a new bytes for vrf generation.
func constructSeed(seed, role []byte) []byte {
	return bytes.Join([][]byte{seed, role}, nil)
}
