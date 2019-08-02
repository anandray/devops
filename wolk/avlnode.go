package wolk

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/tendermint/tendermint/libs/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

const (
	maxSize = 256 // max num of children nodes per chunk. should be > 2.
)

// AVLChunk is a chunk stored in chunkstore.
type AVLChunk struct {
	Node     *AVLNode   `json:"n"`
	Children []*AVLNode `json:"c"` // queue of child nodes, lowest ids are leaf, higher ids are closer to the current node, limited by maxSize. To preserve order.
	//MerkleRoot common.Hash `json:"h"`

	// for bookkeeping
	sync.RWMutex
}

func NewAVLChunk(node *AVLNode) *AVLChunk {
	chunk := new(AVLChunk)
	node.Lock()
	chunk.Node = node
	node.Unlock()
	chunk.Children = make([]*AVLNode, 0)
	return chunk
}

func (chunk *AVLChunk) clone() *AVLChunk {
	chunk.Lock()
	defer chunk.Unlock()
	newChunk := NewAVLChunk(chunk.Node)
	newChunk.Children = chunk.Children
	//newChunk.MerkleRoot = chunk.MerkleRoot
	return newChunk
}

func (chunk *AVLChunk) String() string {
	printstring := "chunk:\n"
	printstring += "  node:" + chunk.Node.String() + "\n"
	printstring += "  children:\n"
	for _, child := range chunk.Children {
		id := makeID(child.isLeaf(), child.Key)
		printstring += "  (" + hex.EncodeToString(id.Bytes()) + ") " + child.String() + "\n"
	}
	return printstring
}

// AVLNode represents a node in an AVLTree
type AVLNode struct {
	Key               []byte       `json:"k"`
	Valhash           []byte       `json:"v"`
	Height            int8         `json:"h"`
	Size              int64        `json:"s"`
	StorageBytes      uint64       `json:"sb"`  //storageBytes for this node only
	StorageBytesTotal uint64       `json:"sbt"` //storageBytes for all nodes "below" in the tree
	LeftID            *common.Hash `json:"lid"` //for indexing children
	RightID           *common.Hash `json:"rid"` //for indexing children
	ChunkHash         common.Hash  `json:"ch"`  //to get this node's chunk, if needed

	// for bookkeeping
	sync.RWMutex
	hash      []byte      // hash of node
	leftHash  []byte      // hash of left child
	rightHash []byte      // hash of right child
	id        common.Hash // index of the node
	persisted bool        // true = loaded from chunkstore and unchanged.
}

// NewAVLNode returns a new node from a key, value.
func NewAVLNode(key []byte, valhash []byte, storageBytes uint64) *AVLNode {
	node := &AVLNode{
		Key:          key,
		Valhash:      valhash,
		Height:       0,
		Size:         1,
		StorageBytes: storageBytes,
		persisted:    false,
	}
	node.id = makeID(true, key)
	//dprint("[NewAVLNode] id(%x), key(%x) valhash(%x) storageBytes(%v)", node.id, shortbytes(key), shortbytes(valhash), storageBytes)
	return node
}

