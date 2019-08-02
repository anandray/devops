package wolk

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/log"
)

var avlZeroBytes []byte
var avlNullBytes []byte

const (
	defaultLimit = 100
)

// AVLTree is a BST compliant merkle tree
type AVLTree struct {
	chunkStore ChunkStore
	ndb        *avlNodeDB
	root       *AVLChunk
	chunkHash  common.Hash //root node's chunkhash.. also in root.node.chunkhash. redundant. TODO: fix
	sync.RWMutex
}

// NewAVLTree returns a new, empty tree
func NewAVLTree(cs ChunkStore) *AVLTree {
	tree := new(AVLTree)
	tree.chunkStore = cs
	tree.ndb = newAVLNodeDB(cs)
	tree.root = new(AVLChunk)
	tree.root.Children = make([]*AVLNode, 0)
	return tree
}

func (tree *AVLTree) StorageBytes() (uint64, error) {
	// TODO
	return 0, nil
}

func (tree *AVLTree) GetWithoutProof(k []byte) (v0 []byte, found bool, deleted bool, storageBytes uint64, err error) {
	v0, found, deleted, _, storageBytes, err = tree.Get(k)
	return v0, found, deleted, storageBytes, err
}

// Init loads an existing tree using a ROOT chunkHash.
func (tree *AVLTree) Init(rootHash common.Hash) {
	var emptyHash common.Hash
	if bytes.Compare(rootHash.Bytes(), emptyHash.Bytes()) == 0 || bytes.Compare(rootHash.Bytes(), getEmptySMTChunkHash()) == 0 {
		log.Info("[avltree:Init] tried to load empty tree")
		return
	}
	tree.root = tree.ndb.loadChunk(rootHash)
}

// Load loads a chunk into an existing tree. does not have to be the root chunk.
// returns the root node of that chunk.
func (tree *AVLTree) Load(chunkHash common.Hash) *AVLChunk {
	return tree.ndb.loadChunk(chunkHash)
}

func (tree *AVLTree) ChunkHash() common.Hash {
	//TODO, b/c if this isn't computed, may have to compute it here on the fly
	root := tree.GetRootNode()
	if root == nil {
		return common.BytesToHash(getEmptySMTChunkHash())
	}
	return root.ChunkHash
}

func (tree *AVLTree) MerkleRoot(ctx context.Context) common.Hash {
	root := tree.GetRootNode()
	return common.BytesToHash(root.hash)
}

func (tree *AVLTree) Flush(ctx context.Context, wg *sync.WaitGroup, writeToCloudstore bool) (ok bool, err error) {
	//defer timeTrack(time.Now(), "Flush")
	if tree.root == nil || tree.root.Node == nil {
		log.Info("[avltree:Flush] trying to save empty tree")
		tree.SetEmptyRootChunkHash()
		return true, nil
	}
	if tree.root.Node.persisted == true {
		log.Info("[avltree:Flush] tree has no modifications. No flushing needed.")
		return false, nil
	}
	root := tree.GetRootChunk()
	tree.ndb.SaveBranch(root, wg, writeToCloudstore)
	//dprint("[Flush] saved tree chunkhash(%x), total storagebytes(%v)", tree.root.Node.ChunkHash, sbt)
	//tree.chunkHash = tree.root.Node.ChunkHash
	tree.SetRootChunkHash()

	return true, nil
}

// func (tree *AVLTree) SaveTree(wg *sync.WaitGroup, writeToCloudstore bool) ([]byte, error) {
//
// 	return tree.root.hash, nil //check this
// }

func (tree *AVLTree) Get(k []byte) (v []byte, found bool, deleted bool, p interface{}, storageBytes uint64, err error) {
	// TODO: proofs
	if tree.root == nil {
		return v, found, deleted, p, storageBytes, fmt.Errorf("[avltree:Get] No tree loaded")
	} //else if tree.root.Node.persisted == false {
	// ok to get from memory only?? TODO: think about this.
	//panic("[avltree:Get] WARNING: Getting from a dirty tree in memory, not loaded to cloudstore. Flush first.")
	//log.Info("[avltree:Get] WARNING: Getting from a dirty tree in memory, not flushed to cloudstore.")
	//}
	root := tree.GetRootNode()
	found = true
	_, v, storageBytes = root.get(tree.ndb, k)
	if v == nil {
		found = false
	}
	deleted = false // TODO
	return v, found, deleted, p, storageBytes, nil
}

