package wolk

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	//	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
)

// Key: uint64 => value is chunk with Registered node
type RegisteredNode struct {
	address     common.Address   `json:"owner,omitempty"`
	pubkey      *ecdsa.PublicKey `json:"pubkey,omitempty"`
	valueInt    uint64           `json:"valueInt,omitempty"`
	valueExt    uint64           `json:"valueExt,omitempty"`
	storageip   string           `json:"storageip"`
	consensusip string           `json:"consensusip"`
	region      byte             `json:"region"`
	httpPort    uint16           `json:"httpport"`
	data        SerializedNode
}

func (a *RegisteredNode) GetScheme() string {
	if a.consensusip == "127.0.0.1" {
		return "http"
	}
	return "https"
}

func (a *RegisteredNode) GetPort() uint16 {
	if a.httpPort == 80 || a.httpPort == 443 {
		return uint16(443)
	}
	return a.httpPort
}

func (a *RegisteredNode) GetStorageIP() string {
	return a.storageip
}

func (a *RegisteredNode) GetStorageNode() string {
	return a.storageip
}

func convertSerializedNode(e SerializedNode) (a RegisteredNode) {
	a.address = e.Address
	pubkey, err := crypto.DecompressPubkey(common.Hex2Bytes(e.PubKey))
	if err != nil {
		a.pubkey = pubkey
	} else {
		a.pubkey = nil
	}
	a.valueInt = e.ValueInt
	a.valueExt = e.ValueExt
	a.storageip = e.StorageIP
	a.consensusip = e.ConsensusIP
	a.region = e.Region
	a.httpPort = e.HTTPPort
	return a
}

type SerializedNode struct {
	Address     common.Address `json:"owner,omitempty"`
	PubKey      string         `json:"pubkey,omitempty"`
	ValueInt    uint64         `json:"valueInt,omitempty"`
	ValueExt    uint64         `json:"valueExt,omitempty"`
	StorageIP   string         `json:"storageip"`
	ConsensusIP string         `json:"consensusip"`
	Region      byte           `json:"region"`
	HTTPPort    uint16         `json:"httpport"`
}

func (a *RegisteredNode) EncodeRLP(w io.Writer) (err error) {
	a.data.Address = a.address
	a.data.ValueInt = a.valueInt
	a.data.ValueExt = a.valueExt
	if a.pubkey != nil {
		a.data.PubKey = fmt.Sprintf("%x", crypto.CompressPubkey(a.pubkey))
	} else {
		a.data.PubKey = ""
	}

	a.data.StorageIP = a.storageip
	a.data.ConsensusIP = a.consensusip
	a.data.Region = a.region
	return rlp.Encode(w, a.data)
}

func (a *RegisteredNode) DecodeRLP(s *rlp.Stream) error {
	if err := s.Decode(&a.data); err != nil {
		return err
	}
	a.address = a.data.Address
	if len(a.data.PubKey) > 0 {
		pubkey, err := crypto.DecompressPubkey(common.Hex2Bytes(a.data.PubKey))
		if err != nil {
			log.Error("GOT PUBKEY", "pubkeyraw", a.data.PubKey)
			//		return err
		} else {
			a.pubkey = pubkey
		}
	}
	a.valueInt = a.data.ValueInt
	a.valueExt = a.data.ValueExt
	a.storageip = a.data.StorageIP
	a.consensusip = a.data.ConsensusIP
	a.region = a.data.Region
	a.httpPort = a.data.HTTPPort
	return nil
}

func (a *RegisteredNode) String() string {
	if a != nil {
		sn := NewSerializedNode(a)
		return sn.String()
	}
	return fmt.Sprint("{}")
}

/*
func (a *RegisteredNode) Url() string {
	if a.pubkey == nil {
		return "err"
	}
	tmpnode := discover.NewNode(discover.PubkeyID(a.pubkey), a.consensusip, 30303, 30300)
	return tmpnode.String()
}
*/

func NewSerializedNode(n *RegisteredNode) *SerializedNode {
	sn := SerializedNode{
		Address:     n.address,
		ValueInt:    n.valueInt,
		ValueExt:    n.valueExt,
		StorageIP:   n.storageip,
		ConsensusIP: n.consensusip,
		Region:      n.region,
		HTTPPort:    n.httpPort,
	}
	if n.pubkey != nil {
		sn.PubKey = fmt.Sprintf("%x", crypto.CompressPubkey(n.pubkey))
	} else {
		sn.PubKey = ""
	}
	return &sn
}

func (a *SerializedNode) String() string {
	bytes, err := json.Marshal(a)
	if err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

/*
func (a *SerializedNode) Url() string {
	p, err := crypto.DecompressPubkey(common.FromHex(a.PubKey))
	if err != nil {
		return "err"
	}
	tmpnode := discover.NewNode(discover.PubkeyID(p), net.ParseIP(a.ConsensusIP), 30303, 30300)
	return tmpnode.String()
}
*/

func IndexToAddress(idx uint64) common.Address {
	h := wolkcommon.Computehash([]byte(fmt.Sprintf("%x", idx)))[0:20]
	return common.BytesToAddress(h)
}