func (node *AVLNode) Bytes() []byte {
	bytes, err := json.Marshal(node)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (node *AVLNode) Hash() common.Hash {
	return common.BytesToHash(wolkcommon.Computehash(node.Bytes()))
}

func makeID(isLeaf bool, key []byte) common.Hash {
	if isLeaf {
		k, err := hex.DecodeString(fmt.Sprintf("%x", key))
		if err != nil {
			panic(err)
		}
		return common.BytesToHash(k)
	} else {
		return common.BytesToHash(wolkcommon.Computehash(key))
	}
}

func (node *AVLNode) isLeaf() bool {
	return node.Height == 0
}

// get recursively looks up a key and returns a value. If the node needed is not in memory, get will go out to chunkstore to load it up.
func (node *AVLNode) get(ndb *avlNodeDB, key []byte) (index int64, value []byte, storageBytes uint64) {
	//node.Lock()
	//defer node.Unlock()
	node = node.clone(true)
	if node.isLeaf() {
		switch bytes.Compare(node.Key, key) {
		case -1: // node.Key < key
			//err := fmt.Errorf("[get**] node(%x) is leaf! and node.Key(%x) < key(%x)", shorthash(node.id), shortbytes(node.Key), shortbytes(key))
			//panic(err)
			return 1, nil, uint64(0)
		case 1: //node.Key > key
			//dprint("[get**] node(%x) is leaf! and node.Key > key", shorthash(node.id))
			return 0, nil, uint64(0)
		default: // node.Key == key
			//dprint("[get] node(%x) is leaf! and node.Key == key", node.id)
			return 0, node.Valhash, node.StorageBytes
		}
	}
	// inner node
	if bytes.Compare(key, node.Key) < 0 {
		return ndb.GetNode(*node.LeftID, node.ChunkHash).get(ndb, key)
		//return node.getLeftNode(ndb).get(ndb, key)
	}
	//rightNode := node.getRightNode(ndb)
	rightNode := ndb.GetNode(*node.RightID, node.ChunkHash)
	index, value, storageBytes = rightNode.get(ndb, key)
	index += node.Size - rightNode.Size
	return index, value, storageBytes
}

func (node *AVLNode) getLeftNode(ndb *avlNodeDB) (leftNode *AVLNode) {
	if node.LeftID == nil {
		panic("getLeftNode why is node.LeftID nil?")
	}
	return ndb.GetNode(*node.LeftID, node.ChunkHash)
}

func (node *AVLNode) getRightNode(ndb *avlNodeDB) (rightNode *AVLNode) {
	if node.RightID == nil {
		panic("[getRightNode] why is node.RightID nil?")
	}
	return ndb.GetNode(*node.RightID, node.ChunkHash)
}

func (node *AVLNode) calcHeightAndSize(ndb *avlNodeDB) (height int8, size int64) {
	lnode := node.getLeftNode(ndb)
	rnode := node.getRightNode(ndb)
	height = maxInt8(lnode.Height, rnode.Height) + 1
	size = lnode.Size + rnode.Size
	return height, size
}

func (node *AVLNode) calcBalance(ndb *avlNodeDB) int {
	return int(node.getLeftNode(ndb).Height) - int(node.getRightNode(ndb).Height)
}

/*// Writes the node's hash to the given io.Writer.
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
}*/

// hashBranch hashes up the tree for the merkle roots
// note- this does not change the tree. so persisted does not have to change.
func (node *AVLNode) hashBranch(ndb *avlNodeDB) (hash []byte, count int64, storageBytesTotal uint64) {
	var lhash, rhash []byte
	var lcount, rcount int64
	var lsb, rsb uint64
	node = node.clone(true)
	if node.LeftID != nil {
		lhash, lcount, lsb = ndb.GetNode(*node.LeftID, node.ChunkHash).hashBranch(ndb)
		count += lcount
	}
	if node.RightID != nil {
		rhash, rcount, rsb = ndb.GetNode(*node.RightID, node.ChunkHash).hashBranch(ndb)
		count += rcount
	}
	if node.isLeaf() {
		pln := pLeaf{
			Key:          node.Key,
			ValHash:      node.Valhash,
			Height:       node.Height,
			Size:         node.Size,
			StorageBytes: node.StorageBytes,
		}
		node.hash = pln.Hash()
	} else {
		node.StorageBytes = 0
		pin := pInner{
			Key:               node.Key,
			Height:            node.Height,
			Size:              node.Size,
			StorageBytesTotal: lsb + rsb,
			Left:              lhash,
			Right:             rhash,
		}
		node.hash = pin.Hash(nil)
	}
	node.StorageBytesTotal = lsb + rsb + node.StorageBytes
	node.leftHash = lhash
	node.rightHash = rhash
	//dprint("hashbranch: node.hash(%x) lhash(%x) rhash(%x)", node.hash, lhash, rhash)
	ndb.SetNode(node)
	return node.hash, count + 1, node.StorageBytesTotal
}

func (node *AVLNode) clone(exact bool) *AVLNode {
	node.Lock()
	defer node.Unlock()
	if !exact {
		if node.isLeaf() {
			panic("Attempt to copy a leaf node")
		}
		id := makeID(false, node.Key)
		return &AVLNode{
			Key: node.Key,
			//Valhash: node.Valhash,
			Height:            node.Height,
			Size:              node.Size,
			StorageBytes:      node.StorageBytes,
			StorageBytesTotal: node.StorageBytesTotal,
			LeftID:            node.LeftID,
			RightID:           node.RightID,
			ChunkHash:         node.ChunkHash,
			id:                id,
			persisted:         false,
			//hash:              nil,
		}
	}
	id := makeID(false, node.Key)
	return &AVLNode{
		Key:               node.Key,
		Valhash:           node.Valhash,
		Height:            node.Height,
		Size:              node.Size,
		StorageBytes:      node.StorageBytes,
		StorageBytesTotal: node.StorageBytesTotal,
		LeftID:            node.LeftID,
		RightID:           node.RightID,
		ChunkHash:         node.ChunkHash,
		persisted:         node.persisted,
		id:                id,
		hash:              node.hash,
	}
}

func (node *AVLNode) String() string {
	var lid []byte
	if node.LeftID != nil {
		lid = node.LeftID.Bytes()
	}
	var rid []byte
	if node.RightID != nil {
		rid = node.RightID.Bytes()
	}
	// var lhash, rhash []byte
	// if node.leftHash != nil {
	// 	lhash = node.leftHash.Bytes()
	// }
	// if node.rightHash != nil {
	// 	rhash = node.rightHash.Bytes()
	// }
	printNode := struct {
		Id                cmn.HexBytes
		Hash              cmn.HexBytes
		LeftID            cmn.HexBytes
		RightID           cmn.HexBytes
		Key               cmn.HexBytes
		Valhash           cmn.HexBytes
		Height            int
		Size              int
		StorageBytes      int
		StorageBytesTotal int
		Persisted         bool
		LHash             cmn.HexBytes
		RHash             cmn.HexBytes
	}{
		shortbytes(node.id.Bytes()),
		shortbytes(node.hash),
		shortbytes(lid), //node.LeftID.Bytes(),
		shortbytes(rid), // node.RightID.Bytes(),
		shortbytes(node.Key),
		shortbytes(node.Valhash),
		int(node.Height),
		int(node.Size),
		int(node.StorageBytes),
		int(node.StorageBytesTotal),
		node.persisted,
		node.leftHash,
		node.rightHash,
	}
	jsonbytes, _ := json.MarshalIndent(&printNode, "", "    ")
	return string(jsonbytes)
}

/////////////////////////////////////////////////////////////////////////////

// avlNodeDB handles read/writes for AVLNodes
type avlNodeDB struct {
	sync.Mutex
	chunkStore ChunkStore
	nodeMap    map[common.Hash]*AVLChunk
	//chunkHashes []common.Hash // used chunkhashes in loading...for cleanup?
}

func newAVLNodeDB(cs ChunkStore) *avlNodeDB {
	nodedb := new(avlNodeDB)
	nodedb.chunkStore = cs
	nodedb.nodeMap = make(map[common.Hash]*AVLChunk)
	//nodedb.chunkHashes = make(map[uint8]common.Hash)
	return nodedb
}

// loadChunk gets a chunk from cloudstore
// TODO: optimize so that if past a certain local mem capacity, we should start deleting chunks to just get them again if we need them. How to decide which chunks to delete?
func (nodedb *avlNodeDB) loadChunk(chunkHash common.Hash) *AVLChunk {
	//defer timeTrack(time.Now(), "loadChunk")
	nodedb.Lock()
	defer nodedb.Unlock()
	//dprint("[loadChunk] chunkHash(%x)", chunkHash)
	if len(chunkHash) == 0 || EmptyBytes(chunkHash.Bytes()) {
		panic("nodedb loadChunk requires hash")
	}
	// get the chunk from cloudstore
	//dprint("[loadchunk] chunkHash(%x)", chunkHash)
	chunkBytes, ok, err := nodedb.chunkStore.GetChunk(context.TODO(), chunkHash.Bytes())
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("loadChunk: chunk not found")
	}
	//dprint("[loadchunk] chunkBytes gotten (%v)", len(chunkBytes))
	// unload the chunk
	chunk := unmarshalAVLChunk(chunkBytes)
	chunk.Node.persisted = true
	chunk.Node.ChunkHash = chunkHash
	chunk.Node.id = makeID(chunk.Node.isLeaf(), chunk.Node.Key)
	//dprint("[loadchunk] Node(%+v) len(children)(%v)", chunk.Node, len(chunk.Children))

	// don't overwrite chunks that are currently modified, but not flushed yet
	for _, child := range chunk.Children {
		child.id = makeID(child.isLeaf(), child.Key)
		if existingchunk, ok := nodedb.nodeMap[child.id]; ok {
			if existingchunk.Node.persisted == false {
				continue
			}
		}
		child.persisted = true // FYI: this was loaded, but has no children itself.
		nodedb.nodeMap[child.id] = NewAVLChunk(child)
		//dprint("[loadchunk] id(%x) child(%s)", id, child)
	}
	nodedb.nodeMap[chunk.Node.id] = &chunk
	//nodedb.chunkHashes = append(nodedb.chunkHashes, chunkHash)
	return &chunk
}

