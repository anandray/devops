// Copyright 2018 Wolk Inc.  All rights reserved.
// This file is part of the Wolk Deep Blockchains library.
package client

import (
	"bufio"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/client/chunker"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk"
	"github.com/wolkdb/cloudstore/wolk/cloud"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	DefaultKeyDirectory = "/root/go/src/github.com/wolkdb/keys"
	//DefaultKeyDirectory = "/Users/alina/src/github.com/wolkdb/keys"
	defaultAccountFile = "default"
	privateKeyFile     = "private.key"
	publicKeyFile      = "public.key"
	friendsKeyFile     = "friends.key"
	personalKeyFile    = "personal.key"
	rsaPrivateKeyFile  = "id_rsa"
	rsaPublicKeyFile   = "id_rsa.pub"
)

type WolkClient struct {
	server     string
	httpclient *http.Client
	httpPort   uint16

	verbose    bool
	chunkQueue []cloud.RawChunk

	lastStatusCode int

	name                  string
	privateKey            *ecdsa.PrivateKey // used for blockchain Tx
	address               common.Address    // derived from privateKey
	rsaPrivateKey         *rsa.PrivateKey   // used to share publicDecryptionKey, send mail
	personalDecryptionKey []byte            // shared with no one, used for Collections where encryption=EncryptionPersonal
	friendsDecryptionKey  []byte            // shared with friends, used for Collections where encryption=EncryptionPublic
	keysLoaded            bool
	keyDirectory          string
}

var DefaultTransport http.RoundTripper = &http.Transport{
	Dial: (&net.Dialer{
		// limits the time spent establishing a TCP connection (if a new one is needed)
		Timeout:   5 * time.Second,
		KeepAlive: 3 * time.Second, // 60 * time.Second,
	}).Dial,
	//MaxIdleConns: 5,
	MaxIdleConnsPerHost: 25, // changed from 100 -> 25

	// limits the time spent reading the headers of the response.
	ResponseHeaderTimeout: 5 * time.Second,
	IdleConnTimeout:       4 * time.Second, // 90 * time.Second,

	// limits the time the client will wait between sending the request headers when including an Expect: 100-continue and receiving the go-ahead to send the body.
	ExpectContinueTimeout: 1 * time.Second,

	// limits the time spent performing the TLS handshake.
	TLSHandshakeTimeout: 5 * time.Second,
}

func NewWolkClient(_server string, _httpPort int, _name string) (wclient *WolkClient, err error) {
	wclient = &WolkClient{
		httpclient:   &http.Client{Timeout: time.Second * 5, Transport: DefaultTransport},
		server:       _server,
		httpPort:     uint16(_httpPort),
		verbose:      false,
		keysLoaded:   false,
		keyDirectory: DefaultKeyDirectory,
		name:         _name,
	}
	if _name == "" {
		name, ok, err := wclient.LoadDefaultAccount()
		if err != nil {
			return wclient, fmt.Errorf("[client:NewWolkClient] %s", err)
		} else if !ok {
			return wclient, nil
		}
		log.Trace("NewWolkClient", "name", name)
		wclient.name = name
	} else {
		err = wclient.LoadAccount(_name)
		if err != nil {
			return wclient, fmt.Errorf("[client:NewWolkClient] %s", err)
		}
	}
	return wclient, nil
}

func (wclient *WolkClient) SetVerbose(verbose bool) {
	wclient.verbose = verbose
}

func (wclient *WolkClient) GetSelfName() string {
	return wclient.name
}

func (wclient *WolkClient) getScheme() (scheme string) {
	return "https"
}

func (wclient *WolkClient) HTTPUrl() (url string) {
	return fmt.Sprintf("%s://%s:%d", wclient.getScheme(), wclient.server, wclient.httpPort)
}

func (wclient *WolkClient) GetKeyDirectory() string {
	return wclient.keyDirectory
}

func (wclient *WolkClient) SetDefaultAccount(name string) (err error) {
	keyDir := wclient.GetKeyDirectory()
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		return fmt.Errorf("[client:SetDefaultAccount] No KeyDir %s", keyDir)
	}
	accountDir := filepath.Join(keyDir, name)
	if _, err := os.Stat(accountDir); os.IsNotExist(err) {
		return fmt.Errorf("[client:SetDefaultAccount] No account directory %s", accountDir)
	}

	err = ioutil.WriteFile(filepath.Join(keyDir, defaultAccountFile), []byte(name), 0600)
	if err != nil {
		return fmt.Errorf("[client:SetDefaultAccount] %s", err)
	}
	return nil
}

func (wclient *WolkClient) GetDefaultAccount() (name string, ok bool, err error) {
	keyDir := wclient.GetKeyDirectory()
	//log.Trace("[client:GetDefaultAccount]", "keyDir", keyDir)
	//log.Trace("[client:GetDefaultAccount]", "defaultAccountFile", defaultAccountFile)
	nameFile := filepath.Join(keyDir, defaultAccountFile)
	//log.Trace("[client:GetDefaultAccount]", "nameFile", nameFile)
	if _, err := os.Stat(nameFile); os.IsNotExist(err) {
		return name, false, nil
	}
	nameBytes, err := ioutil.ReadFile(nameFile)
	if err != nil {
		return name, false, fmt.Errorf("[client:GetDefaultAccount] %s", err)
	}
	name = string(nameBytes)
	//fmt.Printf("[client:GetDefaultAccount] name (%s)\n", name)
	return name, true, nil
}

func (wclient *WolkClient) LoadDefaultAccount() (name string, ok bool, err error) {
	name, ok, err = wclient.GetDefaultAccount()
	if err != nil {
		return name, false, fmt.Errorf("[client:LoadDefaultAccount] %s", err)
	} else if !ok {
		log.Error("[client:LoadDefaultAccount] - no default account")
		return name, false, nil
	}
	err = wclient.LoadAccount(name)
	if err != nil {
		return name, false, fmt.Errorf("[client:LoadDefaultAccount] LoadAccount(%s) - %s", name, err)
	}
	return name, true, nil
}

