package wolk

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

const (
	// TreeDepth is the maximum number of bits the SparseMerkleTree can have it its key
	TreeDepth = 160

	// NumBitsAddress is the maximum number of bits the SparseMerkleTree can have it its key [same as TreeDepth]
	NumBitsAddress = TreeDepth
	bytesPerChild  = 112
	nchildren      = 256
	rootNodePrefix = "ROOTNODE"
)

// GlobalDefaultHashes holds all the default hashes from 0 ... TreeDepth - 1 -- computed once in init
var GlobalDefaultHashes [TreeDepth][]byte

// SparseMerkleTree is a provable Key-Value store, with N-bit keys specifying a path down to a leaf node
type SparseMerkleTree struct {
	ChunkStore ChunkStore

	root      *SMTNode
	nodemap   map[string]*SMTNode
	nodemapMu sync.RWMutex

	fetching   map[string]bool
	fetchingMu sync.RWMutex
}

type smtData struct {
	Key          []byte
	ValHash      common.Hash
	MerkleRoot   common.Hash
	StorageBytes uint64
	Deleted      bool
}

type WGroup struct {
	wg     sync.WaitGroup
	err    error
	cancel func()
}

func (w *WGroup) Wait() error {
	w.wg.Wait()
	/*
		if g.cancel != nil {
			g.cancel()
		}
	*/
	return w.err
}

// SMTNode is the key workhorse of SparseMerkleTree, referenced by chunkHash starting with a RootNode
// Each node has up to 256 children, referenced in CHash (children chunk Hashes) and loaded into children.
// Data is stored in TerminalData's smtData nodes.  When loadNode is called, the chunkHash is used to fill out
type SMTNode struct {
	muNode       sync.RWMutex
	chunkHash    common.Hash // a node must always have a chunkHash
	merkleRoot   common.Hash // a node will have a merkleRoot loaded from the chunk [when unloaded => false]
	unloaded     bool        // set to true as soon as a chunkHash specified, but after a GetChunk call (that fills in the content)
	dirty        bool        // initially false, but on any insert is set to true
	level        int         // 0..TreeDepth - 1, top level Root nodes has level = TreeDepth - 1
	TerminalData []*smtData  // holds unique Key-ValHash data for a key
	CHash        map[int]common.Hash
	StorageBytes uint64
	children     map[int]*SMTNode
	mrcache      [9][nchildren]common.Hash
	mu           sync.RWMutex
	wg           *WGroup
}

// NewSMTNode initializes a node for the SMT
func NewSMTNode(level int) *SMTNode {
	return &SMTNode{
		children:     make(map[int]*SMTNode),
		TerminalData: make([]*smtData, 256),
		dirty:        false,
		unloaded:     false,
		level:        level,
		StorageBytes: 0,
	}
}

func getBit(k []byte, i int) bool {
	return (byte(0x01<<uint(i%8))&byte(k[len(k)-1-i/8]) > 0)
}

func setBit(k []byte, i int) {
	if i >= 0 && i < len(k)*8 {
		k[len(k)-1-i/8] |= 0x01 << uint(i%8)
	} else {
		log.Error("setBit out of range!", "len(k)", len(k), "i", i)
	}
}

// NewSparseMerkleTree creates a new SMT
func NewSparseMerkleTree(depth int, cs ChunkStore) *SparseMerkleTree {
	var smt SparseMerkleTree
	smt.ChunkStore = cs
	smt.nodemap = make(map[string]*SMTNode)
	n := NewSMTNode(TreeDepth - 1)
	smt.nodemap[rootNodePrefix] = n
	smt.fetching = make(map[string]bool)
	return &smt
}

// RootNode returns the top node
func (smt *SparseMerkleTree) RootNode() (n *SMTNode) {
	smt.nodemapMu.RLock()
	n = smt.nodemap[rootNodePrefix]
	smt.nodemapMu.RUnlock()
	return n
}

// Init takes a hash from cloudstore, initializes the top level node with the hash
func (smt *SparseMerkleTree) Init(hash common.Hash) {
	var emptyHash common.Hash
	smt.RootNode().chunkHash = hash
	if bytes.Compare(hash.Bytes(), emptyHash.Bytes()) == 0 || bytes.Compare(hash.Bytes(), getEmptySMTChunkHash()) == 0 {
		return
	}
	smt.RootNode().unloaded = true
}

