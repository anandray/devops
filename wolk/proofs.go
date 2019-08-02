package wolk

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
)

// NoSQLProof is the core data type returned in Provable NoSQL Queries, currently at all 3 levels (owner, collection, key)
// for both individual key-value pairs (GetKey, GetBucket) and collection  results (GetBuckets, ScanCollection)
type NoSQLProof struct {
	BlockNumber uint64
	BlockHash   common.Hash // of the block

	KeyChunkHash  common.Hash // this is the block.KeyRoot of BlockNumber
	KeyMerkleRoot common.Hash // contained within the chunk referenced by KeyChunkHash

	Owner            string
	SystemHash       common.Address   // has system collection of the owner
	SystemProof      *SerializedProof // proof that OwnerHash is in the SMT referenced by the OwnerChunkHash
	SystemChunkHash  common.Hash      // this is the SMT for the users Buckets
	SystemMerkleRoot common.Hash

	Collection           string
	CollectionHash       common.Address   // hash of the owner-collection
	CollectionProof      *SerializedProof // contains inclusion proof that CollectionHash is in the SMT referenced by the CollectionChunkHash
	CollectionChunkHash  common.Hash
	CollectionMerkleRoot common.Hash

	// used in GetKey
	Key          []byte                 `json:"Key,omitempty"`
	KeyHash      common.Address         `json:"KeyHash,omitempty"` // hash of the key
	TxHash       common.Hash            `json:"TxHash,omitempty"`  // the SetKey tx itself
	Tx           *SerializedTransaction `json:"Tx,omitempty"`
	KeyProof     *SerializedProof       `json:"KeyProof,omitempty"` // contains inclusion proof that TxHash is in the SMT
	ReplicaProof *ReplicaProof          `json:"ReplicaProof,omitempty"`

	// used in ScanCollection
	ScanProofs []*ScanProof `json:"ScanProofs,omitempty"`
}

func (s *NoSQLProof) AddScanProofTx(txhash common.Hash, tx *SerializedTransaction) error {
	for _, sp := range s.ScanProofs {
		if bytes.Compare(txhash.Bytes(), sp.TxHash.Bytes()) == 0 {
			sp.Tx = tx
			return nil
		}
	}
	return fmt.Errorf("[proofs:AddScanProofTx] ScanProof txhash Not found %x", txhash)
}

func (pr *NoSQLProof) VerifyScanProofs() (err error) {
	if len(pr.Collection) == 0 {
		// Result of GetBuckets
		for _, p := range pr.ScanProofs {
			prf := DeserializeProof(p.KeyProof)
			if !prf.Check(p.TxHash.Bytes(), pr.SystemMerkleRoot.Bytes(), GlobalDefaultHashes, false) {
				log.Error("[client:verifyNoSQLProof] KeyProof FAILED", "txhash", p.TxHash, "mr", pr.SystemMerkleRoot, "pr", pr.String())
				return fmt.Errorf("KeyProof FAILED %v", err)
			}
		}
		if len(pr.ScanProofs) == 0 {
			if bytes.Compare(pr.SystemChunkHash.Bytes(), getEmptySMTChunkHash()) != 0 {
				return fmt.Errorf("[proofs:Verify] Empty buckets list but incorrect system chunk hash")
			}
			p := DeserializeProof(pr.SystemProof)
			if !p.Check(pr.SystemChunkHash.Bytes(), pr.KeyMerkleRoot.Bytes(), GlobalDefaultHashes, false) {
				return fmt.Errorf("[proofs:Verify] Failed system chunk hash proof")
			}
		}
		return nil
	}
	// Result of ScanCollection
	for _, p := range pr.ScanProofs {
		prf := DeserializeProof(p.KeyProof)
		if !prf.Check(p.TxHash.Bytes(), pr.CollectionMerkleRoot.Bytes(), GlobalDefaultHashes, false) {
			log.Error("[proofs:Verify] KeyProof FAILED", "pr", pr.String())
			return fmt.Errorf("[proofs:Verify] KeyProof FAILED %v", err)
		}
	}
	if len(pr.ScanProofs) == 0 {
		if bytes.Compare(pr.CollectionChunkHash.Bytes(), getEmptySMTChunkHash()) != 0 {
			return fmt.Errorf("Incorrect collection chunk hash")
		}
	}
	return nil
}

