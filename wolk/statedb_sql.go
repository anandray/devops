// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
)

// SQL objects
type SQLStateDB struct {
	stateObjects map[common.Hash]*stateObject
	Owners       map[string]*SQLOwner    //key is ownerName
	Databases    map[string]*SQLDatabase //key is ownerName|databaseName
	Tables       map[string]*SQLTable    //key is ownerName|databaseName|tableName
	DataDir      string
	//hash          []byte
}

// DBChunkStore API for SQL
type DBChunkstore interface {
	GetDBChunk(key []byte) (val []byte, ok bool, err error)
	SetDBChunk(val []byte, encrypted int) (key []byte, err error)
}

// Commit writes the state obj to the cloudstore via the SMT
func (statedb *StateDB) CommitSQL(ctx context.Context, wg *sync.WaitGroup, writeToCloudstore bool) (err error) {
	//log.Trace("[statedb_sql:CommitSQL] START")
	for i, stateObject := range statedb.sql.stateObjects {
		//log.Info("[statedb_sql:CommitSQL] updating sql state object!", "stateobj", stateObject)
		err := statedb.updateStateObject(ctx, stateObject)
		if err != nil {
			log.Error("[statedb_sql:CommitSQL] %s", err)
			return fmt.Errorf("[statedb_sql:CommitSQL] %s", err)
		}
		delete(statedb.sql.stateObjects, i)
	}
	//log.Trace("[statedb_sql:CommitSQL] END")
	return nil
}

func (statedb *StateDB) NewOwner(ownerName string) *SQLOwner {
	o := new(SQLOwner)
	o.Name = ownerName
	o.hash = GetHash(ownerName)
	o.chunkID = make([]byte, CHUNK_HASH_SIZE)
	o.chunk = make([]byte, CHUNK_SIZE)
	o.DatabaseNames = make(map[string][]byte)
	o.statedb = statedb
	return o
}

func (statedb *StateDB) NewDatabase(ownerName string, databaseName string, encrypted int) *SQLDatabase {
	db := new(SQLDatabase)
	db.Name = databaseName
	db.Owner = ownerName
	db.Encrypted = encrypted

	db.Schema = make(WVMSchema) // map[int]WVMIndex //should be NewWVMSchema
	db.TableNames = make(map[string]int)
	db.statedb = statedb

	db.birthts = int(time.Now().Unix())
	db.version = 0 // TODO:
	//db.chunkType = []byte("D")
	db.hash = make([]byte, CHUNK_HASH_SIZE)
	db.chunk = make([]byte, CHUNK_SIZE)

	db.schemaHash = make([]byte, CHUNK_HASH_SIZE)
	db.schemaChunk = make([]byte, CHUNK_SIZE)
	dd := statedb.sql.DataDir
	if len(dd) > 0 && string(dd[len(dd)-1]) != "/" {
		dd = dd + "/"
	}
	db.schemaPath = dd + statedb.GetDatabaseKey(ownerName, databaseName) + ".db"
	//log.Info("[statedb_sql:NewDatabase] db.schemapath %s", db.schemaPath)
	return db
}

//TODO: remove this. it's for testing only, in wvm_test.go
func (db *SQLDatabase) GetSchemaPath() (path string) {
	return db.schemaPath
}

func (statedb *StateDB) NewTable(owner string, database string, tableName string) *SQLTable {
	t := new(SQLTable)
	t.statedb = statedb
	t.Owner = owner
	t.Database = database
	t.Name = tableName
	t.columns = make(map[string]*ColumnInfo)

	t.birthts = int(time.Now().Unix())
	t.version = 0 //TODO
	//t.nodeType = make([]byte, ??) //TODO

	t.roothash = make([]byte, CHUNK_HASH_SIZE)
	t.chunk = make([]byte, CHUNK_SIZE)

	return t
}

func (statedb *StateDB) CreateOwner(ownerName string) (ok bool, err error) {

	//log.Info("[statedb_sql:CreateOwner] creating new owner", "name", ownerName)
	owner := statedb.NewOwner(ownerName)
	ok, err = owner.open()
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateOwner] %s", err)
	}
	if ok {
		return false, nil //owner already exists
	}
	statedb.RegisterOwner(owner)
	return true, nil

}

// removes an owner and all its databases and corresponding tables
func (statedb *StateDB) DropOwner(ownerName string) (ok bool, err error) {
	// get the owner
	// remove all the Databases
	// clean out the owner chunk
	// de-register owner chunk
	return false, fmt.Errorf("[statedb_sql:DropOwner] Not implemented")
}

func (statedb *StateDB) GetOwner(ownerName string) (owner *SQLOwner, ok bool, err error) {

	//owner already exists locally
	if _, ok := statedb.sql.Owners[ownerName]; ok {
		//log.Info("[statedb_sql:GetOwner] Owner found locally")
		return statedb.sql.Owners[ownerName], true, nil
	}

	//log.Info("[statedb_sql:GetOwner] Retrieving Owner chunk")
	// go retrieve the owner chunk
	owner = statedb.NewOwner(ownerName)
	ok, err = owner.open()
	if err != nil {
		return owner, false, fmt.Errorf("[statedb_sql:GetOwner] openOwner %s", err)
	}
	if !ok {
		//log.Info("[statedb_sql:GetOwner] owner chunk doesn't exist. Continuing.")
		return owner, false, nil //owner chunk doesn't exist don't error but return ok=false
	}

	//register locally
	statedb.RegisterOwner(owner)
	return owner, true, nil
}

func (statedb *StateDB) RegisterOwner(owner *SQLOwner) {
	ownerKey := statedb.GetOwnerKey(owner.Name)
	statedb.sql.Owners[ownerKey] = owner
	return
}

func (statedb *StateDB) GetOwnerKey(owner string) (key string) {
	return owner
}

func (statedb *StateDB) GetDatabaseKey(owner string, database string) (key string) {
	return fmt.Sprintf("%s|%s", owner, database)
}

func (statedb *StateDB) GetDatabase(ownerName string, databaseName string) (db *SQLDatabase, ok bool, err error) {
	//log.Debug(fmt.Sprintf("[statedb_sql:GetDatabase] Getting Database [%s] with the Owner [%s] from database [%v]", databaseName, ownerName, statedb.sql.Databases))

	// database exists locally
	databaseKey := statedb.GetDatabaseKey(ownerName, databaseName)
	//fmt.Println("[statedb_sql:GetDatabase] databaseKey = ", databaseKey)
	//fmt.Println("[statedb_sql:GetDatabase] statedb = ", statedb)
	if db, ok = statedb.sql.Databases[databaseKey]; ok {
		//fmt.Printf("[statedb_sql:GetDatabase] Database found in local memory\n")
		return db, true, nil
	}

	// go retrieve the db chunk
	db = statedb.NewDatabase(ownerName, databaseName, 0)
	ok, err = db.open()
	if err != nil {
		//log.Debug("[statedb_sql:GetDatabase] Error", "err", err)
		return db, false, fmt.Errorf("[statedb_sql:GetDatabase] %s", err)
	}
	if !ok {
		return db, false, nil //database chunk doesn't exist
	}

	//update the wvmSchema
	db.Schema, err = GetWVMSchema(db.schemaPath)
	//fmt.Printf("[statedb_sql:GetDatabase] schema gotten is: %+v\n", db.Schema)
	if err != nil {
		//log.Debug("[statedb_sql:GetDatabase] Error", "err", err)
		return db, false, fmt.Errorf("[statedb_sql:GetDatabase] %s", err)
	}

	statedb.RegisterDatabase(db)
	//log.Debug(fmt.Sprintf("Owner [%s] Database %s found in databases, it is: %+v\n", ownerName, databaseName, db))
	return db, true, nil
}

