package wolk

import (
	//"bytes"

	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/log"
)

// noSQL API
/* ========================================================================== */
func (wolkstore *WolkStore) GetName(owner string, options *RequestOptions) (address common.Address, ok bool, proof *SMTProof, err error) {
	blockNumber := int(0)
	if options != nil {
		blockNumber = options.BlockNumber
	}

	withProof := options.WithProof()
	state, ok, err := wolkstore.getStateByNumber(context.Background(), blockNumber)
	if err != nil {
		return address, ok, proof, fmt.Errorf("[backend_nosql:GetName] %s", err)
	}
	if !ok {
		return address, ok, proof, fmt.Errorf("[backend_nosql:GetName] no block")
	}
	address, ok, proof, err = state.GetName(context.Background(), owner, withProof)
	if err != nil {
		return address, ok, proof, fmt.Errorf("[backend_nosql:GetName] %s", err)
	}
	if ok && withProof {
		b := state.block
		proof.BlockNumber = b.BlockNumber
		proof.BlockHash = b.Hash()
		if wolkstore.consensusIdx%2 == 1 {
			proof.MerkleRoot = proof.ChunkHash
		}
	}
	return address, ok, proof, nil
}

func (wolkstore *WolkStore) GetKey(owner string, collection string, key string, options *RequestOptions) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {
	txhash, ok, deleted, proof, err = wolkstore.getKey(context.Background(), owner, string(collection), string(key), options)
	return txhash, ok, deleted, proof, err
}

func (wolkstore *WolkStore) GetIndexedKey(owner string, collection string, index string, key []byte, options *RequestOptions) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {
	txhash, ok, deleted, proof, err = wolkstore.getIndexedKey(context.Background(), owner, collection, index, key, options)
	if err != nil {
		panic(err) // TODO
	}
	return txhash, ok, deleted, proof, nil
}

func (wolkstore *WolkStore) GetBucket(owner string, collection string, options *RequestOptions) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {
	txhash, ok, deleted, proof, err = wolkstore.getBucket(context.Background(), owner, string(collection), options)
	return txhash, ok, deleted, proof, err
}

func (wolkstore *WolkStore) GetKeyHistory(owner string, collection string, key string, options *RequestOptions) (history []*NoSQLProof, err error) {
	return wolkstore.getKeyHistory(context.Background(), owner, collection, key, options)
}

func (wolkstore *WolkStore) ScanCollection(owner string, collection string, options *RequestOptions) (txhashList []common.Hash, ok bool, proof *NoSQLProof, err error) {
	if options == nil {
		log.Error("[backend_nosql:ScanCollection] empty options... why?")
		options = NewRequestOptions()
	}
	return wolkstore.scanCollection(context.Background(), owner, collection, options)
}

func (wolkstore *WolkStore) ScanIndexedCollection(owner string, collection string, index string, keyStart []byte, keyEnd []byte, limit int, options *RequestOptions) (txhashes []common.Hash, ok bool, proof *NoSQLProof, err error) {
	txhashes, ok, proof, err = wolkstore.scanIndexedCollection(context.Background(), owner, collection, index, keyStart, keyEnd, limit, options)
	if err != nil {
		panic(err) //TODO
	}
	return txhashes, ok, proof, nil
}

/* ########################################################################## */
/* Nosql Internal Method Do not expose */
/* ========================================================================== */

