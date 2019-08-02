// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	jose "gopkg.in/square/go-jose.v2"
)

// Blocks store JSON Transactions; P2P Gossiping of Transactions are in RLP form.
type Transaction struct {
	Method  []byte `json:"method,omitempty"` // 0=POST/1=PUT/2=DELETE
	Path    []byte `json:"path,omitempty"`   // owner/collection/key [where key=SQL for SQL Mutates, but could be the Table name]
	Payload []byte `json:"body,omitempty"`   // raw JSON body
	Sig     []byte `json:"sig,omitempty"`    // Sign(ecdsaKey SHA-256, append(path, payload) ==> 64 bytes or DER format
	Signer  []byte `json:"signer,omitempty"` // JSON Web key.     [could eliminate with a V=bit, but that assumes SigType=ECDSA]
	// for extensibility, can describe Sig/Signer, e.g. that Sig is 64 bytes (=> DER, JSONWebSignature), Signer is JSONWebKey (=> Vbit)
	// SigType         []byte `json:"sigType,omitempty"`
}

func (tx *Transaction) Owner() (k string) {
	pathpieces := strings.Split(strings.Trim(string(tx.Path), "/"), "/")
	if len(pathpieces) > 0 {
		return pathpieces[0]
	}
	return "unknown"
}

func (tx *Transaction) Collection() (k string) {
	pathpieces := strings.Split(strings.Trim(string(tx.Path), "/"), "/")
	if len(pathpieces) > 1 {
		return pathpieces[1]
	}
	return "unknown"
}

// use this for paths with indexes
// func (tx *Transaction) IndexAndKey() (index string, key string) {
// 	pathpieces := strings.Split(strings.Trim(string(tx.Path), "/"), "/")
// 	if len(pathpieces) > 3 {
// 		return pathpieces[2], strings.Join(pathpieces[2:], "/")
// 	}
// 	return "unknown", "unknown"
// }

// use this for paths without indexes - keys must be strings
func (tx *Transaction) Key() (k string) {
	pathpieces := strings.Split(strings.Trim(string(tx.Path), "/"), "/")
	if len(pathpieces) > 2 {
		return strings.Join(pathpieces[2:], "/")
	}
	return "unknown"
}

func (tx *Transaction) GetTxKey() (p *TxKey, err error) {
	var txp TxKey
	err = json.Unmarshal(tx.Payload, &txp)
	return &txp, err
}

func (tx *Transaction) GetTxBucket() (p *TxBucket, err error) {
	var txp TxBucket
	err = json.Unmarshal(tx.Payload, &txp)
	return &txp, err
}

func (tx *Transaction) GetSQLRequest() (p *SQLRequest, err error) {
	var txp SQLRequest
	err = json.Unmarshal(tx.Payload, &txp)
	return &txp, err
}

func (tx *Transaction) GetPayloadType() string {
	pathpieces := strings.Split(strings.Trim(string(tx.Path), "/"), "/")
	log.Trace("[transaction:GetPayloadType]", "path", string(tx.Path), "pathpieces", pathpieces)
	if len(pathpieces) > 1 && pathpieces[0] == ProtocolName { // wolk://wolk/transfer
		return pathpieces[1]
	}
	if len(pathpieces) <= 2 { // wolk://owner
		return PayloadBucket
	}
	if len(pathpieces) > 2 && pathpieces[2] == "SQL" { // wolk://owner/database/SQL
		log.Trace("[transaction:GetPayloadType]", "result", "SQL")
		return PayloadSQL
	}
	return PayloadKey // wolk://owner
}