func ReadKeyFromFile(loadFileFrom string) (k []byte, err error) {
	jwk, err := ioutil.ReadFile(loadFileFrom)
	if err != nil {
		return k, err
	}

	var j jose.JSONWebKey
	err = j.UnmarshalJSON([]byte(jwk))
	if err != nil {
		return k, fmt.Errorf("UnmarshalJSON(%s): %v", loadFileFrom, err)
	}

	switch key := j.Key.(type) {
	case []byte:
		k = key
		return k, nil
	default:
		return k, fmt.Errorf("Unknown key type: %T", key)
	}
}

func (wclient *WolkClient) LoadAccount(name string) error {

	keyDir := wclient.GetKeyDirectory()
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		return fmt.Errorf("[client:LoadAccount] No KeyDir %s", keyDir)
	}
	accountDir := filepath.Join(keyDir, name)
	if _, err := os.Stat(accountDir); os.IsNotExist(err) {
		return fmt.Errorf("[client:LoadAccount] No account directory %s", accountDir)
	}
	// Private Key - ECDSA key to submit transactions
	fn := filepath.Join(accountDir, privateKeyFile)
	jwk, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	var j jose.JSONWebKey
	err = j.UnmarshalJSON([]byte(jwk))
	if err != nil {
		return fmt.Errorf("[client:LoadAccount] UnmarshalJSON(%s): %v", fn, err)
	}
	// check key type
	switch privateKey := j.Key.(type) {
	case *ecdsa.PrivateKey:
		wclient.privateKey = privateKey
		wclient.address = crypto.GetECDSAAddress(privateKey)
	default:
		return fmt.Errorf("[client:LoadAccount] Unknown privateKey type %T", privateKey)
	}
	// fmt.Printf("ADDRESS: %x\n", wclient.address)

	// Friends Key - key for sharing collections with friends

	friendsKey, err := ReadKeyFromFile(filepath.Join(accountDir, friendsKeyFile))
	if err != nil {
		return fmt.Errorf("[client:LoadAccount] ReadKeyFromFile %v", err)
	}
	wclient.friendsDecryptionKey = friendsKey

	// Personal Key - key for storing data only for you
	personalKey, err := ReadKeyFromFile(filepath.Join(accountDir, personalKeyFile))
	if err != nil {
		return fmt.Errorf("[client:LoadAccount] ReadKeyFromFile %v", err)
	}
	wclient.personalDecryptionKey = personalKey

	// RSA Key - Asymmetric encryption
	fn = filepath.Join(accountDir, rsaPrivateKeyFile)
	jwk, err = ioutil.ReadFile(fn)
	if err != nil {
		return fmt.Errorf("[client:LoadAccount] ReadFile(%s): %v", fn, err)
	}
	err = j.UnmarshalJSON([]byte(jwk))
	if err != nil {
		return fmt.Errorf("[client:LoadAccount] UnmarshalJSON(%s): %v", fn, err)
	}

	switch rsaPrivateKey := j.Key.(type) {
	case *rsa.PrivateKey:
		wclient.rsaPrivateKey = rsaPrivateKey
	default:
		return fmt.Errorf("[client:LoadAccount] Unknown RSA PrivateKey type %T", rsaPrivateKey)
	}

	wclient.keysLoaded = true
	return nil
}

func (wclient *WolkClient) KeysLoaded() bool {
	return wclient.keysLoaded
}

// stores keys in local directory but keeps RSA Public Key in Account; GetAccount shows RSA Public Key
func (wclient *WolkClient) CreateAccount(name string) (txhash common.Hash, err error) {

	keyDir := wclient.GetKeyDirectory()
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		err = os.Mkdir(keyDir, 0777)
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] Failure to Mkdir %v", err)
		}
	}

	accountDir := filepath.Join(keyDir, name)
	if _, err := os.Stat(accountDir); os.IsNotExist(err) {
		fmt.Printf("%s directory does not exist\n", accountDir)
		err = os.Mkdir(accountDir, 0777)
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] Failure to Mkdir %v", err)
		}

		// Private Key - ECDSA key to submit transactions
		privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		var jwk jose.JSONWebKey
		jwk.Key = privateKey
		b, err := jwk.MarshalJSON()
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}
		ioutil.WriteFile(filepath.Join(accountDir, privateKeyFile), b, 0600)

		jwk.Key = &(privateKey.PublicKey)
		b, err = jwk.MarshalJSON()
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}
		err = ioutil.WriteFile(filepath.Join(accountDir, publicKeyFile), b, 0600)
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}

		// Friends Key - key for sharing collections with friends
		friendsKey := crypto.GenerateRandomKey(32)
		jwk.Key = []byte(friendsKey)
		b, err = jwk.MarshalJSON()
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}
		err = ioutil.WriteFile(filepath.Join(accountDir, friendsKeyFile), b, 0600)
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}

		// Personal Key - key for sharing collections with friends
		personalKey := crypto.GenerateRandomKey(32)
		jwk.Key = []byte(personalKey)
		b, err = jwk.MarshalJSON()
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}
		err = ioutil.WriteFile(filepath.Join(accountDir, personalKeyFile), b, 0600)
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}

		// RSA Key - Asymmetric encryption
		rng := rand.Reader
		rsaPrivateKey, err := rsa.GenerateKey(rng, crypto.BitSize)
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}

		jwk.Key = rsaPrivateKey
		b, err = jwk.MarshalJSON()
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}
		ioutil.WriteFile(filepath.Join(accountDir, rsaPrivateKeyFile), b, 0600)

		jwk.Key = rsaPrivateKey.Public()
		b, err = jwk.MarshalJSON()
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}
		ioutil.WriteFile(filepath.Join(accountDir, rsaPublicKeyFile), b, 0600)
		err = wclient.LoadAccount(name)
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}
	} else {
		// it exists already
		fmt.Printf("%s directory exists, loading account...\n", accountDir)
		err = wclient.LoadAccount(name)
		if err != nil {
			return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
		}
		addr, err := wclient.GetName(name, nil)
		if err != nil {
			if wclient.lastStatusCode == 404 {
				fmt.Printf("[client:CreateAccount] Account %s is not registered\n", name)
			} else {
				return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
			}
		} else if bytes.Compare(wclient.address.Bytes(), addr.Bytes()) == 0 {
			return txhash, fmt.Errorf("[client:CreateAccount] Account exists already on blockchain")
		}
	}

	// get the rsa public key
	rsaBytes, err := ioutil.ReadFile(filepath.Join(accountDir, rsaPublicKeyFile))
	if err != nil {
		return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
	}

	// submit the name + RSA Public Key to the blockchain, submitting a system bucket!
	txhash, err = wclient.post(name, wolk.NewTxAccount(name, rsaBytes))
	if err != nil {
		return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
	}

	_, ok, err := wclient.GetDefaultAccount()
	if err != nil {
		log.Error("[client:CreateAccount] GetDefaultAccount", "err", err)
		return txhash, fmt.Errorf("[client:CreateAccount] %s", err)
	} else if !ok {
		wclient.SetDefaultAccount(name)
	}
	return txhash, nil
}