//lists databases from the chunk level
func (statedb *StateDB) ListDatabases(ownerName string) (result map[string][]byte, err error) {
	owner := statedb.NewOwner(ownerName)
	ok, err := owner.open()
	if err != nil {
		log.Error("[statedb_sql:ListDatabases] Error", "err", err)
		return result, fmt.Errorf("[statedb_sql:ListDatabases] %s", err)
	}
	if !ok {
		return result, fmt.Errorf("[statedb_sql:ListDatabases] owner not found")
	}
	return owner.DatabaseNames, nil
}

func (statedb *StateDB) RegisterDatabase(database *SQLDatabase) {
	databaseKey := statedb.GetDatabaseKey(database.Owner, database.Name)
	statedb.sql.Databases[databaseKey] = database //overwrites if already exists
	return
}

func (statedb *StateDB) UnregisterDatabase(ownerName string, databaseName string) {
	// unregister the database in sqlchain
	databaseKey := statedb.GetDatabaseKey(ownerName, databaseName)
	delete(statedb.sql.Databases, databaseKey)
	//fmt.Printf("[statedb_sql:UnregisterDatabase] databaseKey: %s, databases shouldn't have key: %+v\n", databaseKey, statedb.sql.Databases)

	//TODO:
	// unregister the owner if there are no more databases owned by that owner
	// for _, db := range statedb.sql.Databases {
	// 	if db.Owner == ownerName {
	// 		return //do not unregister owner
	// 	}
	// }
	// delete(statedb.sql.Owners, ownerName)

}

// creating a database results in a new entry, e.g. "videos" in the owners ENS e.g. "wolktoken.eth" stored in a single chunk
// e.g.  key 1: wolktoken.eth (up to 64 chars)
//       key 2: videos     => 32 byte hash, pointing to tables of "video'
func (statedb *StateDB) CreateDatabase(ownerName string, databaseName string, encrypted int) (ok bool, err error) {
	//log.Info("[statedb_sql:CreateDatabase] CREATING DB")
	// check owner chunk for existing owner
	_, ok, err = statedb.GetOwner(ownerName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateDatabase] %s", err)
	}
	if !ok { //owner doesn't exist, create a new owner
		_, err := statedb.CreateOwner(ownerName)
		if err != nil {
			return false, fmt.Errorf("[statedb_sql:CreateDatabase] %s", err)
		}
	}
	//log.Info("[statedb_sql:CreateDatabase] CREATING NEW DB")
	// check database chunks for existing database
	_, ok, err = statedb.GetDatabase(ownerName, databaseName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateDatabase] GetDatabase: %s", err)
	}
	if ok {
		//log.Info("[statedb_sql:CreateDatabase] db exists already")
		return false, nil //database exists already
	}

	//log.Info("[statedb_sql:CreateDatabase] Making new DATABASE Chunk")
	// make a new database chunk
	newdb := statedb.NewDatabase(ownerName, databaseName, encrypted)
	newdbName := make([]byte, DATABASE_NAME_LENGTH_MAX)
	copy(newdbName[0:], databaseName)

	chunkHeader, err := statedb.BuildChunkHeader(newdb.Owner, newdb.Name, "", []byte("D"), newdb.Encrypted)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateDatabase] %s", err)
	}
	copy(newdb.chunk[0:CHUNK_START_CHUNKVAL], chunkHeader)

	//the first 32 bytes of the chunk "val" is the database name (the next keys will be the tables)
	copy(newdb.chunk[DB_DBHASH_START:DB_DBHASH_START+DATABASE_NAME_LENGTH_MAX], newdbName[0:DATABASE_NAME_LENGTH_MAX])

	//fmt.Printf("[statedb_sql:GetDatabase] newdb's 1st 32 bytes of chunk val: %+v\n", newdb.chunk[DB_DBHASH_START:DB_DBHASH_START+DATABASE_NAME_LENGTH_MAX])

	//log.Info("[statedb_sql:CreateDatabase] Storing New DB Chunk")
	newdb.hash, err = statedb.SetDBChunk(newdb.chunk, encrypted)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateDatabase] %s", err)
	}

	// make a new database entry in the owner chunk
	if _, ok := statedb.sql.Owners[ownerName]; !ok {
		return false, fmt.Errorf("[statedb_sql:CreateDatabase] Owner %s Doesn't Exist", ownerName)
	}

	//log.Info("[statedb_sql:CreateDatabase] Storing DATABASE")
	err = statedb.sql.Owners[ownerName].store(newdb)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateDatabase] %s", err)
	}

	//register
	statedb.RegisterDatabase(newdb)
	return true, nil
}

// dropping a database removes the ENS entry
func (statedb *StateDB) DropDatabase(ownerName string, databaseName string) (ok bool, err error) {
	db, ok, err := statedb.GetDatabase(ownerName, databaseName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:DropDatabase] %s %s", databaseName, err)
	}
	if !ok {
		return false, nil //database doesn't exist
	}
	log.Info(fmt.Sprintf("[statedb_sql:DropDatabase] database gotten to drop: %s, with tables %+v\n", db.Name, db.TableNames))

	// drop any existing tables that database had
	for dbTableName, _ := range db.TableNames {
		//fmt.Printf("dropping table: %s\n", dbTableName)
		ok, err := statedb.DropTable(ownerName, databaseName, dbTableName)
		if err != nil {
			return false, fmt.Errorf("[statedb_sql:DropDatabase] %s", err)
		}
		if !ok {
			//weird state: table is in database register, but for some reason doesn't actually exist
			return false, fmt.Errorf("[statedb_sql:DropDatabase] table %s is listed as in database %s for owner %s but doesn't exist to drop", dbTableName, databaseName, ownerName)
			//fmt.Printf("[statedb_sql:DropDatabase] table %s is listed as in database %s for owner %s but doesn't exist to drop\n", dbTableName, databaseName, ownerName)
		}
	}

	// TODO: clean out schema? delete file. not really necessary but nice to have cleanup.

	// drop the database from the owner chunk
	owner, ok, err := statedb.GetOwner(ownerName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:DropDatabase] GetOwner %s", err)
	}
	if !ok {
		return false, fmt.Errorf("[statedb_sql:DropDatabase] Owner %s", err)
	}
	err = owner.drop(databaseName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:DropDatabase] drop %s", err)
	}

	//unregister the database locally (TODO: and the owner if they have no more databases?)
	statedb.UnregisterDatabase(ownerName, databaseName)

	return true, nil
}

