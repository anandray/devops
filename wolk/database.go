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

type SQLDatabase struct {
	statedb   *StateDB
	Name      string
	Owner     string
	Encrypted int
	//RawSchema  string
	Schema     WVMSchema      //usage: tables[wvmSchema[3].tableName].roothash  where 3 is wvm schema's rootpage
	TableNames map[string]int //key is tablename, int is just 1

	birthts int
	version int
	hash    []byte
	chunk   []byte

	schemaHash  []byte
	schemaChunk []byte
	schemaPath  string
}

const (
	DB_DBHASH_START = CHUNK_START_CHUNKVAL
	DB_SCHEMA_START = CHUNK_START_CHUNKVAL + CHUNK_HASH_SIZE
	DB_TABLES_START = CHUNK_START_CHUNKVAL + 2*CHUNK_HASH_SIZE
)

type Database interface {
	GetRootHash() []byte

	// Insert: adds key-value pair (value is an entire recrod)
	// ok - returns true if new key added
	// Possible Errors: KeySizeError, ValueSizeError, DuplicateKeyError, NetworkError, BufferOverflowError
	Insert(key []byte, value []byte) (bool, error)

	// Put -- inserts/updates key-value pair (value is an entire record)
	// ok - returns true if new key added
	// Possible Errors: KeySizeError, ValueSizeError, NetworkError, BufferOverflowError
	Put(key []byte, value []byte) (bool, error)

	// Get - gets value of key (value is an entire record)
	// ok - returns true if key found, false if not found
	// Possible errors: KeySizeError, NetworkError
	Get(key []byte) ([]byte, bool, error)

	// Delete - deletes key
	// ok - returns true if key found, false if not found
	// Possible errors: KeySizeError, NetworkError, BufferOverflowError
	Delete(key []byte) (bool, error)

	// Start/Flush - any buffered updates will be flushed to SWARM on FlushBuffer
	// ok - returns true if buffer started / flushed
	// Possible errors: NoBufferError, NetworkError
	StartBuffer() (bool, error)
	FlushBuffer() (bool, error)

	// Close - if buffering, then will flush buffer
	// ok - returns true if operation successful
	// Possible errors: NetworkError
	Close() (bool, error)

	// prints what is in memory
	Print()
}

type OrderedDatabase interface {
	Database
	// Seek -- moves cursor to key k
	// ok - returns true if key found, false if not found
	// Possible errors: KeySizeError, NetworkError
	Seek(k []byte /*K*/) (e OrderedDatabaseCursor, ok bool, err error)
	SeekFirst() (e OrderedDatabaseCursor, err error)
	SeekLast() (e OrderedDatabaseCursor, err error)
}

type OrderedDatabaseCursor interface {
	Next() (k []byte /*K*/, v []byte /*V*/, err error)
	Prev() (k []byte /*K*/, v []byte /*V*/, err error)
	GetCurrent() (k []byte, err error)
}