func (wclient *WolkClient) DumpAccount() {
	fmt.Printf("PersonalKey %x\n", wclient.personalDecryptionKey)
	fmt.Printf("FriendsKey %x\n", wclient.friendsDecryptionKey)
	fmt.Printf("RSAKey [PRIVATE] %s\n", crypto.RSAPrivateKeyToString(wclient.rsaPrivateKey))

	privKeyAsJWK, _ := crypto.ECDSAPrivateKeyToJSONWebKey(wclient.privateKey)
	rsaPrivKeyAsJWK, _ := crypto.RSAPublicKeyToString(&wclient.rsaPrivateKey.PublicKey)
	pubKeyAsJWK, _ := crypto.ECDSAPublicKeyToJSONWebKey(&wclient.privateKey.PublicKey)

	fmt.Printf("ECDSAKey [PRIVATE] %s\n", privKeyAsJWK)
	fmt.Printf("RSAKey [PUBLIC] %s\n", rsaPrivKeyAsJWK)
	fmt.Printf("ECDSAKey [PUBLIC] %s\n", pubKeyAsJWK)
}

func (wclient *WolkClient) GetRSAPublicKey() *rsa.PublicKey {
	return &(wclient.rsaPrivateKey.PublicKey)
}

func (wclient *WolkClient) Encrypt(chunk []byte, options *wolk.RequestOptions) (encryptedChunk []byte, err error) {
	if options != nil {
		encryption := options.Encryption
		if encryption == wolk.EncryptionFriends {
			return wclient.encrypt(chunk, wclient.friendsDecryptionKey)
		} else if encryption == wolk.EncryptionPersonal {
			return wclient.encrypt(chunk, wclient.personalDecryptionKey)
		}
	}
	return chunk, nil
}

func (wclient *WolkClient) encrypt(chunk []byte, key []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return ciphertext, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext = make([]byte, aes.BlockSize+len(chunk))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return ciphertext, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], chunk)
	/*
		padding := len(chunk)
		initCtr := uint32(0)
		enc := crypto.New(key, padding, initCtr, sha256.New()
		encryptedChunk, err := enc.Encrypt(chunk)
		if err == nil {
			return encryptedChunk
		}*/

	return ciphertext, nil
}

func (wclient *WolkClient) Decrypt(ciphertext []byte, options *wolk.RequestOptions) (chunk []byte, err error) {
	if options != nil {
		encryption := options.Encryption
		if encryption == wolk.EncryptionFriends {
			return wclient.decrypt(ciphertext, wclient.friendsDecryptionKey)
		} else if encryption == wolk.EncryptionPersonal {
			return wclient.decrypt(ciphertext, wclient.personalDecryptionKey)
		}
	}
	return ciphertext, nil
}

func (wclient *WolkClient) decrypt(ciphertext []byte, key []byte) (chunk []byte, err error) {
	/*
		padding := len(encryptedChunk)
		initCtr := uint32(0)
		enc := crypto.New(key, padding, initCtr, sha256.New()
		decryptedChunk, err := enc.Decrypt(encryptedChunk)
		if err == nil {
			return decryptedChunk
		}
	*/
	block, err := aes.NewCipher(key)
	if err != nil {
		return chunk, err
	}

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	iv := ciphertext[:aes.BlockSize]
	// CTR mode is the same for both encryption and decryption, so we can
	// also decrypt that ciphertext with NewCTR.
	chunk = make([]byte, len(ciphertext)-aes.BlockSize)
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(chunk, ciphertext[aes.BlockSize:])
	return chunk, nil
}

func (wclient *WolkClient) GetChunk(ctx context.Context, chunkID []byte, options *wolk.RequestOptions) (chunk []byte, ok bool, err error) {
	output, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadChunk, fmt.Sprintf("%x", chunkID)), nil)
	if err != nil {
		return chunk, false, err
	}
	chunkIDtest := wolkcommon.Computehash(output)
	if bytes.Compare(chunkID, chunkIDtest) != 0 {
		return output, false, fmt.Errorf("[client:GetChunk] chunk not found")
	}
	chunk, err = wclient.Decrypt(output, options)
	if err != nil {
		return chunk, ok, err
	}
	return chunk, true, nil
}

func (wclient *WolkClient) GetShare(ctx context.Context, chunkID []byte) (chunk []byte, ok bool, err error) {
	output, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadShare, fmt.Sprintf("%x", chunkID)), nil)
	chunkIDtest := wolkcommon.Computehash(output)
	if bytes.Compare(chunkID, chunkIDtest) != 0 {
		return output, false, fmt.Errorf("[client:GetShare] chunk not found")
	}
	return chunk, true, nil
}

func (wclient *WolkClient) SignRequest(req *http.Request, body []byte) (err error) {
	jwkRequester, err := wclient.GetJWKRequester()
	if err != nil {
		return err
	}
	req.Header.Add("Requester", string(jwkRequester))
	msg := wolk.PayloadBytes(req, body)
	sig, err := crypto.JWKSignECDSA(wclient.privateKey, msg)
	if err != nil {
		return err
	}
	req.Header.Add("Sig", fmt.Sprintf("%x", sig))
	req.Header.Add("Msg", string(msg))
	if wclient.verbose {
		fmt.Printf("Requester: %s\n", string(jwkRequester))
		fmt.Printf("Sig: %x\n", sig)
		fmt.Printf("Msg: %s\n", msg)
	}
	return nil
}