// TODO: deleted
func (tree *AVLTree) GetWithProof(ctx context.Context, k []byte) (v []byte, found bool, deleted bool, proof interface{}, storageBytes uint64, err error) {

	p, _, values, err := tree.getAVLProof(k, cpIncr(k), 2)
	if err != nil {
		return v, found, deleted, proof, storageBytes, fmt.Errorf("[avltree:GetWithProof] %s", err)
	}
	// dprint("proof! %+v", p)
	// for _, dk := range keys {
	// 	dprint("  key: %x", dk)
	// }
	// for _, dv := range values {
	// 	dprint("  val: %x", dv)
	// }

	// found it, and it's the right key
	if len(values) > 0 && bytes.Equal(p.Leaves[0].Key, k) {
		rootnode := tree.GetRootNode()
		sbt := uint64(0)
		if rootnode.isLeaf() {
			sbt = rootnode.StorageBytes
		} else {
			sbt = rootnode.StorageBytesTotal
		}
		return values[0], true, false, p, sbt, nil
	}
	// didn't find it.
	return v, false, false, proof, storageBytes, nil
	// TODO: deleted
}

/*
func (t *ImmutableTree) GetWithProof(key []byte) (value []byte, proof *RangeProof, err error) {
	//dprint("[iavl_proof:GetWithProof] key(%x)", key)
	proof, _, values, err := t.getRangeProof(key, cpIncr(key), 2)
	if err != nil {
		return nil, nil, cmn.ErrorWrap(err, "constructing range proof")
	}
	if len(values) > 0 && bytes.Equal(proof.Leaves[0].Key, key) {
		return values[0], proof, nil
	}
	return nil, proof, nil
}*/

func (tree *AVLTree) Insert(ctx context.Context, key []byte, value []byte, storageBytesNew uint64, deleted bool) (err error) {
	//defer timeTrack(time.Now(), fmt.Sprintf("Insert k(%x) v(%x)", key, value))
	// TODO: deleted
	if value == nil {
		// TODO, is this a delete?
		//return false, fmt.Errorf("[avltree:avltree-set] value is nil")
		panic(fmt.Errorf("[avltree:set] value is nil"))
	}
	if tree.root.Node == nil {
		tnode := NewAVLNode(key, value, storageBytesNew)
		tree.SetRootNode(tnode)
		return nil
	}
	tnode := tree.GetRootNode()
	tnode, _, _ = tree.recursiveSet(tnode, key, value, storageBytesNew)
	tree.SetRootNode(tnode)

	//dprint("[Insert] updated: %v", updated)
	// TODO: clean up orphaned...go through and remove nodes in child arrays. orphaned should only come from balancing. (will we get orphans from balancing now? b/c ids are stable?)
	return nil
}