func (tx *Transaction) GetTxPayload() (p interface{}, err error) {
	payloadType := tx.GetPayloadType()

	log.Trace("[transaction:GetTxPayload]", "payloadType", payloadType, "method", string(tx.Method))
	// DELETE has no payload, just the Path identifies what to delete
	if string(tx.Method) == http.MethodDelete {
		pathpieces := strings.Split(strings.Trim(string(tx.Path), "/"), "/")
		pathLength := len(pathpieces)
		if pathLength > 2 {
			return TxKey{}, nil
		} else if pathLength > 1 {
			return TxBucket{}, nil
		}
		// TODO: support account deletion
		return nil, nil
	}

	switch payloadType {
	case PayloadBucket: // POST: owner/collection or owner/database or owner/bucket
		var txp TxBucket
		log.Trace("[transaction:GetPayloadType] bucket")
		err = json.Unmarshal(tx.Payload, &txp)
		if err != nil {
			return txp, fmt.Errorf("[transaction:GetPayloadType] %s", err)
		}
		return txp, nil
	case PayloadTransfer: // POST: path=wolk/transfer
		var txp TxTransfer
		err = json.Unmarshal(tx.Payload, &txp)
		if err != nil {
			return txp, fmt.Errorf("[transaction:GetPayloadType] %s", err)
		}
		return txp, nil
	case PayloadNode: // POST: wolk/node
		var txp TxNode
		err = json.Unmarshal(tx.Payload, &txp)
		return txp, err
	case PayloadKey: // POST: owner/collection/key
		var txp TxKey
		log.Trace("[transaction:GetPayloadType] KEY", "payload", string(tx.Payload))
		err = json.Unmarshal(tx.Payload, &txp)
		if err != nil {
			return txp, fmt.Errorf("[transaction:GetPayloadType] %s", err)
		}
		return txp, nil
	case PayloadSQL: // POST: owner/database/SQL
		var txp SQLRequest
		log.Info("[transaction:GetPayloadType] SQLRequest", "payload", string(tx.Payload))
		err = json.Unmarshal(tx.Payload, &txp)
		if err != nil {
			return txp, fmt.Errorf("[transaction:GetPayloadType] %s", err)
		}
		return txp, nil
	}
	return tx.Payload, fmt.Errorf("[transaction:GetPayloadType] Unknown payloadType %s", payloadType)
}

type Transactions []interface{}
type WolkTransactions []*Transaction

type TransactionMsg struct {
	TxType  uint64
	Payload []byte
}

const (
	TypeTransaction = 1
)

func DecodeRLPTransaction(txbytes []byte) (tx *Transaction, err error) {
	var txo Transaction
	err = rlp.DecodeBytes(txbytes, &txo)
	if err != nil {
		return tx, err
	}
	return &txo, nil
}

var EMPTYBYTES = common.BytesToHash(make([]byte, 32))
var EMPTYRECIPIENT = common.BytesToAddress(make([]byte, 20))

func (tx *Transaction) String() string {
	if tx != nil {
		stx := NewSerializedTransaction(tx)
		return stx.String()
	} else {
		return fmt.Sprint("{}")
	}

}

func (tx *Transaction) Size() common.StorageSize {
	return 1
}