func (wclient *WolkClient) SetShare(ctx context.Context, chunk []byte) (chunkHash common.Hash, err error) {
	chunkID := wolkcommon.Computehash(chunk)
	chunkHash = common.BytesToHash(chunkID)
	log.Trace("*****SetChunk", "chunkID", fmt.Sprintf("%x", chunkID), "len(chunk)", len(chunk))

	body := bytes.NewReader(chunk)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/wolk/share", wclient.HTTPUrl()), body)
	err = wclient.SignRequest(req, chunk)
	if err != nil {
		return chunkHash, err
	}
	resp, err := wclient.httpclient.Do(req)
	if err != nil {
		return chunkHash, err
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return chunkHash, err
	}
	return chunkHash, nil
}

func (wclient *WolkClient) SetChunk(ctx context.Context, chunk []byte, options *wolk.RequestOptions) (chunkHash common.Hash, err error) {
	input, err := wclient.Encrypt(chunk, options)
	if err != nil {
		return chunkHash, err
	}
	chunkID := wolkcommon.Computehash(input)
	chunkHash = common.BytesToHash(chunkID)
	log.Trace("*****SetChunk", "chunkHash", chunkHash, "len(input)", len(input))

	body := bytes.NewReader(input)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/wolk/chunk", wclient.HTTPUrl()), body)
	if err != nil {
		return chunkHash, err
	}

	err = wclient.SignRequest(req, input)
	if err != nil {
		return chunkHash, err
	}
	resp, err := wclient.httpclient.Do(req)
	if err != nil {
		return chunkHash, err
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return chunkHash, err
	}
	return chunkHash, nil
}

func (wclient *WolkClient) SetChunkBatch(ctx context.Context, chunks []*cloud.RawChunk) (err error) {
	log.Trace("*****SetChunkBatch", "len(chunks)", len(chunks))

	jchunks, err := json.Marshal(chunks)
	body := bytes.NewReader(jchunks)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/wolk/setbatch", wclient.HTTPUrl()), body)
	err = wclient.SignRequest(req, jchunks)
	if err != nil {
		return err
	}
	req.Header.Add("cache-control", "no-cache")
	resp, err := wclient.httpclient.Do(req)
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (wclient *WolkClient) GetChunkBatch(ctx context.Context, reqChunks []*cloud.RawChunk) (respChunks []*cloud.RawChunk, err error) {
	log.Trace("*****GetChunkBatch", "len(chunks)", len(reqChunks))
	jchunks, err := json.Marshal(&reqChunks)
	body := bytes.NewReader(jchunks)

	url := fmt.Sprintf("%s/wolk/getbatch", wclient.HTTPUrl())
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return respChunks, err
	}
	err = wclient.SignRequest(req, jchunks)
	if err != nil {
		return respChunks, err
	}

	resp, err := wclient.httpclient.Do(req)
	if err != nil {
		return respChunks, fmt.Errorf("GetChunkBatch-Do-Err %v", err)
	}
	body2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return respChunks, fmt.Errorf("GetChunkBatch-ReadAll-Err %v", err)
	}
	err = json.Unmarshal(body2, &respChunks)
	if err != nil {
		return respChunks, fmt.Errorf("GetChunkBatch-Unmarshal-Err %v", err)
	}
	return respChunks, nil
}

func (wclient *WolkClient) verifyContentHash(contentHash common.Hash, content []byte) bool {
	// TODO: check that contentHash  matches content (e.g. wcloud cat 35e3ec4c3c13873051ed33b6534d33870c56458eb4fc192a66c905cc7d4c9c7a == "yellow")
	return true
}

/*
Given a NoSQLProof object provided by a Cloudstore Node such as this, returns validation errors or nil:
{"BlockNumber":3507,
"BlockHash":"0x0ee06ec22b45b6f1309c401cb7f1437c93b20fa3227c08b2d366fad4fa2524fe",
"KeyChunkHash":"0x08b4a0e79be3b61e194285a6b7a93737baa5adcbcc37652ed22853f7b2e56165",
"KeyMerkleRoot":"0x7b1c2306262ac9b2a4e4fa3f3b051cef9994f16089f3dda889c70289867e32b1",
"Owner":"0x545df6fd6811dc32397eae8d218fc4b29681eb62",
"SystemHash":"0x64bebb50a024e206d83a835d539d73d20e5b9bae",
"SystemProof":{"SMTTreeDepth":160,"Key":"64bebb50a024e206d83a835d539d73d20e5b9bae","Proof":["dc847e0f0f77cc12bcd060f2a177fa23cd1a1c3d124a13f05367919668a4a903"],"ProofBits":"8000000000000000000000000000000000000000"},
"SystemChunkHash":"0xd0b008121addf63a0bc372c5ef402c293f89a48f5f5f4fda17493883daf17214",
"SystemMerkleRoot":"0x43b5bcb063d54f9924ac7b2a5b5f08cb8abebe3b13c2439b79326de8812d25ae",
"Collection":"fruits","CollectionHash":"0xf2df7ed8d3e41636d4867d3a8a42b324f73f6038",
"CollectionProof":{"SMTTreeDepth":160,"Key":"f2df7ed8d3e41636d4867d3a8a42b324f73f6038","Proof":["2ebda56aa874607093800d11cd4bc477a82b303350ea7c395acc5ce0b911252f"],"ProofBits":"8000000000000000000000000000000000000000"},
"CollectionChunkHash":"0xfa6418d41ca68b274a5f743b1d37fb147b0d5c5b50bd981ed2d6f02ddf53beb3","CollectionMerkleRoot":"0x4336ef5be9f8fc9c389a0aefd78ab2335fbeecec07469cd921860f6e9843f025",
"Key":"banana",
"KeyHash":"0x21fc98cf70265ae93feb6ced2d2b03192146726c",
"TxHash":"0xddec32bbb56126a6309e84c560ae5c6d307cb70f4a3d08294028d7fc0b4561b6",
"Tx":{"transactionType":6,"recipient":"0x545df6fd6811dc32397eae8d218fc4b29681eb62","amount":6,"node":1549845070,"hash":"0x35e3ec4c3c13873051ed33b6534d33870c56458eb4fc192a66c905cc7d4c9c7a","collection":"fruits","key":"banana","sig":"272dd9aa9c4845f7b7c17bfea3c0fe6812a45051f13736ce2f1e72ebf70fe29ff43ae605f4dc8549be663f0b431df3ef168502f1d8dcf750f3ad50e13d7902bffa4b27312a04e2525efb2922fc26479b91b3a7d5c92d9b148881debf43d9f301","txhash":"0xddec32bbb56126a6309e84c560ae5c6d307cb70f4a3d08294028d7fc0b4","blockNumber":0,"signer":"0x545df6fd6811dc32397eae8d218fc4b29681eb62"},
"KeyProof":{"SMTTreeDepth":160,"Key":"21fc98cf70265ae93feb6ced2d2b03192146726c","Proof":["a42bc302cf370c9f8f385bfc43d8e94d03c2cc685bc13aac28625e57502deb78"],"ProofBits":"8000000000000000000000000000000000000000"}}
*/
func (wclient *WolkClient) verifyNoSQLProof(path string, pr *wolk.NoSQLProof, resp *http.Response, proofType string, output []byte) (err error) {
	if proofType == "NoSQLScan" {
		err = pr.VerifyScanProofs()
		if err != nil {
			log.Error("NoSQLScan Proof NOT Verified", "err", err)
			return err
		}
		return nil
	}
	return pr.Verify()
}

