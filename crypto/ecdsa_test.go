package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	jose "gopkg.in/square/go-jose.v2"
)

// generates ECDSA keys, outputting the raw hex and JWK Private key form, and then the JWK Public key form, with the ECDSA address derived from the Public Key
func TestGenerateKeys(t *testing.T) {
	for i := 0; i < 8; i++ {
		// deterministic raw key (64 char hex), for our testing
		k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i))))

		// turn hex string into ecdsa.PrivateKey for the purposes of ethereum discover
		privateKeysECDSA, err := ethcrypto.HexToECDSA(k_str)
		if err != nil {
			t.Fatalf("HexToECDSA %v", err)
		}

		var jwk jose.JSONWebKey
		jwk.Key = privateKeysECDSA
		jwkbytes, err := jwk.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON %v %v", err, privateKeysECDSA.Curve)
		}
		fmt.Printf("%d: Raw: %x\n", i, k_str)
		fmt.Printf("JWK Private Key: %s\n\n", jwkbytes)
	}
}

func TestValidateKeys(t *testing.T) {
	for i := 0; i < 8; i++ {
		// deterministic raw key (64 char hex), for our testing
		k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i))))

		// turn hex string into ecdsa.PrivateKey for the purposes of ethereum discover
		privateKeyECDSA, err := ethcrypto.HexToECDSA(k_str)
		if err != nil {
			t.Fatalf("HexToECDSA %v", err)
		}

		var jwk jose.JSONWebKey
		jwk.Key = privateKeyECDSA
		jwkbytes, err := jwk.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON %v %v", err, privateKeyECDSA.Curve)
		}
		fmt.Printf("%d: Raw: %x\n", i, k_str)
		fmt.Printf("JWK Private Key: %s\n\n", jwkbytes)

		jwkprivkey, jwkpubkey, addr, err := JWKSetupECDSA(privateKeyECDSA)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("jwkprivkey(%+v) ?= orig jwk (%+v)\n", jwkprivkey.Key, jwk.Key)
		fmt.Printf("jwkpubkey(%+v) addr: %x\n", jwkpubkey, addr)

	}

}

func TestCheckAddress(t *testing.T) {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))

	privKey := `{"kty":"EC","crv":"P-256","x":"d22uwkAD9ou144x1m_gLEEjytXyxokHgVnjMQhENxjs","y":"_fX9VT4KPsbl81HrcomYqaOHy1RMfaq7BB-nCnyqFfI","d":"rwPquDfRZZOuqzlatMjIsSHHVnitJZyndFRTZMGcc7U"}`
	pubKey := `{"kty":"EC","crv":"P-256","x":"d22uwkAD9ou144x1m_gLEEjytXyxokHgVnjMQhENxjs","y":"_fX9VT4KPsbl81HrcomYqaOHy1RMfaq7BB-nCnyqFfI"}`
	privateKey, err := JWKToECDSA(privKey)
	if err != nil {
		t.Fatalf("JWKToECDSA %v\n", err)
	}
	addr := GetECDSAAddress(privateKey)
	addr2 := GetSignerAddress(pubKey)
	fmt.Printf("%x\n%x\n", addr, addr2)
}

func TestECDSA(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey %v\n", err)
	}

	_, jwkPublicKeyString, addr, err := JWKSetupECDSA(privateKey)
	if err != nil {
		t.Fatalf("JWKSetupECDSA %v\n", err)
	}
	fmt.Printf("Public Key: %s\n", jwkPublicKeyString)
	fmt.Printf("Address: %x\n", addr)

	message := "test message"
	signature, err := JWKSignECDSA(privateKey, []byte(message))
	if err != nil {
		t.Fatalf("JWKSignECDSA %v", err)
	}

	verified, err := JWKVerifyECDSA([]byte(message), jwkPublicKeyString, signature)
	if err != nil {
		t.Fatalf("JWKVerifyECDSA %v\n", err)
	}
	fmt.Printf("Verified: %v len(signature)=%d\n", verified, len(signature))
}

func TestNodeJS(t *testing.T) {
	jwk := `{"kty":"EC","crv":"P-256","x":"GyzhP0mMVH8_bVODTwB1kJo_qztnh7ncID0WVsE3ga4","y":"ZFXfCDblHP2apsjge5GWnOeFAx8O_836TZ_SAHl_-Eo","key_ops":["verify"],"ext":true}`
	hashbytes := common.FromHex("e7e8b89c2721d290cc5f55425491ecd6831355e91063f20b39c22f9ec6a71f91")
	nodesig := common.FromHex("30460221009c1b4ed7cce4f987e2a794ac503df38470ac0cd563bbff04ad5383f59a9651e2022100c3b6eb2ff11ba7b829cd3ad66a530c7a9f6cdc4faa303918751c3e6aff7c820e")
	r, s, err := ExtractRSFromDERSig(nodesig)
	if err != nil {
		t.Fatalf("Error")
	}
	var j jose.JSONWebKey
	err = j.UnmarshalJSON([]byte(jwk))
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	pk := j.Key.(*ecdsa.PublicKey)
	if ecdsa.Verify(pk, hashbytes, r, s) {
		fmt.Printf("VERIFIED\n")
	} else {
		fmt.Printf("NOT VERIFIED\n")
	}
}

func TestBrowserJS(t *testing.T) {
	jwk := `{"crv":"P-256","ext":true,"key_ops":["verify"],"kty":"EC","x":"lxYtM63Yccu7xtRoZB9lTxbpdGjvJ2mYsZ4XmYUJ2WE","y":"zEBdoEflfZ3QNe_fGIjt3nyQ2clqBXimbaHHr6X1ik8"}`
	hashbytes := common.FromHex("e7e8b89c2721d290cc5f55425491ecd6831355e91063f20b39c22f9ec6a71f91")
	sig := common.FromHex("635ca4f4dcc66449804f7a255ee00ace03e0a650bd40eac85bf3f403693297a94d5eb308cad8be621e1d3c7d47d307f06a7a2f956c9e88c579527d9af7782c0d")

	var j jose.JSONWebKey
	err := j.UnmarshalJSON([]byte(jwk))
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	pk := j.Key.(*ecdsa.PublicKey)
	if ecdsa.Verify(pk, hashbytes, new(big.Int).SetBytes(sig[0:32]), new(big.Int).SetBytes(sig[32:64])) {
		fmt.Printf("VERIFIED\n")
	} else {
		fmt.Printf("NOT VERIFIED [%d]\n", len(sig))
	}
}
