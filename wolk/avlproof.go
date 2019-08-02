package wolk

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/wolkdb/cloudstore/log"
	//"github.com/ethereum/go-ethereum/common"
	cmn "github.com/tendermint/tendermint/libs/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	//"github.com/wolkdb/cloudstore/wolk/cloud"
)

type AVLProof struct {
	// You don't need the right path because it can be derived from what we have.
	LeftPath   PPathToLeaf   `json:"lp"`
	InnerNodes []PPathToLeaf `json:"in"`
	Leaves     []pLeaf       `json:"lfs"`

	// for bookkeeping
	rootVerified bool
	rootHash     []byte // valid iff rootVerified is true
	treeEnd      bool   // valid iff rootVerified is true
}

// proof's inner nodes
type pInner struct {
	Key               cmn.HexBytes `json:"k"`
	Height            int8         `json:"h"`
	Size              int64        `json:"s"`
	StorageBytesTotal uint64       `json:"sbt"`
	Left              cmn.HexBytes `json:"l"`
	Right             cmn.HexBytes `json:"r"`
}

// proof's leaf nodes
type pLeaf struct {
	Key          cmn.HexBytes `json:"k"`
	ValHash      cmn.HexBytes `json:"v"`
	Height       int8         `json:"h"`
	Size         int64        `json:"s"`
	StorageBytes uint64       `json:"sb"`
}

// proof's path with leaf
type pPathWithLeaf struct {
	Path PPathToLeaf `json:"p"`  // just an array of inner nodes
	Leaf pLeaf       `json:"lf"` // single leaf node
}

// proof's path to leaf, without the leaf
type PPathToLeaf []pInner //just and array of inner nodes

func (proof *AVLProof) Verify(rootToVerify []byte) error {
	if proof == nil {
		return fmt.Errorf("[avlproof:Verify] proof is nil!")
	}
	//dprint("[iavl_proof_range:verify] root to verify(%x)", root)
	rootHash := proof.rootHash
	if rootHash == nil {
		//dprint("[iavl_proof_range:verify] roothash of proof was nil. computing.")
		derivedHash, err := proof.computeRootHash()
		if err != nil {
			return fmt.Errorf("[avlproof:Verify] %s", err)
		}
		rootHash = derivedHash
	}
	//dprint("[iavl_proof_range:verify] roothash(%x)", rootHash)

	if !bytes.Equal(rootHash, rootToVerify) {
		log.Error("[avlproof:Verify] roothash doesn't match!", "root computed", hex.EncodeToString(rootHash), "root to verify", hex.EncodeToString(rootToVerify))
		return fmt.Errorf("[avlproof:Verify] roothash doesnt match")
	} else {
		proof.rootVerified = true
		//log.Info("[avlproof:Verify] matched", "root computed", hex.EncodeToString(rootHash), "root to verify", hex.EncodeToString(rootToVerify))
	}
	return nil
}

func (proof *AVLProof) computeRootHash() (rootHash []byte, err error) {
	rootHash, treeEnd, err := proof._computeRootHash()
	if err == nil {
		proof.rootHash = rootHash // memoize
		proof.treeEnd = treeEnd   // memoize
	}
	return rootHash, err
}