func (wclient *WolkClient) GetJWKRequester() (output []byte, err error) {
	if wclient.privateKey == nil {
		return output, fmt.Errorf("No keys created")
	}
	var jwk jose.JSONWebKey
	jwk.Key = &(wclient.privateKey.PublicKey)
	jwkRequester, err := jwk.MarshalJSON()
	if err != nil {
		return output, err
	}
	return jwkRequester, nil
}

func (wclient *WolkClient) patch(path string, txp interface{}) (output []byte, err error) {
	//log.Info("[client:patch]", "path", path, "txp", txp)
	bodyPost, err := json.Marshal(txp)
	if err != nil {
		return output, fmt.Errorf("[client:patch] %s", err)
	}
	return wclient.wrequest(http.MethodPatch, path, bodyPost, nil)
}

func (wclient *WolkClient) post(path string, txp interface{}) (txhash common.Hash, err error) {
	bodyPost, err := json.Marshal(txp)
	if err != nil {
		return txhash, err
	}
	output, err := wclient.wrequest(http.MethodPost, path, bodyPost, nil)
	return common.HexToHash(string(output)), nil
}

func (wclient *WolkClient) put(path string, body []byte) (txhash common.Hash, err error) {
	output, err := wclient.wrequest(http.MethodPut, path, body, nil)
	if err != nil {
		//log.Error("[client:put] wrequest", "err", err)
		return txhash, err
	}
	//log.Trace("[client:put] SUCCESS", "output", string(output))
	return common.HexToHash(string(output)), nil
}

func (wclient *WolkClient) get(path string, options *wolk.RequestOptions) (output []byte, err error) {
	return wclient.wrequest(http.MethodGet, path, []byte(""), options)
}

func (wclient *WolkClient) Delete(path string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	output, err := wclient.wrequest(http.MethodDelete, path, []byte(""), options)
	if err != nil {
		return txhash, err
	}
	return common.HexToHash(string(output)), nil
}

func (wclient *WolkClient) wrequest(method string, path string, body []byte, options *wolk.RequestOptions) (output []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	url := fmt.Sprintf("%s://%s:%d/%s", wclient.getScheme(), wclient.server, wclient.httpPort, path)
	if options != nil && len(options.History) > 0 {
		url = url + "?history=" + options.History
	}
	//log.Info("[client:wrequest]", "url", url, "body", string(body))
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest(method, url, bodyReader)
	req.Cancel = ctx.Done()
	if err != nil {
		log.Error("[client:wrequest] NewRequest | Error", "err", err)
		return output, fmt.Errorf("[client:wrequest] %s", err)
	}

	err = wclient.SignRequest(req, body)
	if err != nil {
		log.Error("[client:wrequest] SignRequest | Error", "err", err)
		return output, fmt.Errorf("[client:wrequest] %s", err)
	}

	if options != nil {
		if options.MaxContentLength > 0 {
			req.Header.Add("MaxContentLength", fmt.Sprintf("%d", options.MaxContentLength))
		}
		if options.BlockNumber > 0 {
			req.Header.Add("BlockNumber", fmt.Sprintf("%d", options.BlockNumber))
		}
		if options.Proof {
			req.Header.Add("Proof", "1")
		}
		if len(options.WaitForTx) > 0 {
			req.Header.Add("WaitForTx", "1")
		}
	}
	//req.Cancel = ctx.Done()
	resp, err := wclient.httpclient.Do(req)
	if err != nil {
		log.Error("[client:wrequest] httpclient.Do | Error", "err", err)
		return output, fmt.Errorf("[client:wrequest] %s", err)
	}
	if wclient.verbose {
		for key, val := range resp.Header {
			fmt.Printf("%s %v\n", key, val)
		}
	}
	output, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("[client:wrequest] ReadAll | Error", "err", err)
		return output, fmt.Errorf("[client:wrequest] %s", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		wclient.lastStatusCode = resp.StatusCode
		return output, fmt.Errorf("Status Code %d - %s", resp.StatusCode, output)
	}
	//log.Info("[client:wrequest]", "output", string(output))
	if options != nil && len(options.History) > 0 {
		var history []*wolk.NoSQLProof
		err := json.Unmarshal(output, &history)
		if err != nil {
			log.Error("[client:wrequest] json.Unmarshal(p)", "err", err)
			return output, fmt.Errorf("History Unmarshal ERR: %v", err)
		}
		for i, pr := range history {
			proofType := "NoSQL"
			err = wclient.verifyNoSQLProof(path, pr, resp, proofType, output)
			if err != nil {
				log.Error("[client:wrequest] History Proof - FAILURE", "err", err)
				return output, fmt.Errorf("History verifyNoSQLProof ERR: %v", err)
			}
			log.Info("[client:wrequest] VERIFIED NoSQL History Proof", "i", i, "proofType", proofType)
		}

	} else if options != nil && options.Proof {
		p := resp.Header.Get("Proof")
		proofType := resp.Header.Get("Proof-Type")
		if len(p) > 0 {
			// int(pr.Proof.SMTTreeDepth))
			if proofType == "SMT" {
				var pr wolk.SMTProof
				defaultHashes := wolk.ComputeDefaultHashes()
				err := json.Unmarshal([]byte(p), &pr)
				if err != nil {
					log.Error("[client:wrequest] json.Unmarshal(p)", "err", err)
					return output, fmt.Errorf("Proof Unmarshal ERR: %v", err)
				}
				proof := wolk.DeserializeProof(pr.Proof)
				// Verify Proof Process: at proof.Key position there is the value pr.TxHash, which hashes up to pr.MerkleRoot
				if proof.Check(pr.TxHash.Bytes(), pr.MerkleRoot.Bytes(), defaultHashes, false) {
					if wclient.verbose {
						log.Info("[client:wrequest] Proof VERIFIED", "pr", pr.String())
					}
				} else {
					log.Error("[client:wrequest] Proof FAILED", "pr", pr.String())
					return output, fmt.Errorf("Proof FAILED %v", err)
				}
			} else if proofType == "NoSQL" || proofType == "NoSQLScan" {
				var pr wolk.NoSQLProof
				err := json.Unmarshal([]byte(p), &pr)
				if err != nil {
					log.Error("[client:wrequest] json.Unmarshal(p)", "err", err)
					return output, fmt.Errorf("Proof Unmarshal ERR: %v", err)
				}
				err = wclient.verifyNoSQLProof(path, &pr, resp, proofType, output)
				if err != nil {
					log.Error("[client:wrequest] NoSQLScan Proof - FAILURE", "err", err)
					return output, fmt.Errorf("verifyNoSQLProof ERR: %v", err)
				}
				log.Info("[client:wrequest] VERIFIED NoSQL Proof", "proofType", proofType)
			} else {
				return output, fmt.Errorf("[client:wrequest] Proof requested, but unknown Proof-Type supplied [%s]", proofType)
			}
		} else {
			return output, fmt.Errorf("[client:wrequest] Proof requested, but not returned")
		}
	}
	return output, nil
}

