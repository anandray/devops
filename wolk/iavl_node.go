package wolk

// NOTE: This file favors int64 as opposed to int for size/counts.
// The Tree on the other hand favors int.  This is intentional.

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// Node represents a node in a Tree.
type Node struct {
	key               []byte `json:"k"`
	valhash           []byte `json:"v"` //is a common.Hash
	height            int8   `json:"h"`
	size              int64  `json:"s"`
	storageBytes      uint64 `json:"sb"`  // storageBytes for this node only
	storageBytesTotal uint64 `json:"sbt"` // storagBytes for all nodes "below" in the tree
	leftHash          []byte `json:"l,omitempty"`
	rightHash         []byte `json:"r,omitempty"`

	hash      []byte `json:"-"`
	leftNode  *Node  `json:"-"`
	rightNode *Node  `json:"-"`
	persisted bool   `json:"-"`
}

// NewNode returns a new node from a key, value and version.
func NewNode(key []byte, valhash []byte, storageBytes uint64) *Node {
	return &Node{
		key:          key,
		valhash:      valhash,
		height:       0,
		size:         1,
		storageBytes: storageBytes,
	}
}

// MakeNode constructs an *Node from an encoded byte slice.
//
// The new node doesn't have its hash saved or set. The caller must set it
// afterwards.
func MakeNode(buf []byte) (*Node, cmn.Error) {
	//dprint("[iavl_node:MakeNode] decoding")
	// Read node header (height, size, version, key).
	height, n, cause := amino.DecodeInt8(buf)
	if cause != nil {
		return nil, cmn.ErrorWrap(cause, "decoding node.height")
	}
	buf = buf[n:]

	size, n, cause := amino.DecodeVarint(buf)
	if cause != nil {
		return nil, cmn.ErrorWrap(cause, "decoding node.size")
	}
	buf = buf[n:]

	sbytes, n, cause := amino.DecodeUint64(buf)
	if cause != nil {
		return nil, cmn.ErrorWrap(cause, "decoding node.storageBytes")
	}
	buf = buf[n:]

	sbytestotal, n, cause := amino.DecodeUint64(buf)
	if cause != nil {
		return nil, cmn.ErrorWrap(cause, "decoding node.storageBytesTotal")
	}
	buf = buf[n:]

	key, n, cause := amino.DecodeByteSlice(buf)
	if cause != nil {
		return nil, cmn.ErrorWrap(cause, "decoding node.key")
	}
	buf = buf[n:]

	node := &Node{
		height:            height,
		size:              size,
		storageBytes:      sbytes,
		storageBytesTotal: sbytestotal,
		key:               key,
	}

	// Read node body.

	if node.isLeaf() {
		val, n, cause := amino.DecodeByteSlice(buf)
		if cause != nil {
			return nil, cmn.ErrorWrap(cause, "decoding node.valhash")
		}
		node.valhash = val
		buf = buf[n:]
	} else { // Read children.
		leftHash, n, cause := amino.DecodeByteSlice(buf)
		if cause != nil {
			return nil, cmn.ErrorWrap(cause, "deocding node.leftHash")
		}
		buf = buf[n:]

		rightHash, n, cause := amino.DecodeByteSlice(buf)
		if cause != nil {
			return nil, cmn.ErrorWrap(cause, "decoding node.rightHash")
		}
		node.leftHash = leftHash
		node.rightHash = rightHash
		buf = buf[n:]
	}
	// merkleRoot, _, cause := amino.DecodeByteSlice(buf)
	// if cause != nil {
	// 	return nil, cmn.ErrorWrap(cause, "decoding node.merkleRoot")
	// }
	// node.merkleRoot = merkleRoot
	//dprint("[iavl_node:MakeNode] merkleroot decoded(%x)", node.merkleRoot)

	//node.printnode()

	return node, nil
}

// String returns a string representation of the node.
func (node *Node) String() string {
	hashstr := "<no hash>"
	if len(node.hash) > 0 {
		hashstr = fmt.Sprintf("%X", node.hash)
	}
	return fmt.Sprintf("Node{%s:%s %X;%X}#%s",
		cmn.ColoredBytes(node.key, cmn.Green, cmn.Blue),
		cmn.ColoredBytes(node.valhash, cmn.Cyan, cmn.Blue),
		node.leftHash, node.rightHash,
		hashstr)
}

