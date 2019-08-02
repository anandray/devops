package wolk

import (
	"bytes"
	"fmt"
	"sync"
)

// MutableTree is a persistent tree which keeps track of versions.
type MutableTree struct {
	*ImmutableTree                  // The current, working tree.
	lastSaved      *ImmutableTree   // The most recently saved tree.
	orphans        map[string]int64 // Nodes removed by changes to working tree.
	ndb            *nodeDB
}

// NewMutableTree returns a new tree with the specified cache size and datastore.
func NewMutableTree(cs ChunkStore, cacheSize int) *MutableTree {

	ndb := newNodeDB(cs, cacheSize)
	head := &ImmutableTree{ndb: ndb}

	return &MutableTree{
		ImmutableTree: head,
		lastSaved:     head.clone(),
		orphans:       map[string]int64{},
		//versions:      map[int64]bool{},
		ndb: ndb,
	}
}

// LoadMutableTree returns a tree loaded with a previously saved roothash
func LoadMutableTree(cs ChunkStore, roothash []byte, cacheSize int) *MutableTree {
	ndb := newNodeDB(cs, cacheSize)
	head := &ImmutableTree{ndb: ndb}
	head.root = ndb.GetNode(roothash)
	//head.root.printnode()
	//head.version = head.root.version

	return &MutableTree{
		ImmutableTree: head,
		lastSaved:     head.clone(),
		orphans:       map[string]int64{},
		//versions:      map[int64]bool{},
		ndb: ndb,
	}
}

// IsEmpty returns whether or not the tree has any keys. Only trees that are
// not empty can be saved.
func (tree *MutableTree) IsEmpty() bool {
	return tree.ImmutableTree.Size() == 0
}

// Hash returns the hash of the latest saved version of the tree, as returned
// by SaveVersion. If no versions have been saved, Hash returns nil.
func (tree *MutableTree) Hash() []byte {
	if tree.lastSaved.root != nil {
		return tree.lastSaved.Hash()
	}
	return nil
}

// WorkingHash returns the hash of the current working tree. Note that this recomputes the root hash.
func (tree *MutableTree) WorkingHash() []byte {
	return tree.ImmutableTree.Hash()
}

// String returns a string representation of the tree.
// func (tree *MutableTree) String() string {
// 	return tree.ndb.String()
// }

// Set sets a key in the working tree. Nil values are not supported.
func (tree *MutableTree) Set(key, value []byte, storageBytesNew uint64) bool {
	orphaned, updated := tree.set(key, value, storageBytesNew)
	tree.addOrphans(orphaned)
	return updated
}

func (tree *MutableTree) set(key []byte, value []byte, storageBytesNew uint64) (orphaned []*Node, updated bool) {
	//dprint("[mutable_tree:set] k(%x), v(%x)", key, value)
	if value == nil {
		panic(fmt.Sprintf("Attempt to store nil value at key '%s'", key))
	}
	if tree.ImmutableTree.root == nil {
		tree.ImmutableTree.root = NewNode(key, value, storageBytesNew)
		return nil, false
	}
	tree.ImmutableTree.root, updated, orphaned = tree.recursiveSet(tree.ImmutableTree.root, key, value, storageBytesNew)
	//dprint("[mutable_tree:set] inserted (%x, %x)", key, value)
	//PrintTree(tree.ImmutableTree, false)
	//dprint("\n")
	return orphaned, updated
}

