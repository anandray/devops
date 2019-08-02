package merkletree

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestStorageProof(t *testing.T) {

	// prepare random edge list
	edgeList := buildDRG(degree, 65536/32+degree)

	sizes := []int{1, 2, 4, 6, 10, 18, 19, 31, 32, 33, 63, 64, 68, 128, 230, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536}
	for _, n := range sizes {
		_, input := generateRandomData(n)

		labels, storageRoot, err := encodeChunk(edgeList, input)
		if err != nil {
			t.Fatalf("storageProof ERR %v", err)
		}

		// for a random challenge node, compute merkle branch from the labels
		labelsMerkelized := Merkelize(dataBlocks(labels))
		challenge := uint(rand.Intn((n/32 + 1)))
		branch, err := Mk_branch(labelsMerkelized, challenge)
		if err != nil {
			t.Fatalf("Mk_branch ERR %v", err)
		}

		// verify the branch
		_, err = Verify_branch(storageRoot.Bytes(), challenge, branch)
		if err != nil {
			t.Fatalf("Verify_branch ERR %v", err)
		}
		//fmt.Printf(" Verify_branch sz %d ==> res %d: %x\n", n, challenge, res)

		// recover the data from the labels
		recoverd := decodeChunk(edgeList, labels, n)
		if bytes.Compare(input, recoverd) != 0 {
			t.Fatalf("MISMATCHED\ninp:%x\nout:%x\n", input, recoverd)
		}
	}
}