// clone creates a shallow copy of a node with its hash set to nil.
func (node *Node) clone() *Node {
	if node.isLeaf() {
		panic("Attempt to copy a leaf node")
	}
	return &Node{
		key:               node.key,
		height:            node.height,
		storageBytes:      node.storageBytes,
		storageBytesTotal: node.storageBytesTotal,
		size:              node.size,
		hash:              nil,
		leftHash:          node.leftHash,
		leftNode:          node.leftNode,
		rightHash:         node.rightHash,
		rightNode:         node.rightNode,
		persisted:         false,
	}
}

func (node *Node) isLeaf() bool {
	return node.height == 0
}

// Check if the node has a descendant with the given key.
func (node *Node) has(t *ImmutableTree, key []byte) (has bool) {
	if bytes.Equal(node.key, key) {
		return true
	}
	if node.isLeaf() {
		return false
	}
	if bytes.Compare(key, node.key) < 0 {
		return node.getLeftNode(t).has(t, key)
	}
	return node.getRightNode(t).has(t, key)
}

// Get a key under the node.
func (node *Node) get(t *ImmutableTree, key []byte) (index int64, value []byte, storageBytes uint64) { //do you want storageBytes or storageBytesTotal?
	if node.isLeaf() {
		switch bytes.Compare(node.key, key) {
		case -1:
			return 1, nil, uint64(0)
		case 1:
			return 0, nil, uint64(0)
		default:
			return 0, node.valhash, node.storageBytes
		}
	}

	if bytes.Compare(key, node.key) < 0 {
		return node.getLeftNode(t).get(t, key)
	}
	rightNode := node.getRightNode(t)
	index, value, storageBytes = rightNode.get(t, key)
	index += node.size - rightNode.size
	return index, value, storageBytes
}

func (node *Node) getByIndex(t *ImmutableTree, index int64) (key []byte, value []byte, storageBytes uint64) {
	if node.isLeaf() {
		if index == 0 {
			return node.key, node.valhash, node.storageBytes
		}
		return nil, nil, uint64(0)
	}
	// TODO: could improve this by storing the
	// sizes as well as left/right hash.
	leftNode := node.getLeftNode(t)

	if index < leftNode.size {
		return leftNode.getByIndex(t, index)
	}
	return node.getRightNode(t).getByIndex(t, index-leftNode.size)
}

// Computes the hash of the node without computing its descendants. Must be
// called on nodes which have descendant node hashes already computed.
func (node *Node) _hash() []byte {
	//dprint("[iavl_node:_hash] do descendant nodes have already computed hashes?")
	if node.hash != nil {
		//dprint("[iavl_node:_hash] node.hash already existed (%x)", node.hash)
		//return node.hash
	}

	h := tmhash.New()
	buf := new(bytes.Buffer)
	if err := node.writeHashBytes(buf); err != nil {
		panic(err)
	}
	h.Write(buf.Bytes())
	node.hash = h.Sum(nil)
	//dprint("[iavl_node:_hash] made new from node(%x) \n node.hash(%x) \n wolkcommon.hash(%x)", buf.Bytes(), node.hash, wolkcommon.Computehash(buf.Bytes()))
	return node.hash
}

// Hash the node and its descendants recursively. This usually mutates all
// descendant nodes. Returns the node hash and number of nodes hashed.
func (node *Node) hashWithCount() (hash []byte, count int64, storageBytesTotal uint64) {
	//dprint("[iavl_node:hashWithCount] computing node.hash by computing (and mutating) all decendant nodes")
	if node.hash != nil {
		//dprint("[iavl_node:hashWithCount] node.hash already existed (%x)", node.hash)
		return node.hash, 0, node.storageBytesTotal
	}

	h := tmhash.New()
	buf := new(bytes.Buffer)
	hashCount, err := node.writeHashBytesRecursively(buf)
	if err != nil {
		panic(err)
	}
	h.Write(buf.Bytes())
	node.hash = h.Sum(nil)
	//dprint("[iavl_node:hashWithCount] made new node.hash(%x) ct(%d)", node.hash, hashCount+1)
	return node.hash, hashCount + 1, node.storageBytesTotal
}

