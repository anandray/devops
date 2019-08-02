
# Wolk HTTP Signing

Every HTTP GET and HTTP POST/URL to a Wolk node requires a 192-character header `Sig` that represents a 96-byte concatenation of:
* 32 byte Public Key
* 64 byte Signature
Obviously, all signatures from the same private key will have the same public key sitting in the first 32 bytes, with the same address computed with the following:
```
func PubkeyToAddress(pubkey *PublicKey) common.Address {
	pkbytes := pubkey.Bytes()
	var addr common.Address
	copy(addr[:], wolkcommon.Computehash(pkbytes[1:])[12:])
	return addr
}
```

The Signature is composed of a "privateKey.Sign" operation using the client's private key with a variable number of bytes that depends on the kind of operation being performed:
1. in *ALL* GET operations, the bytes are the path _starting with "/"_ which can be seen in `client.go`'s `FetchURL` method
```
  msg := []byte(fmt.Sprintf("/%s", path))
	fmt.Printf("Fetch MSG: %x (%s)\n", msg, string(msg))
	sig, err := self.privateKey.Sign(msg)
	if err != nil {
		return output, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return output, err
	}
	req.Header.Add("Sig", fmt.Sprintf("%x", sig))
```
2. in HTTP PUT "operations, (e.g. postBucket /sourabh/test/banana.gif), 8 bytes representing the number of bytes being sent (in uint64/8 bytes form) is followed by the path, starting with "/", which can
```
sig, err := self.privateKey.Sign(msg)
if err != nil {
  fmt.Printf("ERR1 %v\n", err)
  return txhash, err
}
fmt.Printf("Size: %d ownerAddr: %x collection: %s key: %s\n", uint64(sz), ownerAddr, collection, key)
fmt.Printf("MSG:  %x\n", msg)
req.Header.Add("Sig", fmt.Sprintf("%x", sig))
req.Header.Add("Msg", fmt.Sprintf("%x", msg))
```
The above client operation results in a *virtual* transaction submission of `TransactionSetKey` inside the server
```
tx := NewTransactionSetKey(ownerAddr, collection, key, fileHash, uint64(sz))
// we have validated the signature already with a
tx.Sig = common.FromHex(r.Header.Get("Sig"))
s.wcs.SendRawTransaction(tx)
```

3. in Transaction submission operations, the bytes signed is the RLP-Encoded transaction, with 96 bytes standing in the "signature" slot is used.  In `client.go` a call to `tx.SignTx(self.privateKey)` does this inside `transaction.go`:
```
tx.Sig = make([]byte, crypto.SignatureLength)
sig, err := priv.Sign(tx.BytesWithoutSig())
if err != nil {
  return err
}
tx.Sig = make([]byte, crypto.SignatureLength)
copy(tx.Sig, sig)
```
The signature must be in the transaction for block to be verified.

# Wolk HTTP Verification

Wolk nodes check `Sig` headers in HTTP GET and HTTP PUT operations process them in a single `checkSig` function:
```
func (s *HttpServer) checkSig(w http.ResponseWriter, r *http.Request, checkType string, sz uint64, msg []byte) (signer common.Address, balance uint64, bandwidth int64, err error) {
	sig := common.FromHex(r.Header.Get("Sig"))
...
	pubkey := crypto.RecoverPubkey(sig)
  msgsupplied := common.FromHex(r.Header.Get("Msg"))
...
	err = pubkey.VerifySign(msg, sig)
```

1. In the HTTP GET case:
```
msg := []byte(r.URL.Path)
```
2. In the HTTP POST case:
```
msg := MakeSetKeySignBytes(uint64(sz), ownerAddr, collection, key)
```
If the signature is valid, then the virtual transaction is created using this exact signature

3. In Transaction cases, a `ValidateTx` operation is called for both the standard and virtual case (2) above before a transaction is included in the transaction pool:

```
func (tx *Transaction) ValidateTx() (bool, error) {
	if len(tx.Sig) != crypto.SignatureLength {
		return false, fmt.Errorf("Incorrect Sig length %d != %d", len(tx.Sig), crypto.SignatureLength)
	}
	pubkey := crypto.RecoverPubkey(tx.Sig)
	// depending on the transaction type, we have user signing different bytes
	var msg []byte
	if tx.TransactionType == TransactionSetKey {
		msg = MakeSetKeySignBytes(tx.Amount, tx.Recipient, string(tx.Collection), string(tx.Key))
	} else {
		msg = tx.BytesWithoutSig()
	}
	err := pubkey.VerifySign(msg, tx.Sig)
	if err != nil {
		return false, err
	}
	return true, nil
}
```