func (pr *NoSQLProof) Verify() (err error) {
	owner := pr.Owner
	collection := pr.Collection
	key := pr.Key
	defaultHashes := GlobalDefaultHashes

	// Verify Proof Process: at proof.Key position (which is KeyAddress(key))
	//   there is proof.TxHash, which hashes up to proof.CollectionMerkleRoot
	if pr.KeyProof != nil {
		keyProof := DeserializeProof(pr.KeyProof)
		keyHash := KeyToAddress([]byte(key))
		if bytes.Compare(keyProof.Key, keyHash.Bytes()) != 0 {
			log.Error("[proofs:verifyNoSQLProof] Key Mismatch", "keyhash", fmt.Sprintf("%x", keyHash), "keyProof.Key", fmt.Sprintf("%x", keyProof.Key))
			return fmt.Errorf("Key Mismatch")
		}

		if !keyProof.Check(pr.TxHash.Bytes(), pr.CollectionMerkleRoot.Bytes(), defaultHashes, false) {
			log.Error("[proofs:verifyNoSQLProof] KeyProof FAILED", "pr", pr.String())
			return fmt.Errorf("KeyProof FAILED %v", err)
		}
	}

	// Verify Proof Process: at proof.Key position (which is proof.CollectionHash=CollectionHash(owner, collection))
	//   there is proof.CollectionChunkHash, which hashes up to proof.KeyMerkleRoot
	if pr.CollectionProof != nil {
		collectionProof := DeserializeProof(pr.CollectionProof)
		collectionHash := CollectionHash(pr.Owner, collection)
		if bytes.Compare(collectionProof.Key, collectionHash.Bytes()) != 0 {
			log.Error("[client:verifyNoSQLProof] Collection Mismatch", "collectionHash", collectionHash, "collectionProof.Key", fmt.Sprintf("%x", collectionProof.Key))
			return fmt.Errorf("Collection Mismatch (%x, %x)", collectionProof.Key, collectionHash)
		}
		if !collectionProof.Check(pr.CollectionChunkHash.Bytes(), pr.KeyMerkleRoot.Bytes(), defaultHashes, false) {
			log.Error("[client:verifyNoSQLProof] CollectionProof FAILED", "pr", pr.String())
			return fmt.Errorf("CollectionProof FAILED")
		}
	} else if len(collection) == 0 {

	} else {
		return fmt.Errorf("CollectionProof MISSING")
	}

	// Verify proof process: at proof.Key position (which is proof.SystemHash=CollectionHash(owner, SystemCollection))
	//   there is the value proof.SystemChunkHash, which hashes up to proof.KeyMerkleRoot
	systemProof := DeserializeProof(pr.SystemProof)
	systemHash := KeyToAddress([]byte(owner))
	if bytes.Compare(systemProof.Key, systemHash.Bytes()) != 0 {
		return fmt.Errorf("System Mismatch")
	}
	if !systemProof.Check(pr.SystemChunkHash.Bytes(), pr.KeyMerkleRoot.Bytes(), defaultHashes, false) {
		return fmt.Errorf("SystemProof FAILED %v", err)
	}
	return nil
}

func (s *NoSQLProof) String() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

// Names, Checks, Products, etc.
type SMTProof struct {
	BlockNumber uint64
	BlockHash   common.Hash
	ChunkHash   common.Hash
	MerkleRoot  common.Hash
	Key         string
	KeyHash     common.Address // the position in the SMT
	TxHash      common.Hash    // the value at the KeyHash position
	Tx          *SerializedTransaction
	TxSigner    common.Address
	Proof       *SerializedProof
}

func (s *SMTProof) String() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

type Proof struct {
	SMTTreeDepth uint64
	Key          []byte
	Proof        [][]byte
	ProofBits    []byte
}

func ComputeDefaultHashes() (defaultHashes [TreeDepth][]byte) {
	//defaultHashes = make([TreeDepth][]byte)
	empty := make([]byte, 0)
	defaultHashes[0] = wolkcommon.Computehash(empty)
	for level := 1; level < TreeDepth; level++ {
		defaultHashes[level] = wolkcommon.Computehash(defaultHashes[level-1], defaultHashes[level-1])
	}
	return defaultHashes
}