// Copy makes a deep copy of the SMT [not implemented]
func (smt *SparseMerkleTree) Copy() (t *SparseMerkleTree) {
	return smt
}

// Delete deletes a key from the tree.  In practice, it is not used in Wolk storage
func (smt *SparseMerkleTree) Delete(k []byte) error {
	return nil
}

// Hash returns the top level chunk hash of the SMT
func (smt *SparseMerkleTree) Hash() common.Hash {
	return smt.ChunkHash()
}

// MarkDirty marks the top level root node as being dirty.  This is used in genesis initialization situations.
func (smt *SparseMerkleTree) MarkDirty(dirty bool) {
	smt.RootNode().dirty = dirty
}

// GetWithProof returns back a value for a key, the total storageBytes accumulated for representing that key, and a proof if the key is found.
// If key is not found or deleted, flags are returned.
func (smt *SparseMerkleTree) GetWithProof(ctx context.Context, k []byte) (v0 []byte, found bool, deleted bool, p interface{}, storageBytes uint64, err error) {
	return smt.Get(ctx, k, true)
}

// ChunkHash returns top level root hash
func (smt *SparseMerkleTree) ChunkHash() common.Hash {
	return smt.RootNode().chunkHash
}

// MerkleRoot returns the top SMT root where all proofs must resolve; this is inside the top chunk
func (smt *SparseMerkleTree) MerkleRoot(ctx context.Context) common.Hash {
	r := smt.RootNode()
	var prefix []byte
	smt.loadNode(ctx, r, prefix)
	return smt.RootNode().merkleRoot
}

// StorageBytes returns the total number of bytes held by the SMT, accumulated by multiple Insert calls
func (smt *SparseMerkleTree) StorageBytes() (uint64, error) {
	r := smt.RootNode()
	var prefix []byte
	smt.loadNode(context.TODO(), r, prefix)
	return smt.RootNode().StorageBytes, nil
}

func getEmptySMTChunkHash() []byte {
	return common.FromHex("44a00e7ae0499c6e377b95709f3851843976ea6c2a86eff46adfe15608d22005") // was: GlobalDefaultHashes[0]
}

// ScanAll returns keys (common.Address) and values (common.Hash) and (with withProof=true) a full set of proofs for all the keys
func (smt *SparseMerkleTree) ScanAll(ctx context.Context, withProof bool) (res map[common.Address]common.Hash, collectionProofs map[common.Address]*Proof, err error) {
	res = make(map[common.Address]common.Hash)
	collectionProofs = make(map[common.Address]*Proof)
	var stackn []*SMTNode
	var stackp [][]byte
	stackn = append(stackn, smt.RootNode())
	stackp = append(stackp, []byte(""))

	for len(stackp) > 0 {
		// pop stack
		n := stackn[len(stackn)-1]
		prefix := stackp[len(stackp)-1]
		stackn = stackn[0 : len(stackn)-1]
		stackp = stackp[0 : len(stackp)-1]
		err = smt.loadNode(ctx, n, prefix)
		n.muNode.RLock()
		for suff, d := range n.TerminalData {
			if d != nil {
				k := common.BytesToAddress(d.Key)
				res[k] = d.ValHash
				p := new(Proof)
				p.Proof = make([][]byte, 0)
				p.SMTTreeDepth = TreeDepth
				p.Key = d.Key
				p.ProofBits = make([]byte, TreeDepth/8)
				ps := append(prefix, byte(suff))
				for j := len(ps) - 1; j >= 0; j-- {
					idx := int(ps[j])
					var np *SMTNode
					if j == len(ps)-1 {
						np = n
					} else {
						ts := getNodePrefix(ps[0:j])
						np = smt.nodemap[ts]
						np.muNode.RLock()
					}
					for level := 8; level > 0; level-- {
						var sisterIndex int
						if idx&1 > 0 {
							sisterIndex = idx - 1
						} else {
							sisterIndex = idx + 1
						}
						p0 := np.mrcache[level][sisterIndex]
						if bytes.Compare(p0.Bytes(), GlobalDefaultHashes[np.level-level+1]) != 0 {
							p.Proof = append(p.Proof, np.mrcache[level][sisterIndex].Bytes())
							setBit(p.ProofBits, np.level-level+1)
						}
						idx = idx >> 1
					}
					if j != len(ps)-1 {
						np.muNode.RUnlock()
					}
				}
				collectionProofs[k] = p
			}
		}
		for idx, c := range n.children {
			// push stack
			stackn = append(stackn, c)
			stackp = append(stackp, append(prefix, byte(idx)))
		}
		n.muNode.RUnlock()
	}
	return res, collectionProofs, err
}