func (wolkstore *WolkStore) scanCollection(ctx context.Context, owner string, collection string, options *RequestOptions) (txhashList []common.Hash, ok bool, proof *NoSQLProof, err error) {
	state, ok, err := wolkstore.getStateByNumber(ctx, options.BlockNumber)
	if err != nil {
		return txhashList, false, proof, fmt.Errorf("[backend_nosql:scanCollection] %s", err)
	} else if !ok {
		return txhashList, false, proof, fmt.Errorf("[backend_nosql:scanCollection] no block")
	}

	txhashes, ok, proof, err := state.ScanCollection(ctx, owner, collection, options.WithProof())
	if err != nil {
		return txhashList, false, proof, fmt.Errorf("[backend_nosql:scanCollection] %s", err)
	}
	if !ok {
		return txhashList, false, proof, nil
	}
	txhashList = make([]common.Hash, 0)
	for _, txhash := range txhashes {
		txhashList = append(txhashList, txhash)
	}
	if options.WithProof() {
		b := state.block
		proof.BlockNumber = b.BlockNumber
		proof.BlockHash = b.Hash()
		proof.KeyChunkHash = b.KeyRoot
		proof.KeyMerkleRoot = state.keyStorage.MerkleRoot(context.Background()) // WHEW!
		if wolkstore.consensusIdx%2 == 1 {
			proof.CollectionMerkleRoot = common.BytesToHash(GlobalDefaultHashes[0])
			proof.CollectionChunkHash = common.BytesToHash(GlobalDefaultHashes[1])
			proof.SystemMerkleRoot = common.BytesToHash(GlobalDefaultHashes[2])
			proof.SystemChunkHash = proof.KeyMerkleRoot
			proof.KeyMerkleRoot = proof.KeyChunkHash
		}
	}
	return txhashList, true, proof, nil
}

func (wolkstore *WolkStore) scanIndexedCollection(ctx context.Context, owner string, collection string, index string, keyStart []byte, keyEnd []byte, limit int, options *RequestOptions) (txhashList []common.Hash, ok bool, proof *NoSQLProof, err error) {

	state, ok, err := wolkstore.getStateByNumber(ctx, options.BlockNumber)
	if err != nil {
		return txhashList, false, proof, fmt.Errorf("[backend_nosql:scanCollection] %s", err)
	} else if !ok {
		return txhashList, false, proof, fmt.Errorf("[backend_nosql:scanCollection] no block")
	}

	txhashList, ok, proof, err = state.ScanIndexedCollection(ctx, owner, collection, index, keyStart, keyEnd, limit, options.WithProof())
	if err != nil {
		return txhashList, false, proof, fmt.Errorf("[backend_nosql:scanCollection] %s", err)
	}
	if !ok {
		dprint("[backend_nosql:scanIndexedCollection] scan came back NOT OK!")
		return txhashList, false, proof, nil
	}
	//txhashList = make([]common.Hash, 0)
	//for _, txhash := range txhashes {
	//	txhashList = append(txhashList, txhash)
	//}

	// TODO:
	// if options.WithProof() {
	// 	b := state.block
	// 	proof.BlockNumber = b.BlockNumber
	// 	proof.BlockHash = b.Hash()
	// 	proof.KeyChunkHash = b.KeyRoot
	// 	proof.KeyMerkleRoot = state.keyStorage.MerkleRoot(context.Background()) // WHEW!
	// 	if wolkstore.consensusIdx%2 == 1 {
	// 		proof.CollectionMerkleRoot = common.BytesToHash(GlobalDefaultHashes[0])
	// 		proof.CollectionChunkHash = common.BytesToHash(GlobalDefaultHashes[1])
	// 		proof.SystemMerkleRoot = common.BytesToHash(GlobalDefaultHashes[2])
	// 		proof.SystemChunkHash = proof.KeyMerkleRoot
	// 		proof.KeyMerkleRoot = proof.KeyChunkHash
	// 	}

	return txhashList, true, proof, nil
}

func (wolkstore *WolkStore) StorageCollection(owner string, collection string, options *RequestOptions) (storageBytes uint64, ok bool, err error) {
	if options == nil {
		log.Error("[backend_nosql:StorageCollection] empty options... why?")
		options = NewRequestOptions()
	}
	state, ok, err := wolkstore.getStateByNumber(context.Background(), options.BlockNumber)
	if err != nil {
		return storageBytes, false, fmt.Errorf("[backend_nosql:StorageCollection] %s", err)
	}
	if !ok {
		return storageBytes, false, fmt.Errorf("[backend_nosql:StorageCollection] no block")
	}
	return state.StorageCollection(context.Background(), owner, collection)
}