func (wclient *WolkClient) LatestBlockNumber() (blockNumber int, err error) {
	resp, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadBlock, "latest"), nil)
	if err != nil {
		return blockNumber, err
	}
	blockNumber, err = strconv.Atoi(string(resp))
	return blockNumber, err
}

// how is this used? we have LoadAccount and GetName and CreateAccount already...
func (wclient *WolkClient) GetAccount(name string, options *wolk.RequestOptions) (account *wolk.Account, err error) {
	resp, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadAccount, name), options)
	if err != nil {
		return account, err
	}
	account = new(wolk.Account)
	err = json.Unmarshal(resp, account)
	return account, err
}

func (wclient *WolkClient) WaitForTransaction(txhash common.Hash) (tx *wolk.SerializedTransaction, err error) {
	done := false
	tries := 0
	for done == false {
		tx, err := wclient.GetTransaction(txhash)
		if err != nil {
		} else if tx != nil && tx.BlockNumber > 0 {
			log.Info("[client:WaitForTransaction] FOUND", "txhash", txhash)
			return tx, nil
		} else {
			time.Sleep(2500 * time.Millisecond)
			log.Warn("[client:WaitForTransaction] NOT FOUND", "tx", tx)
		}
		tries = tries + 1
		if tries > 150 {
			log.Error("[client:WaitForTransaction]", "txhash", txhash, "err", err, "tries", tries)
			return tx, fmt.Errorf("Not included!")
		}
	}
	return tx, err
}

func (wclient *WolkClient) GetTransaction(txhash common.Hash) (tx *wolk.SerializedTransaction, err error) {
	resp, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadTx, fmt.Sprintf("%x", txhash)), nil)
	if err != nil {
		return tx, err
	}
	tx = new(wolk.SerializedTransaction)
	err = json.Unmarshal(resp, &tx)
	return tx, err
}

func (wclient *WolkClient) GetBlock(blockNumber int) (block *wolk.SerializedBlock, err error) {
	resp, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadBlock, fmt.Sprintf("%d", blockNumber)), nil)
	if err != nil {
		return block, err
	}
	var b wolk.SerializedBlock
	err = json.Unmarshal(resp, &b)
	return &b, err
}

func (wclient *WolkClient) GetPeers() (numpeers int, err error) {
	resp, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadInfo), nil)
	if err != nil {
		return numpeers, err
	}
	n, err := strconv.Atoi(string(resp))
	if err != nil {
		return numpeers, err
	}
	return n, err
}

func (wclient *WolkClient) GetNode(nodenumber int, options *wolk.RequestOptions) (node *wolk.SerializedNode, err error) {
	resp, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadNode, fmt.Sprintf("%d", nodenumber)), nil)
	if err != nil {
		return node, err
	}
	node = new(wolk.SerializedNode)
	err = json.Unmarshal(resp, &node)
	return node, err
}

// NoSQL
func (wclient *WolkClient) SetKey(owner string, collection string, key string, val []byte) (txhash common.Hash, err error) {
	return wclient.put(path.Join(owner, collection, key), val)
}

func (wclient *WolkClient) SetQuota(quota uint64) (output []byte, err error) {
	//NOTE: for now, you can only set Quota for the owner you're logged in as
	return wclient.patch(wclient.name, &wolk.TxBucket{Name: wclient.name, Quota: quota})
}

func (wclient *WolkClient) SetShimURL(owner string, collection string, shimURL string) (output []byte, err error) {
	//NOTE: for now, you can only set Quota for the owner you're logged in as
	return wclient.patch(path.Join(owner, collection), &wolk.TxBucket{Name: collection, ShimURL: shimURL})
}

func (wclient *WolkClient) GetName(name string, options *wolk.RequestOptions) (address common.Address, err error) {
	resp, err := wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadName, name), options)
	if err != nil {
		return address, err
	}
	address = common.HexToAddress(string(resp))
	return address, nil
}