// opens an existing database pulled from the chunk level
func (db *SQLDatabase) open() (ok bool, err error) {

	//check the owner
	owner, ok, err := db.statedb.GetOwner(db.Owner)
	if err != nil {
		return false, fmt.Errorf("[database:open] %s", err)
	}
	if !ok {
		return false, fmt.Errorf("[database:open] owner doesn't exist %s", db.Owner) //owner doesn't exist
	}
	if _, ok = owner.DatabaseNames[db.Name]; !ok {
		return false, nil //this database isn't in the owner chunk
	}

	//retrieve the database chunk
	db.hash = owner.DatabaseNames[db.Name]
	db.chunk, ok, err = db.statedb.GetDBChunk(db.hash)
	if err != nil {
		return false, fmt.Errorf("[database:open] %s", err)
	} else if !ok {
		log.Error("[database:open] db chunk not found", "chunkid", hex.EncodeToString(db.hash))
		return false, nil // do we need to let this fall through?
	} else if len(db.chunk) == 0 || EmptyBytes(db.chunk) {
		log.Error("[database:open] chunk found but empty")
		return false, nil // do we need to let this fall through?
	}
	//fmt.Printf("[database:open] db chunk retrieved: %+v\n", db.chunk)

	//the first 32 bytes of the chunk val should match the dbname
	dbNameBytes := make([]byte, DATABASE_NAME_LENGTH_MAX)
	copy(dbNameBytes[0:], db.Name)
	if bytes.Compare(db.chunk[DB_DBHASH_START:DB_DBHASH_START+DATABASE_NAME_LENGTH_MAX], dbNameBytes[0:DATABASE_NAME_LENGTH_MAX]) != 0 {
		return false, fmt.Errorf("[database:open] Invalid database %x != %x", dbNameBytes, db.chunk[0:CHUNK_HASH_SIZE])
	}

	// pull the schema hash from the db chunk
	copy(db.schemaHash[0:CHUNK_HASH_SIZE], db.chunk[DB_SCHEMA_START:DB_SCHEMA_START+CHUNK_HASH_SIZE])
	if !EmptyBytes(db.schemaHash) { //could be an empty database with no tables yet, then schema would be empty.
		//retrieve the schema chunk
		//log.Info("[database:open] GetDBChunk", "schemahash", hex.EncodeToString(db.schemaHash))
		db.schemaChunk, ok, err = db.statedb.GetDBChunk(db.schemaHash)
		if err != nil {
			return false, fmt.Errorf("[database:open]  %s", err)
		} else if !ok {
			log.Debug("[SQLDatabase:open] Chunk Not found")
			return false, nil
		}
		//write schema chunk to local
		err = WriteWVMSchemaBytes(db.schemaPath, db.schemaChunk)
		if err != nil {
			return false, fmt.Errorf("[database:open]  %s", err)
		}

		// pull the table names from the db chunk
		for i := CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {
			tableName := string(bytes.Trim(db.chunk[i:(i+TABLE_NAME_LENGTH_MAX)], "\x00"))
			//fmt.Printf("[database:open] considering this piece of dbchunk: %s\n", tableName)
			if !EmptyBytes(db.chunk[i:(i + TABLE_NAME_LENGTH_MAX)]) {
				//fmt.Printf("[database:open] table found: %s\n", tableName)
				if _, ok := db.TableNames[tableName]; !ok {
					db.TableNames[tableName] = 1
				}
			}
		}
	}
	return true, nil
}

// updates a databasechunk with a table
// (also updates the corresponding owner with the new database hash)
func (db *SQLDatabase) store(tbl *SQLTable) (err error) {

	if EmptyBytes(db.hash) { //database needs to be created first
		return fmt.Errorf("[database:store] Database %s doesn't exist", db.Name)
	}

	//looks for the table in the database chunk
	found := false
	for i := CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {

		if !EmptyBytes(db.chunk[i:(i + TABLE_NAME_LENGTH_MAX)]) { //not empty, check if it's the table we want

			tableNameFound := string(bytes.Trim(db.chunk[i:(i+TABLE_NAME_LENGTH_MAX)], "\x00"))
			//log.Debug(fmt.Sprintf("Comparing tableName [%s](%+v) to tblNameFound [%s](%+v)", tableName, tableName, tableNameFound, tableNameFound))
			//fmt.Printf("Comparing tableName [%s](%+v) to tblNameFound [%s](%+v)", tbl.Name, tbl.Name, tableNameFound, tableNameFound)
			if strings.Compare(tbl.Name, tableNameFound) == 0 {
				found = true
				break
				//return &sqlcom.SQLError{Message: fmt.Sprintf("[sqlchain:CreateTable] table exists already"), ErrorCode: 500, ErrorMessage: "Table exists already"}
			}

		} else { // found an empty table slot, update the table name in database chunk

			tableNameBytes := make([]byte, TABLE_NAME_LENGTH_MAX)
			copy(tableNameBytes[0:TABLE_NAME_LENGTH_MAX], tbl.Name)
			copy(db.chunk[i:(i+TABLE_NAME_LENGTH_MAX)], tableNameBytes[0:TABLE_NAME_LENGTH_MAX])
			//log.Debug(fmt.Sprintf("Copying tableName [%s] to bufDB [%s]", tableNameBytes[0:32], db.chunk[i:(i+32)]))
			//fmt.Printf("Copying tableName [%s] to bufDB [%s]", tableNameBytes[0:32], db.chunk[i:(i+32)])
			found = true
			break

		}

	}

	if !found {
		return fmt.Errorf("[database:store] Table could not be created -- exceeded Database chunk allocation")
	}

	// update the schema in the database chunk
	db.schemaChunk, err = ReadWVMSchemaBytes(db.schemaPath) // make the chunk from the local schema
	if err != nil {
		return fmt.Errorf("[database.store] %s", err)
	}
	//fmt.Printf("[database.store] db.schemachunk: \n%s\n", string(db.schemaChunk))
	db.schemaHash, err = db.statedb.SetDBChunk(db.schemaChunk, db.Encrypted)
	if err != nil {
		return fmt.Errorf("[database.store] %s", err)
	}
	copy(db.chunk[DB_SCHEMA_START:DB_SCHEMA_START+CHUNK_HASH_SIZE], db.schemaHash[0:CHUNK_HASH_SIZE])

	// write the database chunk
	db.hash, err = db.statedb.SetDBChunk(db.chunk, db.Encrypted)
	if err != nil {
		return fmt.Errorf("[database.store] %s", err)
	}
	//fmt.Printf("[database:store] setting dbchunk %x\n", db.hash)
	db.TableNames[tbl.Name] = 1

	// update the database hash in the owner's databases
	owner, ok, err := db.statedb.GetOwner(db.Owner)
	if err != nil {
		return fmt.Errorf("[database.store] %s", err)
	}
	if !ok {
		return fmt.Errorf("[database.store] owner %s does not exist", db.Owner)
	}
	err = owner.store(db)
	if err != nil {
		return fmt.Errorf("[database.store] %s", err)
	}

	return nil
}

