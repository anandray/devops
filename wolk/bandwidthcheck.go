package wolk

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
)

type BandwidthCheck struct {
	Node      uint64
	Recipient common.Address
	Amount    uint64
	Sig       []byte
}

func (c *BandwidthCheck) Hash() (h common.Hash) {
	data, _ := rlp.EncodeToBytes(&c)
	return common.BytesToHash(wolkcommon.Computehash(data))
}

func (c *BandwidthCheck) BytesWithoutSig() []byte {
	enc, _ := rlp.EncodeToBytes(&c)
	if len(c.Sig) >= crypto.SignatureLength {
		return enc[0 : len(enc)-crypto.SignatureLength]
	} else {
		return enc
	}
}

func NewBandwidthCheck(nodeID uint64, recipient common.Address, amount uint64) *BandwidthCheck {
	check := new(BandwidthCheck)
	check.Node = nodeID
	check.Recipient = recipient
	check.Amount = uint64(amount)
	return check
}

func (c *BandwidthCheck) SignCheck(priv *crypto.PrivateKey) (err error) {
	sig, err := priv.Sign(c.BytesWithoutSig())
	if err != nil {
		return err
	}
	c.Sig = make([]byte, crypto.SignatureLength)
	copy(c.Sig, sig)
	return nil
}

func (c *BandwidthCheck) Bytes() (enc []byte) {
	enc, _ = rlp.EncodeToBytes(&c)
	return enc
}

type serializedCheck struct {
	Node      uint64         `json:"node"`
	Recipient common.Address `json:"recipient"`
	Amount    uint64         `json:"amount"`
	Sig       string         `json:"sig"`
	Hash      string         `json:"hash"`
	Signer    common.Address `json:"signer"`
}

func (c *BandwidthCheck) String() string {
	if c != nil {
		signer, _ := crypto.GetSigner(c.Sig)
		s := &serializedCheck{
			Node:      c.Node,
			Recipient: c.Recipient,
			Amount:    c.Amount,
			Sig:       fmt.Sprintf("%x", c.Sig),
			//Hash:      c.Hash(),
			Signer: signer,
		}
		bytes, err := json.Marshal(s)
		if err != nil {
			return "{}"
		}
		return string(bytes)
	} else {
		return fmt.Sprint("{}")
	}
}

func (c *BandwidthCheck) GetSigner() (common.Address, error) {
	return crypto.GetSigner(c.Sig)
}