func (statedb *StateDB) GetTable(ownerName string, databaseName string, tableName string) (tbl *SQLTable, proof *Proof, ok bool, err error) {

	// table exists locally
	tblKey := statedb.GetTableKey(ownerName, databaseName, tableName)
	if tbl, ok := statedb.sql.Tables[tblKey]; ok { //found it
		//log.Debug(fmt.Sprintf("[statedb_sql:GetTable] Table [%v] with Owner [%s] Database %s found in local tables\n", tblKey, ownerName, databaseName))
		if !EmptyBytes(tbl.roothash) {
			//fmt.Printf("[statedb_sql:GetTable] table roothash is not empty, real registry - found locally\n")
			// for _, c := range tbl.columns {
			// 	fmt.Printf("[statedb_sql:GetTable] t col: %+v\n", c)
			// }
			return tbl, nil, true, nil
		}
	}

	tbl = statedb.NewTable(ownerName, databaseName, tableName)
	proof, ok, err = tbl.open()
	if err != nil {
		return tbl, proof, false, fmt.Errorf("[statedb_sql:GetTable] %s", err)
	}
	if !ok {
		return tbl, proof, false, nil //table doesn't exist
	}
	statedb.RegisterTable(tbl)

	// fmt.Printf("\n[statedb_sql:GetTable] table gotten: %+v\n", tbl)
	// for _, c := range tbl.columns {
	// fmt.Printf("[statedb_sql:GetTable] t col: %+v\n", c)
	//}

	return tbl, proof, true, nil
}

func (statedb *StateDB) RegisterTable(t *SQLTable) {
	tblKey := statedb.GetTableKey(t.Owner, t.Database, t.Name)
	statedb.sql.Tables[tblKey] = t
}

func (statedb *StateDB) UnregisterTable(owner string, database string, tableName string) {
	tblKey := statedb.GetTableKey(owner, database, tableName)
	//TODO: Empty out column info?
	delete(statedb.sql.Tables, tblKey)
}

func (statedb *StateDB) DropTable(ownerName string, databaseName string, tableName string) (ok bool, err error) {

	//og.Debug(fmt.Sprintf("Attempting to Drop table [%s]", tableName))
	_, _ /*proof*/, ok, err = statedb.GetTable(ownerName, databaseName, tableName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:DropTable] tbl %s %s", tableName, err)
	}
	if !ok {
		return false, nil //table doesn't exist
	}

	// drop the table chunk, drop the table from database chunk and update the database hash in the owner chunk
	db, ok, err := statedb.GetDatabase(ownerName, databaseName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:DropTable] db %s %s", databaseName, err)
	}
	if !ok {
		return false, nil //database doesn't exist
	}

	// update db schema
	sql := MakeDropTableQuery(tableName)
	//fmt.Printf("[statedb_sql:DropTable] making drop table query %s\n", sql)
	if err = SetWVMSchema(db.schemaPath, sql); err != nil {
		return false, fmt.Errorf("[statedb_sql:DropTable] %s", err)
	}
	//fmt.Printf("[statedb_sql:DropTable] update wvm schema \n")

	err = db.drop(tableName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:DropTable] database.drop db:%s tbl:%s %s", db.Name, tableName, err)
	}

	statedb.UnregisterTable(ownerName, databaseName, tableName)
	return true, nil
}

func (statedb *StateDB) ListTables(ownerName string, databaseName string) (tableNames map[string]int, err error) {
	db := statedb.NewDatabase(ownerName, databaseName, 0)
	ok, err := db.open()
	if err != nil {
		return tableNames, fmt.Errorf("[statedb_sql:ListTables] %s", err)
	}
	if !ok {
		return tableNames, fmt.Errorf("[statedb_sql:ListTables] Database doesn't exist")
	}
	//fmt.Printf("[statedb_sql:ListTables] db has tables: %+v\n", db.TableNames)
	return db.TableNames, nil
}

// TODO: Review adding owner string, database string input parameters where the goal is to get database.owner/table/key type HTTP urls like:
//       https://swarm.wolk.com/wolkinc.eth => GET: ListDatabases
//       https://swarm.wolk.com/videos.wolkinc.eth => GET; ListTables
//       https://swarm.wolk.com/videos.wolkinc.eth/user => GET: DescribeTable
//       https://swarm.wolk.com/videos.wolkinc.eth/user/sourabhniyogi => GET: Get
// TODO: check for the existence in the owner-database combination before creating.
// TODO: need to make sure the types of the columns are correct
func (statedb *StateDB) CreateTable(ownerName string, databaseName string, tableName string, columns []Column, primaryColumnName string) (ok bool, err error) {

	_, _ /*proof*/, ok, err = statedb.GetTable(ownerName, databaseName, tableName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateTable] GetTable: tbl %s err %s", tableName, err)
	}
	if ok {
		return false, nil //table exists already
	}

	db, ok, err := statedb.GetDatabase(ownerName, databaseName)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateTable] GetDatabase: db %s err %s", databaseName, err)
	}
	if !ok {
		return false, fmt.Errorf("[statedb_sql:CreateTable] Database could not be found")
	}

	//fmt.Printf("[statedb_sql:CreateTable] database gotten: %+v\n", db)
	//fmt.Printf(">> world state before making table chunk:\no:%+v\nd:%+v\nt:%+v\n", statedb.sql.Owners, statedb.sql.Databases, statedb.sql.Tables)

	//log.Debug(fmt.Sprintf("[statedb_sql:CreateTable] Creating Table [%s] - Owner [%s] Database [%s]\n", tableName, ownerName, databaseName))

	//make a new table struct
	tbl := statedb.NewTable(ownerName, databaseName, tableName)
	// for _, c := range columns {
	// 	fmt.Printf("\n[statedb_sql:CreateTable] column %+v\n", c)
	// }

	err = tbl.columnToColumnInfo(columns)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateTable] %s", err)
	}
	for _, c := range tbl.columns {
		log.Info(fmt.Sprintf("[statedb_sql:CreateTable] columnInfo %+v\n", c))
	}

	tbl.primaryColumnName = primaryColumnName
	tbl.encrypted = db.Encrypted
	tbl.writeChunk()

	//update db schema
	//fmt.Printf("[statedb_sql:CreateTable] making table query\n")
	sql, err := tbl.MakeCreateTableQuery()
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateTable] %s", err)
	}
	//fmt.Printf("[statedb_sql:CreateTable] setting wvm schema with: %s \n", sql)
	if err = SetWVMSchema(db.schemaPath, sql); err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateTable] %s", err)
	}
	db.Schema, err = GetWVMSchema(db.schemaPath)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateTable] %s", err)
	}
	//fmt.Printf("[statedb_sql:CreateTable] schema: %+v\n", db.Schema)

	//add table entry to database chunk, updates database root in owner chunk as well.
	err = db.store(tbl)
	if err != nil {
		return false, fmt.Errorf("[statedb_sql:CreateTable] %s", err)
	}

	//register the new table
	statedb.RegisterTable(tbl)

	return true, nil

}

