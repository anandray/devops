package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"

	wolkcommon "github.com/wolkdb/cloudstore/common"
	"golang.org/x/crypto/ed25519"
	jose "gopkg.in/square/go-jose.v2"
)

func TestEdwardsJWK(t *testing.T) {
	for i := 0; i < 8; i++ {
		hexSeed := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i))))
		b, err := hex.DecodeString(hexSeed)
		if err != nil {
			t.Fatalf("DecodeString")
		}
		edwardsPrivateKey := ed25519.NewKeyFromSeed(b)
		edwardsPublicKey := edwardsPrivateKey.Public()
		var jwk jose.JSONWebKey
		jwk.Key = edwardsPrivateKey
		jwkPrivateBytes, err := jwk.MarshalJSON()
		fmt.Printf("%d JWK Private: %s\n", i, string(jwkPrivateBytes))
		jwk.Key = edwardsPublicKey
		jwkPublicBytes, err := jwk.MarshalJSON()
		fmt.Printf("%d JWK Public: %s\n", i, string(jwkPublicBytes))
	}
}

func TestEdwards(t *testing.T) {
	var priv *PrivateKey

	id := 27
	hash := wolkcommon.Computehash([]byte(fmt.Sprintf("%d", id)))

	k_str := fmt.Sprintf("%x", hash)
	priv, _ = HexToEd25519(k_str)

	msgHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Printf("PublicKey Bytes: [%x] \n", priv.PublicKey().Bytes())

	manualSig, _ := priv.Sign(msgHash)
	pubKey := RecoverPubkey(manualSig)
	err := pubKey.VerifySign(msgHash, manualSig)
	if err != nil {
		t.Fatalf("Unverified Edwards")
	}
	fmt.Printf("Verified Edwards\n")
}
