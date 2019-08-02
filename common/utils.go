// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package common

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/big"
	"math/rand"
	"time"

	"github.com/cznic/mathutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func Computehash(data ...[]byte) []byte {
	hasher := sha256.New()
	for _, b := range data {
		hasher.Write(b)
	}
	return hasher.Sum(nil)
}

// helper stuff here for a while
func StringToHash(s string) (k common.Hash) {
	h := make([]byte, 32)
	l := len(s)
	if l > 32 {
		l = 32
	}
	copy(h[0:l], []byte(s[0:l]))
	return common.BytesToHash(h)
}

// helper stuff here for a while
func IntToByte(i int64) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, uint64(i))
	return k
}

func UIntToByte(i uint64) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, uint64(i))
	return k
}

func UInt64ToByte(i uint64) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, uint64(i))
	return k
}

func UInt16ToByte(i uint16) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint16(k, uint16(i))
	return k
}

func Uint64ToBytes32(i uint64) (k []byte) {
	k = make([]byte, 32)
	binary.BigEndian.PutUint64(k[24:32], uint64(i))
	return k
}

func BytesToUint64(inp []byte) uint64 {
	return binary.BigEndian.Uint64(inp)
}

func Bytes32ToUint64(k []byte) (out uint64) {
	h := k[0:8]
	return BytesToUint64(h)
}

func UInt64ToBigInt(i uint64) (b *big.Int) {
	// PETHTODO
	return b
}

func FloatToByte(f float64) (k []byte) {
	bits := math.Float64bits(f)
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, bits)
	return k
}

func BytesToFloat(b []byte) (f float64) {
	bits := binary.BigEndian.Uint64(b)
	f = math.Float64frombits(bits)
	return f
}

func Rng() *mathutil.FC32 {
	x, err := mathutil.NewFC32(math.MinInt32/4, math.MaxInt32/4, false)
	if err != nil {
		panic(err)
	}
	return x
}

func GenerateRandomChunks(num int, sz int) (out [][]byte) {
	out = make([][]byte, num)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < num; i++ {
		out[i] = make([]byte, sz)
		rand.Read(out[i])
	}
	return out
}

func SignBytes(b []byte, priv *ecdsa.PrivateKey) (sig []byte, err error) {
	hash := Computehash(b)
	sig, err = crypto.Sign(hash, priv)
	if err != nil {
		return sig, err
	}
	return sig, nil
}

func SignHash(h common.Hash, priv *ecdsa.PrivateKey) (sig []byte, err error) {
	sig, err = crypto.Sign(h.Bytes(), priv)
	if err != nil {
		return sig, err
	}
	return sig, nil
}

func BucketTypeToString(bucketType uint8) string {
	if bucketType == 5 {
		return "SYSTEM"
	}
	if bucketType == 2 {
		return "SQLDB"
	}
	if bucketType == 1 {
		return "NoSQLColl"
	}
	return "FileBucket"
}