func (tree *MutableTree) recursiveSet(node *Node, key []byte, value []byte, storageBytesNew uint64) (newSelf *Node, updated bool, orphaned []*Node) {

	//dprint("[recursiveSet] **** on node.key(%x), inserting key(%x) value(%x) sb(%v)", shortbytes(node.key), shortbytes(key), shortbytes(value), storageBytesNew)
	if node.isLeaf() {
		switch bytes.Compare(key, node.key) {
		case -1:
			return &Node{
				key:          node.key,
				valhash:      node.valhash,
				height:       1,
				size:         2,
				storageBytes: node.storageBytes,
				leftNode:     NewNode(key, value, storageBytesNew),
				rightNode:    node,
				//version:      version,
			}, false, []*Node{}
		case 1:
			return &Node{
				key:          key,
				valhash:      value,
				height:       1,
				size:         2,
				storageBytes: storageBytesNew,
				leftNode:     node,
				rightNode:    NewNode(key, value, storageBytesNew),
				//version:      version,
			}, false, []*Node{}
		default:
			return NewNode(key, value, storageBytesNew), true, []*Node{node}
		}
	} else {
		orphaned = append(orphaned, node)
		node = node.clone()

		if bytes.Compare(key, node.key) < 0 {
			var leftOrphaned []*Node
			node.leftNode, updated, leftOrphaned = tree.recursiveSet(node.getLeftNode(tree.ImmutableTree), key, value, storageBytesNew)
			node.leftHash = nil // leftHash is yet unknown
			orphaned = append(orphaned, leftOrphaned...)
		} else {
			var rightOrphaned []*Node
			node.rightNode, updated, rightOrphaned = tree.recursiveSet(node.getRightNode(tree.ImmutableTree), key, value, storageBytesNew)
			node.rightHash = nil // rightHash is yet unknown
			orphaned = append(orphaned, rightOrphaned...)
		}

		if updated {
			return node, updated, orphaned
		}
		node.calcHeightAndSize(tree.ImmutableTree)

		// debugging
		//dprint("[recursiveSet] before balanced:")
		//PrintTree(tree.ImmutableTree, false)
		//dprint("[recursiveSet] before balanced: node")
		//node.printnode()
		newNode, balanceOrphaned := tree.balance(node)
		//dprint("[recursiveSet] balanced")
		//PrintTree(tree.ImmutableTree, false)
		//dprint("[recursiveSet] after balanced: newSelf")
		//newNode.printnode()
		// end debugging

		return newNode, updated, append(orphaned, balanceOrphaned...)
	}
}

// Remove removes a key from the working tree. TODO
func (tree *MutableTree) Remove(key []byte) (val []byte, removed bool) {
	// 	val, orphaned, removed := tree.remove(key)
	// 	tree.addOrphans(orphaned)
	return val, removed
}

// GetImmutable loads an ImmutableTree for querying
func (tree *MutableTree) GetImmutable() (*ImmutableTree, error) {
	rootHash := tree.root.hash
	if rootHash == nil {
		return nil, fmt.Errorf("[mutable_tree:GetImmutable] roothash doesn't exist")
	} else if len(rootHash) == 0 {
		return &ImmutableTree{
			ndb: tree.ndb,
		}, nil
	}
	return &ImmutableTree{
		root: tree.ndb.GetNode(rootHash),
		ndb:  tree.ndb,
	}, nil
}

// Rollback resets the working tree to the latest saved version, discarding
// any unsaved modifications.
func (tree *MutableTree) Rollback() {
	if tree.lastSaved.root == nil {
		tree.ImmutableTree = &ImmutableTree{ndb: tree.ndb} // no lastsaved
	} else {
		tree.ImmutableTree = tree.lastSaved.clone()
	}
	tree.orphans = map[string]int64{}
}

// SaveTree saves a new tree version to disk, based on the current state of
// the tree. Returns the hash.
func (tree *MutableTree) SaveTree(wg *sync.WaitGroup, writeToCloudstore bool) ([]byte, error) {

	if tree.lastSaved.root != nil { // TODO: negative test this
		newHash := tree.WorkingHash()
		oldHash := tree.lastSaved.root.hash
		if bytes.Equal(oldHash, newHash) {
			dprint("[mutable_tree:SaveTree] no changes")
			return oldHash, nil
		}
	}

	if tree.root == nil {
		dprint("[mutable_tree:SaveVersion] SAVE EMPTY TREE ??\n")
		return nil, nil
	} else { // Save the current tree.
		//dprint("[iavl_mutable_tree:SaveVersion] SAVE TREE %v\n", version)
		//dprint("tree root node:", tree.root)
		//tree.root.printnode()

		tree.ndb.SaveBranch(tree.root, wg, writeToCloudstore)
		//tree.ndb.SaveOrphans(version, tree.orphans) // TODO
		//dprint("[iavl_mutable_tree:SaveVersion] root(%x) calling SaveRoot..", tree.root)
		//tree.ndb.SaveRoot(tree.root, version)
	}

	// Set new working tree.
	tree.ImmutableTree = tree.ImmutableTree.clone()
	tree.lastSaved = tree.ImmutableTree.clone()
	tree.orphans = map[string]int64{}
	return tree.root.hash, nil
}