func (wclient *WolkClient) GetKey(owner string, collection string, key string, options *wolk.RequestOptions) (resp []byte, sz uint64, err error) {
	resp, err = wclient.get(path.Join(owner, collection, key), options)
	if err != nil {
		return resp, sz, err
	}

	// rsa:
	//  SetKey(wolk://owner/collection/key, Value)
	//   1. client.go/wolk.js: owner RSAPublicKey is used  map Value into EncryptedValue, so Wolk node does not know what is being sent
	//   2. http.go:   receives *EncryptedValue* from (1), and records in collection
	//  GetKey(wolk://owner/collection/key)
	//   1. http.go:   returns *EncryptedValue*
	//   2. client.go: if owner, then user can use RSAPrivateKey to decrypt data in GetKey, ScanCollection
	// symmetric:
	//  SetKey(wolk://owner/collection/key, Value)
	//   1. client.go/wolk.js:
	//   2. http.go:

	return resp, uint64(len(resp)), err
}

// File
func ParseWolkUrl(wolkurl string) (owner string, collection string, key string, err error) {
	u, err := url.Parse(wolkurl)
	if err != nil {
		return owner, collection, key, err
	}
	if u.Scheme != wolk.ProtocolName {
		return owner, collection, key, fmt.Errorf("Not wolk url")
	}
	owner = u.Host
	pieces := strings.Split(u.Path, "/")
	if len(pieces) < 1 {
		return owner, collection, key, fmt.Errorf("Invalid url")
	}
	if len(pieces) > 1 {
		collection = strings.Trim(pieces[1], "/")
	} else {
		collection = ""
	}
	if len(pieces) > 2 {
		key = strings.Join(pieces[2:], "/")
	} else {
		key = ""
	}
	return owner, collection, key, nil
}

func (wclient *WolkClient) AddFile(fn string, options *wolk.RequestOptions) (fileHash common.Hash, fileSize uint64, err error) {
	file, err := os.Open(fn)
	if err != nil {
		return fileHash, fileSize, err
	}
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return fileHash, fileSize, fmt.Errorf("File does not exist %s", fn)
	}
	defer file.Close()
	stats, _ := file.Stat()
	sz := stats.Size()
	reader := bufio.NewReaderSize(file, int(sz))
	encryption := wolk.EncryptionNone
	if options != nil {
		encryption = options.Encryption
	}
	fileHash, err = wclient.addContent(reader, uint64(sz), encryption)
	return fileHash, fileSize, err
}

func (wclient *WolkClient) AddBlob(blob []byte, options *wolk.RequestOptions) (fileHash common.Hash, err error) {
	encryption := wolk.EncryptionNone
	if options != nil {
		encryption = options.Encryption
	}
	return wclient.addContent(bytes.NewReader(blob), uint64(len(blob)), encryption)
}

func (wclient *WolkClient) addContent(reader io.Reader, sz uint64, encryption string) (fileHash common.Hash, err error) {
	putter := chunker.NewPutter(wclient.HTTPUrl()+"/wolk", encryption)
	ctx := context.Background()
	fileHashBytes, _, err := chunker.TreeSplit(ctx, reader, int64(sz), putter)
	fileHash = common.BytesToHash(fileHashBytes)
	return fileHash, nil
}

func (wclient *WolkClient) AddRSABlob(blob []byte, publicKey *rsa.PublicKey) (chunk []byte, err error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, blob, []byte("rsablob"))
}

func (wclient *WolkClient) GetRSABlob(chunkID []byte) (chunk []byte, ok bool, err error) {
	encryptedChunk, ok, err := wclient.GetChunk(context.TODO(), chunkID, nil)
	if err != nil {
		return chunk, ok, err
	} else if !ok {
		return chunk, ok, err
	}
	chunk, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, wclient.rsaPrivateKey, encryptedChunk, []byte("rsablob"))
	if err != nil {
		return chunk, ok, err
	}
	return chunk, ok, nil
}

func (wclient *WolkClient) CatFile(fileHash common.Hash, options *wolk.RequestOptions) (output []byte, err error) {
	return wclient.get(path.Join(wolk.ProtocolName, wolk.PayloadFile, fmt.Sprintf("%x", fileHash)), options)
}

func (wclient *WolkClient) CatFileLocal(fileHash common.Hash, options *wolk.RequestOptions) (output []byte, err error) {
	encryption := wolk.EncryptionNone
	if options != nil {
		encryption = options.Encryption
	}
	frange := options.Range
	startrange := int64(0)
	endrange := int64(-1)
	if strings.Compare(frange, "all") != 0 {
		ft := strings.Split(frange, "-")
		if len(ft) != 2 {
			return output, fmt.Errorf("Invalid range %s", frange)
		}
		from, err1 := strconv.Atoi(ft[0])
		to, err2 := strconv.Atoi(ft[1])
		if err1 != nil || err2 != nil {
			return output, fmt.Errorf("Invalid range %s", frange)
		}
		startrange = int64(from)
		endrange = int64(to)
	}

	url := wclient.HTTPUrl() + "/wolk"
	getter := chunker.NewGetter(url, encryption)
	qc := make(chan bool)
	reader := chunker.TreeJoin(context.TODO(), fileHash.Bytes(), getter, 0)
	sz, err := reader.Size(context.TODO(), qc)
	if err != nil {
		return output, fmt.Errorf("chunk read error %v", err)
	}

	if endrange == -1 {
		endrange = sz
	}
	if endrange < startrange {
		return output, fmt.Errorf(fmt.Sprintf("error size = %d start = %d to = %d", sz, startrange, endrange))
	} else if endrange > sz {
		return output, fmt.Errorf(fmt.Sprintf("error size = %d to = %d", sz, endrange))
	} else {
		fmt.Printf("sz: %d EndRange %d StartRange %d\n", sz, endrange, startrange)
		output = make([]byte, endrange-startrange)
		_, err := reader.ReadAt(output, int64(startrange))
		return output, err
	}
}

func (wclient *WolkClient) PutFile(localfile string, path string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	if _, err := os.Stat(localfile); os.IsNotExist(err) {
		return txhash, fmt.Errorf("File does not exist %s", localfile)
	}
	body, err := ioutil.ReadFile(localfile)
	if err != nil {
		return txhash, err
	}
	return wclient.put(path, body)
}

