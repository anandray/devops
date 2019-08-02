package wolk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/wolkdb/cloudstore/log"

	chunker "github.com/ethereum/go-ethereum/swarm/storage"
	//"github.com/wolkdb/cloudstore/client/chunker"
)

func TestChunker(t *testing.T) {
	fmt.Println("")
	wolk, _, _, _, nodelist := newWolkPeer(t, defaultConsensus)
	defer ReleaseNodes(t, nodelist)
	// for _, node := range wolk {
	// 	node.Storage.Start()
	// }

	wcs := NewPutGetter(wolk[0].Storage)
	// This test can validate files up to a relatively short length, as tree chunker slows down drastically.
	// Validation of longer files is done by TestLocalStoreAndRetrieve in swarm package.
	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289, 524288, 524288 + 1, 524288 + 4097, 1000000, 7 * 524288, 7*524288 + 1, 7*524288 + 4097}
	sizes = sizes[1:18]
	inputs := make(map[uint64][]byte)
	for _, n := range sizes {
		st := time.Now()
		data, input := generateRandomData(n)
		inputs[uint64(n)] = input

		ctx := context.TODO()

		addr, _, err := chunker.TreeSplit(ctx, data, int64(n), wcs)
		if err != nil {
			t.Fatalf(err.Error())
		}
		log.Info("[wolk_test:TestChunker] Write (TreeSplit)", "size", n, "w", time.Since(st))
		st2 := time.Now()
		reader := chunker.TreeJoin(context.TODO(), addr, wcs, 0)
		output := make([]byte, n)
		r, err := reader.Read(output)
		if r != n || err != io.EOF {
			t.Fatalf("read error  read: %v  n = %v  err = %v\n", r, n, err)
		}
		if input != nil {
			if !bytes.Equal(output, input) {
				t.Fatalf("input and output mismatch\n IN: %v\nOUT: %v\n", input, output)
			}
		}
		log.Info("[wolk_test:TestChunker] Read (TreeJoin)", "size", n, "rw", time.Since(st), "r", time.Since(st2))
		// testing partial read
		for i := 1; i < n; i += 10000 {
			readableLength := n - i
			output := make([]byte, readableLength)
			r, err := reader.ReadAt(output, int64(i))
			fmt.Printf(" n %d i %d readableLength %d, r %d\n", n, i, readableLength, r)
			if r != readableLength || err != io.EOF {
				t.Fatalf("readAt error with offset %v read: %v  n = %v  err = %v\n", i, r, readableLength, err)
			}
			if input != nil {
				if !bytes.Equal(output, input[i:]) {
					t.Fatalf("input and output mismatch\n IN: %v\nOUT: %v\n", input[i:], output)
				}
			}
		}
	}
}
