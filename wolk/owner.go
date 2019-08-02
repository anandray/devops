// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.

package wolk

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/wolkdb/cloudstore/log"
)

type SQLOwner struct {
	statedb       *StateDB
	Name          string
	DatabaseNames map[string][]byte //key is databaseName, value is database chunk hash

	hash    []byte
	chunkID []byte //hash of ownerName|databaseName
	chunk   []byte
}

//Pulls an owner chunk and fills in a list of DBs locally.
func (o *SQLOwner) open() (ok bool, err error) {
	log.Info("[owner:open]", "ownername", o.Name, "dbs", o.DatabaseNames)
	log.Info("[owner:open]", "roothash key", []byte(o.Name), "in hex", hex.EncodeToString([]byte(o.Name)))

	o.chunkID, _, err = o.statedb.GetRootHash([]byte(o.Name))
	if err != nil {
		return false, fmt.Errorf("[owner:open] GetRootHash %s", err)
	}
	//owner is empty - new owner
	if EmptyBytes(o.chunkID) {
		log.Error("[owner:open] no chunkID found. returning NOT OK")
		// copy(o.chunk[0:CHUNK_HASH_SIZE], []byte(o.hash)) //storing owner chunk w/ no dbs
		// o.chunkID, err = o.statedb.SeDBChunk(u, o.chunk, 0)
		// if err != nil {
		// 	return fmt.Errorf("[sqlchain:openOwner] SetDBChunk %s", err))
		// }
		return false, nil
	}
	log.Info("[owner:open] Retrieved ChunkID via GetRootHash", "name", o.Name, "chunkID", o.chunkID, "chunkID", hex.EncodeToString(o.chunkID))

	//owner found
	o.chunk, ok, err = o.statedb.GetDBChunk(o.chunkID)
	if err != nil {
		//log.Error("[owner:open] owner chunk err", "err", err)
		return false, fmt.Errorf("[owner:open] %s", err)
	} else if !ok {
		//log.Error("[owner:open] owner chunk not found. returning NOT OK")
		return false, nil
	} else if len(o.chunk) == 0 {
		//log.Error("[owner:open] owner chunk found but empty. returning NOT OK")
		return false, nil
	}
	if len(o.chunk) == 0 || EmptyBytes(o.chunk) {
		log.Error("[owner:open] owner not found. returning NOT OK.")
		return false, nil //owner not found
	}
	log.Info("[owner:open] after GetDBChunk", "chunkID", o.chunkID, "chunk", o.chunk)

	// if bytes.Compare(o.chunk[0:CHUNK_HASH_SIZE], o.hash[0:CHUNK_HASH_SIZE]) != 0 {
	// 	// the first 32 bytes of the owner.databaseChunk should match
	// 	return false, fmt.Errorf("[owner:open] %x != retrieved %x", o.hash, o.chunk[0:CHUNK_HASH_SIZE])
	// }
	log.Info("[owner:open] header of Chunk retrieved", "header", o.chunk[0:CHUNK_HASH_SIZE])
	//if EmptyBytes(o.chunk[0:CHUNK_HASH_SIZE]) {
	//	log.Info("[owner:open] NEW OWNER ?????")
	//}

	//fill in db names and db chunk hashes belonging to this owner
	//TODO: grab encrypted bit
	for i := CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {
		if i == 1408 {
			break
		}
		if !EmptyBytes(o.chunk[i:(i + DATABASE_NAME_LENGTH_MAX)]) {
			log.Info("[owner:open] dbnames", "i", i, "end", i+DATABASE_NAME_LENGTH_MAX)
			log.Info("[owner:open] dbnames", "name in bytes", bytes.Trim(o.chunk[i:(i+DATABASE_NAME_LENGTH_MAX)], "\x00"), "name in hex", hex.EncodeToString(bytes.Trim(o.chunk[i:(i+DATABASE_NAME_LENGTH_MAX)], "\x00")))
			dbName := string(bytes.Trim(o.chunk[i:(i+DATABASE_NAME_LENGTH_MAX)], "\x00"))
			log.Info("[owner:open] dbnames", "name string", dbName)
			if _, ok := o.DatabaseNames[dbName]; !ok {
				dbHash := make([]byte, CHUNK_HASH_SIZE)
				copy(dbHash[:], o.chunk[(i+CHUNK_HASH_SIZE):(i+CHUNK_HASH_SIZE+32)])
				o.DatabaseNames[dbName] = dbHash
				log.Info("[owner:open] dbnames", "hash", dbHash, "hash in hex", hex.EncodeToString(dbHash))
			}
		}
	}

	log.Info("[owner:open] SUCCESS", "dbs", o.DatabaseNames)
	return true, nil
}