// SaveChunk saves a chunk to cloudstore. assumes childqueue has already been computed.
func (nodedb *avlNodeDB) SaveChunk(chunk *AVLChunk, wg *sync.WaitGroup, writeToCloudstore bool) (chunkHash common.Hash) {
	nodedb.Lock()
	defer nodedb.Unlock()
	chunk.Node.ChunkHash = *new(common.Hash)
	chunkBytes := marshalAVLChunk(chunk)
	//dprint("[SaveChunk] chunkBytes(%s)", chunkBytes)
	//for id, child := range chunk.Children {
	//	dprint("[SaveChunk] id(%x) child(%s)", id, child)
	//}
	if writeToCloudstore {
		wg.Add(1)
		cloudChunk := new(cloud.RawChunk)
		cloudChunk.Value = chunkBytes
		_, err := nodedb.chunkStore.PutChunk(context.TODO(), cloudChunk, wg)
		if err != nil {
			panic(err)
		}
	}
	for _, child := range chunk.Children {
		nodedb.nodeMap[child.id].Node.Lock()
		nodedb.nodeMap[child.id].Node.persisted = true
		nodedb.nodeMap[child.id].Node.Unlock()
	}
	newChunkHash := common.BytesToHash(wolkcommon.Computehash(chunkBytes))
	chunk.Node.ChunkHash = newChunkHash
	chunk.Node.persisted = true
	chunk.Node.id = makeID(chunk.Node.isLeaf(), chunk.Node.Key)
	nodedb.nodeMap[chunk.Node.id] = chunk

	//id, _ := hex.DecodeString("6a44088f")
	//if bytes.Contains(chunk.Node.id.Bytes(), id) { //|| bytes.Contains(chunk.Node.LeftID.Bytes(), id) || bytes.Contains(chunk.Node.RightID.Bytes(), id) {
	//dprint("[savechunk] found culprit being set!! (%s)", chunk.Node)
	//}
	return newChunkHash
}