func (proof *AVLProof) _computeRootHash() (rootHash []byte, treeEnd bool, err error) {
	if len(proof.Leaves) == 0 {
		return nil, false, fmt.Errorf("[avlproof:_computeRootHash] invalid proof. no leaves")
	}
	if len(proof.InnerNodes)+1 != len(proof.Leaves) {
		return nil, false, fmt.Errorf("[avlproof:_computeRootHash] InnerNodes vs Leaves length mismatch, leaves should be 1 more.")
	}

	// Start from the left path and prove each leaf.
	// shared across recursive calls
	var leaves = proof.Leaves
	var innersq = proof.InnerNodes
	var COMPUTEHASH func(path PPathToLeaf, rightmost bool) (hash []byte, treeEnd bool, done bool, err error)

	// rightmost: is the root a rightmost child of the tree?
	// treeEnd: true iff the last leaf is the last item of the tree.
	// Returns the (possibly intermediate, possibly root) hash.
	COMPUTEHASH = func(path PPathToLeaf, rightmost bool) (hash []byte, treeEnd bool, done bool, err error) {

		// Pop next leaf.
		nleaf, rleaves := leaves[0], leaves[1:]
		leaves = rleaves
		//dprint("[proof:_computeRootHash] pop next leaf(%+v)(%x)", nleaf, nleaf)
		// Compute hash.
		hash = (pPathWithLeaf{
			Path: path,
			Leaf: nleaf,
		}).computeRootHash()
		//dprint("[proof:_computeRootHash] hash computed(%x)", hash)

		// If we don't have any leaves left, we're done.
		if len(leaves) == 0 {
			rightmost = rightmost && path.isRightmost()
			return hash, rightmost, true, nil
		}

		// Prove along path (until we run out of leaves).
		for len(path) > 0 {

			// Drop the leaf-most (last-most) inner nodes from path
			// until we encounter one with a left hash.
			// We assume that the left side is already verified.
			// rpath: rest of path
			// lpath: last path item
			rpath, lpath := path[:len(path)-1], path[len(path)-1]
			path = rpath
			if len(lpath.Right) == 0 {
				continue
			}

			// Pop next inners, a PathToLeaf (e.g. []proofInnerNode).
			inners, rinnersq := innersq[0], innersq[1:]
			innersq = rinnersq

			// Recursively verify inners against remaining leaves.
			derivedRoot, treeEnd, done, err := COMPUTEHASH(inners, rightmost && rpath.isRightmost())
			if err != nil {
				return nil, treeEnd, false, fmt.Errorf("[avlproof:_computeRootHash] recursive COMPUTEHASH call err %s", err)
			}
			if !bytes.Equal(derivedRoot, lpath.Right) {
				return nil, treeEnd, false, fmt.Errorf("[avlproof:_computeRootHash] INVALID ROOT: intermediate root hash %X doesn't match, got %X", lpath.Right, derivedRoot)
			}
			if done {
				return hash, treeEnd, true, nil
			}
		}

		// We're not done yet (leaves left over). No error, not done either.
		// Technically if rightmost, we know there's an error "left over leaves
		// -- malformed proof", but we return that at the top level, below.
		return hash, false, false, nil
	}

	// Verify!
	path := proof.LeftPath
	rootHash, treeEnd, done, err := COMPUTEHASH(path, true)
	if err != nil {
		return nil, treeEnd, cmn.ErrorWrap(err, "root COMPUTEHASH call")
	} else if !done {
		return nil, treeEnd, cmn.ErrorWrap(ErrInvalidProof, "left over leaves -- malformed proof")
	}

	// Ok!
	return rootHash, treeEnd, nil
}

func (proof *AVLProof) VerifyItem(key, value []byte) error {
	//dprint("[iavl_proof_range:VerifyItem] proof(%+v)", proof)
	leaves := proof.Leaves
	if proof == nil {
		return fmt.Errorf("[avlProof:VerifyItem] proof is nil")
	}
	if !proof.rootVerified {
		return fmt.Errorf("[avlProof:VerifyItem] must call Verify(root) first.")
	}
	i := sort.Search(len(leaves), func(i int) bool {
		return bytes.Compare(key, leaves[i].Key) <= 0
	})
	if i >= len(leaves) || !bytes.Equal(leaves[i].Key, key) {
		log.Error("[iavl_proof_range:VerifyItem] leaf key not found in proof", "key", hex.EncodeToString(key))
		return fmt.Errorf("[avlproof:VerifyItem] leaf key not found in proof")
	}
	value = bytes.Trim(value, "\x00")
	leafvalue := bytes.Trim(leaves[i].ValHash, "\x00")
	//valueHash := tmhash.Sum(value)
	//dprint("value (%x)", value)
	if !bytes.Equal(leafvalue, value) {
		//dprint("[iavl_proof_range:VerifyItem] leaf value hash not same.\n leaf val(%x) \n value(%x) \n valueHash(%x)", leaves[i].ValueHash, value, valueHash)
		dprint("[avlproof:VerifyItem] leaf value hash not same.\n leaf value(%x) \n valHash to verify(%x)", leaves[i].ValHash, value)
		return fmt.Errorf("[avlproof:VerifyItem] leaf value hash not same")
	}
	return nil
}