// recursiveSet is recursively called on nodes in the tree to insert a new kv pair.
func (tree *AVLTree) recursiveSet(node *AVLNode, key []byte, value []byte, storageBytesNew uint64) (newSelf *AVLNode, updated bool, orphaned []*AVLNode) {
	//dprint("[recursiveSet] **** on node.nodeid(%x), inserting key(%x) value(%x) sb(%v)", node.id, shortbytes(key), shortbytes(value), storageBytesNew)
	//node.Lock()
	//defer node.Unlock()
	node = node.clone(true)
	if node.isLeaf() {
		//dprint("[recursiveSet] node is leaf")
		switch bytes.Compare(key, node.Key) {
		case -1: // key < node.Key
			//dprint("[recursiveSet] case -1")
			leftNode := NewAVLNode(key, value, storageBytesNew)
			newSelf = &AVLNode{
				Key:          node.Key,
				Valhash:      node.Valhash,
				Height:       1,
				Size:         2,
				StorageBytes: node.StorageBytes,
				LeftID:       &leftNode.id,
				RightID:      &node.id,
				persisted:    false,
				id:           makeID(false, node.Key),
			}
			tree.ndb.SetNode(leftNode)
			tree.ndb.SetNode(newSelf)
			tree.ndb.SetNode(node)
			return newSelf, false, []*AVLNode{}
		case 1: // key > node.Key
			//dprint("[recursiveSet] case 1")
			rightNode := NewAVLNode(key, value, storageBytesNew)
			newSelf = &AVLNode{
				Key:          key,
				Valhash:      value,
				Height:       1,
				Size:         2,
				StorageBytes: storageBytesNew,
				LeftID:       &node.id,
				RightID:      &rightNode.id,
				persisted:    false,
				id:           makeID(false, key),
			}
			tree.ndb.SetNode(rightNode)
			tree.ndb.SetNode(newSelf)
			tree.ndb.SetNode(node)
			return newSelf, false, []*AVLNode{}
		default: // key == node.Key
			//dprint("[recursiveSet] default, returning newnode")
			newSelf := NewAVLNode(key, value, storageBytesNew)
			tree.ndb.SetNode(newSelf)
			tree.ndb.SetNode(node)
			return newSelf, true, []*AVLNode{node}
		}

	} else { // is inner node
		//dprint("[recursiveSet] is inner node")
		orphaned = append(orphaned, node)
		node = node.clone(false)

		newLeftNode := new(AVLNode)
		newRightNode := new(AVLNode)
		if bytes.Compare(key, node.Key) < 0 {
			//dprint("[recursiveSet] key(%x) < node.Key(%x)", key, node.Key)
			var leftOrphaned []*AVLNode
			//var newLeftNode *AVLNode
			newLeftNode, updated, leftOrphaned = tree.recursiveSet(node.getLeftNode(tree.ndb), key, value, storageBytesNew)
			tree.ndb.SetNode(newLeftNode)
			//node := tree.ndb.GetNode(node.id, EMPTYBYTES)
			//dprint("[recursiveSet] newLeftNode:(%s)", newLeftNode)
			//dprint("[recursiveSet] node:(%s)", node)
			node.LeftID = &newLeftNode.id
			//dprint("[recursiveSet] node w/ id chg:(%s)", node)
			tree.ndb.SetNode(node)
			orphaned = append(orphaned, leftOrphaned...)

		} else {
			//dprint("[recursiveSet] key(%x) >= node.Key(%x)", key, node.Key)
			var rightOrphaned []*AVLNode
			//var newRightNode *AVLNode
			//dprint("[recursiveSet] **inner / >= / rightNode(%s)", rightNode)
			newRightNode, updated, rightOrphaned = tree.recursiveSet(node.getRightNode(tree.ndb), key, value, storageBytesNew)
			tree.ndb.SetNode(newRightNode)
			//node := tree.ndb.GetNode(node.id, EMPTYBYTES)
			//dprint("[recursiveSet] newRighttNode:(%s)", newRightNode)
			//newRightNode_tmp := tree.ndb.GetNode(newRightNode.id, EMPTYBYTES)
			//dprint("[recursiveset] after getting again newRightNode(%s)", newRightNode_tmp)

			//dprint("[recursiveSet] node:(%s)", node)
			node.RightID = &newRightNode.id
			//dprint("[recursiveSet] node w/ id chg:(%s)", node)
			tree.ndb.SetNode(node)
			orphaned = append(orphaned, rightOrphaned...)

			//dprint("[recursiveset] print tree before leaving else - check newRightNode")
			//tree.PrintTree(node, false, false)
			//dprint("\n")
		}

		if updated { // leaf node added only - don't have to rebalance
			//dprint("[recursiveset] when does this update happen?")
			return node, updated, orphaned
		}
		//if newRightNode.id != EMPTYBYTES {
		//dprint("[recursiveset] before calch&s newRightNode(%s)", newRightNode)
		//newRightNode = tree.ndb.GetNode(newRightNode.id, EMPTYBYTES)
		//dprint("[recursiveset] after getting again newRightNode(%s)", newRightNode)
		//}
		//nrn := tree.ndb.GetNode(*node.RightID, node.ChunkHash)
		//dprint("[recursiveset] straight up get right node of node: nrn(%s) ptr(%v)", nrn, node.RightID)
		node = node.clone(true)
		node.Height, node.Size = node.calcHeightAndSize(tree.ndb)
		tree.ndb.SetNode(node)
		newSelf, balanceOrphaned := tree.balance(node)
		//tree.ndb.SetNode(newSelf)
		//if bytes.Contains(node.id.Bytes(), id) {
		//dprint("[recursiveSet] returning from newSelf:")
		//dprint("\n[recursiveSet] after balanced: newSelf(%s)", newSelf)
		//tree.PrintTree(newSelf, false, true)
		//}
		//
		return newSelf, updated, append(orphaned, balanceOrphaned...)
	}
}

