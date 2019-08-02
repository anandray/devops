package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/reedsolomon"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/merkletree"
)

const (
	jmp     = 255
	rounds  = 8
	debug   = true
	BitSize = 4096
)

var base = big.NewInt(65537)
var bigOne = big.NewInt(1)
var bigTwo = big.NewInt(2)

type ChunkPiece struct {
	rep   uint8
	bytes []byte
}

type Replicator struct {
	Permutation    []int
	PermutationInv []int
	Nodes          []uint64
	PrivateKey     *rsa.PrivateKey
}

func NewReplicator(rsaKey *rsa.PrivateKey, chunkID []byte, q uint, nreplicas int, numnodes int, chunklen int) *Replicator {
	nodes := make([]uint64, numnodes)
	used := make(map[uint64]bool)
	twototheq := uint64(1 << q)
	// replica randomness is public but only generatable by the owner through its signature; each of the 16 can verify if the owner is cheating
	h0 := wolkcommon.Computehash(chunkID)
	for i := 0; i < numnodes; {
		loworder64 := wolkcommon.BytesToUint64(h0[24:32])
		res := loworder64 % twototheq
		if _, ok := used[res]; !ok {
			nodes[i] = res
			used[res] = true
			i++
		}
		h0 = wolkcommon.Computehash(h0)
	}
	// fixed randomness here
	mrand.Seed(42)
	nblocks := chunklen/jmp + 1
	if nblocks%2 > 0 {
		nblocks++
	}
	halfblocks := nblocks / 2
	permutation := mrand.Perm(halfblocks)
	for i, _ := range permutation {
		permutation[i] = halfblocks - i - 1
	}
	permutationinv := make([]int, halfblocks)
	for i, idx := range permutation {
		permutationinv[idx] = i
	}
	if debug {
		fmt.Printf("permutation %v\n", permutation)
		fmt.Printf("permutationinv %v\n", permutationinv)
	}
	return &Replicator{
		Permutation:    permutation,
		PermutationInv: permutationinv,
		Nodes:          nodes,
		PrivateKey:     rsaKey,
	}
}

const (
	N         = 16
	k         = 12
	nreplicas = 4
	q         = 10
)

/*
chunkID is signed => 4 replicas determined through Damgard et al (2018) + 12 parity chunks with RS = 16 replica
  signature decides 16 nodes in R responsible for these 16 replicas
*/
func SetRSAChunk(rsaKey *rsa.PrivateKey, chunk []byte) (data [][]byte, merkleRoot []byte, err error) {
	chunkID := wolkcommon.Computehash(chunk)
	encodeStart := time.Now()
	repl := NewReplicator(rsaKey, chunkID, q, nreplicas, N, len(chunk))
	if debug {
		fmt.Printf("ChunkID # %x Nodes: %v\n", chunkID, repl.Nodes)
	}
	data = make([][]byte, N)
	var wg sync.WaitGroup
	for replica := 0; replica < nreplicas; replica++ {
		wg.Add(1)
		go func(replica int) {
			data[replica] = repl.encode_replica(chunk, replica)
			// fmt.Printf("  Replica %d: %x [Node %d]\n", replica, wolkcommon.Computehash(data[replica]), repl.Nodes[replica])
			wg.Done()
		}(replica)
	}
	wg.Wait()

	// (1) SetChunk: Compute parity chunks -- only cases where you should get an error is, if the data shards aren't of equal size
	// N=number of nodes storing a file split in to dataChunks + k "parity" chunks
	dataChunks := N - k
	enc, err := reedsolomon.New(dataChunks, k)
	for i := dataChunks; i < N; i++ {
		data[i] = make([]byte, len(data[0]))
	}
	err = enc.Encode(data)
	if err != nil {
		return data, merkleRoot, err
	}

	// Build Merkle root out of the RSA-encoded chunks
	mtree := merkletree.Merkelize(data)
	log.Info(fmt.Sprintf("RS(N=%d, k=%d): %3d bytes encoded in %s", N, k, len(chunk), time.Since(encodeStart)))
	return data, mtree[1], nil
}