//----------------------------------------

// `verify` checks that the leaf node's hash + the inner nodes merkle-izes to
// the given root. If it returns an error, it means the leafHash or the
// PathToLeaf is incorrect.
func (pwl pPathWithLeaf) verify(root []byte) error {
	//dprint("[proof_path:pwl:verify] pwl.Leaf(%+v), hash(%x)", pwl.Leaf, pwl.Leaf.Hash())
	leafHash := pwl.Leaf.Hash()
	return pwl.Path.verify(leafHash, root)
}

// `computeRootHash` computes the root hash with leaf node.
// Does not verify the root hash.
func (pwl pPathWithLeaf) computeRootHash() []byte {
	//dprint("[proof_path:pwl:computeRootHash] pwl.Leaf(%+v), hash(%x)", pwl.Leaf, pwl.Leaf.Hash())

	leafHash := pwl.Leaf.Hash()
	return pwl.Path.computeRootHash(leafHash)
}

//----------------------------------------

// `verify` checks that the leaf node's hash + the inner nodes merkle-izes to
// the given root. If it returns an error, it means the leafHash or the
// PathToLeaf is incorrect.
func (pl PPathToLeaf) verify(leafHash []byte, root []byte) error {
	hash := leafHash
	for i := len(pl) - 1; i >= 0; i-- {
		pin := pl[i]
		hash = pin.Hash(hash)
		//dprint("[proof_path:pathtoleaf:verify] hash(%x)", hash)
	}
	if !bytes.Equal(root, hash) {
		log.Error("[avlproof:PPathtoleaf:verify] root != hash", "root", hex.EncodeToString(root), "hash", hex.EncodeToString(hash))
		return fmt.Errorf("[avlproof:PPathToLeaf:verify] root != hash")
	}
	return nil
}

// `computeRootHash` computes the root hash assuming some leaf hash.
// Does not verify the root hash.
func (pl PPathToLeaf) computeRootHash(leafHash []byte) []byte {
	hash := leafHash
	//dprint("[avlproof:computeRootHash] leafHash:\t%x", hash)
	for i := len(pl) - 1; i >= 0; i-- {
		pin := pl[i]
		//dprint("pin(%+v)", pin)
		//dprint("[avlproof:computeRootHash] %d:\t%x\t(%x)\n", i, hash, pin.Hash(hash))
		hash = pin.Hash(hash)
	}
	return hash
}

func (pl PPathToLeaf) isLeftmost() bool {
	for _, node := range pl {
		if len(node.Left) > 0 {
			return false
		}
	}
	return true
}

func (pl PPathToLeaf) isRightmost() bool {
	for _, node := range pl {
		if len(node.Right) > 0 {
			return false
		}
	}
	return true
}

func (pl PPathToLeaf) isEmpty() bool {
	return pl == nil || len(pl) == 0
}

func (pl PPathToLeaf) dropRoot() PPathToLeaf {
	if pl.isEmpty() {
		return pl
	}
	return PPathToLeaf(pl[:len(pl)-1])
}

