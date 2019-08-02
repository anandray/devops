package wolk

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
)

const (
	// message type
	VOTE = iota
	BLOCK_PROPOSAL
	FORK_PROPOSAL
	BLOCK
)

type VoteMessage struct {
	Voter       uint8       `json:"voter"`
	BlockNumber uint64      `json:"blockNumber"`
	ParentHash  common.Hash `json:"parentHash"`
	Seed        []byte      `json:"seed"`
	BlockHash   common.Hash `json:"blockHash"` // filled on certvote step
	Step        uint64      `json:"step"`      // 0..inf
	Sub         uint64      `json:"subuser"`
	VRF         []byte      `json:"vrf"`
	Proof       []byte      `json:"proof"`
	Signature   []byte      `json:"signature"`
	Proposals   []*Proposal `json:"proposals"` //vote verification should not include proposal
	mu          sync.RWMutex
}

func (voteMsg *VoteMessage) absorbProposalVotes(kp *Proposal, step uint64) bool {
	voteMsg.mu.Lock()
	defer voteMsg.mu.Unlock()

	for _, p := range voteMsg.Proposals {
		if bytes.Compare(kp.RootHash.Bytes(), p.RootHash.Bytes()) == 0 {
			p.absorbProposalVotes(kp, step)
		}
	}
	return true
}

func (voteMsg *VoteMessage) MakeCertificate(h common.Hash) (n *VoteMessage) {
	voteMsg.mu.RLock()
	defer voteMsg.mu.RUnlock()
	n = new(VoteMessage)
	n.Voter = voteMsg.Voter
	n.BlockNumber = voteMsg.BlockNumber
	n.ParentHash = voteMsg.ParentHash
	n.BlockHash = voteMsg.BlockHash
	n.Seed = voteMsg.Seed
	n.Step = voteMsg.Step
	n.Sub = voteMsg.Sub
	n.VRF = voteMsg.VRF
	n.Proof = voteMsg.Proof
	n.Signature = voteMsg.Signature
	for _, bp := range voteMsg.Proposals {
		//		if bytes.Compare(bp.RootHash.Bytes(), h.Bytes()) == 0 {
		n.Proposals = append(n.Proposals, bp.MakeCertificate())
		//		}
	}
	return
}

func (voteMsg *VoteMessage) Copy() (n *VoteMessage) {
	voteMsg.mu.RLock()
	defer voteMsg.mu.RUnlock()
	n = new(VoteMessage)
	n.Voter = voteMsg.Voter
	n.BlockNumber = voteMsg.BlockNumber
	n.ParentHash = voteMsg.ParentHash
	n.BlockHash = voteMsg.BlockHash
	n.Seed = voteMsg.Seed
	n.Step = voteMsg.Step
	n.Sub = voteMsg.Sub
	n.VRF = voteMsg.VRF
	n.Proof = voteMsg.Proof
	n.Signature = voteMsg.Signature
	for _, bp := range voteMsg.Proposals {
		n.Proposals = append(n.Proposals, bp.Copy())
	}
	return
}

func (voteMsg *VoteMessage) addProposal(p *Proposal) {
	voteMsg.mu.Lock()
	defer voteMsg.mu.Unlock()
	for _, bp := range voteMsg.Proposals {
		if bytes.Compare(p.RootHash.Bytes(), bp.RootHash.Bytes()) == 0 {
			return
		}
	}
	voteMsg.Proposals = append(voteMsg.Proposals, p)
	return
}

func (v *VoteMessage) dump(hdr string) {
	fmt.Printf("**** %s: ROUND %d STEP %d ParentHash %x Seed %x\n", hdr, v.BlockNumber, v.Step, v.ParentHash, v.Seed)
	for _, p := range v.Proposals {
		p.dump()
	}
	fmt.Printf("\n")
}

func (voteMsg *VoteMessage) SubUser() uint64 {
	if voteMsg != nil {
		return voteMsg.Sub
	}
	return 0
}

type Votes []*VoteMessage
type VotesOrderedbyHash Votes

func (votes VotesOrderedbyHash) Len() int      { return len(votes) }
func (votes VotesOrderedbyHash) Swap(i, j int) { votes[i], votes[j] = votes[j], votes[i] }
func (votes VotesOrderedbyHash) Less(i, j int) bool {
	if bytes.Compare(votes[i].Hash().Bytes(), votes[j].Hash().Bytes()) < 0 {
		return true
	}
	return false
}

type CertList struct {
	BlockHashes []common.Hash `json:"blockHashes"`
}

func (c *CertList) Bytes() (enc []byte) {
	enc, _ = rlp.EncodeToBytes(&c)
	return enc
}

func (c *CertList) AddcertHash(hash common.Hash) (modified bool) {
	if !c.Exist(hash) {
		c.BlockHashes = append(c.BlockHashes, hash)
		return true
	}
	return false
}

func (c *CertList) RemovecertHash(hash common.Hash) (modified bool) {
	if c.Exist(hash) {
		c.BlockHashes = remove(c.BlockHashes, hash)
		return true
	}
	return false
}

func remove(clist []common.Hash, cert common.Hash) []common.Hash {
	for i := len(clist) - 1; i >= 0; i-- {
		if bytes.Equal(clist[i].Bytes(), cert.Bytes()) {
			clist = append(clist[:i], clist[i+1:]...)
		}
	}
	return clist
}

func removeDuplicates(clist []common.Hash) []common.Hash {
	existed := map[common.Hash]bool{}
	for cert := range clist {
		existed[clist[cert]] = true
	}

	certs := []common.Hash{}
	for uniqueCert, _ := range existed {
		certs = append(certs, uniqueCert)
	}
	return certs
}