// updates an owner chunk with a database and stores the owner chunk.
func (o *SQLOwner) store(db *SQLDatabase) (err error) {
	log.Info("[owner:store]", "owner name", o.Name, "dbs", o.DatabaseNames)
	//new blank owner

	if EmptyBytes(o.chunkID) {
		//log.Info(fmt.Sprintf("[owner:store] Empty owner setting chunk hash to: [%+v] [%x]", o.hash, o.hash))
		log.Info("[owner:store] empty owner. setting", "ownerhash", hex.EncodeToString(o.hash))
		copy(o.chunk[0:CHUNK_HASH_SIZE], o.hash) //make new owner chunk header
	}

	newdbName := make([]byte, DATABASE_NAME_LENGTH_MAX)
	copy(newdbName[0:], db.Name)

	//find the index of the owner chunk where the database is or the new one should be
	found := false
	i := 0
	for i = CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {
		if !EmptyBytes(o.chunk[i:(i + DATABASE_NAME_LENGTH_MAX)]) { // found the slot for the existing database
			dbNameFound := string(bytes.Trim(o.chunk[i:(i+DATABASE_NAME_LENGTH_MAX)], "\x00"))
			if strings.Compare(db.Name, dbNameFound) == 0 {
				found = true
				break
			}
		} else { // found an empty slot for the new database
			copy(o.chunk[i:(i+DATABASE_NAME_LENGTH_MAX)], newdbName[0:DATABASE_NAME_LENGTH_MAX])
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("[owner:store] Database could not be created -- exceeded Owner chunk allocation")
	}

	//place the encrypted bit and hash in the owner chunk
	//log.Debug(fmt.Sprintf("Saving Database with encrypted bit of %d at position: %d", encrypted, i+DATABASE_NAME_LENGTH_MAX))
	if db.Encrypted > 0 {
		o.chunk[i+DATABASE_NAME_LENGTH_MAX] = 1
	} else {
		o.chunk[i+DATABASE_NAME_LENGTH_MAX] = 0
	}
	//log.Debug(fmt.Sprintf("Buffer has encrypted bit of %d ", o.chunk[i+DATABASE_NAME_LENGTH_MAX]))

	copy(o.chunk[(i+CHUNK_HASH_SIZE):(i+CHUNK_HASH_SIZE+32)], db.hash[0:CHUNK_HASH_SIZE])

	// this could be a function of the top level domain .pri/.eth -- also encrypted is hardcoded here to 0
	o.chunkID, err = o.statedb.SetDBChunk(o.chunk, 0)
	if err != nil {
		return fmt.Errorf("[owner:store] %s", err)
	}
	err = o.statedb.StoreRootHash([]byte(o.Name), o.chunkID)
	if err != nil {
		return fmt.Errorf("[owner:store] %s", err)
	}

	o.DatabaseNames[db.Name] = db.hash
	return nil
}

// drops a database from the owner chunk
// essentially removes the database chunk by removing the owner's database entry.
func (o *SQLOwner) drop(databaseName string) (err error) {
	dropdbName := make([]byte, DATABASE_NAME_LENGTH_MAX) // 32 byte version of the database name
	copy(dropdbName[0:], databaseName)

	// check for the database entry
	for i := CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {

		if bytes.Compare(o.chunk[i:(i+DATABASE_NAME_LENGTH_MAX)], dropdbName) != 0 {
			continue
		}

		// found it, zero out the database chunk
		copy(o.chunk[i:(i+64)], make([]byte, 64))
		o.chunkID, err = o.statedb.SetDBChunk(o.chunk, 0) // TODO: .eth disc
		if err != nil {
			return fmt.Errorf("[owner:drop] %s", err)
		}
		err = o.statedb.StoreRootHash([]byte(o.Name), o.chunkID)
		if err != nil {
			return fmt.Errorf("[owner:drop] %s", err)
		}

		//drop the database from the owner's list of DatabaseNames
		delete(o.DatabaseNames, databaseName)
		return nil
	}
	return fmt.Errorf("[owner:drop] Database not found in owner chunk")
}
