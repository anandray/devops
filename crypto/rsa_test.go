package crypto

import (
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"testing"

	wolkcommon "github.com/wolkdb/cloudstore/common"
)

func TestRSAKey(t *testing.T) {
	// crypto/rand.Reader is a good source of entropy for randomizing the encryption function.
	rng := crand.Reader
	privateKey, err := rsa.GenerateKey(rng, BitSize)
	err = privateKey.Validate()
	if err != nil {
		t.Fatalf("Failure to Validate %v", err)
	}

	secretMessage := []byte("send reinforcements, we're going to advance")
	label := []byte("orders")

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &privateKey.PublicKey, secretMessage, label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
		return
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, privateKey, ciphertext, label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from decryption: %s\n", err)
		return
	}

	if bytes.Compare(plaintext, secretMessage) != 0 {
		t.Fatalf("Failure to decrypt")
	} else {
		// Since encryption is a randomized function, ciphertext will be different each time.
		// fmt.Printf("Ciphertext: %x\n", ciphertext)
		// fmt.Printf("Plaintext: %s\n", string(plaintext))
	}

	savePrivateFileTo := "./id_rsa_test"
	savePublicFileTo := "./id_rsa_test.pub"

	err = WriteRSAPublicKeyToFile(&privateKey.PublicKey, savePublicFileTo)
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = WriteRSAPrivateKeyToFile(privateKey, savePrivateFileTo)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestRSAErasure(t *testing.T) {
	t.SkipNow()

	rsaKey, err := rsa.GenerateKey(crand.Reader, 4096)
	if err != nil {
		t.Fatalf("GenerateKey %v", err)
	}
	sz := 4 * 1024
	chunks := wolkcommon.GenerateRandomChunks(1, sz)
	chunk := chunks[0]
	chunkID := wolkcommon.Computehash(chunk)
	data, merkleRoot, err := SetRSAChunk(rsaKey, chunk)

	// (2) Simulate nodes failing to respond in time, by deleting a certain number of chunks
	target_erasures := 0
	for erasures := 0; erasures < target_erasures; {
		idx := rand.Intn(N)
		if data[idx] != nil {
			data[idx] = nil
			erasures++
		}
	}

	// (3) GetChunk: Reconstruct the missing shards
	chunkres, merkleRoot2, err := GetRSAChunk(rsaKey, chunkID, data, chunk)
	if bytes.Compare(merkleRoot, merkleRoot2) != 0 {
		t.Fatalf("Verification Failure %x != %x\n", merkleRoot, merkleRoot2)
	} else if bytes.Compare(chunk, chunkres) != 0 {
		t.Fatalf(" [chunk] FAIL\n")
	}
}

func TestReplica(t *testing.T) {
	t.SkipNow()
	rsaKey, err := rsa.GenerateKey(crand.Reader, 4096)
	if err != nil {
		t.Fatalf("GenerateKey %v", err)
	}

	sz := 4 * 1024
	chunks := wolkcommon.GenerateRandomChunks(1, sz)

	chunk := chunks[0]
	chunkID := wolkcommon.Computehash(chunk)

	replica := 3
	repl := NewReplicator(rsaKey, chunkID, q, nreplicas, N, len(chunk))

	data := repl.encode_replica(chunk, replica)

	original, err := repl.decode_replica(data, replica)
	if err != nil {
		t.Fatalf("Decode_replica %v", err)
	}
	if bytes.Compare(chunk, original) != 0 {
		t.Fatalf("mismatch")
	}
}