// Rotate right and return the new node and orphan.
func (tree *AVLTree) rotateRight(node *AVLNode) (*AVLNode, *AVLNode) {

	node = node.clone(false)
	orphaned := node.getLeftNode(tree.ndb)
	newSelf := orphaned.clone(false)

	newSelfRightID := newSelf.RightID
	newSelf.RightID = &node.id
	node.LeftID = newSelfRightID
	tree.ndb.SetNode(node)
	tree.ndb.SetNode(newSelf)

	//dprint("[rotateRight] node(%+v)", node)
	//dprint("[rotateRight] newSelf(%+v)", newSelf)
	node.Height, node.Size = node.calcHeightAndSize(tree.ndb)
	newSelf.Height, newSelf.Size = newSelf.calcHeightAndSize(tree.ndb)
	tree.ndb.SetNode(node)
	tree.ndb.SetNode(newSelf)
	return newSelf, orphaned //note: orphaned is not set in nodeMap
}

// Rotate left and return the new node and orphan.
func (tree *AVLTree) rotateLeft(node *AVLNode) (*AVLNode, *AVLNode) {
	//dprint("[rotateLeft] node(%+v)", node)
	node = node.clone(false)
	//dprint("[rotateLeft] after clone: node(%+v)", node)
	orphaned := node.getRightNode(tree.ndb)
	newSelf := orphaned.clone(false)
	//dprint("[rotateLeft] newSelf before(%s)", newSelf)

	newSelfLeftID := newSelf.LeftID
	newSelf.LeftID = &node.id
	node.RightID = newSelfLeftID
	tree.ndb.SetNode(node)
	tree.ndb.SetNode(newSelf)

	//dprint("[rotateLeft] newSelf after(%s)", newSelf)
	//dprint("[rotateLeft] node after(%s)", node)
	node.Height, node.Size = node.calcHeightAndSize(tree.ndb)
	newSelf.Height, newSelf.Size = newSelf.calcHeightAndSize(tree.ndb)
	tree.ndb.SetNode(node)
	tree.ndb.SetNode(newSelf)
	return newSelf, orphaned //note orphaned is not set in nodeMap
}

func (tree *AVLTree) balance(node *AVLNode) (newSelf *AVLNode, orphaned []*AVLNode) {

	node = node.clone(true)
	//dprint("[balance] node(%+v)", node)
	if node.persisted {
		panic("Unexpected balance() call on persisted self") //don't balance unmodified node
	}
	balance := node.calcBalance(tree.ndb)

	if balance > 1 {
		left := node.getLeftNode(tree.ndb)
		if left.calcBalance(tree.ndb) >= 0 {
			// Left Left Case
			//dprint("[balance] rotateRight node(%+v)", node)
			newSelf, orphaned := tree.rotateRight(node)
			return newSelf, []*AVLNode{orphaned}
		}
		// Left Right Case
		//dprint("[balance] nil'ed out LeftID. rotateLeft node(%s)", node)
		//dprint("[balance] nil'ed out LeftID. rotateLeft left(%s)", left)
		newleft, leftOrphaned := tree.rotateLeft(left)
		node.LeftID = &newleft.id
		tree.ndb.SetNode(node)
		//dprint("[balance] rewrote LeftID. node(%s)", node)
		newSelf, rightOrphaned := tree.rotateRight(node)
		return newSelf, []*AVLNode{left, leftOrphaned, rightOrphaned}
	}
	if balance < -1 {
		right := node.getRightNode(tree.ndb)
		if right.calcBalance(tree.ndb) <= 0 {
			// Right Right Case
			//dprint("[balance] right right case")
			//dprint("[balance] rotateLeft node(%+v)", node)
			newSelf, orphaned := tree.rotateLeft(node)
			return newSelf, []*AVLNode{orphaned}
		}
		// Right Left Case
		//dprint("[balance] right left case")
		//dprint("[balance] node's right node (%s)", right)
		//dprint("[balance] nil'ed out RightID. rotateRight node(%+v)", node)
		//dprint("[balance] nil'ed out RightID. rotateRight left(%s)", right)
		newRight, rightOrphaned := tree.rotateRight(right)
		//dprint("[balance] node's new right node, after rotating right(%s)", newRight)
		node.RightID = &newRight.id
		tree.ndb.SetNode(node)
		//dprint("node is now(%s)", node)
		newSelf, leftOrphaned := tree.rotateLeft(node)
		//dprint("[balance] node's newNode, after rotating left(%s)", newSelf)
		return newSelf, []*AVLNode{right, leftOrphaned, rightOrphaned}
	}
	// Nothing changed
	tree.ndb.SetNode(node)
	return node, []*AVLNode{}
}

