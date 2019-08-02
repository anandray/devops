package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	TestJWKPrivateKey = `{"kty":"EC","crv":"P-256","x":"unfVM277VRp37x1V1WytNjln_tKiB1Js0CDNC9Jw2Ko","y":"4IGM9HdIyTbGhXrz-2y_M6ypbV_h-i7ZF8W4EKKh-sk","d":"G9CAGZnh4Abiz1x-L9bbNJib4pxChht78lDYPE6Ipec"}`
	useDER            = true
)

/*
https://crypto.stackexchange.com/questions/1795/how-can-i-convert-a-der-ecdsa-signature-to-asn-1
describe node's usage of DER generating extra bytes in a signature like
 30460221009c1b4ed7cce4f987e2a794ac503df38470ac0cd563bbff04ad5383f59a9651e2022100c3b6eb2ff11ba7b829cd3ad66a530c7a9f6cdc4faa303918751c3e6aff7c820e
conforming to
 0x30
 b1
 0x02
 b2
 (vr)
 0x02
 b3
 (vs)
via pieces of
 30
 b1 = 0x46 = 70
 02
 b2 = 0x21 = 33
 vr = 009c1b4ed7cce4f987e2a794ac503df38470ac0cd563bbff04ad5383f59a9651e2
 0x02
 b3 = 0x21 = 33
 vs = 00fcf4adae25ded15ae1d09171489bc557d08a24e2f04d896ed18bd160d828c973
*/
func ExtractRSFromDERSig(dersig []byte) (r *big.Int, s *big.Int, err error) {
	if len(dersig) < 64 {
		return r, s, fmt.Errorf("DER is too short")
	}
	if dersig[0] != 0x30 {
		return r, s, fmt.Errorf("Incorrect DER prefix [first byte must be 0x30]")
	}
	b1 := int(dersig[1])
	if b1 != len(dersig)-2 {
		return r, s, fmt.Errorf("Incorrect byte count in sig[1] Observed: %d but expect %d", b1, len(dersig)-2)
	}
	if dersig[2] != 0x02 {
		return r, s, fmt.Errorf("Incorrect separator in sig[2]")
	}
	b2 := dersig[3]
	r_bytes := dersig[4 : 4+b2]
	b3 := dersig[5+b2]
	s_bytes := dersig[6+b2 : 6+b2+b3]
	// fmt.Printf("b1: %d b2: %d b3: %d r_bytes: %x s_bytes: %x\n", b1, b2, b3, r_bytes, s_bytes)
	r = new(big.Int).SetBytes(r_bytes)
	s = new(big.Int).SetBytes(s_bytes)
	return r, s, nil
}

// ECDSA Operations
func JWKToECDSA(privKey string) (pk *ecdsa.PrivateKey, err error) {
	var privateKey jose.JSONWebKey
	err = privateKey.UnmarshalJSON([]byte(privKey))
	if err != nil {
		return pk, fmt.Errorf("[crypto:ecdsa:JWKToECDSA] %s", err)
	}
	pk = privateKey.Key.(*ecdsa.PrivateKey)
	return pk, nil
}

func JWKSetupECDSA(privateKey *ecdsa.PrivateKey) (jwkPrivateKey *jose.JSONWebKey, jwkPublicKeyString string, addr common.Address, err error) {
	jwkPrivateKey = new(jose.JSONWebKey)
	jwkPrivateKey.Key = privateKey

	jwkPublicKey := new(jose.JSONWebKey)
	jwkPublicKey.Key = &(privateKey.PublicKey)
	jwkPublicKeyBytes, err := jwkPublicKey.MarshalJSON()
	if err != nil {
		log.Error("[crypto:ecdsa:JWKSetupECDSA] MarshalJSON", "err", err)
		return jwkPrivateKey, jwkPublicKeyString, addr, fmt.Errorf("[crypto:ecdsa:JWKSetupECDSA] %s", err)
	}
	jwkPublicKeyString = string(jwkPublicKeyBytes)
	addr = GetECDSAAddress(privateKey)
	log.Trace("[crypto:JWKSetupECDSA]", "jwkPublicKeyString", jwkPublicKeyString, "privateKey", privateKey, "addr", addr)
	return jwkPrivateKey, jwkPublicKeyString, addr, nil
}