func (statedb *StateDB) GetTableKey(owner string, database string, tableName string) (key string) {
	return fmt.Sprintf("%s|%s|%s", owner, database, tableName)
}

func parseData(data string) (*SQLRequest, error) {
	udata := new(SQLRequest)
	if err := json.Unmarshal([]byte(data), udata); err != nil {
		return nil, fmt.Errorf("[statedb_sql:parseData] Unmarshal %s", err)
	}
	return udata, nil
}

func (statedb *StateDB) Query(ownerName string, databaseName string, query string) (rows []Row, affectedRows int, proof *Proof, err error) {
	//log.Info("[statedb_sql:Query] QUERY START", "query", query)
	db, ok, err := statedb.GetDatabase(ownerName, databaseName) //should pull schema from chunks and write to local
	if err != nil {
		return rows, 0, nil, fmt.Errorf("[statedb_sql:Query] %s", err)
	}
	if !ok {
		return rows, 0, nil, fmt.Errorf("[statedb_sql:Query] Database does not exist")
	}
	//fmt.Printf("[statedb_sql:Query] wvm schema object is: %+v\n", db.Schema)

	q, err := ParseQuery(query) //also gives you q.Type which is Select/Insert/Update/Delete/CreateTable
	if err != nil {
		return rows, 0, nil, fmt.Errorf("[statedb_sql:Query] %s", err)
	}

	if q.Type == "CreateTable" {
		primaryColumnName, err := checkTableColumns(q.Columns)
		if err != nil {
			return rows, 0, nil, fmt.Errorf("[statedb_sql:Query] %s", err)
		}
		ok, err := statedb.CreateTable(ownerName, databaseName, q.TableName, q.Columns, primaryColumnName)
		if err != nil {
			return rows, 0, nil, fmt.Errorf("[statedb_sql:Query] %s", err)
		}
		if !ok {
			return rows, 0, nil, fmt.Errorf("[statedb_sql:Query] Table already exists")
		}
		return rows, 1, nil, nil
	}

	tbl, proof, ok, err := statedb.GetTable(ownerName, databaseName, q.TableName)
	if err != nil {
		return rows, 0, proof, fmt.Errorf("[statedb_sql:Query] %s", err)
	}
	if !ok {
		return rows, 0, proof, fmt.Errorf("[statedb_sql:Query] Table does not exist")
	}
	//log.Info(fmt.Sprintf("[statedb_sql:Query] Got Table. o:%s, db:%s, tbl:%s", tbl.Owner, tbl.Database, tbl.Name))
	ops, err := CompileOpcodes(db.schemaPath, query)
	if err != nil {
		return rows, 0, proof, fmt.Errorf("[statedb_sql:Query] '%s' %s", query, err)
	}
	// log.Info(fmt.Sprintf("[statedb_sql:Query] - Compiled Opcodes (%+v)", ops))

	// DEBUG
	// for _, op := range ops {
	// 	fmt.Printf("[swarmdb:Query] op: %+v\n", op)
	// }

	vm := NewWVM(statedb, db)
	if q.Type == "Update" {
		vm.SetFlags(q.Type)
	}
	err = vm.ProcessOps(ops)
	if err != nil {
		return rows, 0, proof, fmt.Errorf("[statedb_sql:Query] %s", err)
	}
	log.Info("[statedb_sql:Query] output of processed ops", "rows", vm.Rows)
	indexedColumns := make(map[int]string)
	if q.HasRequestColumns() {
		for _, col := range q.Columns {
			indexedColumns[col.QueryID] = col.ColumnName
		}
	} else {
		for _, col := range tbl.columns {
			indexedColumns[int(col.IDNum)] = col.columnName
		}
	}
	for _, row := range vm.Rows {
		newRow := NewRow()
		index := 0
		for _, rowitem := range row {
			newRow.Set(indexedColumns[index], rowitem) //columnname, value
			index++
		}
		rows = append(rows, newRow)
	}
	log.Info(fmt.Sprintf("[statedb_sql:Query] result = %+v\n", rows))

	err = tbl.FlushBuffer()
	if err != nil {
		log.Error("[statedb_sql:Query] flushbuffer table", "err", err)
	}
	return rows, vm.AffectedRowCount, proof, nil
}

func GetHash(in string) []byte {
	return wolkcommon.Computehash([]byte(in))
}

func checkTableColumns(Columns []Column) (primaryColumnName string, err error) {
	//error checking for Creating a table
	if len(Columns) == 0 {
		return "", fmt.Errorf("[statedb_sql:CheckTable] empty columns")
	}

	for _, col := range Columns {
		if col.Primary > 0 {
			if len(primaryColumnName) > 0 {
				return "", &SQLError{Message: "[statedb_sql:CheckTable] More than one primary column", ErrorCode: 406, ErrorMessage: "Multiple Primary keys specified in Create Table"}
			}
			primaryColumnName = col.ColumnName
		}
		if !CheckColumnType(col.ColumnType) {
			return "", fmt.Errorf("[statedb_sql:CheckTable] bad columntype %s", col.ColumnType)
		}
		if !CheckIndexType(col.IndexType) {
			return "", fmt.Errorf("[statedb_sql:CheckTable] bad indextype %s", col.IndexType)
		}
	}
	if len(primaryColumnName) == 0 {
		return "", fmt.Errorf("[statedb_sql:CheckTable] no primary column indicated")
	}
	return primaryColumnName, nil
}