func (tree *AVLTree) VerifyRangeProof(proof *AVLProof) bool {
	err := proof.Verify(tree.MerkleRoot(context.TODO()).Bytes())
	if err != nil {
		log.Error("[avltree:VerifyProof] proof of root is false", "info", err)
		return false
	}
	return true
}

func (tree *AVLTree) VerifyProof(key []byte, value []byte, proof *AVLProof) bool {

	// verify root
	err := proof.Verify(tree.MerkleRoot(context.TODO()).Bytes())
	if err != nil {
		log.Error("[avltree:VerifyProof] proof of root is false", "info", err)
		return false
	}
	// verify item
	err = proof.VerifyItem(key, value)
	if err != nil {
		log.Error("[avltree:VerifyProof] proof of item is false", "info", err, "key", hex.EncodeToString(key), "val", value)
		return false
	}
	return true
}

func (tree *AVLTree) Scan(keyStart, keyEnd []byte, limit int) (proof interface{}, keys, values [][]byte, err error) {
	p, keys, values, err := tree.getAVLProof(keyStart, keyEnd, limit)
	if err != nil {
		return proof, keys, values, fmt.Errorf("[avltree:GetScanProof] %s", err)
	}
	return p, keys, values, nil
}

// keyStart is inclusive and keyEnd is exclusive.
// If keyStart or keyEnd don't exist, the leaf before keyStart
// or after keyEnd will also be included, but not be included in values.
// If keyEnd-1 exists, no later leaves will be included.
// If keyStart >= keyEnd and both not nil, panics.
// Limit is never exceeded.
func (tree *AVLTree) getAVLProof(keyStart, keyEnd []byte, limit int) (proof *AVLProof, keys, values [][]byte, err error) {
	if keyStart != nil && keyEnd != nil && bytes.Compare(keyStart, keyEnd) >= 0 {
		panic("if keyStart and keyEnd are present, need keyStart < keyEnd.")
	}
	if limit < 0 {
		panic("limit must be greater or equal to 0 -- 0 means no limit")
	}
	if tree.root == nil {
		return nil, nil, nil, nil
	}
	rootnode := tree.GetRootNode()
	rootnode.hash, _, rootnode.StorageBytesTotal = rootnode.hashBranch(tree.ndb) // Ensure that all hashes are calculated.
	rootnode = tree.ndb.GetNode(rootnode.id, EMPTYBYTES)
	tree.SetRootNode(rootnode)
	//dprint("[getAVLProof] tree.root hash: (%x)", rootnode.hash)
	//tree.PrintTree(nil, false, false)
	// Get the first key/value pair proof, which provides us with the left key.
	path, left, err := rootnode.PathToLeaf(tree.ndb, keyStart)
	if err != nil { //TODO...not sure about this
		// Key doesn't exist, but instead we got the prev leaf (or the
		// first or last leaf), which provides proof of absence).
		err = nil
	}
	startOK := keyStart == nil || bytes.Compare(keyStart, left.Key) <= 0
	endOK := keyEnd == nil || bytes.Compare(left.Key, keyEnd) < 0
	// If left.key is in range, add it to key/values.
	if startOK && endOK {
		keys = append(keys, left.Key) // == keyStart
		values = append(values, left.Valhash)
	}
	// Either way, add to proof leaves.
	var leaves = []pLeaf{pLeaf{
		Height:       left.Height,
		Size:         left.Size,
		StorageBytes: left.StorageBytes,
		Key:          left.Key,
		ValHash:      left.Valhash,
	}}

	// 1: Special case if limit is 1.
	// 2: Special case if keyEnd is left.key+1.
	_stop := false
	if limit == 1 {
		_stop = true // case 1
	} else if keyEnd != nil && bytes.Compare(cpIncr(left.Key), keyEnd) >= 0 {
		_stop = true // case 2
		//dprint("[avltree:getAVLProof] case 2 - limit is not 1")
	}
	if _stop {
		return &AVLProof{
			LeftPath: path,
			Leaves:   leaves,
		}, keys, values, nil
	}

	// Get the key after left.key to iterate from.
	afterLeft := cpIncr(left.Key)

	// Traverse starting from afterLeft, until keyEnd or the next leaf
	// after keyEnd.
	var innersq = []PPathToLeaf(nil)
	var inners = PPathToLeaf(nil)
	var leafCount = 1 // from left above.
	var pathCount = 0
	// var keys, values [][]byte defined as function outs.
	//dprint("[avltree:getAVLProof] innersq(%+v) inners(%+v) leafCount(%d) pathCount(%d)", innersq, inners, leafCount, pathCount)

	rootnode.traverseInRange(tree.ndb, afterLeft, nil, true, false, 0,
		func(node *AVLNode, depth uint8) (stop bool) {

			//dprint("[avltree:getAVLProof] leafCount(%d) pathCount(%d)", leafCount, pathCount)
			//dprint("%s", node.String())
			// Track when we diverge from path, or when we've exhausted path,
			// since the first innersq shouldn't include it.
			if pathCount != -1 {
				if len(path) <= pathCount {
					// We're done with path counting.
					pathCount = -1
				} else {
					pn := path[pathCount]

					//dprint("pn.Height(%v) vs node.Height(%v)", pn.Height, node.Height)
					//dprint("pn.Left(%x) vs node.leftHash(%x)", pn.Left, node.leftHash)
					//dprint("pn.Right(%x) vs node.rightHash(%x)", pn.Right, node.rightHash)

					if pn.Height != node.Height ||
						pn.Left != nil && !bytes.Equal(pn.Left, node.leftHash) ||
						pn.Right != nil && !bytes.Equal(pn.Right, node.rightHash) {
						//dprint("diverged")

						// We've diverged, so start appending to inners.
						pathCount = -1
					} else {
						//dprint("pathcount++")
						pathCount += 1
					}
				}
			}

			if node.Height == 0 {
				// Leaf node.
				// Append inners to innersq.
				//dprint("[avltree:getAVLProof] leafnode k(%x) v(%x)", node.Key, node.Valhash)
				//dprint("[avltree:getAVLProof] inners(%+v)", inners)
				innersq = append(innersq, inners)
				//dprint("[avltree:getAVLProof] innersq(%+v)", innersq)
				inners = PPathToLeaf(nil)
				//dprint("[avltree:getAVLProof] inners(%+v)", inners)
				// Append leaf to leaves.
				leaves = append(leaves, pLeaf{
					Key:          node.Key,
					ValHash:      node.Valhash,
					Height:       node.Height,
					Size:         node.Size,
					StorageBytes: node.StorageBytes,
				})

				leafCount += 1
				// Maybe terminate because we found enough leaves.
				if limit > 0 && limit <= leafCount {
					return true
				}
				// Terminate if we've found keyEnd or after.
				if keyEnd != nil && bytes.Compare(node.Key, keyEnd) >= 0 {
					return true
				}
				// Value is in range, append to keys and values.
				keys = append(keys, node.Key)
				values = append(values, node.Valhash)
				// Terminate if we've found keyEnd-1 or after.
				// We don't want to fetch any leaves for it.
				if keyEnd != nil && bytes.Compare(cpIncr(node.Key), keyEnd) >= 0 {
					return true
				}
			} else {
				// Inner node.
				//dprint("inner node k(%x) v(%x) pathcount(%d)", node.Key, node.Valhash, pathCount)
				if pathCount >= 0 {
					//dprint("pathcount >= 0, skip redundant path items")
					// Skip redundant path items.
				} else {
					//dprint("appending as pin")
					inners = append(inners, pInner{
						Key:               node.Key, //keyadded
						Height:            node.Height,
						Size:              node.Size,
						StorageBytesTotal: node.StorageBytesTotal,
						Left:              nil, // left is nil for range proof inners
						Right:             node.rightHash,
					})
				}
			}
			return false
		},
	)

	return &AVLProof{
		LeftPath:   path,
		InnerNodes: innersq,
		Leaves:     leaves,
	}, keys, values, nil

}