func (wolkstore *WolkStore) getKey(ctx context.Context, owner string, collection string, key string, options *RequestOptions) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {
	if options == nil {
		log.Error("[backend_nosql:getKey] empty options... why?")
		options = NewRequestOptions()
	}
	log.Trace("[backend_nosql:getKey]", "owner", owner, "collection", collection, "key", key, "options", options.String())
	state, ok, err := wolkstore.getStateByNumber(ctx, options.BlockNumber)
	if err != nil {
		log.Error("[backend_nosql:getKey] GetBlockByNumber")
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:GetKey] %s", err)
	} else if !ok {
		log.Error("[backend_nosql:getKey] GetBlockByNumber NOT OK")
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:GetKey] no block")
	}

	txhash, ok, deleted, proof, err = state.GetKey(ctx, owner, collection, key, options.WithProof())
	if err != nil {
		log.Error(fmt.Sprintf("[backend_nosql:getKey] GetKey Error in GetKey | %+v ", err))
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:GetKey] %s", err)
	}
	if !ok {
		//log.Debug("[backend:getKey] trying shim")
		//txhash, ok, err = wolkstore.shim(owner, collection, key)
		return txhash, false, deleted, proof, nil
	}

	log.Trace("[backend_nosql:getKey] Found", "txhash", txhash, "owner", owner, "collection", collection, "key", key, "options", options.String())
	if options.WithProof() {
		b := state.block
		proof.BlockNumber = b.BlockNumber
		proof.BlockHash = b.Hash()
		/*Rodney: remove alternating false proofs.
		if wolkstore.consensusIdx%2 == 1 {
			proof.KeyMerkleRoot = proof.KeyChunkHash
			proof.CollectionMerkleRoot = proof.CollectionChunkHash
			proof.SystemMerkleRoot = proof.SystemChunkHash
		}
		*/
	}
	log.Info("[backend:getKey] OK!!!!", "txhash", txhash, "owner", owner, "collection", collection, "key", key, "options", options.String())
	return txhash, ok, deleted, proof, nil
}

func (wolkstore *WolkStore) getIndexedKey(ctx context.Context, owner string, collection string, index string, key []byte, options *RequestOptions) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {
	if options == nil {
		return txhash, false, false, nil, fmt.Errorf("[backend_nosql:getIndexedKey] RequestOptions must be included for getIndexedKey")
	}
	state, ok, err := wolkstore.getStateByNumber(ctx, options.BlockNumber)
	if err != nil {
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:getIndexedKey] %s", err)
	} else if !ok {
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:getIndexedKey] no block")
	}
	txhash, ok, deleted, proof, err = state.GetIndexedKey(ctx, owner, collection, index, key, options.WithProof())
	if err != nil {
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:getIndexedKey] %s", err)
	}
	if !ok {
		dprint("[backend_nosql:getKey] get indexed key was NOT OK")
		return txhash, false, false, nil, nil
	}
	return txhash, true, deleted, proof, nil
}

func (wolkstore *WolkStore) getBucket(ctx context.Context, owner string, collection string, options *RequestOptions) (txhash common.Hash, ok bool, deleted bool, proof *NoSQLProof, err error) {
	if options == nil {
		log.Error("[backend_nosql:getBucket] empty options... why?")
		options = NewRequestOptions()
	}
	state, ok, err := wolkstore.getStateByNumber(ctx, options.BlockNumber)
	if err != nil {
		log.Error("[backend_nosql:getBucket] getStateByNumber")
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:getBucket] %s", err)
	} else if !ok {
		log.Error("[backend_nosql:getBucket] getStateByNumber NOT OK")
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:getBucket] no block")
	}

	txhash, ok, deleted, proof, err = state.GetBucket(ctx, owner, collection, options.WithProof())
	if err != nil {
		log.Error(fmt.Sprintf("[backend_nosql:getBucket] GetKey Error in getBucket | %+v ", err))
		return txhash, ok, deleted, proof, fmt.Errorf("[backend_nosql:getBucket] %s", err)
	}
	if !ok {
		return txhash, false, deleted, proof, nil
	} else if ok {
		log.Trace("[backend:getBucket] found", "txhash", txhash, "owner", owner, "collection", collection, "options", options.String())
		if options.WithProof() {
			b := state.block
			proof.BlockNumber = b.BlockNumber
			proof.BlockHash = b.Hash()
		}
		return txhash, ok, deleted, proof, nil
	}
	return txhash, ok, deleted, proof, nil
}