func GetRSAChunk(rsaKey *rsa.PrivateKey, chunkID []byte, data [][]byte, orig []byte) (chunk []byte, merkleRoot []byte, err error) {
	decodeStart := time.Now()
	dataChunks := N - k
	enc, err := reedsolomon.New(dataChunks, k)
	err = enc.Reconstruct(data)
	if err != nil {
		return chunk, merkleRoot, err
	}

	// If more than *k* shards are missing, the encoding will fail
	mtree2 := merkletree.Merkelize(data)
	var wg sync.WaitGroup

	repl := NewReplicator(rsaKey, chunkID, q, nreplicas, N, len(orig))
	for replica := 0; replica < nreplicas; replica++ {
		wg.Add(1)
		go func(replica int) {
			original, err := repl.decode_replica(data[replica], replica)
			if err != nil {

			} else {
				// fmt.Printf("  Replica %d %x (%d bytes)\n", replica, wolkcommon.Computehash(data[replica]), len(data[replica]))
				chunk = original
			}
			wg.Done()
		}(replica)
	}
	wg.Wait()

	log.Info(fmt.Sprintf("RS(N=%d, k=%d): %3d bytes decoded in %s", N, k, len(chunk), time.Since(decodeStart)))
	return chunk, mtree2[1], nil
}

func deserialize_chunkpiece(in []byte) (cp ChunkPiece) {
	cp.rep = in[0]
	d := 2 + 256
	cp.bytes = make([]byte, 256)
	copy(cp.bytes[:], in[2:d])
	return cp
}

func serialize_chunkpiece(cp *ChunkPiece) []byte {
	out := make([]byte, 258)
	out[0] = byte(cp.rep)
	copy(out[2+(256-len(cp.bytes)):], cp.bytes[:])
	return out
}

func H(s []int, inp []*big.Int) (out []*big.Int) {
	out = make([]*big.Int, len(s))
	for i, _ := range s {
		out[i] = inp[s[i]]
	}
	return out
}

// privateKey.D = new(big.Int).ModInverse(base, totient)
/*
From Ganesh 11/27 mail:
 0. m' = m | r ; split m’ into L || R is \ell where L, R is \ell/2
 5. s = L . H(R) is \ell/2, t = R . H(s) is \ell/2
 8. output = s || t  is \ell
*/
func (self *Replicator) encode_replica(chunk []byte, replica int) (enc []byte) {
	zp := make([]*big.Int, 0)
	x := new(big.Int)
	var chunkpieces []ChunkPiece
	for i := 0; i < len(chunk); i += jmp {
		i1 := i + jmp
		if i1 > len(chunk) {
			i1 = len(chunk)
		}
		x.SetBytes(chunk[i:i1])
		zp = append(zp, new(big.Int).Exp(x, self.PrivateKey.D, self.PrivateKey.N))
		var chunkpiece ChunkPiece
		chunkpiece.rep = uint8(i1 - i)
		chunkpieces = append(chunkpieces, chunkpiece)
	}
	if len(zp)%2 > 0 {
		zp = append(zp, big.NewInt(int64(replica+1)))
		var chunkpiece ChunkPiece
		chunkpiece.rep = 0
		chunkpieces = append(chunkpieces, chunkpiece)
	}
	for i := 0; i < len(chunkpieces); i++ {
		if debug {
			fmt.Printf("pre permuted z[%d] = %v => %d\n", i, zp[i], chunkpieces[i].rep)
		}
	}
	nblocks := len(zp)
	halfblocks := nblocks / 2
	for j := 0; j < rounds; j++ {
		// s = L . H(R) is halfblocks
		H_R := H(self.Permutation, zp[halfblocks:nblocks])
		s := make([]*big.Int, halfblocks)
		for i, Hr := range H_R {
			s[i] = new(big.Int).Mul(Hr, zp[i])
			s[i].Mod(s[i], self.PrivateKey.N)
			if debug {
				fmt.Printf(" Round %d: s[i=%d] = zp[i] * H(R[i]) = %v * %v = %v\n", j, i, zp[i], Hr, s[i])
			}
		}

		// t = R . H(s) is halfblocks
		H_s := H(self.Permutation, s)
		t := make([]*big.Int, halfblocks)
		for i, Hs := range H_s {
			t[i] = new(big.Int).Mul(Hs, zp[i+halfblocks])
			t[i].Mod(t[i], self.PrivateKey.N)
			if debug {
				fmt.Printf(" Round %d: t[i=%d] = R[i] * H(S[i]) = %v * %v = %v\n", j, i, zp[i+halfblocks], Hs, t[i])
			}
		}
		zp = append(s, t...)
	}
	if debug {
		for i, zi := range zp {
			fmt.Printf("encoded z[%d] = %v\n", i, zi)
		}
	}

	enc = make([]byte, 0)
	for i, _ := range chunkpieces {
		chunkpieces[i].bytes = zp[i].Bytes()
		cpb := serialize_chunkpiece(&chunkpieces[i])
		//fmt.Printf(" replica chunk # %d  rep: %d -- %d bytes\n", i, chunkpieces[i].rep, len(cpb))
		enc = append(enc, cpb...)
	}
	return enc
}