// unmarshalAVLNode unpacks a []byte into an AVLNode + nodeMap
func unmarshalAVLChunk(chunkData []byte) (chunk AVLChunk) {
	err := json.Unmarshal(chunkData, &chunk)
	if err != nil {
		panic(err)
	}
	return chunk
}

// marshalAVLNode packs up an AVLNode + nodeMap into a bytes for chunking
func marshalAVLChunk(chunk *AVLChunk) []byte {
	bytes, err := json.Marshal(chunk)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (nodedb *avlNodeDB) GetNode(id common.Hash, parentChunkHash common.Hash) (node *AVLNode) {
	found := false
	node, found = nodedb.getNode(id)
	if found {
		return node
	}
	if parentChunkHash == EMPTYBYTES {
		err := fmt.Errorf("parentChunkHash is empty and id(%x) is not found in nodeMap", id)
		panic(err)
	}
	nodedb.loadChunk(parentChunkHash)
	node, found = nodedb.getNode(id)
	if !found {
		err := fmt.Errorf("node.id(%x) not found. parent chunkhash(%x)", id, parentChunkHash)
		panic(err)
	}
	return node
}

// Get gets a node from the cache, if it is there. if it is not there does not mean it doesn't exist; it may need to be loaded.
func (nodedb *avlNodeDB) getNode(id common.Hash) (node *AVLNode, found bool) {
	nodedb.Lock()
	defer nodedb.Unlock()

	chunk, ok := nodedb.nodeMap[id]
	if ok {
		//dprint("[GetNode] node gotten(%s)", chunk.Node)
		//chunk := NewAVLChunk(nmchunk.Node)
		return chunk.Node, true
	}
	//dprint("[GetNode] node is not found id(%x)", id)
	return node, false
}

// Set inserts a node into a nodedb cache. If the nodeid is already set, it will be overwritten.
func (nodedb *avlNodeDB) SetNode(node *AVLNode) {
	nodedb.Lock()
	defer nodedb.Unlock()
	if node == nil {
		panic("SetNode: can't set a nil node")
	}
	node.id = makeID(node.isLeaf(), node.Key)
	if n, ok := nodedb.nodeMap[node.id]; ok {
		// dprint("[SetNode] FYI setting a node that's already there")
		if reflect.DeepEqual(n, node) {
			// dprint("[SetNode]   did not need to set nodeid(%x)", node.id)
			return
		}
	}
	//id, _ := hex.DecodeString("4E6498FF")
	//if bytes.Contains(node.id.Bytes(), id) {
	//	dprint("[SetNode] found culprit being set!! (%s)", node)
	//	dprint("[SetNode] rightid(%v) (%x)", node.RightID, *node.RightID)
	//}

	chunk := NewAVLChunk(node)
	chunk.Node.persisted = false
	nodedb.nodeMap[node.id] = chunk

	//if bytes.Contains(node.id.Bytes(), id) {
	//	dprint("[SetNode] chunk:(%+v)", chunk)
	//}
}

func (nodedb *avlNodeDB) GetChunk(id common.Hash, parentChunkHash common.Hash) (chunk *AVLChunk) {
	found := false
	chunk, found = nodedb.getChunk(id)
	if found {
		return chunk
	}
	if parentChunkHash == EMPTYBYTES {
		err := fmt.Errorf("parentChunkHash is empty and id(%x) is not found in nodeMap", id)
		panic(err)
	}
	nodedb.loadChunk(parentChunkHash)
	chunk, found = nodedb.getChunk(id)
	if !found {
		err := fmt.Errorf("chunk id(%x) not found. parent chunkhash(%x)", id, parentChunkHash)
		panic(err)
	}
	return chunk
}

func (nodedb *avlNodeDB) getChunk(id common.Hash) (chunk *AVLChunk, found bool) {
	nodedb.Lock()
	defer nodedb.Unlock()
	c, ok := nodedb.nodeMap[id]
	if ok {
		return c, true
	}
	return chunk, false
}

func (nodedb *avlNodeDB) setChunk(chunk *AVLChunk) {
	nodedb.Lock()
	defer nodedb.Unlock()
	if chunk.Node == nil {
		panic("SetChunk: can't set a nil chunk")
	}
	//dprint("[SetChunk] (%s)", chunk.String())
	chunk.Node.id = makeID(chunk.Node.isLeaf(), chunk.Node.Key)
	nodedb.nodeMap[chunk.Node.id] = chunk
}

// SaveBranch saves the given node and all of its descendants.
// NOTE: This function changes the given node and all the descendants as well.
func (nodedb *avlNodeDB) SaveBranch(chunk *AVLChunk, wg *sync.WaitGroup, writeToCloudstore bool) (storageBytesTotal uint64, children []*AVLNode) {

	chunk = chunk.clone()
	// TODO: optimize, don't save unchanged chunks
	if chunk.Node.persisted {
		//dprint("savebranch: node is persisted, don't save. pass back what you have.")
		return chunk.Node.StorageBytesTotal, chunk.Children
	}
	//chunk.Node.persisted = false

	var lchunk *AVLChunk // left chunk
	var rchunk *AVLChunk
	var lq []*AVLNode // left child ids queue
	var rq []*AVLNode
	var lsb uint64 // left storage bytes
	var rsb uint64

	ct := 0
	if chunk.Node.LeftID != nil {
		lchunk = nodedb.GetChunk(*chunk.Node.LeftID, chunk.Node.ChunkHash)
		lsb, lq = nodedb.SaveBranch(lchunk, wg, writeToCloudstore)
		ct++
	}
	if chunk.Node.RightID != nil {
		rchunk = nodedb.GetChunk(*chunk.Node.RightID, chunk.Node.ChunkHash)
		rsb, rq = nodedb.SaveBranch(rchunk, wg, writeToCloudstore)
		ct++
	}

	// compute storage bytes total
	if !chunk.Node.isLeaf() {
		chunk.Node.StorageBytes = 0 // inner nodes are 0 storageBytes
		//dprint("not leaf")
	}
	chunk.Node.StorageBytesTotal = lsb + rsb + chunk.Node.StorageBytes
	//dprint("check sbt: chunk(%s)", chunk)

	// compute children
	for len(lq)+len(rq)+ct > maxSize {
		// pop the child queues until maxSize is met
		lq = popFront(lq)
		rq = popFront(rq)
	}
	cq := make([]*AVLNode, 0)
	for i := 0; i < len(lq) || i < len(rq); i++ {
		if i < len(lq) {
			cq = append(cq, lq[i])
		}
		if i < len(rq) {
			cq = append(cq, rq[i])
		}
	}
	// push the immediate children onto the child queue
	if chunk.Node.LeftID != nil {
		cq = append(cq, lchunk.Node)
	}
	if chunk.Node.RightID != nil {
		cq = append(cq, rchunk.Node)
	}
	chunk.Children = cq

	// compute chunkhash, save chunk
	nodedb.setChunk(chunk)
	chunk.Node.ChunkHash = nodedb.SaveChunk(chunk, wg, writeToCloudstore)
	//cq = append(cq, chunk.Node.id)
	//dprint("SaveBranch: cq")
	//for i := 0; i < len(cq); i++ {
	//	dprint("  i(%d) id(%x)", i, cq[i])
	//}
	return chunk.Node.StorageBytesTotal, cq
}

func popFront(in []*AVLNode) (out []*AVLNode) {
	return in[1:]
}

// DeleteNode is unused at the moment. TODO for cleanup.
func (nodedb *avlNodeDB) DeleteNode(node *AVLNode) {
	nodedb.Lock()
	defer nodedb.Unlock()
	if node == nil {
		panic("Why are you trying to remove a nil node?")
	}
	if _, ok := nodedb.nodeMap[node.id]; !ok {
		dprint("[DeleteNode] why are you deleting something that isn't there?")
	}
	delete(nodedb.nodeMap, node.id)
}

func (nodedb *avlNodeDB) SetID(node *AVLNode) common.Hash {
	nodedb.Lock()
	defer nodedb.Unlock()
	node.id = makeID(node.isLeaf(), node.Key)
	return node.id
}

func (nodedb *avlNodeDB) String() string {
	nodedb.Lock()
	defer nodedb.Unlock()
	printstring := "nodeMap:\n"
	for id, _ := range nodedb.nodeMap {
		printstring += "  " + nodedb.nodeMap[id].Node.String() + "\n"
	}
	return printstring
}