func (smt *SparseMerkleTree) setFetching(prefix []byte, fetching bool) {
	smt.fetchingMu.Lock()
	smt.fetching[fmt.Sprintf("%x", prefix)] = fetching
	smt.fetchingMu.Unlock()
}

func (smt *SparseMerkleTree) isFetching(prefix []byte) (fetching bool) {
	smt.fetchingMu.RLock()
	fetching = smt.fetching[fmt.Sprintf("%x", prefix)]
	smt.fetchingMu.RUnlock()
	return fetching
}

func (n *SMTNode) getUnloaded() (unloaded bool) {
	n.muNode.RLock()
	unloaded = n.unloaded
	n.muNode.RUnlock()
	return unloaded
}

// ChunkHash returns the chunk hash of the SMTNode
func (n *SMTNode) ChunkHash() (h common.Hash) {
	n.muNode.RLock()
	h = n.chunkHash
	n.muNode.RUnlock()
	return h
}

func getNodePrefix(p []byte) string {
	if len(p) == 0 {
		return rootNodePrefix
	}
	return fmt.Sprintf("%x", p)
}

func (smt *SparseMerkleTree) loadNode(ctx context.Context, n *SMTNode, prefix []byte) (err error) {
	if n.getUnloaded() == false {
		return nil
	}

	n.mu.Lock()
	if n.wg == nil {
		n.wg = new(WGroup)
		n.mu.Unlock()
		n.wg.wg.Add(1)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := smt.exec(ctx, n, prefix)
		n.mu.Lock()
		n.wg.err = err
		n.wg.wg.Done()
		n.wg = nil
		n.mu.Unlock()
		return err
	} else {
		wg := n.wg
		n.mu.Unlock()
		err := wg.Wait()
		return err
	}
}

