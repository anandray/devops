package wolk

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

type Indexer struct {
	localdb         *cloud.CloudLevelDB
	lastBlockNumber uint64
	certsMu         sync.RWMutex
}

func NewIndexer(dataDir string) (indexer *Indexer, err error) {
	indexer = new(Indexer)
	localdbpath := fmt.Sprintf("%s/localdb", dataDir)
	localdb, err := cloud.NewLevelDB(localdbpath, nil)
	if err != nil {
		log.Error("NewLevelDB err", err)
		return indexer, err
	}
	indexer.localdb = localdb
	log.Info("[indexer:Indexer] STARTED indexer1")
	blockNumber, err := indexer.GetLastKnownBlockNumber()
	if err != nil {
		return indexer, err
	}
	log.Info("[indexer:Indexer] STARTED indexer2", "blocknumber", blockNumber)
	return indexer, nil
}

func (indexer *Indexer) setLocal(chunkKey []byte, chunk []byte) (err error) {
	log.Trace("setLocal", "k", fmt.Sprintf("%s", chunkKey), "len(chunk)", len(chunk))
	return indexer.localdb.SetChunk(nil, chunkKey, chunk)
}

func (indexer *Indexer) getLocal(chunkKey []byte) (val []byte, ok bool, err error) {
	log.Trace("[indexer:getLocal] tring to get from localdb", "chunk key", common.BytesToHash(chunkKey))
	val, ok, err = indexer.localdb.GetChunk(chunkKey)
	log.Trace("[indexer:getLocal]", "k", fmt.Sprintf("%s", chunkKey), "ok", ok, "len(val)", len(val), "err", err)
	return val, ok, err
}

func (indexer *Indexer) deleteLocal(chunkKey []byte) (err error) {
	return indexer.localdb.Delete(chunkKey)
}

func (indexer *Indexer) processFinalized(cert *VoteMessage) (missedCerts map[uint64]common.Hash, err error) {
	missingCerts := make(map[uint64]common.Hash)
	// reach backward from this finalized block to the previous finalized block, deleting any tentative certs not on the path
	log.Info("[indexer:processFinalized] START")
	nblocks := 0
	nextBlockHash := cert.ParentHash
	b := cert.BlockNumber - 1
	for done := false; !done; {
		certs, found, err := indexer.GetCertificate(b)
		if err != nil {
			return missingCerts, err
		} else if !found {
			missingCerts[b] = nextBlockHash
			log.Info("[indexer:processFinalized] Exit 0: parent cert not found", "b", b, "BlockHash", nextBlockHash, "missCert len", len(missingCerts), "nblocks", nblocks)
			return missingCerts, fmt.Errorf("certs not found")
		}

		// get the block and index it in the indexer as well?

		certfound := false
		for _, c := range certs {
			if bytes.Compare(c.BlockHash.Bytes(), nextBlockHash.Bytes()) == 0 {
				certfound = true
				nextBlockHash = c.ParentHash
				_, cv := c.CountVotes(voteNext, c.Seed, b)
				if cv > expectedTokensFinal {
					log.Info("[indexer:processFinalized] FINALIZED REACHED", "b", b, "missCert len", len(missingCerts))
					done = true
					return missingCerts, nil
				}
				b-- //parent not finalized, looking up grandparent..
			} else {
				log.Info("[indexer:processFinalized] DELETING CERT", "b", b, "c", c.BlockHash)
				indexer.deleteCertificate(c)
			}
		}
		// if *NO* cert is found, add to the map of missing certs
		if !certfound {
			missingCerts[b] = nextBlockHash
			log.Info("[indexer:processFinalized] MISSING CERT", "b", b, "BlockHash", nextBlockHash, "missCert len", len(missingCerts))
			done = true
			log.Info("[indexer:processFinalized] Exit 1: cert not found", "b", b, "BlockHash", nextBlockHash, "missCert len", len(missingCerts))
			return missingCerts, nil
		}
		if nblocks > 50 {
			//this case is not handled yet!
			log.Info("[indexer:processFinalized] OVERFLOW")
			done = true
			log.Info("[indexer:processFinalized] Exit 2: overflow", "b", b, "BlockHash", nextBlockHash, "missCert len", len(missingCerts), "nblocks", nblocks)
			return missingCerts, nil
		}
		nblocks++
	}
	log.Info("[indexer:processFinalized] Exit 3", "b", b, "BlockHash", nextBlockHash, "missCert len", len(missingCerts))
	return missingCerts, nil
}