func (self *Proof) Check(v []byte, root []byte, defaultHashes [TreeDepth][]byte, verbose bool) bool {
	// the leaf value to start off hashing!  The value is hash(RLPEncode([]))
	debug := verbose
	cur := v
	p := 0
	smtTreeDepth := int(self.SMTTreeDepth)
	// fmt.Printf("Check: %d smtTreeDepth %x key proof: %+v proofBits: %x\n", self.SMTTreeDepth, self.Key, self.Proof, self.ProofBits)
	for i := 0; i < smtTreeDepth; i++ {
		if getBit(self.ProofBits, i) {
			// (uint64(1<<i) & self.proofBits) > 0
			if getBit(self.Key, i) {
				if debug {
					fmt.Printf("C%v | [P,*] bit%v=1 | H(P[%d]:%x, C[%d]:%x) => ", i+1, i, p, self.Proof[p], i, cur)
				}
				cur = wolkcommon.Computehash(self.Proof[p], cur)
			} else { // i-th bit is "0", so hash with H([]) on the right
				if debug {
					fmt.Printf("C%v | [*,P] bit%v=0 | H(C[%d]:%x, P[%d]:%x) => ", i+1, i, i, cur, p, self.Proof[p])
				}
				cur = wolkcommon.Computehash(cur, self.Proof[p])
			}
			p++
		} else {
			if getBit(self.Key, i) { // i-th bit is "1", so hash with H([]) on the left
				if debug {
					fmt.Printf("C%v | [D,*] bit%v=1 | H(D[%d]:%x, C[%d]:%x) => ", i+1, i, i, defaultHashes[i], i, cur)
				}
				cur = wolkcommon.Computehash(defaultHashes[i], cur)
			} else {
				if debug {
					fmt.Printf("C%v | [*,D] bit%v=0 | H(C[%d]:%x, D[%d]:%x) => ", i+1, i, i, cur, i, defaultHashes[i])
				}
				cur = wolkcommon.Computehash(cur, defaultHashes[i])
			}
		}
		if debug {
			fmt.Printf(" %x\n", cur)
		}
	}
	res := bytes.Compare(cur, root) == 0
	if verbose {
		if res {
			log.Trace("CheckProof success", "root", root)
		} else {
			log.Error("CheckProof FAILURE", "root", fmt.Sprintf("%x", root))
		}
	}
	return res
}

func (self *Proof) String() string {
	bytes, err := json.Marshal(self)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func (p *Proof) Bytes() (out []byte) {
	out = append(out, p.ProofBits...)
	for _, h := range p.Proof {
		out = append(out, h...)
	}
	return out
}

type SerializedProof struct {
	SMTTreeDepth uint64
	Key          string
	Proof        []string
	ProofBits    string
}

type ScanProof struct {
	KeyHash  common.Address // hash of the key
	TxHash   common.Hash    // the SetKey tx itself
	Tx       *SerializedTransaction
	KeyProof *SerializedProof // contains inclusion proof that TxHash is in the SMT
}

func DeserializeProof(sp *SerializedProof) (p *Proof) {
	p = new(Proof)
	p.SMTTreeDepth = sp.SMTTreeDepth
	p.Key = common.Hex2Bytes(sp.Key)
	p.ProofBits = common.Hex2Bytes(sp.ProofBits)
	p.Proof = make([][]byte, len(sp.Proof))
	for i, str := range sp.Proof {
		p.Proof[i] = common.Hex2Bytes(str)
	}
	return p
}

func NewSerializedProof(p *Proof) *SerializedProof {
	sp := new(SerializedProof)
	if p != nil {
		sp.SMTTreeDepth = p.SMTTreeDepth
		if len(p.Key) > 0 {
			sp.Key = fmt.Sprintf("%x", p.Key)
		} else {
			log.Error("[proofs:NewSerializedProof] empty key... how?")
		}
		sp.Proof = make([]string, len(p.Proof))
		for i, h := range p.Proof {
			sp.Proof[i] = fmt.Sprintf("%x", h)
		}
		sp.ProofBits = fmt.Sprintf("%x", p.ProofBits)
	} else {
		log.Error("[proofs:NewSerializedProof] nil proof")
	}
	return sp
}