// Writes the node's hash to the given io.Writer. This function expects
// child hashes to be already set.
func (node *Node) writeHashBytes(w io.Writer) cmn.Error {
	//dprint("[iavl_node:writeHashBytes] WRITE HASH BYTES")
	err := amino.EncodeInt8(w, node.height)
	if err != nil {
		return cmn.ErrorWrap(err, "writing height")
	}
	err = amino.EncodeVarint(w, node.size)
	if err != nil {
		return cmn.ErrorWrap(err, "writing size")
	}
	err = amino.EncodeUint64(w, node.storageBytes)
	if err != nil {
		return cmn.ErrorWrap(err, "writing storageBytes")
	}
	err = amino.EncodeUint64(w, node.storageBytesTotal)
	if err != nil {
		return cmn.ErrorWrap(err, "writing storageBytes")
	}

	// Key is not written for inner nodes, unlike writeBytes.
	err = amino.EncodeByteSlice(w, node.key) // ?
	if err != nil {
		return cmn.ErrorWrap(err, "writing key")
	}

	if node.isLeaf() {

		// Indirection needed to provide proofs without values.
		// (e.g. proofLeafnode.valhashHash)
		// valueHash := tmhash.Sum(node.valhash)
		// err = amino.EncodeByteSlice(w, valueHash)
		//if err != nil {
		//	return cmn.ErrorWrap(err, "writing valueHash")
		//}
		err = amino.EncodeByteSlice(w, node.valhash) //?
		if err != nil {
			return cmn.ErrorWrap(err, "writing valhash")
		}
	} else {
		if node.leftHash == nil || node.rightHash == nil {
			panic("Found an empty child hash")
		}
		err = amino.EncodeByteSlice(w, node.leftHash)
		if err != nil {
			return cmn.ErrorWrap(err, "writing left hash")
		}
		err = amino.EncodeByteSlice(w, node.rightHash)
		if err != nil {
			return cmn.ErrorWrap(err, "writing right hash")
		}
	}

	//err = amino.EncodeByteSlice(w, node.merkleRoot)
	//if err != nil {
	//	return cmn.ErrorWrap(err, "writing merkleroot")
	//}

	return nil
}

// Writes the node's hash to the given io.Writer.
// This function has the side-effect of calling hashWithCount.
func (node *Node) writeHashBytesRecursively(w io.Writer) (hashCount int64, err cmn.Error) {
	//dprint("[iavl_node:writeHashBytesRecursively]")
	var leftStorageBytes uint64
	var rightStorageBytes uint64
	var leftHash []byte
	var leftCount int64
	var rightHash []byte
	var rightCount int64

	if node.leftNode != nil {
		leftHash, leftCount, leftStorageBytes = node.leftNode.hashWithCount()
		node.leftHash = leftHash
		hashCount += leftCount
	}
	if node.rightNode != nil {
		rightHash, rightCount, rightStorageBytes = node.rightNode.hashWithCount()
		node.rightHash = rightHash
		hashCount += rightCount
	}
	node.storageBytesTotal = node.storageBytes + leftStorageBytes + rightStorageBytes
	if err = node.writeHashBytes(w); err != nil {
		panic(err)
	}

	return
}