/*
To invert: lets say \ell = 10 then
 L = [l[0], l[1], l[2], l[3], l[4]]
 R = [r[0], r[1], r[2], r[3], r[4]]
where the forward permutation and backward permutation is "reverse"
 H  = [4 3 2 1 0]
 Hi = [4 3 2 1 0]
then the output = s || t is:
   s = ( l[0] * r[4],
	       l[1] * r[3],
				 l[2] * r[2],
				 l[3] * r[1],
				 l[4] * r[0] )
	 t = ( r[0] * l[4] * r[0],
	       r[1] * l[3] * r[1],
				 r[2] * l[2] * r[2],
				 r[3] * l[1] * r[3],
				 r[4] * l[0] * r[4] )
so to go from the output back to the input L || R
  r[0] = t[0] / s[Hi[0]] = t[0] / s[4]
  r[1] = t[1] / s[Hi[1]] = t[1] / s[3]
  r[2] = t[2] / s[Hi[2]] = t[2] / s[2]
  r[3] = t[3] / s[Hi[3]] = t[3] / s[1]
  r[4] = t[4] / s[Hi[4]] = t[4] / s[0]
and now that you have R, you can compute L
  l[0] = s[0] / r[Hi[0]] = s[0] / r[4]
  l[1] = s[1] / r[Hi[1]] = s[1] / r[3]
  l[2] = s[2] / r[Hi[2]] = s[2] / r[2]
  l[3] = s[3] / r[Hi[3]] = s[3] / r[1]
  l[4] = s[4] / r[Hi[4]] = s[4] / r[0]
*/
func (self *Replicator) decode_replica(enc []byte, replica int) (dec []byte, err error) {
	N := self.PrivateKey.N
	zp := make([]*big.Int, 0)

	chunkpieces := make([]ChunkPiece, 0)
	for i := 0; i < len(enc); i += 258 {
		cp := deserialize_chunkpiece(enc[i : i+258])
		chunkpieces = append(chunkpieces, cp)
		zp = append(zp, new(big.Int).SetBytes(cp.bytes))
	}
	nblocks := len(zp)
	if nblocks%2 > 0 {
		zp = append(zp, big.NewInt(int64(replica)))
		nblocks++
	}
	halfblocks := nblocks / 2
	s := make([]*big.Int, halfblocks)
	t := make([]*big.Int, halfblocks)
	if debug {
		for i, zi := range zp {
			fmt.Printf("decoding input zp[%d]=%v\n", i, zi)
		}
	}
	for j := 0; j < rounds; j++ {
		copy(s[:], zp[0:halfblocks])
		copy(t[:], zp[halfblocks:nblocks])
		if debug {
			for i, si := range s {
				fmt.Printf(" Round %d input s[%d]=%v\n", j, i, si)
			}
			for i, ti := range t {
				fmt.Printf(" Round %d input t[%d]=%v\n", j, i, ti)
			}
		}

		// R =  t / H'(s)
		R := make([]*big.Int, halfblocks)
		for i, ti := range t {
			inv := new(big.Int).ModInverse(s[self.PermutationInv[i]], N)
			R[i] = new(big.Int).Mul(ti, inv)
			R[i].Mod(R[i], N)
			if debug {
				fmt.Printf("t[%d]=%v\n  self.PermutationInv[%d]=%v\n  s=%v\n  R[%d]=%v\n\n", i, ti, i, self.PermutationInv[i], s[self.PermutationInv[i]], i, R[i])
			}
		}
		// L =  s / H'(R)
		L := make([]*big.Int, halfblocks)
		for i, si := range s {
			inv := new(big.Int).ModInverse(R[self.PermutationInv[i]], N)
			L[i] = new(big.Int).Mul(si, inv)
			L[i].Mod(L[i], N)
			if debug {
				fmt.Printf("s[%d]=%v\n  self.PermutationInv[%d]=%v\n  R=%v\n  L[%d]=%v\n\n", i, si, i, self.PermutationInv[i], R[self.PermutationInv[i]], i, L[i])
			}
		}
		zp = append(L, R...)
	}
	for i, zi := range zp {
		if debug {
			fmt.Printf("decoding output zp[%d]=%v\n", i, zi)
		}
	}
	dec = make([]byte, 0)
	for i, cp := range chunkpieces {
		decoded := new(big.Int).Exp(zp[i], base, self.PrivateKey.N)
		b := decoded.Bytes()
		if cp.rep > 0 {
			res := make([]byte, cp.rep)
			if debug {
				fmt.Printf("i%d => len(b)=%d cp.rep=%d\n", i, len(b), cp.rep)
			}
			if len(b) <= int(cp.rep) {
				copy(res[int(cp.rep)-(len(b)):], b[:])
			} else { //  Problem 256 != 255
				return dec, fmt.Errorf("Problem %d != %d", len(b), cp.rep)
			}
			dec = append(dec, res...)
		}
	}
	return dec, nil
}

