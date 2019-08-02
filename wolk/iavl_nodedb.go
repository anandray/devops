package wolk

import (
	"bytes"
	"container/list"
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
	//wolkcommon "github.com/wolkdb/cloudstore/common"
)

// const (
// 	int64Size = 8
// 	hashSize  = tmhash.Size
// )

type nodeDB struct {
	mtx sync.Mutex // Read/write lock.
	wdb ChunkStore

	nodeCache      map[string]*list.Element // Node cache.
	nodeCacheSize  int                      // Node cache size limit in elements.
	nodeCacheQueue *list.List               // LRU queue of cache elements. Used for deletion.
}

func newNodeDB(cs ChunkStore, cacheSize int) *nodeDB {
	ndb := &nodeDB{
		wdb:            cs,
		nodeCache:      make(map[string]*list.Element),
		nodeCacheSize:  cacheSize,
		nodeCacheQueue: list.New(),
	}
	return ndb
}

// GetNode gets a node from cache or disk.
func (ndb *nodeDB) GetNode(hash []byte) *Node {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	// dprint("[iavl_nodedb:GetNode] in GetNode with (%x)", hash)

	if len(hash) == 0 {
		panic("nodeDB.GetNode() requires hash")
	}

	// Check the cache.
	if elem, ok := ndb.nodeCache[string(hash)]; ok {
		// dprint("[iavl_nodedb:GetNode] Node(%x) is found in the cache.", hash)
		// Already exists. Move to back of nodeCacheQueue.
		ndb.nodeCacheQueue.MoveToBack(elem)
		return elem.Value.(*Node)
	}

	// Doesn't exist, load.
	//dprint("[iavl_nodedb:GetNode] GetChunk(%x)", hash)
	buf, ok, err := ndb.wdb.GetChunk(context.TODO(), hash)
	if err != nil {
		log.Error("[iavl_nodedb:GetNode] wdb", "err", err)
	} else if !ok {
		err = fmt.Errorf("[iavl_nodedb:GetNode] wdb NOT FOUND")
		log.Error("[iavl_nodedb:GetNode] wdb NOT FOUND", "key", hex.EncodeToString(hash))
		panic(err)

	}
	// dprint("[iavl_nodedb:GetNode] wdb val(%x), cmp to iavl val(%x)", buf, iavlbuf)

	node, err := MakeNode(buf)
	if err != nil {
		panic(fmt.Sprintf("Error reading Node. bytes: %x, error: %v", buf, err))
	}
	// dprint("[iavl_nodedb:GetNode] after making node(%+v)", node)
	node.hash = hash
	node.persisted = true
	//ndb.cacheNode(node)  //TODO: turn on cache

	return node
}

// SaveNode saves a node to disk.
func (ndb *nodeDB) SaveNode(node *Node, wg *sync.WaitGroup, writeToCloudstore bool) {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()

	// dprint("\n")
	if node.hash == nil {
		panic("Expected to find node.hash, but none found.")
	}
	if node.persisted {
		panic("Shouldn't be calling save on an already persisted node.")
	}
	//	node.nodekey = ndb.nodeKey(node.hash)

	// Save node bytes to db.
	buf := new(bytes.Buffer)
	//	if err := node.writeBytes(buf); err != nil {
	if err := node.writeHashBytes(buf); err != nil {
		panic(err)
	}

	if writeToCloudstore {
		wg.Add(1)
		chunk := new(cloud.RawChunk)
		chunk.Value = buf.Bytes()
		// dprint("[iavl_nodedb:SaveNode] do the chunk IDs match?\n (%x)\n (%x)", wolkcommon.Computehash(chunk.Value), node.hash)
		_, err := ndb.wdb.PutChunk(context.TODO(), chunk, wg)
		if err != nil {
			log.Error("[iavl_nodedb:SaveNode]", "err", err)
			return
		}
	}
	dprint("[iavl_nodedb:SaveNode] putchunk id(%x)", node.hash)
	// dprint("[iavl_nodedb:SaveNode] Nodehash(%x),\n  buf.Bytes(%x)\n  chunkID(%x)\n", node.hash, buf.Bytes(), chunkID)

	node.persisted = true
	//ndb.cacheNode(node) // TODO: turn on cache
}

// Has checks if a hash exists in the database.
// func (ndb *nodeDB) Has(hash []byte) bool {
// 	key := ndb.nodeKey(hash)
//
// 	if ldb, ok := ndb.db.(*dbm.GoLevelDB); ok {
// 		exists, err := ldb.DB().Has(key, nil)
// 		if err != nil {
// 			panic("Got error from leveldb: " + err.Error())
// 		}
// 		return exists
// 	}
// 	return ndb.db.Get(key) != nil
// }

// SaveBranch saves the given node and all of its descendants.
// NOTE: This function clears leftNode/rigthNode recursively and
// calls _hash() on the given node.
// TODO refactor, maybe use hashWithCount() but provide a callback.
func (ndb *nodeDB) SaveBranch(node *Node, wg *sync.WaitGroup, writeToCloudstore bool) (hash []byte, storageBytesTotal uint64) {
	if node.persisted {
		return node.hash, node.storageBytesTotal
	}

	var leftStorageBytes uint64
	var rightStorageBytes uint64
	if node.leftNode != nil {
		node.leftHash, leftStorageBytes = ndb.SaveBranch(node.leftNode, wg, writeToCloudstore)
	}
	if node.rightNode != nil {
		node.rightHash, rightStorageBytes = ndb.SaveBranch(node.rightNode, wg, writeToCloudstore)
	}

	// calculate storage bytes
	if !node.isLeaf() {
		node.storageBytes = 0 // TODO: storageBytes for inner nodes
	}
	node.storageBytesTotal = leftStorageBytes + rightStorageBytes + node.storageBytes

	node._hash()
	// dprint("[iavl_nodedb:SaveBranch] node's hash (%x)", node.hash)
	ndb.SaveNode(node, wg, writeToCloudstore)

	node.leftNode = nil
	node.rightNode = nil

	return node.hash, node.storageBytesTotal
}