func (pl PPathToLeaf) hasCommonRoot(pl2 PPathToLeaf) bool {
	if pl.isEmpty() || pl2.isEmpty() {
		return false
	}
	leftEnd := pl[len(pl)-1]
	rightEnd := pl2[len(pl2)-1]

	return bytes.Equal(leftEnd.Left, rightEnd.Left) &&
		bytes.Equal(leftEnd.Right, rightEnd.Right)
}

func (pl PPathToLeaf) isLeftAdjacentTo(pl2 PPathToLeaf) bool {
	for pl.hasCommonRoot(pl2) {
		pl, pl2 = pl.dropRoot(), pl2.dropRoot()
	}
	pl, pl2 = pl.dropRoot(), pl2.dropRoot()

	return pl.isRightmost() && pl2.isLeftmost()
}

// returns -1 if invalid.
func (pl PPathToLeaf) Index() (idx int64) {
	for i, node := range pl {
		if node.Left == nil {
			continue
		} else if node.Right == nil {
			if i < len(pl)-1 {
				idx += node.Size - pl[i+1].Size
			} else {
				idx += node.Size - 1
			}
		} else {
			return -1
		}
	}
	return idx
}

//----------------------------------------

func (node *AVLNode) PathToLeaf(ndb *avlNodeDB, key []byte) (p PPathToLeaf, val *AVLNode, err error) {
	path := new(PPathToLeaf)
	val, err = node.pathToLeaf(ndb, key, path)
	return *path, val, fmt.Errorf("[PathToLeaf] %s", err)
}

func (node *AVLNode) pathToLeaf(ndb *avlNodeDB, key []byte, path *PPathToLeaf) (*AVLNode, error) {
	node = node.clone(true)
	if node.Height == 0 {
		if bytes.Equal(node.Key, key) {
			return node, nil
		}
		return node, fmt.Errorf("[pathToLeaf] key does not exist!")
	}

	if bytes.Compare(key, node.Key) < 0 {
		// left side
		pin := pInner{
			Height:            node.Height,
			Size:              node.Size,
			StorageBytesTotal: node.StorageBytesTotal,
			Key:               node.Key,
			Left:              nil,
			Right:             node.getRightNode(ndb).hash,
		}
		*path = append(*path, pin)
		n, err := node.getLeftNode(ndb).pathToLeaf(ndb, key, path)
		return n, err
	}
	// right side
	pin := pInner{
		Height:            node.Height,
		Size:              node.Size,
		StorageBytesTotal: node.StorageBytesTotal,
		Key:               node.Key,
		Left:              node.getLeftNode(ndb).hash,
		Right:             nil,
	}
	*path = append(*path, pin)
	n, err := node.getRightNode(ndb).pathToLeaf(ndb, key, path)
	return n, fmt.Errorf("[pathToLeaf] %s", err)
}

func (node *AVLNode) traverseInRange(ndb *avlNodeDB, start, end []byte, ascending bool, inclusive bool, depth uint8, cb func(*AVLNode, uint8) bool) bool {
	afterStart := start == nil || bytes.Compare(start, node.Key) < 0
	startOrAfter := start == nil || bytes.Compare(start, node.Key) <= 0
	beforeEnd := end == nil || bytes.Compare(node.Key, end) < 0
	if inclusive {
		beforeEnd = end == nil || bytes.Compare(node.Key, end) <= 0
	}

	// Run callback per inner/leaf node.
	stop := false
	if !node.isLeaf() || (startOrAfter && beforeEnd) {
		stop = cb(node, depth)
		if stop {
			return stop
		}
	}
	if node.isLeaf() {
		return stop
	}

	if ascending {
		// check lower nodes, then higher
		if afterStart {
			stop = node.getLeftNode(ndb).traverseInRange(ndb, start, end, ascending, inclusive, depth+1, cb)
		}
		if stop {
			return stop
		}
		if beforeEnd {
			stop = node.getRightNode(ndb).traverseInRange(ndb, start, end, ascending, inclusive, depth+1, cb)
		}
	} else {
		// check the higher nodes first
		if beforeEnd {
			stop = node.getRightNode(ndb).traverseInRange(ndb, start, end, ascending, inclusive, depth+1, cb)
		}
		if stop {
			return stop
		}
		if afterStart {
			stop = node.getLeftNode(ndb).traverseInRange(ndb, start, end, ascending, inclusive, depth+1, cb)
		}
	}

	return stop
}

