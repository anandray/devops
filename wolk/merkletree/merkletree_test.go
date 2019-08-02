package merkletree

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

)

func generateRandomData(l int) (r io.Reader, slice []byte) {
	slice = make([]byte, l)
	rand.Seed(time.Now().Unix())
	if _, err := rand.Read(slice); err != nil {
		panic("rand error")
	}
	r = io.LimitReader(bytes.NewReader(slice), int64(l))
	return
}

func TestMerkleTree(t *testing.T) {
	nitems := int64(6)
	o := make([][]byte, nitems)
	o[0] = common.FromHex("9a1f45935e904a65f23b9568d5e2613f422991ae05063e695b536bb388e41f0f")
	o[1] = common.FromHex("93f4538e7a2fa11445303e594b20c148076437ac7921d7f828f73ec7dd00edb1")
	o[2] = common.FromHex("537aa610c9485bc71376b89cf766673f61905fa640473f12738550850d767ffa")
	o[3] = common.FromHex("196be8b1e390231492624c3a2ad553eb717017c84478d7ca387e5ff7c623054a")
	o[4] = common.FromHex("a69e04f83afac91bef3695055e1431bf3741b4fa5092d513225a6b449db7f534")
	o[5] = common.FromHex("3cea1a477e1a047e6fc7c9b50b408f11206a9112702a2a9c9c6b736d5810c1cf")
	for x := int64(0); x < nitems; x++ {
		fmt.Printf("o[%d]=%x\n", x,  o[x])
	}

	index := uint(0)
	fmt.Printf("value: %x\n", o[index])

	// build merkle tree
	mtree := Merkelize(o)
	root := mtree[1]
	fmt.Printf("root: %x (%d)\n", root, len(mtree))

	// generate merkle proof
	b, err := Mk_branch(mtree, index)
	if err != nil {
		t.Fatalf("mk_branch: %v\n", err)
	}
	for i, x := range b {
		fmt.Printf("%d %x\n", i, x)
	}

	// verify merkle proof
	res, err := Verify_branch(root, uint(index), b)
	if err != nil {
		t.Fatalf("Merkle tree failure %v", err)
	} else {
		fmt.Printf("Merkle tree works %x\n", res)
	}
}