func (tree *AVLTree) SetRootNode(node *AVLNode) {
	tree.Lock()
	defer tree.Unlock()
	tree.root.Node = node
}

func (tree *AVLTree) GetRootNode() (node *AVLNode) {
	tree.Lock()
	defer tree.Unlock()
	if tree.root.Node == nil {
		return nil
	}
	//dprint("[avltree:GetRootNode] tree.chunkHash(%x)", tree.chunkHash)
	return tree.root.Node.clone(true)
}

func (tree *AVLTree) GetRootChunk() (chunk *AVLChunk) {
	tree.Lock()
	defer tree.Unlock()
	chunk = NewAVLChunk(tree.root.Node)
	chunk.Children = tree.root.Children
	//chunk.MerkleRoot = tree.root.MerkleRoot
	return chunk
}

func (tree *AVLTree) SetRootChunkHash() {
	tree.Lock()
	defer tree.Unlock()
	tree.chunkHash = tree.root.Node.ChunkHash
}

func (tree *AVLTree) SetEmptyRootChunkHash() {
	tree.Lock()
	defer tree.Unlock()
	tree.chunkHash = common.BytesToHash(getEmptySMTChunkHash())
}

///////// helper functions

// PrintTree prints the whole tree. if chstore is false, it will not go out to cloud.ChunkStore to get unloaded chunks. if strict is true, it will error if it cannot find a chunk, otherwise it will load what it can (with a warning) and move on.
func (tree *AVLTree) PrintTree(root *AVLNode, strict bool, chstore bool) {
	if root == nil {
		root = tree.root.Node
	}
	fmt.Printf("\n")
	tree.printNode(root, 0, strict, chstore)
	fmt.Printf("\n")
}