func (wolkstore *WolkStore) getKeyHistory(ctx context.Context, owner string, collection string, k string, options *RequestOptions) (history []*NoSQLProof, err error) {
	done := false
	nlooks := 0
	blockNumber := options.BlockNumber
	for bn := uint64(blockNumber); !done; {
		state, ok, err := wolkstore.getStateByNumber(ctx, blockNumber)
		if err != nil {
			return history, fmt.Errorf("[backend_nosql:GetKeyHistory] err %s", err)
		} else if !ok {
			return history, fmt.Errorf("[backend_nosql:GetKeyHistory] block %d not found", blockNumber)
		}

		txhash, ok, _, proof, err := state.GetKey(ctx, owner, collection, k, true)
		if err != nil {
			return history, fmt.Errorf("[backend_nosql:GetKeyHistory] err %s", err)
		} else if !ok {
			done = true
		} else {
			log.Info("[backend:getKeyHistory] GetKey", "txhash", txhash, "proof", proof)
			var tx *Transaction
			tx, bn, _, ok, err = wolkstore.GetTransaction(ctx, txhash)
			if err != nil {
				return history, fmt.Errorf("[backend_nosql:GetKeyHistory] %s", err)
			} else if !ok {
				return history, fmt.Errorf("[backend_nosql:GetKeyHistory] txn(%x) not gotten", txhash)
			}
			proof.Tx = NewSerializedTransaction(tx)
			b := state.block
			proof.BlockNumber = b.BlockNumber
			proof.BlockHash = b.Hash()
			proof.KeyChunkHash = b.KeyRoot
			proof.KeyMerkleRoot = state.keyStorage.MerkleRoot(context.Background()) // WHEW!
			history = append(history, proof)
			bn--
			log.Info("[backend:getKeyHistory] added to history", "bn", bn, "proof.BlockNumber", proof.BlockNumber)
			nlooks++
			if nlooks > 10 { // This is hardcoded to 5 looks?
				done = true
			}
		}
	}
	return history, nil
}

func (wolkstore *WolkStore) GetBuckets(owner string, blockNumber int, options *RequestOptions) (txhashList []common.Hash, ok bool, proof *NoSQLProof, err error) {
	if options == nil {
		log.Error("[backend_nosql:ScanCollection] empty options... why?")
		options = NewRequestOptions()
	}
	withProof := options.WithProof()
	log.Trace("[backend_nosql:getBuckets] start", "owner", owner, "blocknum", blockNumber)
	state, ok, err := wolkstore.getStateByNumber(context.Background(), blockNumber)
	if err != nil {
		return txhashList, ok, proof, fmt.Errorf("[backend_nosql:getBuckets] GetBlockByNumber %s", err)
	}
	if !ok {
		return txhashList, ok, proof, fmt.Errorf("[backend_nosql:getBuckets] GetBlockByNumber - no block")
	}
	txhashes, ok, proof, err := state.GetBuckets(context.Background(), owner, withProof)
	if err != nil {
		log.Error("[backend_buckets:getBuckets] ERR", "err", err)
		return txhashList, ok, proof, fmt.Errorf("[backend:getBuckets] GetBuckets %s", err)
	}
	if !ok {
		log.Error("[backend_buckets:getBuckets] NOT OK", "owner", owner)
		return txhashList, false, proof, nil
	}
	for _, txhash := range txhashes {
		txhashList = append(txhashList, txhash)
	}
	if withProof {
		if wolkstore.consensusIdx%2 == 1 {
			proof.SystemChunkHash = proof.CollectionChunkHash
			proof.CollectionChunkHash = proof.KeyMerkleRoot
			proof.KeyMerkleRoot = proof.KeyChunkHash
			proof.CollectionMerkleRoot = proof.CollectionChunkHash
			proof.SystemMerkleRoot = proof.SystemChunkHash
		}
	}
	return txhashList, true, proof, nil
}

