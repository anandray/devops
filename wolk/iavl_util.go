package wolk

import (
	"bytes"
	"fmt"
	"sort"
	//wolkcommon "github.com/wolkdb/cloudstore/common"
)

// PrintTree prints the whole tree in an indented form.
// WARNING: this uses 'gets' from storage
func PrintTree(tree *ImmutableTree, strict bool) {
	ndb, root := tree.ndb, tree.root
	printNode(ndb, root, 0, strict)
}

// PrintNode prints a node, indented. strict is whether tree is required to be perfectly hashed.
// WARNING: this uses 'gets' from storage
func printNode(ndb *nodeDB, node *Node, indent int, strict bool) {
	indentPrefix := ""
	for i := 0; i < indent; i++ {
		indentPrefix += "    "
	}

	if node == nil {
		fmt.Printf("%s<nil>\n", indentPrefix)
		return
	}
	if node.rightNode != nil {
		printNode(ndb, node.rightNode, indent+1, strict)
	} else if node.rightHash != nil {
		rightNode := ndb.GetNode(node.rightHash)
		printNode(ndb, rightNode, indent+1, strict)
	}
	var hash []byte
	if strict {
		hash = node._hash()
	} else {
		if node.leftHash == nil || node.rightHash == nil {
			// don't compute the hash
		} else {
			hash = node._hash()
		}
	}
	if hash == nil {
		hash = EMPTYBYTES.Bytes()
	}
	//fmt.Printf("%sh:%X (%v)(%v)\n", indentPrefix, hash, node.height, node.storageBytesTotal)
	//if node.isLeaf() {
	//fmt.Printf("%s%X:%X (%v)\n", indentPrefix, node.key, node.valhash, node.height)
	//}
	//fmt.Printf("%sh:%X k(%x) v(%x) h(%v) s(%v) sb(%v) sbt(%v)\n", indentPrefix, shortbytes(hash), shortbytes(node.key), shortbytes(node.valhash), node.height, node.size, node.storageBytes, node.storageBytesTotal)
	fmt.Printf("%sh:%X k(%x) v(%x) h(%v) s(%v) sb(%v) sbt(%v)\n", indentPrefix, hash, node.key, node.valhash, node.height, node.size, node.storageBytes, node.storageBytesTotal)

	if node.leftNode != nil {
		printNode(ndb, node.leftNode, indent+1, strict)
	} else if node.leftHash != nil {
		leftNode := ndb.GetNode(node.leftHash)
		printNode(ndb, leftNode, indent+1, strict)
	}

}

// printtree prints a tree from the root, whatever is availble without using 'gets' from storage.
func (tree *ImmutableTree) printtree() {
	dprint("---tree---")
	root := tree.root
	root.printnoderecursive()
	dprint("---endtree---")
}

// printnode prints the Node struct parameters
func (node *Node) printnode() {
	dprint("---node---")

	// debug
	// nodeid := node.key
	// if !node.isLeaf() {
	// 	nodeid = wolkcommon.Computehash(node.key)
	// }
	// var lnodeid, rnodeid []byte
	// if node.leftNode != nil {
	// 	lnodeid = node.leftNode.key
	// 	if !node.leftNode.isLeaf() {
	// 		lnodeid = wolkcommon.Computehash(node.leftNode.key)
	// 	}
	// }
	// if node.rightNode != nil {
	// 	rnodeid = node.rightNode.key
	// 	if !node.rightNode.isLeaf() {
	// 		rnodeid = wolkcommon.Computehash(node.rightNode.key)
	// 	}
	// }
	// dprint("nodeID(%x)", shortbytes(nodeid))
	// dprint("leftNodeID(%x)", shortbytes(lnodeid))
	// dprint("rightNodeID(%x)", shortbytes(rnodeid))
	// end debug

	dprint("hash(%x)", node.hash)
	dprint("key(%x)", node.key)
	dprint("valhash(%x)", node.valhash)

	//dprint("version(%d)", node.version)
	dprint("height(%d)", node.height)
	dprint("size(%d)", node.size)
	dprint("size(%d)", node.storageBytes)
	dprint("size(%d)", node.storageBytesTotal)

	dprint("lefthash(%x)", node.leftHash)
	dprint("righthash(%x)", node.rightHash)
	if node.persisted {
		dprint("persisted(true)")
	} else {
		dprint("persisted(false)")
	}
	dprint("---end node---")
}

// printnoderecursive is a helper for printtree
func (node *Node) printnoderecursive() {
	if node == nil {
		return
	}
	dprint("---node---")
	dprint("hash(%x)", node.hash)
	dprint("key(%x)", node.key)
	dprint("valhash(%x)", node.valhash)

	//dprint("version(%d)", node.version)
	dprint("height(%d)", node.height)
	dprint("size(%d)", node.size)
	dprint("size(%d)", node.storageBytes)
	dprint("size(%d)", node.storageBytesTotal)

	dprint("lefthash(%x)", node.leftHash)
	if node.leftNode != nil {
		dprint("leftNode(%x)")
		node.leftNode.printnoderecursive()
	}
	dprint("righthash(%x)", node.rightHash)
	if node.rightNode != nil {
		dprint("rightNode(%x)")
		node.rightNode.printnoderecursive()
	}
	if node.persisted {
		dprint("persisted(true)")
	} else {
		dprint("persisted(false)")
	}
	dprint("---end node---")
}

func maxInt8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}

func cp(bz []byte) (ret []byte) {
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}

// Returns a slice of the same length (big endian)
// except incremented by one.
// Appends 0x00 if bz is all 0xFF.
// CONTRACT: len(bz) > 0
func cpIncr(bz []byte) (ret []byte) {
	ret = cp(bz)
	for i := len(bz) - 1; i >= 0; i-- {
		if ret[i] < byte(0xFF) {
			ret[i]++
			return
		}
		ret[i] = byte(0x00)
		if i == 0 {
			return append(ret, 0x00)
		}
	}
	return []byte{0x00}
}

type byteslices [][]byte

func (bz byteslices) Len() int {
	return len(bz)
}

func (bz byteslices) Less(i, j int) bool {
	switch bytes.Compare(bz[i], bz[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("should not happen")
	}
}

func (bz byteslices) Swap(i, j int) {
	bz[j], bz[i] = bz[i], bz[j]
}

func sortByteSlices(src [][]byte) [][]byte {
	bzz := byteslices(src)
	sort.Sort(bzz)
	return bzz
}

// for debugging non-test code
func dprint(in string, args ...interface{}) {
	if in == "\n" {
		fmt.Println()
	} else {
		fmt.Printf("[debug] "+in+"\n", args...)
	}
}