func (indexer *Indexer) SetLastKnownBlockNumber(bn uint64) (err error) {
	if bn <= indexer.lastBlockNumber {
		return nil
	}
	lastBN := fmt.Sprintf("lastbn")
	err = indexer.setLocal([]byte(lastBN), wolkcommon.UInt64ToByte(bn))
	if err != nil {
		return err
	}
	log.Trace(fmt.Sprintf("INDEX:SetLastKnownBlocknumber %d", bn))
	indexer.lastBlockNumber = bn
	return nil
}

func (indexer *Indexer) GetLastKnownBlockNumber() (blockNumber uint64, err error) {
	lastBN := fmt.Sprintf("lastbn")
	val, ok, err := indexer.getLocal([]byte(lastBN))
	if ok {
		blockNumber = wolkcommon.BytesToUint64(val)
		return blockNumber, nil
	}
	return 0, nil
}

func (indexer *Indexer) GetBlockHash(blockNumber uint64) (h common.Hash, err error) {
	b, ok, err := indexer.GetBlockByNumber(blockNumber)
	if err != nil {
		return h, err
	}
	if !ok {
		return h, fmt.Errorf("Blocknumber not found")
	}
	return b.Hash(), nil
}

func (indexer *Indexer) GetBlockByNumber(blockNumber uint64) (b *Block, ok bool, err error) {
	if blockNumber <= 0 {
		return b, false, fmt.Errorf("No 0 block")
	}
	keyBlockNumber := fmt.Sprintf("bn-%d", blockNumber)
	log.Trace("[indexer:GetBlockByNumber] getting", "key block number", keyBlockNumber)
	encoded, ok, err := indexer.getLocal([]byte(keyBlockNumber))
	if err != nil {
		log.Error("[indexer:GetBlockByNumber]", "error", err)
		return b, false, fmt.Errorf("[indexer:GetBlockByNumber] %s", err)
	} else if !ok {
		log.Trace("[indexer:GetBlockByNumber] Local Chunk not found", "chunk", keyBlockNumber)
		return b, false, nil
	}
	log.Trace("[indexer:GetBlockByNumber] CHECK GetBlockByBlockNumber", "encoded", fmt.Sprintf("%x", encoded))
	b = new(Block)
	err = rlp.Decode(bytes.NewReader(encoded), b)
	if err != nil {
		log.Error("[indexer:GetBlockByNumber] Decode", "Error", err)
		return b, false, fmt.Errorf("[indexer:GetBlockByNumber] %s", err)
	}
	log.Trace("[indexer:GetBlockByNumber] looks like block gotten success", "block", b)
	return b, true, nil
}

func (indexer *Indexer) setFinalizedRound(blockNumber uint64, bhash common.Hash) (err error) {
	if blockNumber < 1 {
		return nil
	}
	finalizedKey := fmt.Sprintf("FP-%d", blockNumber)
	err = indexer.setLocal([]byte(finalizedKey), bhash.Bytes())
	if err != nil {
		log.Error("[indexer:setFinalizedRound] ERR", "BN", blockNumber, "BHash", bhash, "ERROR", err)
		return err
	}
	log.Info("[indexer:setFinalizedRound]", "BN", blockNumber, "BHash", bhash)
	return nil
}

func (indexer *Indexer) getFinalizedPath(blockNumber uint64) (bhash common.Hash, err error) {
	finalizedKey := fmt.Sprintf("FP-%d", blockNumber)
	hash, ok, err := indexer.getLocal([]byte(finalizedKey))
	if err != nil {
		log.Error("[indexer:getFinalizedHash]", "error", err)
		return bhash, err
	} else if !ok {
		log.Warn("[indexer:getFinalizedHash] NOT Found", "BN", blockNumber)
		return bhash, fmt.Errorf("[backend:GetBlockByHash] %s", "NOT Found")
	}
	return common.BytesToHash(hash), err
}

