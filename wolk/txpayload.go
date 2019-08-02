// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
)

const (
	MinimumQuota = 25000000000
	TestBalance  = 25000
)

const (
	PayloadAccount = "account"

	PayloadTransfer = "transfer"
	PayloadSQL      = "sql"
	PayloadNode     = "node"
	PayloadBucket   = "bucket"
	PayloadKey      = "key"

	// non-transactional
	PayloadBandwidth   = "bandwidth"
	PayloadChunk       = "chunk"
	PayloadChunkSearch = "chunksearch"
	PayloadShare       = "share"
	PayloadBatch       = "batch"
	PayloadShareBatch  = "sbatch"
	PayloadFile        = "file"

	PayloadGenesis     = "genesis"
	PayloadBlock       = "block"
	PayloadVerifyBlock = "verify"

	PayloadVote = "vote"

	PayloadInfo = "info"
	PayloadName = "name"
	PayloadTx   = "tx"
)

type TxTransfer struct {
	Recipient string `json:"recipient,omitempty"`
	Amount    uint64 `json:"amount,omitempty"`
}

type TxNode struct {
	Recipient   common.Address `json:"recipient,omitempty"`
	Amount      uint64         `json:"amount,omitempty"`
	StorageIP   []byte         `json:"storageip,omitempty"`
	ConsensusIP []byte         `json:"consensusip,omitempty"`
	Region      uint8          `json:"region,omitempty"`
}

type TxKey struct {
	ValHash         common.Hash `json:"hash,omitempty"`
	Amount          uint64      `json:"amount,omitempty"`
	BucketIndexName string      `json:"index,omitempty"`
	Key             []byte      `json:"key,omitempty"`
	Data            []byte      `json:"data,omitempty"` // for bookkeeping only, this should not be stored in the transaction
}

type TxBucket struct {
	Name          string           `json:"name,omitempty"`
	BucketType    uint8            `json:"bucketType,omitempty"`   // 0=cloudstore bucket, 1=nosql collection, 2=sql database
	RSAPublicKey  []byte           `json:"rsaPublicKey,omitempty"` // JWK of RSA Public Key ==> could be a SET of JSONWebKeys
	Quota         uint64           `json:"quota,omitempty"`
	RequesterPays uint8            `json:"requesterPays,omitempty"`
	Writers       []common.Address `json:"writers,omitempty"` // TODO: support * (address 0x0) vs specific users
	Schema        string           `json:"schemaURL,omitempty"`
	Indexes       []*BucketIndex   `json:"indexes,omitempty"` // AVL only. TODO: consolidate with Indexes?
	Encryption    string           `json:"encryption,omitempty"`
	ShimURL       string           `json:"shimURL,omitempty"`
}

// AVL only
type BucketIndex struct {
	IndexName string `json:"iname"`
	IndexType string `json:"itype"` // String, Num, Float
	Primary   bool   `json:"p,omitempty"`
	//RootHash  common.Hash `json:"rootHash"`
	//t         *AVLTree
}

// CreateCollection(wolk://owner/person, SchemaURL: wolk://wolk/schema/Person, Encryption: EncryptionNone)
// CreateCollection(wolk://owner/webapplication, SchemaURL: wolk://wolk/schema/WebApplication, Encryption: EncryptionNone)
// CreateCollection(wolk://owner/friendrequest, SchemaURL: wolk://wolk/schema/BefriendAction, Encryption: EncryptionRSA)
// CreateCollection(wolk://owner/actions, SchemaURL: wolk://wolk/schema/Action, Encryption: EncryptionDecryption)

type BucketItem struct {
	Key        string         `json:"key"`
	ValHash    common.Hash    `json:"valHash"`
	Size       uint64         `json:"size"`
	CreateTime uint64         `json:"CreateTime"`
	UpdateTime uint64         `json:"UpdateTime"`
	Writer     common.Address `json:"Writer"`
}

const (
	BucketFile   = 0
	BucketNoSQL  = 1
	BucketSQL    = 2
	BucketSystem = 5
)

const (
	EncryptionNone     = "none"
	EncryptionRSA      = "rsa"
	EncryptionFriends  = "friends"
	EncryptionPersonal = "personal"
)

const (
	indexTypeText   = "Text"
	indexTypeNumber = "Number"
	indexTypeFloat  = "Float"
)

// getPrimaryIndex returns the primary BucketIndex of the Bucket
func (b *TxBucket) GetPrimaryIndex() (idx *BucketIndex, err error) {
	for _, idx := range b.Indexes {
		if idx.Primary {
			return idx, nil
		}
	}
	return idx, fmt.Errorf("No primary key defined")
}

// getIndex returns the corresponding BucketIndex with the name input.
func (b *TxBucket) GetIndex(indexName string) (idx *BucketIndex, err error) {
	//dprint("[txpayload:getIndex] indexName(%s)", indexName)
	for _, idx := range b.Indexes {
		if strings.Compare(idx.IndexName, indexName) == 0 {
			return idx, nil
		}
	}
	return idx, fmt.Errorf("No index found")
}

func (b *TxBucket) ValidWriter(addr common.Address, bucketOwner common.Address) bool {
	if bytes.Compare(addr.Bytes(), bucketOwner.Bytes()) == 0 {
		return true
	}
	for _, w := range b.Writers {
		if bytes.Compare(w.Bytes(), addr.Bytes()) == 0 {
			return true
		}
	}
	return false
}