func sqlrequestChecks(req *SQLRequest) (err error) {

	if len(req.Owner) == 0 {
		return fmt.Errorf("[statedb_sql:sqlrequestChecks] owner missing ")
		//out["err"] = "Owner Missing"
	}
	if len(req.Database) == 0 && req.RequestType != RT_LIST_DATABASES {
		return fmt.Errorf("[statedb_sql:sqlrequestChecks] database missing ")
		//out["err"] = "Database Missing"
	}
	if len([]byte(req.Database)) > DATABASE_NAME_LENGTH_MAX {
		//return & SQLError{Message: "[statedb_sql:sqlrequestChecks] Database Name > MAX", ErrorCode: 500, ErrorMessage: "Database Name too long (max is 32 chars)"}
		//out["err"] = "Database Name too long (max is 32 chars)"
		return fmt.Errorf("[statedb_sql:sqlrequestChecks] %s is %d len, which is > MAX %d", req.Database, len([]byte(req.Database)), DATABASE_NAME_LENGTH_MAX)
	}
	if strings.Contains(strings.ToUpper(req.RequestType), "TABLE") {
		if len(req.Table) == 0 && req.RequestType != RT_LIST_TABLES {
			return fmt.Errorf("[statedb_sql:sqlrequestChecks] tablename missing ")
			//out["err"] = "Table Name Missing"
		}
		if len([]byte(req.Table)) > TABLE_NAME_LENGTH_MAX {
			return fmt.Errorf("[statedb_sql:sqlrequestChecks] Tablename length")
			//out["err"] = "Table Name too long (max is 32 chars)"
		}
		if len(req.Columns) > COLUMNS_PER_TABLE_MAX {
			return fmt.Errorf("[statedb_sql:sqlrequestChecks] Max Allowed Columns for a table is %d and you submit %d", COLUMNS_PER_TABLE_MAX, len(req.Columns))
			//out["err"] = fmt.Sprintf("Max Allowed Columns exceeded - [%d] supplied, max is [%d]", len(req.Columns), COLUMNS_PER_TABLE_MAX)
		}
	}
	return nil

}

func (statedb *StateDB) SelectHandler(req *SQLRequest, withProof bool) (resp SQLResponse, proof *Proof, err error) {

	//todo: implement 'withproof'

	if err = sqlrequestChecks(req); err != nil {
		return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
	}

	switch req.RequestType {

	case RT_CREATE_DATABASE:
		//log.Info("[statedb_sql:SelectHandler] create db:", "req", req)
		ok, err := statedb.CreateDatabase(req.Owner, req.Database, req.Encrypted)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:Selecthandler] Database exists already")
		}
		//log.Info("[statedb_sql:SelectHandler] created db. response: %+v\n", SQLResponse{AffectedRowCount: 1})
		return SQLResponse{AffectedRowCount: 1}, proof, nil

	case RT_DROP_DATABASE:

		ok, err := statedb.DropDatabase(req.Owner, req.Database)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return SQLResponse{AffectedRowCount: 0}, proof, nil //database doesn't exist
		}
		return SQLResponse{AffectedRowCount: 1}, proof, nil

	case RT_LIST_DATABASES:

		log.Info("[statedb_sql:SelectHandler] list databases, resp should be empty", "resp", resp)
		databases, err := statedb.ListDatabases(req.Owner)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		log.Info("[statedb_sql:SelectHandler] list databases", "resp", resp)
		//rearrange to [] Row formatting for resp.Data output
		for dbname, _ := range databases {
			log.Info("[statedb_sql:SelectHandler] list databases", "db", dbname)
			row := NewRow()
			row["database"] = dbname
			resp.Data = append(resp.Data, row)
		}
		resp.MatchedRowCount = len(databases)
		return resp, proof, nil

	case RT_CREATE_TABLE:

		primaryColumnName, err := checkTableColumns(req.Columns) //error checking
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		ok, err := statedb.CreateTable(req.Owner, req.Database, req.Table, req.Columns, primaryColumnName)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Table already exists")
		}
		return SQLResponse{AffectedRowCount: 1}, proof, nil

	case RT_DROP_TABLE:

		ok, err := statedb.DropTable(req.Owner, req.Database, req.Table)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if ok {
			return SQLResponse{AffectedRowCount: 1}, proof, nil
		} else {
			return SQLResponse{AffectedRowCount: 0}, proof, nil
		}

	case RT_DESCRIBE_TABLE:

		tbl, proof, ok, err := statedb.GetTable(req.Owner, req.Database, req.Table)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] [%s] not found", req.Table)
		}
		tblcols, err := tbl.DescribeTable()
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if len(tblcols) == 0 {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Table [%s] not found", req.Table)
		}
		//rearrange for standard output
		for _, colInfo := range tblcols {
			r := NewRow()
			r["ColumnName"] = colInfo.ColumnName
			r["IndexType"] = colInfo.IndexType
			r["Primary"] = colInfo.Primary
			r["ColumnType"] = colInfo.ColumnType
			resp.Data = append(resp.Data, r)

		}

		return resp, proof, nil

	case RT_LIST_TABLES:

		tables, err := statedb.ListTables(req.Owner, req.Database)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		for tblname, _ := range tables {
			row := NewRow()
			row["table"] = tblname
			resp.Data = append(resp.Data, row)
		}
		resp.MatchedRowCount = len(tables)
		//log.Debug(fmt.Sprintf("returning resp %+v tablenames %+v Mrc %+v", resp, tables, len(tables)))
		return resp, proof, nil

	case RT_PUT:

		tbl, _ /*proof*/, ok, err := statedb.GetTable(req.Owner, req.Database, req.Table)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Put - Table does not exist")
		}
		//fmt.Printf("[statedb_sql:SelectHandler:RT_PUT] tbl cols gotten: %+v\n", tbl.columns)
		req.Rows, err = tbl.assignRowColumnTypes(req.Rows)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}

		//error checking for primary column, and valid columns
		for _, row := range req.Rows {
			//log.Debug(fmt.Sprintf("[statedb_sql:SelectHandler] checking row %v\n", row))
			if _, ok := row[tbl.primaryColumnName]; !ok {
				return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Put row %+v needs primary column '%s' value", row, tbl.primaryColumnName)
			}
			for columnName, _ := range row {
				if _, ok := tbl.columns[columnName]; !ok {
					return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Put row %+v has unknown column %s", row, columnName)
				}
			}
		}

		//put the rows in
		//for colname, col := range tbl.columns {
		//fmt.Printf("[statedb_sql:SelectHandler] col: %s, %+v\n", colname, col)
		//}
		successfulRows := 0
		for _, row := range req.Rows {
			//fmt.Printf("[statedb_sql:SelectHandler] row to be put in: %+v\n", row)
			err = tbl.Put(row, true)
			if err != nil {
				return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
			}
			successfulRows++
		}
		return SQLResponse{AffectedRowCount: successfulRows}, proof, nil

	case RT_GET:

		if isNil(req.Key) {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Get - Missing Key")
		}
		tbl, proof, ok, err := statedb.GetTable(req.Owner, req.Database, req.Table)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Get - Table does not exist")
		}
		if _, ok := tbl.columns[tbl.primaryColumnName]; !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Get - Primary Key Not found in Column Definition")
		}
		primaryColumnType := tbl.columns[tbl.primaryColumnName].columnType
		//fmt.Printf("[sqlchain.SelectHandler] RT_GET key: %x\n", req.Key)
		convertedKey, err := convertValueToBytes(primaryColumnType, req.Key)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		//fmt.Printf("[sqlchain.SelectHandler] RT_GET convertedkey: %x\n", convertedKey)
		byteRow, ok, err := tbl.Get(convertedKey)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, nil //nothing gotten
		}
		//fmt.Printf("[sqlchain.SelectHandler] RT_GET byteRow: %x\n", byteRow)
		validRow, err := tbl.byteArrayToRow(byteRow)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}

		resp.Data = append(resp.Data, validRow)
		resp.MatchedRowCount = 1
		return resp, proof, nil

	case RT_DELETE:

		if isNil(req.Key) {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Delete is Missing Key")
		}
		tbl, proof, ok, err := statedb.GetTable(req.Owner, req.Database, req.Table)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] Delete - Table does not exist")
		}
		ok, err = tbl.Delete(req.Key)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return SQLResponse{AffectedRowCount: 0}, proof, nil //nothing deleted
		}
		return SQLResponse{AffectedRowCount: 1}, proof, nil

	case RT_START_BUFFER:

		tbl, proof, ok, err := statedb.GetTable(req.Owner, req.Database, req.Table)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] StartBuffer - Table does not exist")
		}
		err = tbl.StartBuffer()
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		//TODO: update to use real "count"
		return SQLResponse{AffectedRowCount: 1}, proof, nil

	case RT_FLUSH_BUFFER:

		tbl, proof, ok, err := statedb.GetTable(req.Owner, req.Database, req.Table)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		if !ok {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] FlushBuffer - Table does not exist")
		}
		err = tbl.FlushBuffer()
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}
		//TODO: update to use real "count"
		return SQLResponse{AffectedRowCount: 1}, proof, nil

	case RT_QUERY:

		if len(req.RawQuery) == 0 {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] RawQuery is blank")
		}

		var rows []Row
		rows, resp.AffectedRowCount, proof, err = statedb.Query(req.Owner, req.Database, req.RawQuery)
		if err != nil {
			return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] %s", err)
		}

		resp.Data = rows
		resp.MatchedRowCount = len(rows)
		log.Info("[statedb_sql:SelectHandler] official resp", "r", resp)
		return resp, proof, nil

	} //end switch
	return resp, proof, fmt.Errorf("[statedb_sql:SelectHandler] RequestType invalid: [%s]", req.RequestType)
}