// Writes the node as a serialized byte slice to the supplied io.Writer.
func (node *Node) writeBytes(w io.Writer) cmn.Error {
	//dprint("[iavl_node:writeBytes] WRITEBYTES")
	//node.printnode()

	var cause error
	cause = amino.EncodeInt8(w, node.height)
	if cause != nil {
		return cmn.ErrorWrap(cause, "writing height")
	}
	cause = amino.EncodeVarint(w, node.size)
	if cause != nil {
		return cmn.ErrorWrap(cause, "writing size")
	}
	cause = amino.EncodeUint64(w, node.storageBytes)
	if cause != nil {
		return cmn.ErrorWrap(cause, "writing storageBytes")
	}
	cause = amino.EncodeUint64(w, node.storageBytesTotal)
	if cause != nil {
		return cmn.ErrorWrap(cause, "writing storageBytesTotal")
	}

	// Unlike writeHashBytes, key is written for inner nodes.
	cause = amino.EncodeByteSlice(w, node.key)
	if cause != nil {
		return cmn.ErrorWrap(cause, "writing key")
	}

	if node.isLeaf() {
		cause = amino.EncodeByteSlice(w, node.valhash)
		if cause != nil {
			return cmn.ErrorWrap(cause, "writing valhash")
		}
	} else {
		if node.leftHash == nil {
			panic("node.leftHash was nil in writeBytes")
		}
		cause = amino.EncodeByteSlice(w, node.leftHash)
		if cause != nil {
			return cmn.ErrorWrap(cause, "writing left hash")
		}

		if node.rightHash == nil {
			panic("node.rightHash was nil in writeBytes")
		}
		cause = amino.EncodeByteSlice(w, node.rightHash)
		if cause != nil {
			return cmn.ErrorWrap(cause, "writing right hash")
		}
	}
	//dprint("[iavl_node:writeBytes] node.merkleRoot to encode(%x)", node.merkleRoot)
	//cause = amino.EncodeByteSlice(w, node.merkleRoot)
	//if cause != nil {
	//	return cmn.ErrorWrap(cause, "writing merkleRoot")
	//}

	return nil
}

func (node *Node) getLeftNode(t *ImmutableTree) *Node {
	if node.leftNode != nil {
		return node.leftNode
	}
	return t.ndb.GetNode(node.leftHash)
}

func (node *Node) getRightNode(t *ImmutableTree) *Node {
	if node.rightNode != nil {
		return node.rightNode
	}
	return t.ndb.GetNode(node.rightHash)
}

// NOTE: mutates height and size
func (node *Node) calcHeightAndSize(t *ImmutableTree) {
	node.height = maxInt8(node.getLeftNode(t).height, node.getRightNode(t).height) + 1
	node.size = node.getLeftNode(t).size + node.getRightNode(t).size
}

func (node *Node) calcBalance(t *ImmutableTree) int {
	return int(node.getLeftNode(t).height) - int(node.getRightNode(t).height)
}

// traverse is a wrapper over traverseInRange when we want the whole tree
func (node *Node) traverse(t *ImmutableTree, ascending bool, cb func(*Node) bool) bool {
	return node.traverseInRange(t, nil, nil, ascending, false, 0, func(node *Node, depth uint8) bool {
		return cb(node)
	})
}

func (node *Node) traverseWithDepth(t *ImmutableTree, ascending bool, cb func(*Node, uint8) bool) bool {
	return node.traverseInRange(t, nil, nil, ascending, false, 0, cb)
}

func (node *Node) traverseInRange(t *ImmutableTree, start, end []byte, ascending bool, inclusive bool, depth uint8, cb func(*Node, uint8) bool) bool {
	afterStart := start == nil || bytes.Compare(start, node.key) < 0
	startOrAfter := start == nil || bytes.Compare(start, node.key) <= 0
	beforeEnd := end == nil || bytes.Compare(node.key, end) < 0
	if inclusive {
		beforeEnd = end == nil || bytes.Compare(node.key, end) <= 0
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
			stop = node.getLeftNode(t).traverseInRange(t, start, end, ascending, inclusive, depth+1, cb)
		}
		if stop {
			return stop
		}
		if beforeEnd {
			stop = node.getRightNode(t).traverseInRange(t, start, end, ascending, inclusive, depth+1, cb)
		}
	} else {
		// check the higher nodes first
		if beforeEnd {
			stop = node.getRightNode(t).traverseInRange(t, start, end, ascending, inclusive, depth+1, cb)
		}
		if stop {
			return stop
		}
		if afterStart {
			stop = node.getLeftNode(t).traverseInRange(t, start, end, ascending, inclusive, depth+1, cb)
		}
	}

	return stop
}

func (node *Node) getStorageBytesTotal() uint64 {
	return node.storageBytesTotal
}

// Only used in testing...
func (node *Node) lmd(t *ImmutableTree) *Node {
	if node.isLeaf() {
		return node
	}
	return node.getLeftNode(t).lmd(t)
}