// Translate uses the index to change the val into the correct type.
func (index *BucketIndex) Translate(val []byte) (field interface{}) {
	switch index.IndexType {
	case indexTypeText:
		field = string(val)
	case indexTypeNumber:
		//field = byteToInt64(val)
		ffield := byteToFloat(val)
		field = int64(ffield)
	case indexTypeFloat:
		field = byteToFloat(val)
	default:
		panic(fmt.Sprintf("unknown index type (%s)", index.IndexType))
	}
	return field
}

func (index *BucketIndex) StringToBytes(in string) (out []byte) {
	switch index.IndexType {
	case indexTypeText:
		out = []byte(in)
	case indexTypeNumber:
		//it, err := strconv.ParseInt(in, 10, 64)
		it, err := strconv.ParseFloat(in, 64)
		if err != nil {
			panic(err)
		}
		//out = int64ToByte(it)
		out = floatToByte(it)
		//log.Info("[txpayload:StringToBytes] indexTypeNumber", "str", in, "int(float)", it, "intbytes", floatToByte(it))
	case indexTypeFloat:
		flt, err := strconv.ParseFloat(in, 64)
		if err != nil {
			panic(err)
		}
		out = floatToByte(flt)
	default:
		panic(fmt.Sprintf("unknown index type (%s)", index.IndexType))
	}
	return out
}

// GetField with take a row (unmarshaled json record) and use the index to return the value of that index in the row.
func (index *BucketIndex) GetField(row map[string]interface{}) (field []byte, ok bool) {
	val, ok := row[index.IndexName]
	if !ok {
		return field, false
	}
	switch index.IndexType {
	case indexTypeText:
		field = []byte(val.(string))
	case indexTypeNumber:
		// switch v := val.(type) {
		// case uint64:
		// 	field = intToByte(val.(uint64))
		// case int64:
		// 	field = int64ToByte(val.(int64))
		// case float64:
		// 	field = floatToByte(val.(float64))
		// default:
		// 	panic(fmt.Sprintf("type not supported(%v)", v))
		// }
		field = floatToByte(val.(float64)) // json always comes out in float64
	case indexTypeFloat:
		field = floatToByte(val.(float64))
	default:
		panic(fmt.Sprintf("unknown index type (%s)", index.IndexType))
	}

	return field, true
}

func ParseBucketList(rawBucketList []byte) (buckets []*TxBucket, err error) {
	err = json.Unmarshal(rawBucketList, &buckets)
	if err != nil {
		return buckets, err
	}
	return buckets, nil
}

func NewTxTransfer(amount uint64, recipient string) *TxTransfer {
	return &TxTransfer{
		Recipient: recipient,
		Amount:    amount,
	}
}

func NewTxAccount(name string, rsaPublicKeyBytes []byte) *TxBucket {
	return &TxBucket{
		Name:         name,
		Quota:        MinimumQuota,
		BucketType:   BucketNoSQL,
		RSAPublicKey: rsaPublicKeyBytes,
	}
}

func NewTxBucket(name string, bucketType uint8, options *RequestOptions) (b *TxBucket) {
	b = &TxBucket{
		Name:       name,
		BucketType: bucketType,
	}
	if options != nil {
		if options.Indexes != nil {
			// check primary key
			foundPrimary := false
			//dprint("options.Indexes(%+v)", options.Indexes)
			for _, idx := range options.Indexes {
				b.Indexes = append(b.Indexes, idx)
				if idx.IndexName == options.PrimaryKey {
					idx.Primary = true
					foundPrimary = true
				} else {
					idx.Primary = false
				}
			}
			if !foundPrimary {
				panic(fmt.Errorf("[txpayload:NewTxBucket] no primary key found"))
			}
		}

		b.RequesterPays = options.RequesterPays
		b.Writers = options.ValidWriters
		b.Schema = options.Schema
		b.Encryption = options.Encryption
		b.ShimURL = options.ShimURL
	}
	//dprint("[txpayload:NewTxBucket] just made a new TxBucket:(%+v)", b)
	return b
}

func NewTxNode(storageip string, consensusip string, region uint8, value uint64) *TxNode {
	return &TxNode{
		StorageIP:   []byte(storageip),
		ConsensusIP: []byte(consensusip),
		Recipient:   EMPTYRECIPIENT,
		Region:      region,
		Amount:      value,
	}
}

func NewTxKey(valHash common.Hash, amount uint64) *TxKey {
	return &TxKey{
		ValHash: valHash,
		Amount:  amount,
	}
}

func NewTxIndexedKey(key []byte, data []byte, amount uint64, bucketIndexName string) *TxKey {
	return &TxKey{
		ValHash:         common.BytesToHash(wolkcommon.Computehash(data)),
		Amount:          amount,
		BucketIndexName: bucketIndexName,
		Key:             key,
		Data:            data,
	}
}

func (txp *TxKey) String() string {
	bytes, err := json.Marshal(txp)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func (txp *TxKey) Byte() []byte {
	bytes, err := json.Marshal(txp)
	if err != nil {
		return nil
	}
	return bytes
}

func (b *TxBucket) String() string {
	bytes, err := json.Marshal(b)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func (txp *TxTransfer) String() string {
	bytes, err := json.Marshal(txp)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