// Traverse all keys.
// func (ndb *nodeDB) traverse(fn func(key, value []byte)) {
// 	itr := ndb.db.Iterator(nil, nil)
// 	defer itr.Close()
//
// 	for ; itr.Valid(); itr.Next() {
// 		fn(itr.Key(), itr.Value())
// 	}
// }

// Traverse all keys with a certain prefix.
// func (ndb *nodeDB) traversePrefix(prefix []byte, fn func(k, v []byte)) {
// 	itr := dbm.IteratePrefix(ndb.db, prefix)
// 	defer itr.Close()
//
// 	for ; itr.Valid(); itr.Next() {
// 		fn(itr.Key(), itr.Value())
// 	}
// }

func (ndb *nodeDB) uncacheNode(hash []byte) {
	if elem, ok := ndb.nodeCache[string(hash)]; ok {
		ndb.nodeCacheQueue.Remove(elem)
		delete(ndb.nodeCache, string(hash))
	}
}

// Add a node to the cache and pop the least recently used node if we've
// reached the cache size limit.
func (ndb *nodeDB) cacheNode(node *Node) {
	elem := ndb.nodeCacheQueue.PushBack(node)
	ndb.nodeCache[string(node.hash)] = elem

	if ndb.nodeCacheQueue.Len() > ndb.nodeCacheSize {
		oldest := ndb.nodeCacheQueue.Front()
		hash := ndb.nodeCacheQueue.Remove(oldest).(*Node).hash
		delete(ndb.nodeCache, string(hash))
	}
}

////////////////// Utility and test functions /////////////////////////////////

// func (ndb *nodeDB) leafNodes() []*Node {
// 	leaves := []*Node{}
//
// 	ndb.traverseNodes(func(hash []byte, node *Node) {
// 		if node.isLeaf() {
// 			leaves = append(leaves, node)
// 		}
// 	})
// 	return leaves
// }
//
// func (ndb *nodeDB) nodes() []*Node {
// 	nodes := []*Node{}
//
// 	ndb.traverseNodes(func(hash []byte, node *Node) {
// 		nodes = append(nodes, node)
// 	})
// 	return nodes
// }

// func (ndb *nodeDB) orphans() [][]byte {
// 	orphans := [][]byte{}
//
// 	ndb.traverseOrphans(func(k, v []byte) {
// 		orphans = append(orphans, v)
// 	})
// 	return orphans
// }

// func (ndb *nodeDB) roots() map[int64][]byte {
// 	roots, _ := ndb.getRoots()
// 	return roots
// }

// Not efficient.
// NOTE: DB cannot implement Size() because
// mutations are not always synchronous.
// func (ndb *nodeDB) size() int {
// 	size := 0
// 	ndb.traverse(func(k, v []byte) {
// 		size++
// 	})
// 	return size
// }

// func (ndb *nodeDB) traverseNodes(fn func(hash []byte, node *Node)) {
// 	nodes := []*Node{}
//
// 	ndb.traversePrefix(nodeKeyFormat.Key(), func(key, value []byte) {
// 		node, err := MakeNode(value)
// 		if err != nil {
// 			panic(fmt.Sprintf("Couldn't decode node from database: %v", err))
// 		}
// 		nodeKeyFormat.Scan(key, &node.hash)
// 		nodes = append(nodes, node)
// 	})
//
// 	sort.Slice(nodes, func(i, j int) bool {
// 		return bytes.Compare(nodes[i].key, nodes[j].key) < 0
// 	})
//
// 	for _, n := range nodes {
// 		fn(n.hash, n)
// 	}
// }

// func (ndb *nodeDB) String() string {
// 	var str string
// 	index := 0
//
// 	ndb.traversePrefix(rootKeyFormat.Key(), func(key, value []byte) {
// 		str += fmt.Sprintf("%s: %x\n", string(key), value)
// 	})
// 	str += "\n"
//
// 	ndb.traverseOrphans(func(key, value []byte) {
// 		str += fmt.Sprintf("%s: %x\n", string(key), value)
// 	})
// 	str += "\n"
//
// 	ndb.traverseNodes(func(hash []byte, node *Node) {
// 		if len(hash) == 0 {
// 			str += fmt.Sprintf("<nil>\n")
// 		} else if node == nil {
// 			str += fmt.Sprintf("%s%40x: <nil>\n", nodeKeyFormat.Prefix(), hash)
// 		} else if node.valhash == nil && node.height > 0 {
// 			str += fmt.Sprintf("%s%40x: %s   %-16s h=%d version=%d\n",
// 				nodeKeyFormat.Prefix(), hash, node.key, "", node.height, node.version)
// 		} else {
// 			str += fmt.Sprintf("%s%40x: %s = %-16s h=%d version=%d\n",
// 				nodeKeyFormat.Prefix(), hash, node.key, node.valhash, node.height, node.version)
// 		}
// 		index++
// 	})
// 	return "-" + "\n" + str + "-"
// }