func (indexer *Indexer) IndexBlock(b *Block) (err error) {
	keyBlockNumber := fmt.Sprintf("bn-%d", b.BlockNumber)
	encoded, _ := rlp.EncodeToBytes(b)
	err = indexer.setLocal([]byte(keyBlockNumber), encoded)
	if err != nil {
		return err
	}

	log.Trace("[indexer:GetBlockByNumber] CHECK IndexBlock ", "encoded", fmt.Sprintf("%x", encoded))

	keyBlockHash := fmt.Sprintf("bh-%x", b.Hash())
	err = indexer.setLocal([]byte(keyBlockHash), encoded)
	if err != nil {
		return err
	}
	if len(b.Transactions) > 0 {
		log.Trace("[indexer:IndexBlock]", "bn", b.BlockNumber, "len(tx)", len(b.Transactions))
	}
	for _, tx := range b.Transactions {
		keyTxHash := fmt.Sprintf("tx-%x", tx.Hash())
		err = indexer.setLocal([]byte(keyTxHash), tx.Bytes())
		if err != nil {
			log.Error("[indexer:IndexBlock] - Error", "error", err)
			return nil
		}

		keyTxBNHash := fmt.Sprintf("txbn-%x", tx.Hash())
		err = indexer.setLocal([]byte(keyTxBNHash), wolkcommon.UInt64ToByte(b.BlockNumber))
		if err != nil {
			log.Error("[indexer:IndexBlock] - Error", "error", err)
			return nil
		}
	}

	indexer.SetLastKnownBlockNumber(b.BlockNumber)
	return nil
}

func (indexer *Indexer) StoreCertificate(c *VoteMessage) (err error) {
	if c == nil {
		return fmt.Errorf("No Certificate Provided")
	}
	indexer.certsMu.Lock()
	defer indexer.certsMu.Unlock()

	addKey := true
	knowncertHash, _, err := indexer.GetCertificateHash(c.BlockNumber)
	if knowncertHash == nil {
		err1 := indexer.storeRawCert(c)
		err2 := indexer.UpdateCertificateHash(c.BlockNumber, c.BlockHash, addKey)
		if err1 != nil || err2 != nil {
			log.Error("[indexer:StoreCertificate]", "bn", c.BlockNumber, "step", (c.Step), "err1", err1, "err2", err2)
			return fmt.Errorf("%v - %v", err1, err2)
		}
		log.Info("[indexer:StoreCertificate] First CERT", "bn", c.BlockNumber, "bh", c.BlockHash, "n", 0)
		return nil
	}

	oldcerts, _, err := indexer.GetCertificate(c.BlockNumber)
	//prcoessing updated votes when possible, exit immediately when found
	for _, oldcert := range oldcerts {
		if bytes.Equal(oldcert.BlockHash.Bytes(), c.BlockHash.Bytes()) {
			//knowncert
			_, oldcv := oldcert.CountVotes(voteNext, oldcert.Seed, oldcert.BlockNumber)
			_, newcv := c.CountVotes(voteNext, c.Seed, c.BlockNumber)
			if newcv < oldcv {
				//TODO: absorb when possible
				log.Info("[indexer:StoreCertificate] FIRST REPEAT CERT - Ignored", "bn", c.BlockNumber, "bh", c.BlockHash, "Found", oldcv, "Received", newcv)
				return nil
			} else {
				err = indexer.storeRawCert(c)
				// does not require certHash update
				if err != nil {
					log.Error("[indexer:StoreCertificate]", "bn", c.BlockNumber, "step", (c.Step), "err", err)
					return err
				}
				log.Info("[indexer:StoreCertificate] BETTER CERT", "bn", c.BlockNumber, "bh", c.BlockHash, "newcv", newcv, "oldcv", oldcv)
				return nil
			}
		}
	}

	//received new cert
	err1 := indexer.storeRawCert(c)
	err2 := indexer.UpdateCertificateHash(c.BlockNumber, c.BlockHash, addKey)
	if err1 != nil || err2 != nil {
		log.Error("[indexer:StoreCertificate]", "bn", c.BlockNumber, "step", (c.Step), "err1", err1, "err2", err2)
		return fmt.Errorf("%v - %v", err1, err2)
	}
	log.Info("[indexer:StoreCertificate] New CERT", "bn", c.BlockNumber, "bh", c.BlockHash, "n", len(oldcerts))
	return nil
}

