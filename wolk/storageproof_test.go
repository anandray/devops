package wolk

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	wolkcommon "github.com/wolkdb/cloudstore/common"
)

func generateDeterministicData(n int) []byte {
	numItems := (n / 32) + 2
	out := make([]byte, numItems*32)
	for i := 0; i < numItems; i++ {
		h := wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i)))
		copy(out[i*32:(i+1)*32], h[:])
	}
	return out[0:n]
}

func TestStorageProof(t *testing.T) {
	genesisConfig := new(GenesisConfig)
	cfg, genesisConfig := genConfigs(t)
	wstore := WolkStore{}
	storage, err := NewStorage(cfg, genesisConfig, &wstore)
	if err != nil {
		t.Fatalf("NewStorage %v\n", err)
	}
	sizes := []int{1, 2, 4, 6, 10, 18, 19, 31, 32, 33, 63, 64, 68, 128, 230, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 1000000, 12000000, 50000000}

	for _, n := range sizes {
		input := generateDeterministicData(n)

		// SetReplica: store a chunk of data
		st := time.Now()
		k, storageRoot, err := storage.SetReplica(context.Background(), input)
		if err != nil {
			t.Fatalf("storageProof ERR %v", err)
		}
		fmt.Printf("[storageproof_test:TestStorageProof] SetReplica chunkSize=%d [%s]\n", n, time.Since(st))

		// GetReplica: recovers the chunk of data -- each storer of the share will have a different storageRoot
		output, storageRoot2, err := storage.GetReplica(k, 0, maxFileSize)
		st = time.Now()
		if err != nil {
			t.Fatalf("GetShare ERR %v", err)
		} else if bytes.Compare(input, output) != 0 {
			t.Fatalf("MISMATCHED DATA\ninp:%x\nout:%x\n", input, output)
		} else if bytes.Compare(storageRoot, storageRoot2) != 0 {
			t.Fatalf("MISMATCHED ROOT\ninp:%x\nout:%x\n", storageRoot, storageRoot2)
		}
		fmt.Printf("[storageproof_test:TestStorageProof] GetReplicaProof chunkSize=%d [%s]\n", n, time.Since(st))

		// GetReplicaProof: With a random challenge, the storer must provide a merkle branch "proof of storage" that hashes to the storageRoot
		challenge := uint(rand.Intn((len(output)/32 + 1)))
		st = time.Now()
		proof, err := storage.GetReplicaProof(k, challenge)
		if err != nil {
			t.Fatalf("GetReplicaProof ERR %v", err)
		}
		fmt.Printf("[storageproof_test:TestStorageProof] GetReplicaProof chunkSize=%d [%s]\n", n, time.Since(st))

		// The Verifier can take the Storage Proof
		st = time.Now()
		err = storage.VerifyReplicaProof(storageRoot, proof)
		if err != nil {
			t.Fatalf("[storageproof_test:TestStorageProof] VerifyReplicaProof failure %v\n", err)
		}
		fmt.Printf("[storageproof_test:TestStorageProof] VerifyReplicaProof chunkSize=%d [%s]\n", n, time.Since(st))
	}

}