func (pin pInner) Hash(childHash []byte) []byte {
	if len(childHash) > 0 && childHash != nil {
		if len(pin.Left) == 0 {
			pin.Left = childHash
		} else {
			pin.Right = childHash
		}
	}
	bytes, err := json.Marshal(pin)
	if err != nil {
		panic(err)
	}
	return wolkcommon.Computehash(bytes)
}

func (pln pLeaf) Hash() []byte {
	bytes, err := json.Marshal(pln)
	if err != nil {
		panic(err)
	}
	return wolkcommon.Computehash(bytes)
}

// helper functions
/*
// String returns a string representation of the proof.
func (proof *AVLProof) String() string {
	if proof == nil {
		return "<nil-AVLProof>"
	}
	return proof.StringIndented("")
}

func (proof *AVLProof) StringIndented(indent string) string {
	istrs := make([]string, 0, len(proof.InnerNodes))
	for _, ptl := range proof.InnerNodes {
		istrs = append(istrs, ptl.stringIndented(indent+"    "))
	}
	lstrs := make([]string, 0, len(proof.Leaves))
	for _, leaf := range proof.Leaves {
		lstrs = append(lstrs, leaf.stringIndented(indent+"    "))
	}
	return fmt.Sprintf(`RangeProof{
%s  LeftPath: %v
%s  InnerNodes:
%s    %v
%s  Leaves:
%s    %v
%s  (rootVerified): %v
%s  (rootHash): %X
%s  (treeEnd): %v
%s}`,
		indent, proof.LeftPath.stringIndented(indent+"  "),
		indent,
		indent, strings.Join(istrs, "\n"+indent+"    "),
		indent,
		indent, strings.Join(lstrs, "\n"+indent+"    "),
		indent, proof.rootVerified,
		indent, proof.rootHash,
		indent, proof.treeEnd,
		indent)
}

func (pl PPathToLeaf) stringIndented(indent string) string {
	if len(pl) == 0 {
		return "empty-PathToLeaf"
	}
	strs := make([]string, len(pl))
	for i, pin := range pl {
		if i == 20 {
			strs[i] = fmt.Sprintf("... (%v total)", len(pl))
			break
		}
		strs[i] = fmt.Sprintf("%v:%v", i, pin.stringIndented(indent+"  "))
	}
	return fmt.Sprintf(`PathToLeaf{
%s  %v
%s}`,
		indent, strings.Join(strs, "\n"+indent+"  "),
		indent)
}

func (pin pInner) String() string {
	return pin.stringIndented("")
}

func (pin pInner) stringIndented(indent string) string {
	return fmt.Sprintf(`proofInnerNode{
%s  Key:               %X
%s  Height:            %v
%s  Size:              %v
%s  StorageBytesTotal: %v
%s  Left:              %X
%s  Right:             %X
%s}`,
		indent, pin.Key,
		indent, pin.Height,
		indent, pin.Size,
		indent, pin.StorageBytesTotal,
		indent, pin.Left,
		indent, pin.Right,
		indent)
}

func (pln pLeaf) String() string {
	return pln.stringIndented("")
}

func (pln pLeaf) stringIndented(indent string) string {
	return fmt.Sprintf(`proofLeafNode{
%s  Key:               %X
%s  ValHash:           %X
%s  Height:            %v
%s  Size:              %v
%s  StorageBytes:      %v
%s}`,
		indent, pln.Key,
		indent, pln.ValHash,
		indent, pln.Height,
		indent, pln.Size,
		indent, pln.StorageBytes,
		indent)
}

*/