// store a raw certificate
func (indexer *Indexer) storeRawCert(c *VoteMessage) (err error) {
	certKey := fmt.Sprintf("cert-%d-%x", c.BlockNumber, c.BlockHash) //cert-round-blockhash
	encoded, _ := rlp.EncodeToBytes(c)
	err = indexer.setLocal([]byte(certKey), encoded)
	if err != nil {
		log.Error("[indexer:StoreRawCert]", "bn", c.BlockNumber, "bh", c.BlockHash, "step", (c.Step), "err", err)
		return err
	}
	return nil
}

//delete both certHash and cert
func (indexer *Indexer) deleteCertificate(c *VoteMessage) (err error) {
	log.Info("[indexer:deleteCertificate]", "bn", c.BlockNumber, "h", c.BlockHash)
	addKey := false //delete
	err = indexer.UpdateCertificateHash(c.BlockNumber, c.BlockHash, addKey)
	if err != nil {
		return err
	}
	certKey := fmt.Sprintf("cert-%d-%x", c.BlockNumber, c.BlockHash) //cert-round-blockhash
	indexer.deleteLocal([]byte(certKey))
	return nil
}

// update/delete a certificate from knownHash TODO: add mutex
func (indexer *Indexer) UpdateCertificateHash(round uint64, certHash common.Hash, AddKey bool) (err error) {
	var knowncerts *CertList
	knowcertKey := fmt.Sprintf("knowncert-%d", round)
	knowncerts, _, err = indexer.GetCertificateHash(round)
	if err != nil {
		return err
	}

	if knowncerts == nil {
		knowncerts = new(CertList)
	}
	modified := false
	switch {
	case AddKey == true:
		if knowncerts == nil {
			knowncerts.BlockHashes = []common.Hash{}
		}
		modified = knowncerts.AddcertHash(certHash)

	default:
		//Deleting key
		if knowncerts == nil {
			return nil
		}
		modified = knowncerts.RemovecertHash(certHash)
	}

	if !modified {
		log.Info("[indexer:UpdateCertificateHash] Ingored", "bn", round, "certHash", certHash, "AddKey?", AddKey)
	}

	encoded, _ := rlp.EncodeToBytes(knowncerts)
	err = indexer.setLocal([]byte(knowcertKey), encoded)
	if err != nil {
		log.Error("[indexer:UpdateCertificateHash] ERROR", "bn", round, "certHash", certHash, "AddKey?", AddKey, "err", err)
		return err
	}
	log.Info("[indexer:UpdateCertificateHash] Update", "bn", round, "certHash", certHash, "AddKey?", AddKey)
	return nil
}

// returning a lists of known certificate hashes for a given round
func (indexer *Indexer) GetCertificateHash(round uint64) (c *CertList, ok bool, err error) {
	knowcertKey := fmt.Sprintf("knowncert-%d", round)
	val, ok, err := indexer.getLocal([]byte(knowcertKey))
	if err != nil {
		return c, ok, err //TODO
	} else if ok {
		c, err = DecodeRLPCertList(val)
		if err != nil {
			return c, ok, err //TODO
		}
	}
	return c, ok, nil
}

// returning a target certificate
func (indexer *Indexer) GetTargetCertificate(round uint64, certHash common.Hash) (certs []*VoteMessage, err error) {
	certs = make([]*VoteMessage, 0)
	certKey := fmt.Sprintf("cert-%d-%x", round, certHash) //cert-round-blockhash
	val, ok, err := indexer.getLocal([]byte(certKey))
	if err != nil {
		return certs, err
	} else if ok {
		cert, err := DecodeRLPVote(val)
		if err != nil {
		} else if cert != nil {
			certs = append(certs, cert)
		}
	}
	return certs, err
}

//returning finalized or tentative certificate(s) for a given round
func (indexer *Indexer) GetCertificate(round uint64) (certs []*VoteMessage, ok bool, err error) {
	certs = make([]*VoteMessage, 0)
	certList, ok, err := indexer.GetCertificateHash(round)
	if err != nil || !ok {
		return certs, false, err
	}

	for _, certHash := range certList.BlockHashes {
		certKey := fmt.Sprintf("cert-%d-%x", round, certHash) //cert-round-blockhash
		val, ok, err := indexer.getLocal([]byte(certKey))
		if err != nil {
			continue
		} else if ok {
			cert, err := DecodeRLPVote(val)
			if err != nil {
				//continue
			} else if cert != nil {
				certs = append(certs, cert)
			}
		}
	}
	return certs, len(certs) > 0, nil
}

