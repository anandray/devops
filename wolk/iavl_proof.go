package wolk

import (
	"bytes"
	"fmt"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	// ErrInvalidProof is returned by Verify when a proof cannot be validated.
	ErrInvalidProof = fmt.Errorf("invalid proof")

	// ErrInvalidInputs is returned when the inputs passed to the function are invalid.
	ErrInvalidInputs = fmt.Errorf("invalid inputs")

	// ErrInvalidRoot is returned when the root passed in does not match the proof's.
	ErrInvalidRoot = fmt.Errorf("invalid root")
)

//----------------------------------------

type proofInnerNode struct {
	Key               cmn.HexBytes `json:"k"`
	Height            int8         `json:"h"`
	Size              int64        `json:"s"`
	StorageBytes      uint64       `json:"sb"`
	StorageBytesTotal uint64       `json:"sbt"`

	Left  cmn.HexBytes `json:"l"`
	Right cmn.HexBytes `json:"r"`
}

func (pin proofInnerNode) String() string {
	return pin.stringIndented("")
}

func (pin proofInnerNode) stringIndented(indent string) string {
	return fmt.Sprintf(`proofInnerNode{
%s  Height:            %v
%s  Size:              %v
%s  StorageBytes:      %v
%s  StorageBytesTotal: %v
%s  Key:               %X
%s  Left:              %X
%s  Right:             %X
%s}`,
		indent, pin.Height,
		indent, pin.Size,
		indent, pin.StorageBytes,
		indent, pin.StorageBytesTotal,
		indent, pin.Key,
		indent, pin.Left,
		indent, pin.Right,
		indent)
}

func (pin proofInnerNode) Hash(childHash []byte) []byte {
	// dprint("HASH [pINNERn] height, size, version, key, childhash(%x), right/left", childHash)
	hasher := tmhash.New()
	buf := new(bytes.Buffer)

	err := amino.EncodeInt8(buf, pin.Height)
	if err == nil {
		err = amino.EncodeVarint(buf, pin.Size)
	}
	if err == nil {
		err = amino.EncodeUint64(buf, pin.StorageBytes)
	}
	if err == nil {
		err = amino.EncodeUint64(buf, pin.StorageBytesTotal)
	}
	if err == nil { //keyadded
		err = amino.EncodeByteSlice(buf, pin.Key)
	}

	if len(pin.Left) == 0 {
		if err == nil {
			err = amino.EncodeByteSlice(buf, childHash) // childHash is left
		}
		if err == nil {
			err = amino.EncodeByteSlice(buf, pin.Right)
		}
	} else {
		if err == nil {
			err = amino.EncodeByteSlice(buf, pin.Left)
		}
		if err == nil {
			err = amino.EncodeByteSlice(buf, childHash) // childHash is right
		}
	}
	if err != nil {
		panic(fmt.Sprintf("Failed to hash proofInnerNode: %v", err))
	}

	hasher.Write(buf.Bytes())
	// dprint("HASH    returned(%x)", hasher.Sum(nil))
	return hasher.Sum(nil)
}

//----------------------------------------

type proofLeafNode struct {
	Key               cmn.HexBytes `json:"k"`
	ValHash           cmn.HexBytes `json:"v"`
	Height            int8         `json:"h"`
	Size              int64        `json:"s"`
	StorageBytes      uint64       `json:"sb"`
	StorageBytesTotal uint64       `json:"sbt"`
}

func (pln proofLeafNode) String() string {
	return pln.stringIndented("")
}

func (pln proofLeafNode) stringIndented(indent string) string {
	return fmt.Sprintf(`proofLeafNode{
%s  Height:            %v
%s  Size:              %v
%s  StorageBytes:      %v
%s  StorageBytesTotal: %v
%s  Key:               %X
%s  ValHash:           %X
%s}`,
		indent, pln.Height,
		indent, pln.Size,
		indent, pln.StorageBytes,
		indent, pln.StorageBytesTotal,
		indent, pln.Key,
		indent, pln.ValHash,
		indent)
}

func (pln proofLeafNode) Hash() []byte {

	hasher := tmhash.New()
	buf := new(bytes.Buffer)

	err := amino.EncodeInt8(buf, pln.Height)
	if err == nil {
		err = amino.EncodeVarint(buf, pln.Size)
	}
	if err == nil {
		err = amino.EncodeUint64(buf, pln.StorageBytes)
	}
	if err == nil {
		err = amino.EncodeUint64(buf, pln.StorageBytesTotal)
	}
	if err == nil {
		err = amino.EncodeByteSlice(buf, pln.Key)
	}
	if err == nil {
		err = amino.EncodeByteSlice(buf, pln.ValHash)
	}

	hasher.Write(buf.Bytes())
	// dprint("HASH [pln] 0/1, height, size, version, key, valhash, returned(%x)", hasher.Sum(nil))
	return hasher.Sum(nil)
}

//----------------------------------------

// If the key does not exist, returns the path to the next leaf left of key (w/
// path), except when key is less than the least item, in which case it returns
// a path to the least item.
func (node *Node) PathToLeaf(t *ImmutableTree, key []byte) (PathToLeaf, *Node, error) {
	path := new(PathToLeaf)
	val, err := node.pathToLeaf(t, key, path)
	return *path, val, err
}

// pathToLeaf is a helper which recursively constructs the PathToLeaf.
// As an optimization the already constructed path is passed in as an argument
// and is shared among recursive calls.
func (node *Node) pathToLeaf(t *ImmutableTree, key []byte, path *PathToLeaf) (*Node, error) {
	// dprint("[proof:node:pathtoLEAF] HASH making path! key(%x)", key)
	if node.height == 0 {
		if bytes.Equal(node.key, key) {
			return node, nil
		}
		return node, cmn.NewError("key does not exist")
	}

	if bytes.Compare(key, node.key) < 0 {
		// left side
		pin := proofInnerNode{
			Height:            node.height,
			Size:              node.size,
			StorageBytes:      node.storageBytes,
			StorageBytesTotal: node.storageBytesTotal,
			Key:               node.key, //keyadded
			Left:              nil,
			Right:             node.getRightNode(t).hash,
		}
		*path = append(*path, pin)
		n, err := node.getLeftNode(t).pathToLeaf(t, key, path)
		// dprint("[proof:node:pathtoLEAF] HASH left pin(%s)", pin.String())
		return n, err
	}
	// right side
	pin := proofInnerNode{
		Height:            node.height,
		Size:              node.size,
		StorageBytes:      node.storageBytes,
		StorageBytesTotal: node.storageBytesTotal,
		Key:               node.key, //keyadded
		Left:              node.getLeftNode(t).hash,
		Right:             nil,
	}
	*path = append(*path, pin)
	n, err := node.getRightNode(t).pathToLeaf(t, key, path)
	// dprint("[proof:node:pathtoLEAF] HASH right pin(%s)", pin.String())
	return n, err
}