const (
	DATABASE_NAME_LENGTH_MAX = 31
	TABLE_NAME_LENGTH_MAX    = 32
	COLUMN_NAME_LENGTH_MAX   = 24
	DATABASES_PER_USER_MAX   = 30
	COLUMNS_PER_TABLE_MAX    = 30

	hashChunkSize            = 4000
	epochSeconds             = 600
	EpochSeconds             = 600
	CHUNK_HASH_SIZE          = 32
	CHUNK_START_SIG          = 0
	CHUNK_END_SIG            = 65
	CHUNK_START_MSGHASH      = 65
	CHUNK_END_MSGHASH        = 97
	CHUNK_START_PAYER        = 97
	CHUNK_END_PAYER          = 129
	CHUNK_START_CHUNKTYPE    = 129
	CHUNK_END_CHUNKTYPE      = 130
	CHUNK_START_MINREP       = 130
	CHUNK_END_MINREP         = 131
	CHUNK_START_MAXREP       = 131
	CHUNK_END_MAXREP         = 132
	CHUNK_START_BIRTHTS      = 132
	CHUNK_END_BIRTHTS        = 140
	CHUNK_START_LASTUPDATETS = 140
	CHUNK_END_LASTUPDATETS   = 148
	CHUNK_START_ENCRYPTED    = 148
	CHUNK_END_ENCRYPTED      = 149
	CHUNK_START_VERSION      = 149
	CHUNK_END_VERSION        = 157
	CHUNK_START_RENEW        = 157
	CHUNK_END_RENEW          = 158
	CHUNK_START_KEY          = 158
	CHUNK_END_KEY            = 190
	CHUNK_START_OWNER        = 190
	CHUNK_END_OWNER          = 222
	CHUNK_START_DB           = 222
	CHUNK_END_DB             = 254
	CHUNK_START_TABLE        = 254
	CHUNK_END_TABLE          = 286
	//CHUNK_START_EPOCHTS      = 254
	//CHUNK_END_EPOCHTS        = 286
	CHUNK_START_ITERATOR = 416
	CHUNK_END_ITERATOR   = 512
	CHUNK_START_CHUNKVAL = 512
	CHUNK_END_CHUNKVAL   = 4096
)

func (statedb *StateDB) GetDBChunkStore() ChunkStore {
	return statedb.Storage
}

/*
func (statedb *StateDB) GenerateChunkstoreChunkID(blockchainID uint64, chunkID []byte) []byte {
	byteBlockchainID := deep.UInt64ToByte(blockchainID)
	newKey := make([]byte, len(byteBlockchainID)+len(chunkID))
	copy(newKey[:len(byteBlockchainID)], byteBlockchainID)
	copy(newKey[len(byteBlockchainID):], chunkID)
	//log.Info(fmt.Sprintf("KeyManagement: BlockchainID [%x] | k [%x] | result[%x]", byteBlockchainID, chunkID, newKey))
	return newKey
}
*/

func (statedb *StateDB) GetDBChunk(key []byte) (val []byte, ok bool, err error) {

	//log.Info("[statedb_sql:GetDBChunk] get", "key", key)
	if EmptyBytes(key) || BytesToInt(key) == 0 {
		log.Error("[statedb_sql:GetDBChunk] trying to get an empty key.")
		return val, false, nil
	}
	val, ok, err = statedb.Storage.GetChunk(context.TODO(), key)
	if err != nil {
		return val, false, fmt.Errorf("[statedb_sql:GetDBChunk] %s", err)
	} else if !ok {
		log.Error("[statedb_sql:GetDBChunk] chunk key not found. Continuing.")
		val = make([]byte, CHUNK_SIZE)
	} else if len(val) == 0 || EmptyBytes(val) { //Val is either not found or empty
		log.Error("[statedb_sql:GetDBChunk] chunk val was empty. Continuing.")
		val = make([]byte, CHUNK_SIZE)
	}
	return val, true, nil
}

func (statedb *StateDB) SetDBChunk(val []byte, encrypted int) (key []byte, err error) {

	key = wolkcommon.Computehash(val)
	//TODO: put in Storage
	//	_, err = statedb.ChunkStore.SetChunk(nil, val)
	_, err = statedb.Storage.QueueChunk(nil, val)
	if err != nil {
		return key, fmt.Errorf("[statedb_sql:SetDBChunk] %s", err)
	}
	return key, nil
}