// GetCertificates for votes in the range and returns ALL those found
func (indexer *Indexer) GetCertificates(blockNumberStart uint64, blockNumberEnd uint64) (certs []*VoteMessage, err error) {
	certs = make([]*VoteMessage, 0)
	errs := make([]uint64, 0)
	for b := blockNumberStart; b <= blockNumberEnd; b++ {
		var bestcert *VoteMessage
		fpHash, err := indexer.getFinalizedPath(b) //this is the finalized path
		if err == nil {
			//if fpcert found, skip countvote; fallback otherwise
			fpcerts, _ := indexer.GetTargetCertificate(b, fpHash)
			if len(fpcerts) == 1 {
				bestcert = fpcerts[0]
				certs = append(certs, bestcert)
				continue
			}
		}
		bcerts, found, _ := indexer.GetCertificate(b)
		if found {
			// only pick highest count
			bestcv := int(0)
			for _, c := range bcerts {
				_, cv := c.CountVotes(voteNext, c.Seed, c.BlockNumber)
				if cv > bestcv {
					bestcv = cv
					bestcert = c
				}
			}
			certs = append(certs, bestcert)
		} else {
			errs = append(errs, b) //all bn with missing certs
		}

	}
	if len(errs) > 0 {
		//log.Info("[indexer:GetCertificates] missing cert", "start", blockNumberStart, "end", blockNumberEnd, "cert not found", errs)
	}
	return certs, nil
}

type certificateSummary struct {
	BlockNumber uint64
	CountVotes  uint64
	RootHash    common.Hash
	BlockHash   common.Hash
	ParentHash  common.Hash
	Empty       bool
}

func (indexer *Indexer) getCertificateSummary(round uint64) (certs []*certificateSummary, err error) {
	roundStart := uint64(2)
	if round > 10 {
		roundStart = round - 9
	}
	certificates, err := indexer.GetCertificates(roundStart, round)
	if err != nil {
		log.Error("getCertificateSummary", "err", err)
		return certs, err
	}

	for _, c := range certificates {
		cs := new(certificateSummary)
		cs.BlockNumber = c.BlockNumber
		cs.BlockHash = c.BlockHash
		cs.ParentHash = c.ParentHash
		rootHash, cv, isEmpty := c.IsEmpty()
		cs.CountVotes = uint64(cv)
		cs.Empty = isEmpty
		cs.RootHash = rootHash
		certs = append(certs, cs)
	}
	return certs, nil
}

func (indexer *Indexer) GetTransaction(txhash common.Hash) (tx *Transaction, blockNumber uint64, ok bool, err error) {
	keyTxHash := fmt.Sprintf("tx-%x", txhash)
	log.Trace("[indexer:GetTransaction]", "txhash", txhash)
	val, ok, err := indexer.getLocal([]byte(keyTxHash))
	if err != nil {
		log.Error("[indexer:GetTransaction]", "err", err)
		return nil, blockNumber, ok, fmt.Errorf("[indexer:GetTransaction] %s", err)
	} else if !ok {
		log.Debug("[indexer:GetTransaction] Not Ok. Chunk not found", "k", keyTxHash)
		return nil, blockNumber, false, nil
	}

	tx, err = DecodeRLPTransaction(val)
	if err != nil {
		return tx, blockNumber, ok, err
	}

	// Read blockNumber
	keyTxBNHash := fmt.Sprintf("txbn-%x", txhash)
	val2, ok, err := indexer.getLocal([]byte(keyTxBNHash))
	if err != nil {
		return nil, blockNumber, ok, fmt.Errorf("[indexer:GetTransaction] %s", err)
	} else if !ok {
		return nil, blockNumber, ok, fmt.Errorf("[indexer:GetTransaction] Mapping to bn not found")
	}
	blockNumber = wolkcommon.BytesToUint64(val2)
	log.Trace("[indexer:GetTransaction] FOUND", "k", keyTxHash, "bn", blockNumber)
	return tx, blockNumber, true, nil
}