type RSADecoder struct {
}

func (wstore *RSADecoder) Decode(chunkID []byte, repl *Replicator, data [][]byte) (chunk []byte, merkleRoot []byte, err error) {
	dataChunks := N - k
	enc, err := reedsolomon.New(dataChunks, k)
	err = enc.Reconstruct(data)
	if err != nil {
		log.Error("RSADecoder.Reconstruct", "err", err)
		return chunk, merkleRoot, err
	}

	// If more than *k* shards are missing, the encoding will fail
	mtree2 := merkletree.Merkelize(data)
	var wg sync.WaitGroup
	for replica := 0; replica < nreplicas; replica++ {
		wg.Add(1)
		go func(replica int) {
			original, err := repl.decode_replica(data[replica], replica)
			if err != nil {

			} else {
				//				fmt.Printf("  Replica %d %x (%d bytes)\n", replica, wolkcommon.Computehash(data[replica]), len(data[replica]))
				chunk = original
			}
			wg.Done()
		}(replica)
	}
	wg.Wait()
	return chunk, mtree2[1], nil
}

func computeRSAChunks(rsaKey *rsa.PrivateKey, chunk []byte) (merkleRoot []byte, err error) {
	q := uint(4) // TODO - must be from recent block
	chunkKey := wolkcommon.Computehash(chunk)
	repl := NewReplicator(rsaKey, chunkKey, q, nreplicas, N, len(chunk))
	data := make([][]byte, N)
	var wg sync.WaitGroup
	for replica := 0; replica < nreplicas; replica++ {
		wg.Add(1)
		go func(replica int) {
			data[replica] = repl.encode_replica(chunk, replica)
			fmt.Printf("  Replica %d: %x [Node %d]\n", replica, wolkcommon.Computehash(data[replica]), repl.Nodes[replica])
			wg.Done()
		}(replica)
	}
	wg.Wait()

	// (1) SetChunk: Compute parity chunks -- only cases where you should get an error is, if the data shards aren't of equal size
	// N=number of nodes storing a file split in to dataChunks + k "parity" chunks
	dataChunks := N - k
	enc, err := reedsolomon.New(dataChunks, k)
	for i := dataChunks; i < N; i++ {
		data[i] = make([]byte, len(data[0]))
	}
	err = enc.Encode(data)
	if err != nil {
		return merkleRoot, err
	}

	// Build Merkle root out of the RSA-encoded chunks
	mtree := merkletree.Merkelize(data)
	log.Trace("SetRSAChunk Encoded", "chunkKey", fmt.Sprintf("%x", chunkKey), "q", q, "nreplicas", nreplicas, "N", N, "k", k, "len(chunk)", len(chunk), "mr", fmt.Sprintf("%x", mtree[1]))
	return merkleRoot, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func WriteRSAPrivateKeyToFile(privateKey *rsa.PrivateKey, saveFileTo string) error {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)
	err := ioutil.WriteFile(saveFileTo, privatePEM, 0600)
	if err != nil {
		return err
	}

	return nil
}

func ReadRSAPrivateKeyFromFile(loadFileFrom string) (pk *rsa.PrivateKey, err error) {

	k0, err := ioutil.ReadFile(loadFileFrom)
	if err != nil {
		return pk, err
	}
	block, _ := pem.Decode(k0)

	if strings.Compare(block.Type, "RSA PRIVATE KEY") != 0 {
		return pk, fmt.Errorf("Unknown type %s", block.Type)
	}
	pk, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return pk, err
	}
	return pk, nil
}

func GetRSAPublicKeyBytes(pk *rsa.PublicKey) (b []byte, err error) {
	// Get ASN.1 DER format
	pubDER := x509.MarshalPKCS1PublicKey(pk)
	pubBlock := pem.Block{
		Type:    "RSA PUBLIC KEY",
		Headers: nil,
		Bytes:   pubDER,
	}

	// Private key in PEM format
	pubPEM := pem.EncodeToMemory(&pubBlock)
	return pubPEM, nil
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func WriteRSAPublicKeyToFile(pk *rsa.PublicKey, saveFileTo string) error {
	pubPEM, err := GetRSAPublicKeyBytes(pk)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(saveFileTo, pubPEM, 0600)
	if err != nil {
		return err
	}
	return nil
}

func RSAPrivateKeyToString(pk *rsa.PrivateKey) string {
	if pk != nil {
		return fmt.Sprintf("{N:%s}", pk.N.String())
	}
	return "{}"
}