// TODO: Implement ShimURL concept
func (wolkstore *WolkStore) shim(owner string, collection string, key string) (txhash common.Hash, ok bool, err error) {
	if strings.Compare(owner, "arc") != 0 {
		log.Error("[backend_nosql:shim] owner is not arc", "owner", owner)
		return txhash, false, nil
	}
	shimUrl := "https://dweb.archive.org"

	url := fmt.Sprintf("%s/%s", shimUrl, string(key))
	log.Info("[backend_nosql:shim] Shim FETCH", "url", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error("[backend_nosql:shim] Error doing shim request", "error", err)
		return txhash, false, fmt.Errorf("[backend_nosql:shim] ERROR executing shim request for [%s] -> [%s]", url, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Millisecond)
	defer cancel()

	httpclient := &http.Client{Timeout: time.Second * 15}
	resp, err := httpclient.Do(req)
	if err != nil {
		log.Error("[backend_nosql:shim] FETCH", "err", err)
		return txhash, false, fmt.Errorf("[backend_nosql:shim] %s", err)
	}
	defer resp.Body.Close()

	reader, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("[backend_nosql:shim] FETCH", "err", err)
		return txhash, false, fmt.Errorf("[backend_nosql:shim] %s", err)
	}
	sz := len(reader)

	// Use the gateway's headers ( Sig, Requester ) and validate for inclusion
	req.Header.Add("Sig", resp.Header.Get("Sig"))
	req.Header.Add("Requester", resp.Header.Get("Requester"))
	req.Method = http.MethodPut
	// req.URL.Path = path.Join(owner, collection, key)

	var b *TxBucket
	err = json.Unmarshal(reader, &b)
	if err != nil {
		log.Error("[backend_nosql:shim] Unable to unmarshal bucket sent in [%+v]", b)
	}
	log.Info("[backend_nosql:shim] FETCH SUCC", "sz", sz)

	shimTxP := NewTransactionImplicit(req, reader)
	log.Info("[backend_nosql:shim] AFTER NewTransactionImplicit attempting to Validate")

	/* TODO: put back

	verified, err := shimTxP.ValidateTx()
	if err != nil {
		log.Error("[backend_nosql:shim] ERROR Validating TX", "error", err)
		return shimTxP.Hash(), false, err
	}
	if !verified {
		log.Error("[backend_nosql:shim] ERROR Unable to verify content", "error", err)
		return shimTxP.Hash(), false, fmt.Errorf("Could not verify received content")
	}
	*/

	log.Info("[backend_nosql:shim] About to SendingTransaction")
	txhash, txHashErr := wolkstore.SendRawTransaction(shimTxP)
	log.Info("[backend_nosql:shim] completed SendingTransaction")
	if txHashErr != nil {
		log.Error("[backend_nosql:shim] SendRawTransaction ERR", "err", txHashErr)
		return txhash, false, fmt.Errorf("[backend_nosql:shim] %s", txHashErr)
	}

	//var wgPutFile *sync.WaitGroup
	//wgPutFile.Add(1)
	contentHashBytes, err := wolkstore.Storage.PutFile(ctx, (reader), nil)
	//wgPutFile.Wait()
	contentHash := common.BytesToHash(contentHashBytes)
	log.Info("[backend_nosql:shim] SendRawTransaction", "txhash", txhash, "contentHash", contentHash, "sz", sz)
	return txhash, true, nil
}