func (wclient *WolkClient) PutFile2(localfile string, path string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	start := time.Now()
	if _, err := os.Stat(localfile); os.IsNotExist(err) {
		return txhash, fmt.Errorf("File does not exist %s", localfile, "time", time.Since(start))
	}
	b, err := ioutil.ReadFile(localfile)
	if err != nil {
		log.Error("PutFile2 ReadFile error", "err", err, "time", time.Since(start))
		return txhash, err
	}
	body := bytes.NewReader(b)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/wolk/file", wclient.HTTPUrl()), body)
	if err != nil {
		log.Error("PutFile2 NewRequest error", "err", err, "time", time.Since(start))
		return txhash, err
	}
	//err = wclient.SignRequest(req, chunk)
	resp, err := wclient.httpclient.Do(req)
	if err != nil {
		log.Error("PutFile2 http request error", "err", err, "time", time.Since(start))
		return txhash, err
	}
	output, err := ioutil.ReadAll(resp.Body)

	//return wclient.post(fmt.Sprintf("%s/wolk/file", wclient.HTTPUrl()), b)
	log.Info("PutFile2", "output", fmt.Sprintf("%s", output), "err", err, "time", time.Since(start))
	return common.HexToHash(string(output)), err
}

func (wclient *WolkClient) PutFile3(body []byte, path string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	/*
	   if _, err := os.Stat(localfile); os.IsNotExist(err) {
	           return txhash, fmt.Errorf("File does not exist %s", localfile)
	   }
	   body, err := ioutil.ReadFile(localfile)
	   if err != nil {
	           return txhash, err
	   }
	*/
	return wclient.put(path, body)
}

func (wclient *WolkClient) GetFile(path string, options *wolk.RequestOptions) (output []byte, err error) {
	log.Info("[client.go:GetFile] ", "path", path)
	output, err = wclient.get(path, options)
	if err != nil {
		log.Error("[client.go:GetFile] Error ", "err", err)
		return output, err
	}
	return output, err
}

func (wclient *WolkClient) GetCollections(owner string, options *wolk.RequestOptions) (items []*wolk.TxBucket, err error) {
	resp, err := wclient.get(owner, options)
	if err != nil {
		log.Error("[client:GetCollections] get", "err", err)
		return items, err
	}
	fmt.Printf("%s\n", resp)
	// Unmarshal into buckets
	items = make([]*wolk.TxBucket, 0)
	err = json.Unmarshal(resp, &items)
	return items, nil
}

func (wclient *WolkClient) GetCollection(owner string, collection string, options *wolk.RequestOptions) (b *wolk.TxBucket, items []*wolk.BucketItem, err error) {
	resp, err := wclient.get(path.Join(owner, collection), options)
	if err != nil {
		log.Error("[client:GetCollection] get", "owner", owner, "collection", collection, "err", err)
		return b, items, err
	}
	// Unmarshal into buckets
	b = new(wolk.TxBucket)
	err = json.Unmarshal(resp, &b)

	resp, err = wclient.get(path.Join(owner, collection), options)
	if err != nil {
		log.Error("[client:GetCollection] get", "err", err)
		return b, items, err
	}
	// Unmarshal into buckets
	items = make([]*wolk.BucketItem, 0)
	err = json.Unmarshal(resp, &items)
	return b, items, nil
}

// SQL
func (wclient *WolkClient) ReadSQL(owner string, database string, req *wolk.SQLRequest, options *wolk.RequestOptions) (response *wolk.SQLResponse, err error) {
	//var reqBytes []byte
	//reqBytes, err = json.Marshal(req)
	//if err != nil {
	//	return response, fmt.Errorf("[client:ReadSQL] %s", err)
	//}
	//log.Info("[client:ReadSQL]", "string(reqBytes)", string(reqBytes))
	resp, err := wclient.patch(path.Join(owner, database, "SQL"), req)
	if err != nil {
		log.Error("[client:ReadSQL]", "err", err)
		return response, fmt.Errorf("[client:ReadSQL] %s", err)
	}
	if len(resp) == 0 {
		return response, fmt.Errorf("[client:ReadSQL] no response")
	}

	//log.Info("[client:ReadSQL] before unmarshal", "response", string(resp))
	err = json.Unmarshal(resp, &response)
	if err != nil {
		log.Error("[client:ReadSQL] not a SQLResponse!", "err", err)
		var unknownResp map[string]interface{}
		err := json.Unmarshal(resp, &unknownResp)
		if err != nil {
			log.Error("[client:ReadSQL] not a json response!", "err", err)
		}
		return response, fmt.Errorf("[client:ReadSQL] unknown response (%x) (%v) (%s)", resp, resp, resp)
	}
	//log.Info("[client:ReadSQL]", "response", response)

	return response, nil
}

func (wclient *WolkClient) MutateSQL(owner string, database string, req *wolk.SQLRequest) (txhash common.Hash, err error) {
	return wclient.post(path.Join(owner, database, "SQL"), req)
}

func (wclient *WolkClient) CreateBucket(bucketType uint8, owner string, bucket string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	return wclient.post(path.Join(owner, bucket), wolk.NewTxBucket(bucket, bucketType, options))
}

func (wclient *WolkClient) DeleteBucket(owner string, bucket string) (txhash common.Hash, err error) {
	return wclient.Delete(path.Join(owner, bucket), nil)
}

func (wclient *WolkClient) DeleteKey(owner string, collection string, key string) (txhash common.Hash, err error) {
	return wclient.Delete(path.Join(owner, collection, key), nil)
}

func (wclient *WolkClient) Transfer(recipient string, amount uint64) (txhash common.Hash, err error) {
	return wclient.post(path.Join(wolk.ProtocolName, wolk.PayloadTransfer, recipient), wolk.NewTxTransfer(amount, recipient))
}

func (wclient *WolkClient) RegisterNode(nodenumber uint64, ip string, consensusip string, region uint8, valueInternal uint64) (txhash common.Hash, err error) {
	tx := wolk.NewTxNode(ip, consensusip, region, valueInternal)
	return wclient.post(path.Join(wolk.ProtocolName, wolk.PayloadNode, fmt.Sprintf("%d", nodenumber)), tx)
}