func (c *CertList) Exist(hash common.Hash) bool {
	if c == nil {
		c = new(CertList)
		return false
	}
	for _, h := range c.BlockHashes {
		if bytes.Equal(h.Bytes(), hash.Bytes()) {
			return true
		}
	}
	return false
}

func DecodeRLPCertList(bytes []byte) (c *CertList, err error) {
	var clist CertList
	err = rlp.DecodeBytes(bytes, &clist)
	if err != nil {
		return c, err
	}
	return &clist, nil
}

func DecodeRLPVote(bytes []byte) (v *VoteMessage, err error) {
	var vote VoteMessage
	err = rlp.DecodeBytes(bytes, &vote)
	if err != nil {
		return v, err
	}
	return &vote, nil
}

func (v *VoteMessage) Serialize() ([]byte, error) {
	return json.Marshal(v)
}

func (v *VoteMessage) Deserialize(data []byte) error {
	return json.Unmarshal(data, v)
}

func (v *VoteMessage) VerifySignature() error {
	pubkey, err := v.RecoverPubkey()
	if err != nil {
		return err
	}
	msgHash := v.ShortHash()
	return pubkey.VerifySign(msgHash.Bytes(), v.Signature)
}

func (v *VoteMessage) Sign(priv *crypto.PrivateKey) ([]byte, error) {
	msgHash := v.ShortHash()
	sign, err := priv.Sign(msgHash.Bytes())
	if err != nil {
		return nil, err
	}
	v.Signature = sign
	return sign, nil
}

func (v *VoteMessage) GetRootHash() (h common.Hash) {
	for _, p := range v.Proposals {
		return p.RootHash
	}
	return h
}

func (v *VoteMessage) CountVotes(step uint64, seed []byte, bn uint64) (h common.Hash, cnt int) {
	cnt = 0
	for _, p := range v.Proposals {
		t := p.countVotes(step, seed, bn)
		if t > cnt {
			h = p.RootHash
			cnt = t
		}
	}
	return h, cnt
}

//NOT sure if it's correct
func (v *VoteMessage) IsEmpty() (h common.Hash, cnt int, isEmpty bool) {
	cnt = 0
	for _, p := range v.Proposals {
		t := p.countVotes(voteNext, v.Seed, v.BlockNumber)
		if t > cnt {
			h = p.RootHash
			cnt = t
			isEmpty = p.IsEmpty()
		}
	}
	return h, cnt, isEmpty
}

func (v *VoteMessage) ShortHash() common.Hash {
	unsignedmsg := &VoteMessage{
		BlockNumber: v.BlockNumber,
		ParentHash:  v.ParentHash,
		Seed:        v.Seed,
		BlockHash:   v.BlockHash,
		Step:        v.Step,
		Sub:         v.Sub,
		VRF:         v.VRF,
		Proof:       v.Proof,
		Voter:       v.Voter,
		Signature:   make([]byte, 0),
	}
	enc, _ := rlp.EncodeToBytes(&unsignedmsg)
	return common.BytesToHash(wolkcommon.Computehash(enc))
}

func (votes Votes) Hash() common.Hash {
	enc, _ := rlp.EncodeToBytes(&votes)
	return common.BytesToHash(wolkcommon.Computehash(enc))
}

func (v *VoteMessage) Hash() common.Hash {
	enc, _ := rlp.EncodeToBytes(&v)
	return common.BytesToHash(wolkcommon.Computehash(enc))
}

func (v *VoteMessage) RecoverPubkey() (*crypto.PublicKey, error) {
	return crypto.RecoverPubkey(v.Signature)
}

func (v *VoteMessage) Bytes() (enc []byte) {
	enc, _ = rlp.EncodeToBytes(&v)
	return enc
}

func (v *VoteMessage) Size() uint64 {
	return uint64(len(v.Bytes()))
}

func (v *VoteMessage) String() string {
	sv := NewSerializedVote(v)
	return sv.String()
}

type SerializedVote struct {
	VoteHash    common.Hash `json:"voteHash"`
	Voter       uint8       `json:"voter"`
	BlockNumber uint64      `json:"blockNumber"`
	BlockHash   common.Hash `json:"blockHash"`
	ParentHash  common.Hash `json:"parentHash"`
	Step        uint64      `json:"step"`
	Sub         uint64      `json:"subuser"`
	VRF         string      `json:"vrf"`
	Proof       string      `json:"proof"`
	Signature   string      `json:"signature"`
	Proposals   []*Proposal `json:"proposals"`
	Size        uint64      `json:"size"`
}

func (sv *SerializedVote) DeserializeVote() *VoteMessage {
	v := new(VoteMessage)
	v.BlockNumber = sv.BlockNumber
	v.ParentHash = sv.ParentHash
	v.BlockHash = sv.BlockHash
	v.Step = sv.Step
	v.Sub = sv.Sub
	v.VRF = common.FromHex(sv.VRF)
	v.Proof = common.FromHex(sv.Proof)
	v.Voter = sv.Voter
	v.Signature = common.FromHex(sv.Signature)
	v.Proposals = sv.Proposals
	return v
}

func NewSerializedVote(v *VoteMessage) *SerializedVote {
	return &SerializedVote{
		VoteHash:    v.Hash(),
		BlockNumber: v.BlockNumber,
		ParentHash:  v.ParentHash,
		BlockHash:   v.BlockHash,
		Step:        v.Step,
		Sub:         v.Sub,
		VRF:         fmt.Sprintf("%x", v.VRF),
		Proof:       fmt.Sprintf("%x", v.Proof),
		Voter:       v.Voter,
		Signature:   fmt.Sprintf("%x", v.Signature),
		Proposals:   v.Proposals,
		Size:        v.Size(),
	}
}

func (sv *SerializedVote) String() string {
	bytes, err := json.Marshal(sv)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