//TODO: use ChunkHeader struct
func (statedb *StateDB) BuildChunkHeader(ownerName string, databaseName string, tableName string, nodeType []byte, encrypted int) (ch []byte, err error) {

	ch = make([]byte, CHUNK_START_CHUNKVAL)
	copy(ch[CHUNK_START_OWNER:CHUNK_END_OWNER], []byte(ownerName))
	copy(ch[CHUNK_START_DB:CHUNK_END_DB], []byte(databaseName))
	copy(ch[CHUNK_START_TABLE:CHUNK_END_TABLE], []byte(tableName))
	//	copy(ch[CHUNK_START_KEY:CHUNK_END_KEY], key)
	copy(ch[CHUNK_START_PAYER:CHUNK_END_PAYER], "aaaa" /*TODO:u.Address*/)
	copy(ch[CHUNK_START_CHUNKTYPE:CHUNK_END_CHUNKTYPE], nodeType) // O = OWNER | D = Database | T = Table | x,h,d,k = various data nodes
	//copy(ch[CHUNK_START_RENEW:CHUNK_END_RENEW], IntToByte(u.AutoRenew))
	//copy(ch[CHUNK_START_MINREP:CHUNK_END_MINREP], IntToByte(u.MinReplication))
	//copy(ch[CHUNK_START_MAXREP:CHUNK_END_MAXREP], IntToByte(u.MaxReplication))
	copy(ch[CHUNK_START_ENCRYPTED:CHUNK_END_ENCRYPTED], IntToByte(encrypted))

	rawMetadata := ch[CHUNK_END_MSGHASH:CHUNK_START_CHUNKVAL]
	msgHash := SignHash(rawMetadata)

	//TODO: msg_hash --
	copy(ch[CHUNK_START_MSGHASH:CHUNK_END_MSGHASH], msgHash)

	//km := statedb.dbchunkstore.GetKeyManager()
	/*	sdataSig, errSign := statedb.Cloudstore.SignMessage(msg_hash, statedb.dbchunkstore.config.ParentPrivateKey)
			if errSign != nil {
				return ch, & SQLError{Message: `[BuildChunkHeader] SignMessage ` + errSign.Error(), ErrorCode: 455, ErrorMessage: "Keymanager Unable to Sign Message"}
			}
		//TODO: Sig -- document this
		copy(ch[CHUNK_START_SIG:CHUNK_END_SIG], sdataSig)
		//log.Debug(fmt.Sprintf("Metadata is [%+v]", ch))
	*/

	return ch, err

	//From buildSData in table.go:
	// mergedBodycontent = make([]byte, CHUNK_SIZE)
	// copy(mergedBodycontent[:], metadataBody)
	// copy(mergedBodycontent[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL], value) // expected to be the encrypted body content
	//
	// log.Debug(fmt.Sprintf("Merged Body Content: [%v]", mergedBodycontent))
	// return mergedBodycontent, err
}

func (statedb *StateDB) GetRootHash(key []byte) (roothash []byte, proof *Proof, err error) {

	keyhash := common.BytesToHash(wolkcommon.Computehash(key))
	//k := wolkcommon.BytesToUint64(keyhash[24:32])
	log.Info("[statedb_sql:GetRootHash] get", "key", hex.EncodeToString(key), "uint64 key", key)
	log.Info("[statedb_sql:GetRootHash] get", "key", string(key), "keyhash", keyhash)
	stateObj, proof, ok, err := statedb.getStateObject(context.Background(), keyhash)
	if err != nil {
		return roothash, proof, fmt.Errorf("[statedb_sql:GetRootHash] %s", err)
	}
	if !ok {
		return nil, proof, nil //no root hash found
	}
	if proof == nil {
		log.Error("[statedb_sql:GetRootHash] proof is nil. Continuing.", "key", key)
	} else {
		log.Info("[statedb_sql:GetRootHash] proof retrieved", "key", key, "proof", proof)
	}
	log.Info(fmt.Sprintf("[statedb_sql:GetRootHash] retrieved stateObject (%+v)", stateObj))
	return stateObj.Val().Bytes(), proof, nil

}

func (statedb *StateDB) StoreRootHash(key []byte, roothash []byte) (err error) {

	keyhash := common.BytesToHash(wolkcommon.Computehash(key))
	//k := wolkcommon.BytesToUint64(keyhash[24:32])
	//log.Info("[statedb_sql:StoreRootHash] store", "key", hex.EncodeToString(key), "uint64 key", k, "roothash", hex.EncodeToString(roothash))
	//log.Info("[statedb_sql:StoreRootHash] store", "key", string(key), "keyhash", keyhash)
	stateObj, _, ok, err := statedb.getStateObject(context.Background(), keyhash)
	if err != nil {
		return fmt.Errorf("[statedb_sql:StoreRootHash] %s", err)
	}
	if !ok {
		return fmt.Errorf("[statedb_sql:StoreRootHash] state object did not exist for key (%v)", keyhash)
	}

	//TODO: change SetVal to SetDatabase/Owner/Table appropriately
	stateObj.SetOwner(common.BytesToHash(roothash))
	//log.Info("[statedb_sql:StoreRootHash] owner set", "owner roothash", common.BytesToHash(roothash))
	statedb.sql.stateObjects[keyhash] = stateObj
	//log.Info(fmt.Sprintf("[statedb_sql:StoreRootHash] done. stateObject (%+v)", stateObj))
	return nil

}

func (statedb *StateDB) getStateObject(ctx context.Context, key common.Hash) (so *stateObject, proof *Proof, ok bool, err error) {

	//log.Info("[statedb_sql:getStateObject] looking for", "key", key)
	cachedObj, isSet := statedb.sql.stateObjects[key]
	if isSet && cachedObj != nil {
		log.Info("[statedb_sql:getStateObject] found cached obj", "cachedobj val", cachedObj.Val())
		return cachedObj, proof, true, nil // no proof for cached objs?
	}
	//log.Info("[statedb_sql:getStateObject] keystorage key", "first 20 bytes", key.Bytes()[0:20])
	val, ok, _, proof, _, err := statedb.keyStorage.Get(ctx, key.Bytes()[0:20], true)
	if err != nil {
		log.Error("[statedb_sql:getStateObject] keystorage", "err", err)
		return nil, proof, false, fmt.Errorf("[statedb_sql:getStateObject] %s", err)
	}
	if !ok {
		log.Error("[statedb_sql:getStateObject] NOT OK, stateobject not found in keystorage. Continuing.")
		//return nil, proof, false, nil
	} else {
		//log.Info("[statedb_sql:getStateObject] SUCCESS")
	}
	return NewStateObject(statedb, key, common.BytesToHash(val)), proof, true, nil
}

func (statedb *StateDB) updateStateObject(ctx context.Context, s *stateObject) (err error) {

	stateBytes, err := rlp.EncodeToBytes(s)
	if err != nil {
		return fmt.Errorf("[statedb_sql:updateStateObject] %s", err)
	}
	_, err = statedb.Storage.QueueChunk(nil, stateBytes)
	if err != nil {
		return fmt.Errorf("[statedb_sql:updateStateObject] %s", err)
	}

	// TODO: figure this model out in "stateObject.go"
	oldStorageBytes := s.StorageBytesOld()
	newStorageBytes := s.StorageBytes()
	diff := newStorageBytes - oldStorageBytes
	smt := statedb.keyStorage
	//log.Info("[statedb_sql:updateStateObject] key", "key", s.key, "len(key)", len(s.key))
	//log.Info("[statedb_sql:updateStateObject] key", "1st 20 bytes", s.key.Bytes()[0:20], "diff", diff, "s.val.Bytes()", s.val.Bytes(), "stateBytes", stateBytes)
	err = smt.Insert(ctx, s.key.Bytes()[0:20], s.val.Bytes(), diff, false)
	if err != nil {
		return fmt.Errorf("[statedb_sql:updateStateObject] %s", err)
	}
	return nil
}

