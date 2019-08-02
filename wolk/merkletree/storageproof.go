package merkletree

import (
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
)

const (
	degree = 2
)

func padInput(d []byte) (o []byte) {
	padding := 32 - (len(d) % 32)
	if padding > 0 && padding < 32 {
		return append(d, make([]byte, padding)...)
	}
	return d
}

func numDataBlocks(d []byte) (lend int) {
	n := len(d)
	lend = n / 32
	if n%32 > 0 {
		lend++
	}
	return lend
}

func buildDRG(degree int, numNodes int) (edgeList [][]int) {
	edgeList = make([][]int, numNodes)
	for i := 1; i < numNodes; i++ {
		edgeList[i] = make([]int, degree)
		for j := 0; j < degree; j++ {
			edgeList[i][j] = rand.Intn(i)
		}
	}
	return edgeList
}

func decodeChunk(edgeList [][]int, labels []byte, chunkSize int) (d []byte) {
	//fmt.Printf("recoverData len(labels)=%d len(edgeList)=%d\n", len(labels), len(edgeList))
	numBlocks := numDataBlocks(labels)
	d = make([]byte, numBlocks*32) // (labels)-32*degree)
	for start := degree; start < numBlocks; start++ {
		res := make([]byte, 0)
		for _, end := range edgeList[start] {
			e0 := (end) * 32
			e1 := (end + 1) * 32
			res = append(res, labels[e0:e1]...)
		}
		b0 := (start - degree) * 32
		b1 := (start - degree + 1) * 32
		c0 := start * 32
		c1 := (start + 1) * 32
		if c1 > len(labels) {
			c1 = len(labels)
		}
		l := labels[c0:c1]
		if len(l) < 32 {
			extra := make([]byte, 32-len(l))
			l = append(l, extra...)
		}
		n := xorhash(wolkcommon.Computehash(res), l)
		copy(d[b0:b1], n[:])
		//fmt.Printf("d[%d]=%x == xorhash(%x, labels[%d]=%x)\n", start-degree, d[start-degree], wolkcommon.Computehash(res), start, labels[start])
	}
	if len(d) > chunkSize {
		d = d[0:chunkSize]
	}
	return d
}

func computeLabels(edgeList [][]int, inp []byte, rootHash []byte) (dlabels []byte) {
	input := padInput(inp)
	lend := numDataBlocks(input)
	labels := make([][]byte, lend+degree)

	// compute labels using DRG
	for start := 0; start < lend+degree; start++ {
		if start == 0 {
			labels[start] = rootHash
		} else {
			res := make([]byte, 0)
			for _, end := range edgeList[start] {
				res = append(res, labels[end]...)
			}
			l := wolkcommon.Computehash(res)
			if start >= degree {
				b := (start - degree) * 32
				b2 := b + 32
				if b2 > len(input) {
					b2 = len(input)
				}
				d := input[b:b2]
				if len(d) < 32 {
					extra := make([]byte, 32-len(d))
					d = append(d, extra...)
				}
				l = xorhash(l, d)
			}
			labels[start] = l
		}
	}
	dlabels = make([]byte, 0)
	for _, l := range labels {
		dlabels = append(dlabels, l...)
	}
	return dlabels
}

func dataBlocks(input []byte) (d [][]byte) {
	numBlocks := numDataBlocks(input)
	d = make([][]byte, numBlocks)
	for i := 0; i < numBlocks; i++ {
		b0 := i * 32
		b1 := (i + 1) * 32
		if b1 > len(input) {
			b1 = len(input)
		}
		d[i] = input[b0:b1]
	}
	return d
}

func encodeChunk(edgeList [][]int, inp []byte) (labels []byte, storageRoot common.Hash, err error) {
	// pad the input
	input := padInput(inp)

	// Merkelize the data
	mr := Merkelize(dataBlocks(input))

	// compute labels using DRG edgeList
	labels = computeLabels(edgeList, input, mr[1])

	// merkleize labels
	labels_mr := Merkelize(dataBlocks(labels))
	storageRoot = common.BytesToHash(labels_mr[1])
	return labels, storageRoot, nil
}

func xorhash(a, b []byte) (r []byte) {
	r = make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		r[i] = a[i] ^ b[i]
	}
	return r
}