func (tree *AVLTree) printNode(node *AVLNode, indent int, strict bool, chstore bool) {
	indentPrefix := ""
	for i := 0; i < indent; i++ {
		indentPrefix += "    "
	}
	if node == nil {
		fmt.Printf("%s<nil>\n", indentPrefix)
		return
	}
	if node.RightID != nil {
		//dprint("found RightID %v", *node.RightID)
		if chstore {
			tree.printNode(tree.ndb.GetNode(*node.RightID, node.ChunkHash), indent+1, strict, chstore)
		} else {
			rnode, f := tree.ndb.getNode(*node.RightID)
			if !f && !strict {
				fmt.Printf("%sh:x i(%x) WARNING: node.RightID not found in nodemap\n", indentPrefix+"    ", shorthash(*node.RightID))
				fmt.Printf("%sh:x i(%x) l(%x) r(%x) h(%v) s(%v) sbt(%v)\n", indentPrefix, shorthash(node.id), shorthash(*node.LeftID), shorthash(*node.RightID), node.Height, node.Size, node.StorageBytesTotal)
				//fmt.Printf("%sh:(%x) l(%x) r(%x)\n", indentPrefix, node.hash, node.leftHash, node.rightHash)
				return
			} else if !f && strict {
				fmt.Printf("%sh:x i(%x) l(%x) r(%x) h(%v) s(%v) sbt(%v)\n", indentPrefix, shorthash(node.id), shorthash(*node.LeftID), shorthash(*node.RightID), node.Height, node.Size, node.StorageBytesTotal)
				//fmt.Printf("%sh:(%x) l(%x) r(%x)\n", indentPrefix, node.hash, node.leftHash, node.rightHash)
				err := fmt.Errorf("node.RightID(%x) not found in nodemap", *node.RightID)
				panic(err)
			}
			tree.printNode(rnode, indent+1, strict, chstore)
		}
	}
	if node.isLeaf() {
		fmt.Printf("%sh:(%x) i(%x) k(%x) v(%x) h(%v) s(%v) sb(%v)\n", indentPrefix, node.hash, shorthash(node.id), shortbytes(node.Key), shortbytes(node.Valhash), node.Height, node.Size, node.StorageBytes)
		//fmt.Printf("%sh:(%x) l(%x) r(%x)\n", indentPrefix, shortbytes(node.hash), shortbytes(node.leftHash), shortbytes(node.rightHash))
	} else {
		fmt.Printf("%sh:(%x) i(%x) l(%x) r(%x) h(%v) s(%v) sbt(%v)\n", indentPrefix, node.hash, shorthash(node.id), shorthash(*node.LeftID), shorthash(*node.RightID), node.Height, node.Size, node.StorageBytesTotal)
		//fmt.Printf("%sh:(%x) l(%x) r(%x)\n", indentPrefix, shortbytes(node.hash), shortbytes(node.leftHash), shortbytes(node.rightHash))
	}
	// TODO: hashup
	// var hash []byte
	// if strict {
	// 	hash = node._hash()
	// } else {
	// 	if node.leftHash == nil || node.rightHash == nil {
	// 		// don't compute the hash
	// 	} else {
	// 		hash = node._hash()
	// 	}
	// }
	// if hash == nil {
	// 	hash = EMPTYBYTES.Bytes()
	// }
	//	fmt.Printf("%sh:%X (%v)(%v)\n", indentPrefix, hash, node.height, node.storageBytesTotal)fmt.Printf("%sh:%X k(%x) v(%x) h(%v) s(%v) sb(%v) sbt(%v)\n", indentPrefix, hash, node.key, node.valhash, node.height, node.size, node.storageBytes, node.storageBytesTotal)
	if node.LeftID != nil {
		if chstore {
			tree.printNode(tree.ndb.GetNode(*node.LeftID, node.ChunkHash), indent+1, strict, chstore)
		} else {
			lnode, f := tree.ndb.getNode(*node.LeftID)
			if !f && !strict {
				//fmt.Printf("%sh:x i(%x) l(%x) r(%x) h(%v) s(%v) sbt(%v)\n", indentPrefix, shorthash(node.id), shorthash(*node.LeftID), shorthash(*node.RightID), node.Height, node.Size, node.StorageBytesTotal)
				fmt.Printf("%sh:x i(%x) WARNING: node.LeftID not found in nodemap\n", indentPrefix+"    ", shorthash(*node.LeftID))
				return
			} else if !f && strict {
				//fmt.Printf("%sh:x i(%x) l(%x) r(%x) h(%v) s(%v) sbt(%v)\n", indentPrefix, shorthash(node.id), shorthash(*node.LeftID), shorthash(*node.RightID), node.Height, node.Size, node.StorageBytesTotal)
				err := fmt.Errorf("node.LeftID(%x) not found in nodemap", *node.LeftID)
				panic(err)
			}
			tree.printNode(lnode, indent+1, strict, chstore)
		}
	}
}

