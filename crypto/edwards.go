package crypto

import (
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/vrf"
	"golang.org/x/crypto/ed25519"
)

const (
	TestPrivateKey  = "044852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d"
	TestAddress     = "0xeae271d4ad7ce77b61dd081a58299190ef86e1c7"
	SignatureLength = 96 // for Edwards / Consensus
)

type PublicKey struct {
	pk ed25519.PublicKey
}

type PrivateKey struct {
	sk ed25519.PrivateKey
}

func (pub *PublicKey) Bytes() []byte {
	return pub.pk
}

func (pub *PublicKey) Address() common.Address {
	pubBytes := pub.Bytes()
	return common.BytesToAddress(wolkcommon.Computehash(pubBytes[1:])[12:])
}

func (priv *PrivateKey) ToSeed() string {
	s := priv.sk.Seed()
	return common.Bytes2Hex(s)
}

func (priv *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{priv.sk.Public().(ed25519.PublicKey)}
}

func (pub *PublicKey) VerifySign(m, sign []byte) error {
	signature := sign[ed25519.PublicKeySize:]
	if ok := ed25519.Verify(pub.pk, m, signature); !ok {
		return fmt.Errorf("signature invalid")
	}
	return nil
}

func (pub *PublicKey) VerifyVRF(proof, m []byte) error {
	vrf.ECVRF_verify(pub.pk, proof, m)
	return nil
}

func (priv *PrivateKey) Sign(m []byte) ([]byte, error) {
	sign, err := priv.sk.Sign(rand.Reader, m, crypto.Hash(0))
	if err != nil {
		return nil, err
	}
	pubkey := priv.sk.Public().(ed25519.PublicKey)
	return append(pubkey, sign...), nil
}

func (priv *PrivateKey) Evaluate(m []byte) (value, proof []byte, err error) {
	proof, err = vrf.ECVRF_prove(priv.PublicKey().pk, priv.sk, m)
	if err != nil {
		return
	}
	value = vrf.ECVRF_proof2hash(proof)
	return
}

func recoverPubkey(sign []byte) *PublicKey {
	pubkey := sign[:ed25519.PublicKeySize]
	return &PublicKey{pubkey}
}

func RecoverPubkey(sign []byte) (pubkey *PublicKey, err error) {
	if len(sign) < ed25519.PublicKeySize {
		return pubkey, fmt.Errorf("Invalid pubkey")
	}
	pk := sign[:ed25519.PublicKeySize]
	return &PublicKey{pk}, nil
}
func GetSigner(sign []byte) (common.Address, error) {
	pkbytes := sign[:ed25519.PublicKeySize]
	var addr common.Address
	copy(addr[:], wolkcommon.Computehash(pkbytes[1:])[12:])
	return addr, nil
}

func PubkeyToAddress(pubkey *PublicKey) common.Address {
	pkbytes := pubkey.Bytes()
	var addr common.Address
	copy(addr[:], wolkcommon.Computehash(pkbytes[1:])[12:])
	return addr
}

func ByteToPublicKey(pubByte []byte) *PublicKey {
	var pubkey PublicKey
	pubkey.pk = make([]byte, len(pubByte))
	copy(pubkey.pk[:], pubByte)
	return &pubkey
}

func HexToPublicKey(s string) *PublicKey {
	var pubkey PublicKey
	pubByte := common.FromHex(s)
	pubkey.pk = make([]byte, len(pubByte))
	copy(pubkey.pk[:], pubByte)
	return &pubkey
}

func HexToPrivateKey(hexSeed string) (*PrivateKey, error) {
	return HexToEd25519(hexSeed)
}

func HexToEd25519(hexSeed string) (*PrivateKey, error) {
	b, err := hex.DecodeString(hexSeed)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	var k PrivateKey
	k.sk = ed25519.NewKeyFromSeed(b)

	return &k, nil
}

func NewKeyPair() (*PublicKey, *PrivateKey, error) {
	pk, sk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return &PublicKey{pk}, &PrivateKey{sk}, nil
}