func (statedb *StateDB) ApplySQLTransaction(req *SQLRequest) error {
	log.Info("[statedb_sql:ApplySQLTransaction] applying", "req", req)
	resp, _, err := statedb.SelectHandler(req, false) //no proof
	if err != nil {
		return fmt.Errorf("[statedb_sql:ApplySQLTransaction] %s", err)
	}
	log.Info(fmt.Sprintf("[backend:ApplySQLTransaction] response: %+v", resp))
	return nil
}

// CommitTo writes the state to cloudstore via the SMT
// func (statedb StateDB) Commit(cs deep.StorageLayer, blockNumber uint64, parentHash, parentAnchorHash common.Hash) (dh deep.Header, err error) {
//
// 	defer statedb.clearJournal()
// 	// Commit objects to the SMT
// 	for index, stateObject := range statedb.sql.stateObjects {
// 		err := statedb.updateStateObject(stateObject)
// 		if err != nil {
// 			return dh, fmt.Errorf("[statedb_sql:Commit] %s", err)
// 		}
// 		//log.Info("Successfully updated stateObject.  Deleting now: ", "stateObject", stateObject, "index", index)
// 		delete(statedb.sql.stateObjects, index)
// 	}
// 	// Commit the data in the SMT to Cloudstore here
// 	err = statedb.Storage.Flush()
// 	if err != nil {
// 		return dh, fmt.Errorf("[statedb_sql:Commit] %s", err)
// 	}
// 	var h Header
// 	h.BlockNumber = blockNumber
// 	h.ParentHash = parentHash
// 	h.ParentAnchorHash = parentAnchorHash
// 	h.Time = uint64(time.Now().Unix())
// 	h.KeyRoot = statedb.Storage.keyStorage.MerkleRoot()
// 	h.KeyChunkHash = statedb.Storage.keyStorage.ChunkHash()
// 	//hdeep := h.(deep.Header)
//
// 	return &h, nil
// }

// NOT USED yet
// func (statedb *StateDB) DecryptDataParameter(data []byte) (b []byte, err error) {
// 	key, _ := hex.DecodeString("3d51e84f0270019e9238f6946bd35a8f") //TODO: retrieve don't hardcode
// 	var hashFunc = wolkcommon.Computehash
// 	encObj :=  New(0, uint32(0), hashFunc)
// 	decryptedHexData, _ := encObj.Decrypt(data, key)
// 	decoded, decodeErr := hex.DecodeString(fmt.Sprintf("%x", decryptedHexData))
// 	if decodeErr != nil {
// 		return b, fmt.Errorf("[sqlchain:DecryptDataParameter] DecodeString %s", decodeErr)
// 	}
// 	return decoded, nil
// }

//for comparing rows in two different sets of data
//only 1 cell in the row has to be different in order for the rows to be different
func isDuplicateRow(row1 Row, row2 Row) bool {

	//if row1.primaryKeyValue == row2.primaryKeyValue {
	//	return true
	//}

	for k1, r1 := range row1 {
		if _, ok := row2[k1]; !ok {
			return false
		}
		if r1 != row2[k1] {
			return false
		}
	}

	for k2, r2 := range row2 {
		if _, ok := row1[k2]; !ok {
			return false
		}
		if r2 != row1[k2] {
			return false
		}
	}

	return true
}

type ChunkHeader struct {
	MsgHash        []byte
	Sig            []byte
	Payer          []byte
	ChunkType      []byte
	MinReplication int
	MaxReplication int
	Birthts        int
	LastUpdatets   int
	Encrypted      int
	Version        int
	AutoRenew      int
	Key            []byte
	Owner          []byte
	Database       []byte
	Table          []byte
	//Epochts       []byte -- Do we need this in our Chunk?
	//Trailing Bytes
}

func ParseChunkHeader(chunk []byte) (ch ChunkHeader, err error) {
	/*
		if len(bytes.Trim(chunk, "\x00")) != CHUNK_SIZE {
			return ch, & SQLChainError{ Message: fmt.Sprintf("[types:ParseChunkHeader]"), ErrorCode: 480, ErrorMessage: fmt.Sprintf("Chunk of invalid size.  Expecting %d bytes, chunk is %d bytes", CHUNK_SIZE, len(chunk)) }
		}
	*/
	//fmt.Printf("Chunk is of size: %d and looking at %d to %d\n", len(chunk), CHUNK_START_MINREP, CHUNK_END_MINREP)
	//log.Debug(fmt.Sprintf("Chunk is of size: %d and looking at %d to %d ==> %+v\n%+v", CHUNK_SIZE, CHUNK_START_MINREP, CHUNK_END_MINREP, chunk[CHUNK_START_MINREP:CHUNK_END_MINREP], chunk))
	ch.MsgHash = chunk[CHUNK_START_MSGHASH:CHUNK_END_MSGHASH]
	ch.Sig = chunk[CHUNK_START_SIG:CHUNK_END_SIG]
	ch.Payer = chunk[CHUNK_START_PAYER:CHUNK_END_PAYER]
	ch.ChunkType = chunk[CHUNK_START_CHUNKTYPE:CHUNK_END_CHUNKTYPE]
	ch.MinReplication = int(BytesToInt(chunk[CHUNK_START_MINREP:CHUNK_END_MINREP]))
	ch.MaxReplication = int(BytesToInt(chunk[CHUNK_START_MAXREP:CHUNK_END_MAXREP]))
	ch.Birthts = int(BytesToInt(chunk[CHUNK_START_BIRTHTS:CHUNK_END_BIRTHTS]))
	ch.LastUpdatets = int(BytesToInt(chunk[CHUNK_START_LASTUPDATETS:CHUNK_END_LASTUPDATETS]))
	ch.Encrypted = int(BytesToInt(chunk[CHUNK_START_ENCRYPTED:CHUNK_END_ENCRYPTED]))
	ch.Version = int(BytesToInt(chunk[CHUNK_START_VERSION:CHUNK_END_VERSION]))
	ch.AutoRenew = int(BytesToInt(chunk[CHUNK_START_RENEW:CHUNK_END_RENEW]))
	ch.Key = chunk[CHUNK_START_KEY:CHUNK_END_KEY]
	ch.Owner = chunk[CHUNK_START_OWNER:CHUNK_END_OWNER]
	ch.Database = chunk[CHUNK_START_DB:CHUNK_END_DB]
	ch.Table = chunk[CHUNK_START_TABLE:CHUNK_END_TABLE]
	//ch.Epochts = chunk[CHUNK_START_EPOCHTS:CHUNK_END_EPOCHTS])
	return ch, err
}