const frontBytes = true

func shorthash(hash common.Hash) (short []byte) {
	if len(hash) == 0 {
		hash = EMPTYBYTES
	}
	numbytes := 4
	if frontBytes {
		return hash.Bytes()[0:numbytes]
	}
	start := len(hash.Bytes()) - numbytes
	end := len(hash.Bytes())
	return hash.Bytes()[start:end]

}
func shortbytes(in []byte) (out []byte) {
	if len(in) == 0 {
		in = EMPTYBYTES.Bytes()
	}
	numbytes := 4
	if len(in) < 32 {
		in = padBytes(in, 32)
	}
	if frontBytes {
		return in[0:numbytes]
	}
	start := len(in) - numbytes
	end := len(in)
	return in[start:end]
}

func padBytes(a []byte, numBytes int) (padded_a []byte) {
	if len(a) >= numBytes {
		return a
	}
	padded_a = make([]byte, numBytes)
	if frontBytes {
		copy(padded_a[0:len(a)], a)
	} else {
		copy(padded_a[numBytes-len(a):numBytes], a)
	}
	return padded_a
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	dprint("%s took %s", name, elapsed)
}

func floatToByte(in float64) (out []byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, in)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func byteToFloat(in []byte) (out float64) {
	buf := bytes.NewReader(in)
	err := binary.Read(buf, binary.BigEndian, &out)
	if err != nil {
		panic(err)
	}
	return out
}

func intToByte(in uint64) (out []byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, in)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func byteToInt(in []byte) (out uint64) {
	buf := bytes.NewReader(in)
	err := binary.Read(buf, binary.BigEndian, &out)
	if err != nil {
		panic(err)
	}
	return out
}

func int64ToByte(in int64) (out []byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, in)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func byteToInt64(in []byte) (out int64) {
	buf := bytes.NewReader(in)
	err := binary.Read(buf, binary.BigEndian, &out)
	if err != nil {
		panic(err)
	}
	return out
}