// Rotate right and return the new node and orphan.
func (tree *MutableTree) rotateRight(node *Node) (*Node, *Node) {
	//version := tree.version + 1

	// TODO: optimize balance & rotate.
	node = node.clone()
	orphaned := node.getLeftNode(tree.ImmutableTree)
	newNode := orphaned.clone()

	newNoderHash, newNoderCached := newNode.rightHash, newNode.rightNode
	newNode.rightHash, newNode.rightNode = node.hash, node
	node.leftHash, node.leftNode = newNoderHash, newNoderCached

	node.calcHeightAndSize(tree.ImmutableTree)
	newNode.calcHeightAndSize(tree.ImmutableTree)

	return newNode, orphaned
}

// Rotate left and return the new node and orphan.
func (tree *MutableTree) rotateLeft(node *Node) (*Node, *Node) {
	//version := tree.version + 1

	// TODO: optimize balance & rotate.
	node = node.clone()
	orphaned := node.getRightNode(tree.ImmutableTree)
	newNode := orphaned.clone()
	//dprint("[rotateLeft] newSelf before:")
	newNode.printnode()

	newNodelHash, newNodelCached := newNode.leftHash, newNode.leftNode
	newNode.leftHash, newNode.leftNode = node.hash, node
	node.rightHash, node.rightNode = newNodelHash, newNodelCached

	//dprint("[rotateLeft] newSelf after")
	//newNode.printnode()
	//dprint("[rotateLeft] node after")
	//node.printnode()

	node.calcHeightAndSize(tree.ImmutableTree)
	newNode.calcHeightAndSize(tree.ImmutableTree)

	return newNode, orphaned
}

// NOTE: assumes that node can be modified
// TODO: optimize balance & rotate
func (tree *MutableTree) balance(node *Node) (newSelf *Node, orphaned []*Node) {
	if node.persisted {
		panic("Unexpected balance() call on persisted node")
	}
	balance := node.calcBalance(tree.ImmutableTree)

	if balance > 1 {
		if node.getLeftNode(tree.ImmutableTree).calcBalance(tree.ImmutableTree) >= 0 {
			// Left Left Case
			newNode, orphaned := tree.rotateRight(node)
			return newNode, []*Node{orphaned}
		}
		// Left Right Case
		var leftOrphaned *Node

		left := node.getLeftNode(tree.ImmutableTree)
		node.leftHash = nil
		node.leftNode, leftOrphaned = tree.rotateLeft(left)
		newNode, rightOrphaned := tree.rotateRight(node)

		return newNode, []*Node{left, leftOrphaned, rightOrphaned}
	}
	if balance < -1 {
		if node.getRightNode(tree.ImmutableTree).calcBalance(tree.ImmutableTree) <= 0 {
			// Right Right Case
			newNode, orphaned := tree.rotateLeft(node)
			return newNode, []*Node{orphaned}
		}
		// Right Left Case
		//dprint("[balance] right left case")
		var rightOrphaned *Node
		right := node.getRightNode(tree.ImmutableTree)
		//dprint("[balance] node's right node")
		//right.printnode()
		node.rightHash = nil
		node.rightNode, rightOrphaned = tree.rotateRight(right)
		//dprint("[balance] node's new right node, after rotating right")
		//node.rightNode.printnode()
		//dprint("node is now")
		//node.printnode()
		newNode, leftOrphaned := tree.rotateLeft(node)
		//dprint("[balance] node's newNode, after rotating left")
		//newNode.printnode()

		return newNode, []*Node{right, leftOrphaned, rightOrphaned}
	}
	// Nothing changed
	return node, []*Node{}
}

func (tree *MutableTree) addOrphans(orphans []*Node) {
	for _, node := range orphans {
		if !node.persisted {
			// We don't need to orphan nodes that were never persisted.
			continue
		}
		if len(node.hash) == 0 {
			panic("Expected to find node hash, but was empty")
		}
		//tree.orphans[string(node.hash)] = node.version
		tree.orphans[string(node.hash)] = 1
	}
}