// full hash
func (tx Transaction) Hash() common.Hash {
	return rlpHash([]interface{}{
		tx.Method,
		tx.Path,
		tx.Payload,
		tx.Sig,
		tx.Signer,
	})
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha256.New()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func (tx *Transaction) Hex() string {
	return fmt.Sprintf("%x", tx.Bytes())
}

func (tx *Transaction) Bytes() (enc []byte) {
	enc, _ = rlp.EncodeToBytes(&tx)
	return enc
}

func NewTransactionImplicit(r *http.Request, payload []byte) (tx *Transaction) {
	tx = new(Transaction)
	tx.Method = []byte(r.Method)
	tx.Path = []byte(r.URL.Path)
	tx.Sig = common.FromHex(r.Header.Get("Sig"))
	tx.Signer = []byte(r.Header.Get("Requester"))
	tx.Payload = payload
	//log.Info("[transaction:NewTransactionImplicit]", "tx.Payload", fmt.Sprintf("%s", payload))
	return tx
}

// ownerpath is everything up to and not including the immediate bucket name
func MakeIndexedTransactions(priv *ecdsa.PrivateKey, method string, ownerPath string, bucket *TxBucket, jsonData []byte) (txns []*Transaction, err error) {
	var row map[string]interface{}
	err = json.Unmarshal(jsonData, &row)
	if err != nil {
		return txns, fmt.Errorf("[transaction:MakeIndexedTransactions] %s", err)
	}
	//valChunkHash := common.BytesToHash(wolkcommon.Computehash(jsonData))
	amount := uint64(len(jsonData))
	primaryFound := false
	//dprint("[transaction:MakeIndexedTransactions] indexes (%+v)", bucket.Indexes)
	for _, idx := range bucket.Indexes {
		//dprint("[transaction:MakeIndexedTransactions] idx(%+v)", idx)
		key, ok := idx.GetField(row)
		if !ok {
			log.Info("[transaction:MakeIndexedTransactions] index field not found in row", "field", idx.IndexName, "row", row)
			// this is ok b/c we may be inserting a row that doesn't have all the indexes in it.
			continue
		}
		if idx.Primary {
			primaryFound = true
		}
		txKey := NewTxIndexedKey(key, jsonData, amount, idx.IndexName)
		tx, err := NewTransaction(priv, method, path.Join(ownerPath, bucket.Name, idx.IndexName), txKey)
		if err != nil {
			return txns, fmt.Errorf("[transaction:MakeIndexedTransactions] %s", err)
		}
		txns = append(txns, tx)
	}
	if !primaryFound {
		return txns, fmt.Errorf("[transaction:MakeIndexedTransactions] primary key not found")
	}
	return txns, nil
}

func MakeIndexedTransactionsImplicit(r *http.Request, bucket *TxBucket, rowData []byte) (txns []*Transaction, err error) {

	var row map[string]interface{}
	err = json.Unmarshal(rowData, &row)
	if err != nil {
		return txns, fmt.Errorf("[transaction:MakeIndexedTransactionsImplicit] %s", err)
	}
	//valChunkHash := common.BytesToHash(wolkcommon.Computehash(jsonData))
	amount := uint64(len(rowData))
	primaryFound := false
	//dprint("[transaction:MakeIndexedTransactions] indexes (%+v)", bucket.Indexes)
	for _, idx := range bucket.Indexes {
		//dprint("[transaction:MakeIndexedTransactions] idx(%+v)", idx)
		key, ok := idx.GetField(row)
		if !ok {
			log.Info("[transaction:MakeIndexedTransactionsImplicit] index field not found in row", "field", idx.IndexName, "row", row)
			// this is ok b/c we may be inserting a row that doesn't have all the indexes in it.
			continue
		}
		//log.Info("[transaction:MakeIndexedTransactionsImplicit] from GetField", "key", key, "row", rowData)
		if idx.Primary {
			primaryFound = true
		}
		txKey := NewTxIndexedKey(key, rowData, amount, idx.IndexName)
		//tx, err := NewTransaction(priv, method, path.Join(ownerPath, bucket.Name, idx.IndexName), txKey)
		tx := NewTransactionImplicit(r, txKey.Byte())
		if err != nil {
			return txns, fmt.Errorf("[transaction:MakeIndexedTransactionsImplicit] %s", err)
		}
		txns = append(txns, tx)
	}
	if !primaryFound {
		return txns, fmt.Errorf("[transaction:MakeIndexedTransactionsImplicit] primary key not found")
	}
	return txns, nil
}

func NewTransaction(priv *ecdsa.PrivateKey, method string, path string, payload interface{}) (tx *Transaction, err error) {
	tx = new(Transaction)
	tx.Method = []byte(method)
	tx.Path = []byte(path)

	log.Trace("[transaction:NewTransaction]", "method", string(tx.Method), "path", string(tx.Path))

	// set tx.Signer
	var jwkPublicKey jose.JSONWebKey
	jwkPublicKey.Key = &(priv.PublicKey)
	jwkPublicKeyString, err := jwkPublicKey.MarshalJSON()
	if err != nil {
		return tx, fmt.Errorf("[transaction:NewTransaction] %s", err)
	}
	tx.Signer = []byte(jwkPublicKeyString)

	// set tx.TransactionType + tx.Payload
	switch p := payload.(type) {

	case *TxTransfer:
		tx.Payload, err = json.Marshal(p)
		if err != nil {
			return tx, fmt.Errorf("[transaction:NewTransaction] %s", err)
		}
	case *TxNode:
		log.Trace("[transaction:NewTransaction] TxNode", "tx", tx, "p", p)
		tx.Payload, err = json.Marshal(p)
		if err != nil {
			return tx, fmt.Errorf("[transaction:NewTransaction] %s", err)
		}
	case *TxBucket:
		log.Trace("[transaction:NewTransaction] TxBucket", "tx", tx, "p", p)
		tx.Payload, err = json.Marshal(p)
		if err != nil {
			return tx, fmt.Errorf("[transaction:NewTransaction] %s", err)
		}
	case *TxKey:
		tx.Payload, err = json.Marshal(p)
		if err != nil {
			return tx, fmt.Errorf("[transaction:NewTransaction] %s", err)
		}
	case *SQLRequest:
		log.Trace("[transaction:NewTransaction] SQLREQUEST", "tx", tx)
		tx.Payload, err = json.Marshal(p)
		if err != nil {
			return tx, fmt.Errorf("[transaction:NewTransaction] %s", err)
		}
	default:
		return tx, fmt.Errorf("[transaction:NewTransaction] Unknown Type (%T)", p)
	}

	// set tx.Sig
	var payloadBytes []byte
	if string(tx.Method) == http.MethodPut {
		// This covers our base for now.
		payloadBytes = append([]byte(tx.Method), []byte(tx.Path)...)
	} else {
		payloadBytes = append(append([]byte(tx.Method), []byte(tx.Path)...), tx.Payload...)
	}

	sig, err := crypto.JWKSignECDSA(priv, payloadBytes)
	if err != nil {
		return tx, fmt.Errorf("[transaction:NewTransaction] %s", err)
	}
	tx.Sig = []byte(sig)

	return tx, nil
}

func (tx *Transaction) GetSignerAddress() (address common.Address, err error) {
	return crypto.GetSignerAddress(string(tx.Signer)), nil
}

func (tx *Transaction) ValidateTx() (bool, error) {
	/*	if p.TransactionType == PayloadSetName {
			if len(txp.Key) == 0 {
				return false, fmt.Errorf("Invalid name")
			}
		}
		if p.TransactionType == PayloadSetKey {
			if len(txp.Collection) == 0 || len(txp.Collection) > 128 {
				return false, fmt.Errorf("Invalid collection name")
			}

			if len(txp.Key) == 0 || len(txp.Key) > 128 {
				return false, fmt.Errorf("Invalid key length")
			}
		} */
	// this is a raw JSON Web Signature (JWS)
	log.Trace("[transaction:ValidateTx]", "tx", tx)
	payloadType := tx.GetPayloadType()
	// the content of the tx.Payload is a JSON string, except for the PUT case
	var payloadBytes []byte
	if string(tx.Method) == http.MethodPut {
		// TODO: bring in the sha256(body) from the client
		payloadBytes = append([]byte(tx.Method), []byte(tx.Path)...)
	} else {
		payloadBytes = append(append([]byte(tx.Method), []byte(tx.Path)...), tx.Payload...)
	}
	verified, err := crypto.JWKVerifyECDSA(payloadBytes, string(tx.Signer), tx.Sig)
	if err != nil {
		log.Error("[transaction:ValidateTx] Verification failure", "err", err)
		return false, err
	} else if !verified {
		log.Error("[transaction:ValidateTx] NOT VERIFIED", "payloadType", payloadType, "signer", string(tx.Signer), "sig", fmt.Sprintf("%x", tx.Sig))
	}
	return verified, nil
}

type SerializedTransaction struct {
	Txhash        common.Hash `json:"txhash"`
	Method        string      `json:"method"`
	Path          string      `json:"path"`
	Payload       string      `json:"payload"`
	Sig           string      `json:"sig"`    //
	Signer        string      `json:"signer"` // JSONWebKey
	BlockNumber   uint64
	Status        string
	SignerAddress common.Address
	Txsize        uint64 `json:"txsize"`
}

func (stx *SerializedTransaction) DeserializeTransaction() (n *Transaction) {
	n = new(Transaction)
	n.Method = []byte(stx.Method)
	n.Path = []byte(stx.Path)
	n.Payload = []byte(stx.Payload)
	n.Sig = common.FromHex(stx.Sig)
	n.Signer = []byte(stx.Signer)
	return n
}

func NewSerializedTransaction(tx *Transaction) *SerializedTransaction {
	signerAddress, err := tx.GetSignerAddress()
	if err != nil {
		signerAddress = common.Address{}
	}
	txsize := uint64(len(tx.Bytes()))
	return &SerializedTransaction{
		Txhash:  tx.Hash(),
		Method:  string(tx.Method),
		Path:    string(tx.Path),
		Payload: string(tx.Payload),
		Sig:     fmt.Sprintf("%x", tx.Sig),
		Signer:  string(tx.Signer),
		//BlockNumber:     tx.blockNumber,
		SignerAddress: signerAddress,
		Txsize:        txsize,
	}
}

func (stx *SerializedTransaction) String() string {
	bytes, err := json.Marshal(stx)
	if err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}
