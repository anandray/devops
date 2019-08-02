package merkletree

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/ethereum/go-ethereum/common"
)

type Proof struct {
	Index uint
	Root  []byte
	Proof []byte
}

func is_a_power_of_2(x int) bool {
	if x == 1 {
		return true
	}
	if x%2 == 1 {
		return false
	}
	return is_a_power_of_2(x / 2)
}

func Merkelize(L [][]byte) [][]byte {
	for is_a_power_of_2(len(L)) == false {
		L = append(L, []byte(""))
	}
	//fmt.Printf("len(L) == %d\n", len(L))
	LH := make([][]byte, len(L))
	for i, v := range L {
		LH[i] = v
	}
	nodes := make([][]byte, len(L))
	nodes = append(nodes, LH...)
	for i := len(L) - 1; i >= 0; i-- {
		nodes[i] = wolkcommon.Computehash(append(nodes[i*2], nodes[i*2+1]...))
		// fmt.Printf("MERKLIZE node %d %x <-- %d:%x %d:%x\n", i, nodes[i], i*2, nodes[i*2], i*2+1, nodes[i*2+1])
	}
	return nodes
}

func MerkelRoot(tree [][]byte) (merkelroot []byte) {
	return tree[1]
}

func Mk_branch(tree [][]byte, index uint) (o [][]byte, err error) {
	if index > uint(len(tree)/2) {
		return o, fmt.Errorf("Invalid idx")
	}
	//fmt.Printf("Len: %d\n", len(tree))
	index += uint(len(tree)) / 2
	o = make([][]byte, 1)
	o[0] = tree[index]
	for index > 1 {
		o = append(o, tree[index^1])
		index = index / 2
	}
	return o, nil
}

func Verify_branch_int(root []byte, index uint, proof [][]byte) (res *big.Int, err error) {
	res_byte, err := Verify_branch(root, index, proof)
	if err != nil {
		return res, err
	}
	res = common.BytesToHash(res_byte).Big()
	return res, nil
}

func Verify_branch(root []byte, index uint, proof [][]byte) (res []byte, err error) {
	q := 1 << uint(len(proof)) // 2**len(proof)
	index += uint(q)
	v := proof[0]
	for _, p := range proof[1:] {
		if index%2 > 0 {
			v = wolkcommon.Computehash(append(p, v...))
		} else {
			v = wolkcommon.Computehash(append(v, p...))
		}
		index = index / 2
	}
	if bytes.Compare(v, root) != 0 {
		return res, fmt.Errorf("Mismatch root, got:[%x] expected:[%x]", v, root)
	}
	return proof[0], nil
}

func GenProof(tree [][]byte, ind uint) (merkelroot []byte, mkproof []byte, index uint, err error) {
	index = ind
	treelen := uint(len(tree) / 2)
	if ind > treelen {
		return merkelroot, mkproof, index, fmt.Errorf("Invalid idx")
	}
	ind += treelen
	mkproof = append(mkproof, tree[ind]...)
	for ind > 1 {
		mkproof = append(mkproof, tree[ind^1]...)
		ind = ind / 2
	}
	return tree[1], mkproof, index, nil
}

func CheckProof(expectedMerkleRoot []byte, mkproof []byte, index uint) (isValid bool, merkleroot []byte, err error) {
	if len(mkproof)%32 != 0 {
		return false, merkleroot, errors.New("Invalid mkproof length")
	}

	merkleroot = append(merkleroot, mkproof[0:32]...)
	merklepath := merkleroot
	for depth := 1; depth < len(mkproof)/32; depth++ {
		rhash := make([]byte, 32)
		copy(rhash, mkproof[depth*32:(depth+1)*32])
		if index%2 > 0 {
			merkleroot = wolkcommon.Computehash(append(rhash, merkleroot...))
		} else {
			merkleroot = wolkcommon.Computehash(append(merkleroot, rhash...))
		}
		index = index / 2
		merklepath = append(merklepath, merkleroot...)
	}
	if bytes.Compare(expectedMerkleRoot, merkleroot) != 0 {
		return false, merkleroot, nil
	} else {
		return true, merkleroot, nil
	}
}

func PrintProof(mkproof []byte, index uint) (merkleroot []byte, proofstr string, err error) {
	if len(mkproof)%32 != 0 {
		return merkleroot, proofstr, errors.New("Invalid mkproof length")
	}

	merkleroot = append(merkleroot, mkproof[0:32]...)
	merklepath := merkleroot
	out := fmt.Sprintf("****\nBlockProof \nH0       %x (Leaf) \n", merkleroot)
	for depth := 1; depth < len(mkproof)/32; depth++ {
		rhash := make([]byte, 32)
		copy(rhash, mkproof[depth*32:(depth+1)*32])
		if index%2 > 0 {
			out = out + fmt.Sprintf("H%d [*,P] H(%x,%x)", depth, rhash, merkleroot)
			merkleroot = wolkcommon.Computehash(append(rhash, merkleroot...))
		} else {
			out = out + fmt.Sprintf("H%d [P,*] H(%x,%x)", depth, merkleroot, rhash)
			merkleroot = wolkcommon.Computehash(append(merkleroot, rhash...))
		}
		index = index / 2
		out = out + fmt.Sprintf(" => %x\n", merkleroot)
		merklepath = append(merklepath, merkleroot...)
	}
	proofstr = proofstr + fmt.Sprintf("BlockRoot: %x\n****\n", merkleroot)
	return merkleroot, out, nil
}

func ToProof(roothash []byte, mkProof []byte, ind uint) (p *Proof) {
	var externalProof []byte
	externalProof = append(externalProof, mkProof[32:]...)
	p = &Proof{Index: ind, Root: roothash, Proof: externalProof}
	return p
}

func (self *Proof) String() string {
	out := fmt.Sprintf("{\"index\":\"%d\",\"root\":\"%x\",\"proof\":[", self.Index, self.Root)
	for prev := 0; prev < len(self.Proof); prev += 32 {
		if prev > 0 {
			out = out + ","
		}
		out = out + common.Bytes2Hex(self.Proof[prev:prev+32])
	}
	out = out + "]}"
	return out
}