// msg is raw (unhashed) data; sig is 128 byte string
func JWKSignECDSA(privateKey *ecdsa.PrivateKey, msg []byte) (sig []byte, err error) {
	hashbytes := wolkcommon.Computehash(msg)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashbytes)
	if err != nil {
		return sig, fmt.Errorf("[crypto:ecdsa:JWKSignECDSA] %s", err)
	}
	if useDER {
		rbin := r.Bytes()
		sbin := s.Bytes()
		b2 := len(rbin)
		b3 := len(sbin)
		b1 := b2 + b3 + 4
		output := make([]byte, 4)
		output[0] = 0x30
		output[1] = byte(b1)
		output[2] = 0x02
		output[3] = byte(b2)
		output = append(output, rbin...)
		output = append(output, byte(0x02))
		output = append(output, byte(b3))
		output = append(output, sbin...)
		return output, nil
	}
	signature := make([]byte, 64)
	rb := make([]byte, 32)
	sb := make([]byte, 32)
	rbin := r.Bytes()
	sbin := s.Bytes()
	copy(rb[32-len(rbin):], rbin[:])
	copy(sb[32-len(sbin):], sbin[:])
	copy(signature[0:32], rb[:])
	copy(signature[32:64], sb[:])

	return signature, nil
}

func JWKVerifyECDSA(msg []byte, requester string, sig []byte) (verified bool, err error) {
	// get the public key
	var publicKey jose.JSONWebKey
	err = publicKey.UnmarshalJSON([]byte(requester))
	if err != nil {
		return false, fmt.Errorf("[crypto:ecdsa:JWKVerifyECDSA] %s", err)
	}
	pk := publicKey.Key.(*ecdsa.PublicKey)

	// based on the length of the provided signature, its simple ECDSA or the DER format 0x30 b1 0x02 b2 (vr) 0x02 b3 (vs)
	// so, get r and s
	var r, s *big.Int
	if len(sig) == 64 {
		r = new(big.Int).SetBytes(sig[0:32])
		s = new(big.Int).SetBytes(sig[32:64])
	} else {
		r, s, err = ExtractRSFromDERSig(sig)
		if err != nil {
			return false, fmt.Errorf("[crypto:ecdsa:JWKVerifyECDSA] %s", err)
		}
	}

	// this is what is signed
	hashbytes := wolkcommon.Computehash(msg)
	// now that we have the publickey pk, the hash of the message, and the r & s values from the signature, do the verification
	if ecdsa.Verify(pk, hashbytes, r, s) {
		return true, nil
	}
	log.Info("[ecdsa:JWKVerifyECDSA] HashBytes", "hashbytes", fmt.Sprintf("%x", hashbytes), "msg", string(msg), "signer", requester, "signature", fmt.Sprintf("%x", sig))

	return false, nil
}

func ECDSAPrivateKeyToJSONWebKey(privateKeysECDSA *ecdsa.PrivateKey) (jwkbytes []byte, err error) {
	var jwk jose.JSONWebKey
	jwk.Key = privateKeysECDSA
	jwkbytes, err = jwk.MarshalJSON()
	if err != nil {
		return jwkbytes, err
	}
	return jwkbytes, nil
}

func ECDSAPublicKeyToJSONWebKey(publicKeysECDSA *ecdsa.PublicKey) (jwkbytes []byte, err error) {
	var jwk jose.JSONWebKey
	jwk.Key = publicKeysECDSA
	jwkbytes, err = jwk.MarshalJSON()
	if err != nil {
		return jwkbytes, err
	}
	return jwkbytes, nil
}

func RSAPublicKeyToString(rsaPublickKey *rsa.PublicKey) (jwkbytes []byte, err error) {
	var jwk jose.JSONWebKey
	jwk.Key = rsaPublickKey
	jwkbytes, err = jwk.MarshalJSON()
	if err != nil {
		return jwkbytes, err
	}
	return jwkbytes, nil
}

func ECRecover(pubKey *ecdsa.PublicKey) (addr common.Address) {
	xb := pubKey.X.Bytes()
	pubBytes := make([]byte, 32)
	copy(pubBytes[32-len(xb):], xb[:])

	if pubKey.Y.Bit(0) == 0 {
		pubBytes2 := make([]byte, 33)
		pubBytes2[0] = 0x02
		copy(pubBytes2[1:], pubBytes[:])
		return common.BytesToAddress(wolkcommon.Computehash(pubBytes2[1:])[12:])
	}
	return common.BytesToAddress(wolkcommon.Computehash(pubBytes[1:])[12:])
}
func GetECDSAAddress(privKey *ecdsa.PrivateKey) (addr common.Address) {
	return ECRecover(&(privKey.PublicKey))
}

func GetSignerAddress(requester string) (addr common.Address) {
	if requester == "" {
		log.Error("[crypto:ecdsa:GetSignerAddress] no requester")
		return addr
	}
	var publicKey jose.JSONWebKey
	err := publicKey.UnmarshalJSON([]byte(requester))
	if err != nil {
		log.Error("[crypto:ecdsa:GetSignerAddress]", "requester", requester, "err", err)
		return addr
	}
	pubKey := publicKey.Key.(*ecdsa.PublicKey)
	return ECRecover(pubKey)
}