func (smt *SparseMerkleTree) exec(ctx context.Context, n *SMTNode, prefix []byte) (err error) {
	errChan := make(chan error, 1)

	go func() {
		errChan <- smt.readNode(ctx, n, prefix)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (smt *SparseMerkleTree) readNode(ctx context.Context, n *SMTNode, prefix []byte) (err error) {
	st := time.Now()
	ps := getNodePrefix(prefix)
	//alreadyFetching := false
	var chunkHash common.Hash
	chunkHash = n.ChunkHash()
	chunk, ok, err := smt.ChunkStore.GetChunk(ctx, chunkHash.Bytes())
	if err != nil {
		log.Error(fmt.Sprintf("[smtnode:load] Error while attempting to retrieve chunk of hash %x | %+v", chunkHash, err))
		return fmt.Errorf("[smtnode:load] Error while attempting to retrieve chunk of hash %x | %+v", chunkHash, err)
	} else if !ok {
		log.Error("[smtnode:load] FAILED GetChunk", "ps", ps, "chunkHash", fmt.Sprintf("%x", chunkHash), "st", st)
		return fmt.Errorf("[smtnode:load] Attempted to retrieve chunk of hash %x but chunk not found", chunkHash)
	}
	n.muNode.Lock()
	defer n.muNode.Unlock()

	nrecs := (len(chunk) - 32) / bytesPerChild
	dh := GlobalDefaultHashes[n.level-8+1]
	for i := 0; i < 256; i++ {
		n.mrcache[8][i] = common.BytesToHash(dh)
	}
	for i := 0; i < nrecs; i++ {
		r := int(chunk[i*bytesPerChild+104])
		if chunk[i*bytesPerChild+105] > 0 {
			d := new(smtData)
			d.Key = chunk[i*bytesPerChild : i*bytesPerChild+20]
			d.StorageBytes = wolkcommon.BytesToUint64(chunk[i*bytesPerChild+32 : i*bytesPerChild+40])
			d.ValHash = common.BytesToHash(chunk[i*bytesPerChild+40 : i*bytesPerChild+72])
			d.MerkleRoot = common.BytesToHash(chunk[i*bytesPerChild+72 : i*bytesPerChild+104])
			if chunk[i*bytesPerChild+106] > 0 {
				d.Deleted = true
			}
			n.TerminalData[r] = d
			n.mrcache[8][r] = d.MerkleRoot
		} else {
			prefixc := append(prefix, byte(r))
			c := NewSMTNode(n.level - 8)
			c.muNode.Lock()
			c.unloaded = true
			c.dirty = false
			c.chunkHash = common.BytesToHash(chunk[i*bytesPerChild : i*bytesPerChild+32])
			c.StorageBytes = wolkcommon.BytesToUint64(chunk[i*bytesPerChild+32 : i*bytesPerChild+40])
			c.merkleRoot = common.BytesToHash(chunk[i*bytesPerChild+40 : i*bytesPerChild+72])
			n.children[r] = c
			n.mrcache[8][r] = c.merkleRoot
			smt.setNodemap(prefixc, c)
			c.muNode.Unlock()
		}
	}
	n.merkleRoot = common.BytesToHash(chunk[len(chunk)-32 : len(chunk)])
	n.unloaded = false
	// now for each of 7...0 levels, hash the level of "leaves" into  n.mrcache
	newleavesCnt := nchildren / 2
	for level := 7; level >= 0; level-- {
		for i := 0; i < newleavesCnt; i++ {
			n.mrcache[level][i] = common.BytesToHash(wolkcommon.Computehash(n.mrcache[level+1][i*2].Bytes(), n.mrcache[level+1][i*2+1].Bytes()))
		}
		newleavesCnt = newleavesCnt / 2
	}
	log.Trace(fmt.Sprintf("[smtnode:load] SUCCESS chunk %x %d", chunkHash, len(chunk)))
	return nil
}

func (smt *SparseMerkleTree) setNodemap(pref []byte, n *SMTNode) {
	smt.nodemapMu.Lock()
	smt.nodemap[fmt.Sprintf("%x", pref)] = n
	smt.nodemapMu.Unlock()
}

func (smt *SparseMerkleTree) getNodemap(pref []byte) (n *SMTNode, ok bool) {
	smt.nodemapMu.RLock()
	n, ok = smt.nodemap[fmt.Sprintf("%x", pref)]
	smt.nodemapMu.RUnlock()
	return n, ok
}

// Insert mutates the SMT in-memory to hold (k,v) along with an incremental number of bytes.  If the k is deleted, there is a deleted flag
func (smt *SparseMerkleTree) Insert(ctx context.Context, k []byte, v []byte, storageBytesNew uint64, deleted bool) (err error) {
	storageBytesNew = 0
	done := false
	p := smt.RootNode()
	for b := 0; b < len(k) && !done; b++ {
		prefix := make([]byte, b)
		copy(prefix, k[0:b])
		st0 := time.Now()
		err = smt.loadNode(ctx, p, prefix)
		if err != nil {
			return err
		}
		if time.Since(st0) > warningTxTime {
			log.Error("[smt:Insert] loadNode", "b", b, "k", fmt.Sprintf("%x", k), "tm", time.Since(st0))
		} else {
			log.Trace("[smt:Insert] loadNode", "b", b, "k", fmt.Sprintf("%x", k), "tm", time.Since(st0))
		}
		idx := int(k[b])

		st0 = time.Now()
		p.muNode.Lock()
		if time.Since(st0) > warningTxTime {
			log.Error("[smt:Insert] step0", "b", b, "k", fmt.Sprintf("%x", k), "tm", time.Since(st0))
		}
		st0 = time.Now()
		if n, ok := p.children[idx]; ok {
			//	fmt.Printf("  1: %d Insert(%x, %x)\n", b, k, v)
			p.dirty = true
			p.muNode.Unlock()
			p = n // RECURSE
		} else {
			d := p.TerminalData[idx]
			qk := make([]byte, len(k))
			copy(qk[:], k[:])
			d2 := &smtData{Key: qk, ValHash: common.BytesToHash(v), StorageBytes: storageBytesNew}
			if d != nil {
				//	fmt.Printf("  2: %d Insert(%x, %x)\n", b, k, v)
				if bytes.Compare(d.Key, k) == 0 {
					if bytes.Compare(d.ValHash.Bytes(), v) == 0 {
						// nothing changed but we do have
						if storageBytesNew > 0 {
							p.dirty = true
							d.StorageBytes += storageBytesNew
						}
					} else {
						p.dirty = true
						if bytes.Compare(v, d.ValHash.Bytes()) != 0 {
							d.ValHash = common.BytesToHash(v)
							d.StorageBytes += storageBytesNew
						}
						if d.Deleted != deleted {
							d.Deleted = deleted
						}
					}
					p.muNode.Unlock()
					if time.Since(st0) > warningTxTime {
						log.Error("[smt:Insert] step1", "k", fmt.Sprintf("%x", k), "tm", time.Since(st0))
					}
					return nil
				}
				p.TerminalData[idx] = nil
				advanced := false
				// Two keys (d, d2) are competing for the same slot, so we need to create:
				//  (1) a sequence of single child
				//  (2) a node with both children
				for ; b < TreeDepth; b++ {
					idx1 := int(d.Key[b])
					idx2 := int(d2.Key[b])
					if d.Key[b] == d2.Key[b] { // case (1)
						n = NewSMTNode(p.level - 8)
						n.muNode.Lock()
						p.children[idx1] = n
						p.StorageBytes = d.StorageBytes + d2.StorageBytes
						p.dirty = true
						newp := p.children[int(d.Key[b])]
						if advanced == false {
							p.muNode.Unlock()
							advanced = true
						}
						p = newp
						smt.setNodemap(k[0:b+1], n)
						n.muNode.Unlock()
					} else { // case (2)
						p.StorageBytes = d.StorageBytes + d2.StorageBytes
						p.TerminalData[idx1] = d
						p.TerminalData[idx2] = d2
						p.dirty = true
						if advanced == false {
							p.muNode.Unlock()
							advanced = true
						}
						if time.Since(st0) > warningTxTime {
							log.Error("[smt:Insert] finish1", "k", fmt.Sprintf("%x", k), "tm", time.Since(st0))
						}
						return nil
					}
				}
			} else {
				//fmt.Printf("  3: %d Insert(%x, %x)\n", b, k, v)
				p.StorageBytes += storageBytesNew
				p.TerminalData[idx] = d2
				p.dirty = true
				p.muNode.Unlock()
				if time.Since(st0) > warningTxTime {
					log.Info("[smt:Insert] finish2", "k", fmt.Sprintf("%x", k), "tm", time.Since(st0))
				}
				return nil
			}
		}
	}
	return nil
}

// GetWithoutProof returns back a value for a key, the total storageBytes accumulated for representing that key, and a proof if the key is found.
// If key is not found or deleted, flags are returned.  However, no proof is generated
func (smt *SparseMerkleTree) GetWithoutProof(ctx context.Context, k []byte) (v0 []byte, found bool, deleted bool, storageBytes uint64, err error) {
	v0, found, deleted, _, storageBytes, err = smt.Get(ctx, k, false)
	return v0, found, deleted, storageBytes, err
}

// Get returns back a value for a key, the total storageBytes accumulated for representing that key, and a proof if the key is found.
// If key is not found or deleted, flags are returned.
func (smt *SparseMerkleTree) Get(ctx context.Context, k []byte, withProof bool) (v0 []byte, found bool, deleted bool, p *Proof, storageBytes uint64, err error) {
	key := make([]byte, len(k))
	copy(key[:], k[:])
	n := smt.RootNode()
	p = new(Proof)
	p.Key = k
	p.SMTTreeDepth = TreeDepth
	p.ProofBits = make([]byte, TreeDepth/8)
	var idx int
	found = false

	proofstack := make([]*SMTNode, 0)
	for b := 0; b < len(key) && !found; b++ {
		prefix := make([]byte, b)
		copy(prefix, key[0:b])
		err = smt.loadNode(ctx, n, prefix)
		if err != nil {
			return v0, false, false, p, storageBytes, err
		}
		idx = int(k[b])
		n.muNode.RLock()
		if c, ok := n.children[idx]; ok {
			n.muNode.RUnlock()
			if withProof {
				proofstack = append(proofstack, n)
			}
			n = c
		} else {
			d := n.TerminalData[idx]
			n.muNode.RUnlock()
			if d != nil {
				if bytes.Equal(k, d.Key) {
					if withProof {
						proofstack = append(proofstack, n)
						for j := len(proofstack) - 1; j >= 0; j-- {
							idx = int(key[j])
							np := proofstack[j]
							np.muNode.RLock()
							for level := 8; level > 0; level-- {
								var sisterIndex int
								if idx&1 > 0 {
									sisterIndex = idx - 1
								} else {
									sisterIndex = idx + 1
								}
								p0 := np.mrcache[level][sisterIndex]
								if bytes.Compare(p0.Bytes(), GlobalDefaultHashes[np.level-level+1]) != 0 {
									p.Proof = append(p.Proof, np.mrcache[level][sisterIndex].Bytes())
									setBit(p.ProofBits, np.level-level+1)
								}
								idx = idx >> 1
							}
							np.muNode.RUnlock()
						}
					}
					return d.ValHash.Bytes(), true, d.Deleted, p, d.StorageBytes, nil
				}
			} else {
				return v0, false, false, p, storageBytes, nil
			}
		}
	}

	return v0, false, false, p, 0, nil
}

// Dump prints the tree to stdout starting from the root.
func (smt *SparseMerkleTree) Dump() {
	smt.RootNode().dump(0, 0)
}

func (n *SMTNode) dump(prefix byte, h int) {
	for i := 0; i < h; i++ {
		fmt.Printf("  ")
	}
	fmt.Printf("[%02x] Level %d ChunkHash %x MerkleRoot %x StorageBytes: %d dirty: %v", prefix, n.level, n.chunkHash, n.merkleRoot, n.StorageBytes, n.dirty)
	if n.unloaded {
		fmt.Printf(" (UNLOADED)")
	}
	fmt.Printf("\n")

	for i := 0; i < 256; i++ {
		d := n.TerminalData[i]
		if d != nil {
			for j := 0; j < h; j++ {
				fmt.Printf("  ")
			}
			fmt.Printf("  [%02x] KEY: %x VAL: %x MR: %x StorageBytes: %d\n", i, d.Key, d.ValHash, d.MerkleRoot, d.StorageBytes)
		} else if c, ok := n.children[i]; ok {
			c.dump(byte(i), h+1)
		}
	}
}

func computeTerminalMerkleRoot(k []byte, v common.Hash, level int) common.Hash {
	cur := make([]byte, 32)
	copy(cur[:], v.Bytes())
	for i := uint64(0); i < uint64(level); i++ {
		idx := int((TreeDepth - 1 - i) / 8)
		if idx >= len(k) {
			log.Error("Problem", "i", i, "idx", idx, "TreeDepth", TreeDepth, "len(n.key)", len(k))
		} else if byte(0x01<<(i%8))&byte(k[(TreeDepth-1-i)/8]) > 0 {
			// i-th bit is "1", so hash with DH(i) on the left
			cur = wolkcommon.Computehash(GlobalDefaultHashes[i], cur)
		} else { // i-th bit is "0", so hash with H(i) on the right
			cur = wolkcommon.Computehash(cur, GlobalDefaultHashes[i])
		}
	}
	return common.BytesToHash(cur)
}

func (n *SMTNode) saveNode(cs ChunkStore, wg *sync.WaitGroup, writeToCloudstore bool) (err error) {
	n.muNode.Lock() //
	defer n.muNode.Unlock()
	chunk := make([]byte, bytesPerChild*256+32)
	i := 0
	for r := 0; r < 256; r++ {
		if c, ok := n.children[r]; ok {
			c.muNode.RLock()
			copy(chunk[i*bytesPerChild:i*bytesPerChild+32], c.chunkHash.Bytes())
			copy(chunk[i*bytesPerChild+32:i*bytesPerChild+40], wolkcommon.UIntToByte(c.StorageBytes))
			copy(chunk[i*bytesPerChild+40:i*bytesPerChild+72], c.merkleRoot.Bytes())
			n.mrcache[8][r] = c.merkleRoot
			c.muNode.RUnlock()
			chunk[i*bytesPerChild+104] = byte(r)
			i++
		} else {
			d := n.TerminalData[r]
			if d != nil {
				d.MerkleRoot = computeTerminalMerkleRoot(d.Key, d.ValHash, n.level-8+1)
				copy(chunk[i*bytesPerChild:i*bytesPerChild+20], d.Key)
				copy(chunk[i*bytesPerChild+32:i*bytesPerChild+40], wolkcommon.UIntToByte(d.StorageBytes))
				copy(chunk[i*bytesPerChild+40:i*bytesPerChild+72], d.ValHash.Bytes())
				copy(chunk[i*bytesPerChild+72:i*bytesPerChild+104], d.MerkleRoot.Bytes())
				n.mrcache[8][r] = d.MerkleRoot
				chunk[i*bytesPerChild+104] = byte(r)
				chunk[i*bytesPerChild+105] = 1
				if d.Deleted {
					chunk[i*bytesPerChild+106] = 1
				}
				i++
			} else {
				n.mrcache[8][r] = common.BytesToHash(GlobalDefaultHashes[n.level-8+1])
			}
		}
	}
	nrecs := i
	// for i := 0; i < 256; i++ {
	// 	fmt.Printf("saveNode -- n.mrcache[8][%x] = %x\n", i, n.mrcache[8][i])
	// }
	// now for each of 8...0 levels, hash the level of "leaves" into  n.mrcache
	newleavesCnt := nchildren / 2
	for level := 7; level >= 0; level-- {
		for j := 0; j < newleavesCnt; j++ {
			n.mrcache[level][j] = common.BytesToHash(wolkcommon.Computehash(n.mrcache[level+1][j*2].Bytes(), n.mrcache[level+1][j*2+1].Bytes()))
			//	fmt.Printf("mrcache[%d][%x] = %x\n", level, i, n.mrcache[level][i])
		}
		newleavesCnt = newleavesCnt / 2
	}
	// copy the merkleRoot as the last 32 bytes
	n.merkleRoot = n.mrcache[0][0]
	copy(chunk[nrecs*bytesPerChild:i*bytesPerChild+32], n.merkleRoot[:])
	chunk = chunk[0 : nrecs*bytesPerChild+32]
	chunkID := wolkcommon.Computehash(chunk)
	if writeToCloudstore {
		if wg != nil {
			wg.Add(1)
		}
		// put chunk into storage queue so that final Flush on storage uses setchunkbatch call
		rc := new(cloud.RawChunk)
		rc.Value = chunk
		_, err := cs.PutChunk(context.TODO(), rc, wg)
		if err != nil {
			log.Error("[smt:flush] PutChunk ERR", "error", err)
			return err
		}
	}
	n.chunkHash = common.BytesToHash(chunkID)
	n.dirty = false
	return nil
}

// Flush recomputes the top level SMT root chunk hash and internal merkle roots
func (smt *SparseMerkleTree) Flush(ctx context.Context, wg *sync.WaitGroup, writeToCloudstore bool) (ok bool, err error) {
	dirtynodes := make(map[int][]*SMTNode)
	minlevel := TreeDepth
	smt.nodemapMu.RLock()
	for _, n := range smt.nodemap {
		if n.dirty {
			dirtynodes[n.level] = append(dirtynodes[n.level], n)
			if n.level < minlevel {
				minlevel = n.level
			}
		}
	}
	smt.nodemapMu.RUnlock()
	for i := minlevel; i <= TreeDepth; i++ {
		for _, n := range dirtynodes[i] {
			err := n.saveNode(smt.ChunkStore, wg, writeToCloudstore)
			if err != nil {
				return false, err
			}
			//			fmt.Printf(" SaveNode @ level %d: %x\n", i, n.chunkHash)
		}
	}
	return true, nil
}
