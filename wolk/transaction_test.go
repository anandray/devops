// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
)

func init() {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
}

func TestInternalTx(t *testing.T) {

	rawtx := `{"transactionType":8,"payload":"eyJrZXkiOiJZbVZ5ZEdsbCIsImhhc2giOiIweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAifQ==","sig":"MEQCIAc6hSWfuYFkP6fMibyv0cEfVu4/ApzpehxBIAvBUG/ZAiCvVw50re9IPVkgjIhwxWY/xleTa/AlOSpxWa2qjsD7lw==","signer":"eyJrdHkiOiJFQyIsImNydiI6IlAtMjU2IiwieCI6ImV5V25HaGE3WUx2VTZOQzBBNDNGZ0lyY05JTElYWkVaR19Fd2NYNG5udU0iLCJ5IjoiNExKWDh1LXdldVdEcDNLeGxNYW51QW5iSUQ5cW9BTFd0N1JaRFRDaWpQYyJ9"}`
	var tx Transaction
	err := json.Unmarshal([]byte(rawtx), &tx)
	if err != nil {
		t.Fatalf("Unmarshal: %v\n", err)
	}

	validated, err := tx.ValidateTx()
	if err != nil {
		fmt.Printf("Error : %+v", err)
	}
	fmt.Printf("ValidateTx? %v\n", validated)

}

func TestSignTx(t *testing.T) {
	jwk := `{"kty":"EC","crv":"P-256","d":"l2_rYJ48HD0GcH2U7aaA7B6T23Nr3VI7NRUu7B_89qI","x":"eyWnGha7YLvU6NC0A43FgIrcNILIXZEZG_EwcX4nnuM","y":"4LJX8u-weuWDp3KxlManuAnbID9qoALWt7RZDTCijPc","key_ops":["sign"],"ext":true}`
	privateKeyObj, _ := crypto.JWKToECDSA(jwk)

	payloadName := NewTxAccount("random", []byte(""))
	rawTx, err := NewTransaction(privateKeyObj, http.MethodPost, path.Join(ProtocolName, PayloadTransfer, "bertie"), payloadName)
	if err != nil {
		fmt.Printf("Error : %+v", err)
	}
	b, err := json.Marshal(rawTx)
	fmt.Printf("Raw: %s\n", string(b))

	fmt.Printf("Serialized: %s\n", rawTx.String())

	validated, err := rawTx.ValidateTx()
	if err != nil {
		fmt.Printf("Error : %+v", err)
	}
	fmt.Printf("ValidateTx? %v\n", validated)
}

func TestTransaction(t *testing.T) {
	balance := uint64(1000000000000000000)
	owner := "arc"
	collection := "mygod"
	key := "12344321"
	privKey, err := crypto.JWKToECDSA(crypto.TestJWKPrivateKey)
	if err != nil {
		t.Fatalf("JWKToECDSA %v\n", err)
	}

	h := common.BytesToHash(wolkcommon.Computehash([]byte("[]")))
	for i := 0; i < 5; i++ {
		var p interface{}
		method := http.MethodPost
		var txpath string
		switch i {
		case 0:
			p = NewTxAccount("random", []byte(""))
			txpath = path.Join(ProtocolName, PayloadBucket, owner)
			break
		case 1:
			p = NewTxTransfer(balance, owner)
			txpath = path.Join(ProtocolName, PayloadTransfer)
			break
		case 2:
			p = NewTxNode("1.2.3.4", "6.7.8.9", 1, 12344321)
			txpath = path.Join(ProtocolName, PayloadNode, fmt.Sprintf("%d", 12))
			break
		case 3:
			p = NewTxKey(h, 42)
			method = http.MethodPut
			txpath = path.Join(owner, collection, key)
			break
		case 4:
			//data := "insert into account (accountID, email) values (2, 'johnny.appleseed@gmail.com')"
			txpath = path.Join(owner, collection, key)
			p = &SQLRequest{} // fill this in with data to be more realistic
			break
		}

		tx, err := NewTransaction(privKey, method, txpath, p)
		if err != nil {
			t.Fatalf("SignTx err %v", err)
		}
		tx_jsonencode, err := json.Marshal(tx)
		if err != nil {
			t.Fatalf("json Marshal: err %v", err)
		}

		txbytes := tx.Bytes()
		fmt.Printf("\nTx %d: %s\nHash: %x\nBytes: %x\n", i, tx.String(), tx.Hash(), tx.Bytes())

		var tx2 Transaction
		err = rlp.Decode(bytes.NewReader(txbytes), &tx2)
		if err != nil {
			t.Fatalf("Decode  err %v", err)
		}
		validated, err := tx2.ValidateTx()
		if err != nil {
			t.Fatalf("ValidateTx err %v", err)
		} else if !validated {
			t.Fatalf("ValidateTx NOT VALIDATED")
		}
		// fmt.Printf("\nRECO %d Tx: %s\nHash: %x\nBytes: %x\n", i, tx2.String(), tx2.Hash(),  tx2.Bytes())
		tx2_jsonencode, err := json.Marshal(tx2)
		if err != nil {
			t.Fatalf("json Marshal: err %v", err)
		}
		if bytes.Compare(tx_jsonencode, tx2_jsonencode) != 0 {
			t.Fatalf("json Marshal MISMATCH")
		}
		actualSigner := crypto.GetECDSAAddress(privKey)
		signer, err := tx2.GetSignerAddress()
		if err != nil {
			t.Fatalf("GetSigner err %v", err)
		}
		if bytes.Compare(signer.Bytes(), actualSigner.Bytes()) != 0 {
			t.Fatalf("Incorrect ADDRESS: %x != %x\n", signer, actualSigner)
		}

	}

}