// drops a table from ENS
// drops table from database chunk, updates database hash
// updates database hash in owner chunk
func (db *SQLDatabase) drop(tableName string) (err error) {

	// nuke the table name in the database chunk and write the updated database
	dropTableName := make([]byte, TABLE_NAME_LENGTH_MAX)
	copy(dropTableName[0:], tableName)

	foundTable := false
	for j := CHUNK_START_CHUNKVAL + 64; j < CHUNK_SIZE; j += 64 {

		if bytes.Compare(db.chunk[j:(j+TABLE_NAME_LENGTH_MAX)], dropTableName) != 0 {
			continue
		}

		// found the table, blank it out
		foundTable = true
		log.Debug(fmt.Sprintf("Found Table: Attempting to delete from DB Chunk"))
		blankTableName := make([]byte, TABLE_NAME_LENGTH_MAX)
		copy(db.chunk[j:(j+TABLE_NAME_LENGTH_MAX)], blankTableName[0:TABLE_NAME_LENGTH_MAX])
		break

	}

	if !foundTable { //did not find the table in the database chunk
		return fmt.Errorf("[database.drop] Table not found")
	}

	// update the schema in the database chunk
	db.schemaChunk, err = ReadWVMSchemaBytes(db.schemaPath) // make the chunk from the local schema
	if err != nil {
		return fmt.Errorf("[database.drop] %s", err)
	}
	//fmt.Printf("[database.drop] db.schemachunk: %s\n", string(db.schemaChunk))
	db.schemaHash, err = db.statedb.SetDBChunk(db.schemaChunk, db.Encrypted)
	if err != nil {
		return fmt.Errorf("[database.drop] %s", err)
	}
	copy(db.chunk[DB_SCHEMA_START:DB_SCHEMA_START+CHUNK_HASH_SIZE], db.schemaHash[0:CHUNK_HASH_SIZE])

	// store updated db chunk
	db.hash, err = db.statedb.SetDBChunk(db.chunk, db.Encrypted)
	if err != nil {
		return fmt.Errorf("[database.drop] %s", err)
	}
	log.Debug(fmt.Sprintf("[database:drop] Update DB Chunk after blanking out [%s]", dropTableName))

	// update the database hash in the owner chunk
	owner, ok, err := db.statedb.GetOwner(db.Owner)
	if err != nil {
		return fmt.Errorf("[database.drop] %s", err)
	}
	if !ok {
		return fmt.Errorf("[database:drop] owner %s does not exist", db.Owner)
	}
	err = owner.store(db)
	if err != nil {
		return fmt.Errorf("[database.drop] %s", err)
	}

	//Drop Table from ENS hash as well as db columns
	tblKey := db.statedb.GetTableKey(db.Owner, db.Name, tableName)
	err = db.statedb.StoreRootHash([]byte(tblKey), make([]byte, 64)) //store an empty root hash
	if err != nil {
		return fmt.Errorf("[database.drop] %s", err)
	}

	//drop tablename from db's list of tablenames
	delete(db.TableNames, tableName)

	return nil

}
