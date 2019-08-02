// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/wolkdb/cloudstore/log"
)

type SQLTable struct {
	statedb  *StateDB
	Owner    string
	Database string
	Name     string

	primaryColumnName string
	columns           map[string]*ColumnInfo
	encrypted         int //why is this here?
	birthts           int
	version           int
	//chunkType         []byte //what is this for? should this be 'chunkType' like in SQLDatabase?

	buffered bool
	roothash []byte
	chunk    []byte
}

type ColumnInfo struct {
	columnName string
	indexType  IndexType
	roothash   []byte // roothash of btree or hashdb tree for this column
	dbaccess   *Tree  // pointer to btree or hashdb tree
	primary    uint8
	columnType ColumnType
	IDNum      uint8
}

func (t *SQLTable) open() (proof *Proof, ok bool, err error) {

	// check the owner
	owner, ok, err := t.statedb.GetOwner(t.Owner)
	if err != nil {
		return proof, false, fmt.Errorf("[table:open] %s", err)
	}
	if !ok {
		return proof, false, nil
	}
	if _, ok = owner.DatabaseNames[t.Database]; !ok {
		return proof, false, nil //this database isn't in the owner chunk
	}

	// check the database
	_, ok, err = t.statedb.GetDatabase(t.Owner, t.Database)
	if err != nil {
		return proof, false, fmt.Errorf("[table:open] %s", err)
	}
	if !ok {
		return proof, false, nil //database chunk doesn't exist
	}
	//fmt.Printf(">> world state before getting table:\no:%+v\nd:%+v\nt:%+v\n", t.statedb.Owners, t.statedb.Databases, t.statedb.Tables)

	// get Table RootHash to retrieve the table descriptor
	tblKey := t.statedb.GetTableKey(t.Owner, t.Database, t.Name)
	t.roothash, proof, err = t.statedb.GetRootHash([]byte(tblKey))
	if err != nil {
		return proof, false, fmt.Errorf("[table:open] %s", err)
	}
	if len(bytes.Trim(t.roothash, "\x00")) == 0 { // when does this happen?
		//return false, & SQLError{Message: fmt.Sprintf("Attempting to Open Table with roothash of [%v]", t.roothash), ErrorCode: 481, ErrorMessage: fmt.Sprintf("Table [%s] has an empty roothash", t.Name)}
		return proof, false, nil //table does not exist
	}

	//log.Debug(fmt.Sprintf("[table:open] opening table @ %s roothash [%x]\n", t.Name, t.roothash))
	//fmt.Printf("[table:open] opening table @ %s roothash [%x]\n", t.Name, t.roothash)

	err = t.readChunk()
	if err != nil {
		return proof, false, fmt.Errorf("[table:open] %s", err)
	}
	//log.Debug(fmt.Sprintf("[table:open] open table %s with Owner %s Database %s Returning with Columns: %v\n", t.Name, t.Owner, t.Database, t.columns))
	//fmt.Printf("[table:open] read chunk %s with Owner %s Database %s Returning with Columns: %+v\n", t.Name, t.Owner, t.Database, t.columns)

	return proof, true, nil
}

func (t *SQLTable) ByteArrayToOrderedRow(byteData []byte) (orderedRow map[int]interface{}, err error) {
	orderedRow = make(map[int]interface{}, len(t.columns))
	o, err := t.byteArrayToRow(byteData)
	if err != nil {
		return orderedRow, fmt.Errorf("[table:ByteArrayToOrderedRow] %s", err)
	}
	//fmt.Printf("[table:ByteArrayToOrderedRow] row: %+v\n", o)
	//fmt.Printf("[table:ByteArrayToOrderedRow] t.columns : %+v\n", t.columns)

	for _, c := range t.columns {
		idnum := int(c.IDNum)
		if v, ok := o[c.columnName]; ok {
			//fmt.Printf("[table:ByteArrayToOrderedRow] id: %v, col %s => %+v\n", idnum, c.columnName, v)
			orderedRow[idnum] = v
		} else {
			//fmt.Printf("[table:ByteArrayToOrderedRow] id: %v, col %s => NULL\n", idnum, c.columnName)
			orderedRow[idnum] = nil
		}
	}
	//fmt.Printf("[table:ByteArrayToOrderedRow] orderedrow: %+v\n", orderedRow)
	return orderedRow, nil
}

// func (t *SQLTable) ByteArrayToRow(byteData []byte) (out  Row, err error) {
// 	return t.byteArrayToRow(byteData)
// }

//TODO: combine this with assignRowColumnTypes
func (t *SQLTable) byteArrayToRow(byteData []byte) (out Row, err error) {

	if len(byteData) == 0 {
		return out, nil
	}

	rawRow := NewRow()
	//fmt.Printf("[table:byteArrayToRow] byteData is: %s\n", string(byteData))
	//fmt.Printf("[table:byteArrayToRow] table columns are: %+v\n", t.columns)
	if err := json.Unmarshal(byteData, &rawRow); err != nil {
		return out, fmt.Errorf("[table:byteArrayToRow] %s", err)
	}

	//convert the rawRow to row with the correct types
	row := NewRow()
	for colName, cell := range rawRow {
		//fmt.Printf("[table:byteArrayToRow] considering col: %s\n", colName)
		if _, ok := t.columns[colName]; !ok {
			return out, fmt.Errorf("[table:byteArrayToRow] Column %s does not exist", colName)
		}
		colDef := t.columns[colName]
		switch a := cell.(type) {
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			switch colDef.columnType {
			case CT_STRING:
				row[colName] = fmt.Sprintf("%d", a)
				break
			case CT_INTEGER:
				row[colName] = a
				break
			case CT_FLOAT:
				row[colName] = float64(a.(int))
			}
			break
		case float64:
			switch colDef.columnType {
			case CT_STRING:
				row[colName] = fmt.Sprintf("%f", cell)
				break
			case CT_INTEGER:
				row[colName] = int(a)
				break
			case CT_FLOAT:
				row[colName] = a
			}
			break
		case string:
			switch colDef.columnType {
			case CT_INTEGER:
				row[colName], err = strconv.Atoi(a)

			case CT_STRING:
				row[colName] = a
			case CT_FLOAT:
				row[colName], err = strconv.ParseFloat(a, 64)
			}
			break
		}
	}
	return row, nil
}

//Gets a row in the table, based on primary key
func (t *SQLTable) Get(key []byte) (out []byte, ok bool, err error) {

	if _, ok := t.columns[t.primaryColumnName]; !ok {
		return out, false, fmt.Errorf("[table:Get] primary column (%s) missing in table columns description", t.primaryColumnName)
	}

	chunkKey, ok, err := t.columns[t.primaryColumnName].dbaccess.Get(key)
	if err != nil {
		return out, false, fmt.Errorf("[table:Get] %s", err)
	}
	if !ok { //didn't get it
		log.Info("[table:Get] key was not in dbaccess tree", "key", key, "key", hex.EncodeToString(key))
		return out, false, nil
	}

	chunk, ok, err := t.statedb.GetDBChunk(chunkKey)
	if err != nil {
		return nil, false, fmt.Errorf("[table:Get] %s", err)
	} else if !ok {
		log.Info("[table:Get] chunk was not found", "chunkkey", hex.EncodeToString(chunkKey))
		return nil, false, nil
	}

	//TODO: Validate Header?
	jsonRecord := chunk[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL]
	jsonRecord = bytes.TrimRight(jsonRecord, "\x00")
	if bytes.Trim(jsonRecord, "\x00") == nil {
		return out, false, nil //returning nil chunk
	}
	out = bytes.Trim(jsonRecord, "\x00")

	return out, true, nil

}

//Deletes a row in the table, based on primary key
func (t *SQLTable) Delete(key interface{}) (ok bool, err error) {

	if _, ok := t.columns[t.primaryColumnName]; !ok {
		return false, &SQLError{Message: fmt.Sprintf("[table:Delete] columns array missing %s ", t.primaryColumnName), ErrorCode: 479, ErrorMessage: fmt.Sprintf("Table Definition Missing Selected Column [%s]", t.primaryColumnName)}
	}
	k, err := convertValueToBytes(t.columns[t.primaryColumnName].columnType, key)
	if err != nil {
		return ok, fmt.Errorf("[table:Delete] %s", err)
	}

	done := false
	for _, column := range t.columns {
		if column.dbaccess == nil {
			return false, fmt.Errorf("[table:Delete] ColumnInfo %s has no dbaccess", column.columnName)
		}
		ok, err := column.dbaccess.Delete(k)
		if err != nil {
			return false, fmt.Errorf("[table:Delete] %s", err)
		}
		if !ok {
			//fmt.Printf("[table:Delete] index delete column: %+v, \n   key: %s failed. No deletion done.\n", column, string(k))
			log.Error("[table:Delete] No deletion done for", "column", column, "key failed", string(k))
		} else {
			done = true
			//fmt.Printf("[table:Delete] index delete column: %+v, \n   key: %s success.\n", column, string(k))
			log.Info("[table:Delete] success", "column", column, "key", string(k))
			break
		}
	}
	if done == false {
		return false, nil //deletion wasn't done
	}

	// TODO: K node deletion
	return true, nil
}

func (t *SQLTable) StartBuffer() (err error) {

	log.Info("[table:StartBuffer]")
	if t.buffered {
		log.Info("[table:StartBuffer] calling table's FlushBuffer", "tbl", t.Name)
		t.FlushBuffer()
	} else {
		t.buffered = true
	}

	for _, col := range t.columns {
		log.Info("[table:StartBuffer] calling col's dbaccess StartBuffer", "col", col)
		_, err := col.dbaccess.StartBuffer()
		if err != nil {
			return fmt.Errorf("[table:StartBuffer] %s", err)
		}
	}
	return nil
}

func (t *SQLTable) FlushBuffer() (err error) {

	log.Info("[table:FlushBuffer]")
	// get the updated tree root hashes of the columns
	for _, col := range t.columns {
		if col.dbaccess == nil {
			log.Error("[table:FlushBuffer] col did not have a dbaccess. Continuing.", "col", col.columnName)
			continue //TODO: should this be an err?
		}
		//log.Info("[table:FlushBuffer] calling col's dbaccess FlushBuffer", "col", col.columnName)
		_, err := col.dbaccess.FlushBuffer()
		if err != nil {
			return fmt.Errorf("[table:FlushBuffer] %s", err)
		}
		roothash := col.dbaccess.GetRootHash()
		col.roothash = roothash
		//log.Info("[table:FlushBuffer]", "col", col.columnName, "roothash", hex.EncodeToString(col.roothash))
	}

	// make and write the table chunk
	//log.Info("[table:FlushBuffer] calling table writeChunk")
	err = t.writeChunk()
	if err != nil {
		return fmt.Errorf("[table:FlushBuffer] %s", err)
	}

	// update database and owner chunks to reflect changed table chunk roothash
	db, ok, err := t.statedb.GetDatabase(t.Owner, t.Database)
	if err != nil {
		return fmt.Errorf("[table:FlushBuffer] %s", err)
	}
	if !ok {
		return fmt.Errorf("[table:FlushBuffer] Database does not exist")
	}

	err = db.store(t)
	if err != nil {
		return fmt.Errorf("[table:FlushBuffer] %s", err)
	}

	return nil
}

// fills in a SQLTable struct from table chunk bytes
// assums t.roothash is already found, table exists
func (t *SQLTable) readChunk() (err error) {

	var ok bool
	t.chunk, ok, err = t.statedb.GetDBChunk(t.roothash)
	if err != nil {
		return fmt.Errorf("[table:readChunk] %s", err)
	} else if !ok {
		return fmt.Errorf("[table:readChunk] Chunk not found")
	} else if len(t.chunk) == 0 {
		return fmt.Errorf("[table:readChunk] Chunk found but empty")
	}
	//log.Info("[table:readChunk] chunk found.")

	// get table descriptors
	t.encrypted = BytesToInt(t.chunk[4000:4024])
	//t.birthts = //TODO
	//t.version = //TODO
	//t.nodeType = //TODO

	// get columns
	//id := 1
	for i := 2048; i < 4000; i = i + 64 { //TODO: replace with constants

		buf := make([]byte, 64)
		copy(buf, t.chunk[i:i+64])

		if buf[0] == 0 { //empty slot
			break
		}

		//make new column
		cinfo := new(ColumnInfo)
		cinfo.columnName = string(bytes.Trim(buf[:23], "\x00"))
		cinfo.IDNum = uint8(buf[24])
		cinfo.primary = uint8(buf[26])
		cinfo.columnType, _ = ByteToColumnType(buf[28]) //:29
		cinfo.indexType = ByteToIndexType(buf[30])
		cinfo.roothash = buf[32:] //sometimes retrieving nothing?

		if cinfo.primary == 1 {
			t.primaryColumnName = cinfo.columnName
		}
		//log.Info(fmt.Sprintf("[table:readChunk] columnName: %s, primary: %d, idnum: %d, roothash: %x, columnType: %s", cinfo.columnName, cinfo.primary, cinfo.IDNum, cinfo.roothash, cinfo.columnType))
		t.columns[cinfo.columnName] = cinfo
	}

	if err = t.getDbaccess(); err != nil { //cinfo.dbaccess for all cols
		return fmt.Errorf("[table:readChunk] %s", err)
	}

	return nil
}

func (t *SQLTable) getDbaccess() (err error) {

	_, ok := t.columns[t.primaryColumnName]
	if !ok {
		return fmt.Errorf("[table:getDbaccess] Primary Column %s doesn't exist %+v", t.primaryColumnName, t.columns)
	}
	primaryColumnType := t.columns[t.primaryColumnName].columnType

	for _, c := range t.columns {
		//log.Info("[table:getDBaccess] get", "colname", c.columnName, "c.roothash", c.roothash)
		secondary := true
		if t.primaryColumnName == c.columnName {
			secondary = false
		}

		switch c.indexType {
		case IT_BPLUSTREE:
			bplustree, err := NewBPlusTreeDB(t.statedb, c.roothash, ColumnType(c.columnType), secondary, ColumnType(primaryColumnType), t.encrypted)
			if err != nil {
				return fmt.Errorf("[table:getDbaccess] %s", err)
			}
			c.dbaccess = bplustree
			//log.Info("[table:getDBaccess] dbaccess result:", "bplustree", bplustree)
		case IT_HASHTREE: //no hash tree for now
			return fmt.Errorf("[table:getDbaccess] no hash trees for now")
		}
	}
	return nil
}

// makes a table chunk in bytes from SQLTable struct
func (t *SQLTable) writeChunk() (err error) {

	// table name; 1st 32 bytes of chunk is table name
	bytesName := make([]byte, TABLE_NAME_LENGTH_MAX)
	copy(bytesName[0:], t.Name)
	copy(t.chunk[0:TABLE_NAME_LENGTH_MAX], bytesName[0:TABLE_NAME_LENGTH_MAX])

	// write columns into the table chunk
	i := 0
	for _, c := range t.columns {

		//fmt.Printf("[table:writeChunk] writing col: %+v\n", c)
		b := make([]byte, 1)

		//TODO: replace #s with consts
		// column name
		bytesColumnName := make([]byte, COLUMN_NAME_LENGTH_MAX)
		copy(bytesColumnName[0:], c.columnName)
		copy(t.chunk[2048+i*64:], bytesColumnName[0:COLUMN_NAME_LENGTH_MAX])
		//fmt.Printf("[table:writeChunk] writing col: %s (%v)\n", bytesColumnName, bytesColumnName)

		// idnum bit
		b[0] = byte(c.IDNum)
		copy(t.chunk[2048+i*64+24:], b)

		// primary bit
		b[0] = byte(c.primary)
		copy(t.chunk[2048+i*64+26:], b)

		// columnType bit
		ctInt, _ := ColumnTypeToInt(c.columnType)
		b[0] = byte(ctInt)
		copy(t.chunk[2048+i*64+28:], b)

		// indexType bit
		itInt := IndexTypeToInt(c.indexType)
		b[0] = byte(itInt)
		copy(t.chunk[2048+i*64+30:], b)

		// roothash of tree
		if !EmptyBytes(c.roothash) { // empty for new tables
			copy(t.chunk[2048+i*64+32:], c.roothash)
		}

		i++
	}

	//TODO: Could (Should?) be less bytes, but leaving space in case more is to be there
	copy(t.chunk[4000:4024], IntToByte(t.encrypted))
	//t.birthts  //TODO
	//t.version //TODO
	//t.nodeType //TODO

	// store the chunk
	t.roothash, err = t.statedb.SetDBChunk(t.chunk, t.encrypted)
	if err != nil {
		return fmt.Errorf("[table:writeChunk] %s", err)
	}
	tblKey := t.statedb.GetTableKey(t.Owner, t.Database, t.Name)
	err = t.statedb.StoreRootHash([]byte(tblKey), []byte(t.roothash))
	if err != nil {
		return fmt.Errorf("[table:writeChunk] %s", err)
	}

	return nil

}

// converts []Columns - array of input columns to tbl.columns
// opposite of DescribeTable()
func (t *SQLTable) columnToColumnInfo(inputColumns []Column) (err error) {

	id := uint8(1)
	for _, col := range inputColumns {
		cinfo := new(ColumnInfo)
		cinfo.columnName = col.ColumnName
		cinfo.primary = uint8(col.Primary)
		cinfo.columnType = col.ColumnType
		cinfo.indexType = col.IndexType
		cinfo.roothash = make([]byte, CHUNK_HASH_SIZE) //TODO: check this

		if cinfo.primary == 1 {
			t.primaryColumnName = cinfo.columnName
			cinfo.IDNum = 0
		} else {
			cinfo.IDNum = id
			id++
		}

		t.columns[cinfo.columnName] = cinfo
	}

	// for key, val := range t.columns {
	// 	fmt.Printf("[table:columnToColumnInfo] col: %s, %+v\n", key, val)
	// }
	// fmt.Println()

	if err = t.getDbaccess(); err != nil { //cinfo.dbaccess for all cols
		return fmt.Errorf("[table:columnToColumnInfo] %s", err)
	}

	return nil
}

// converts tbl.columns into []Columns - array of output columns
// opposite of columnToColumnInfo
func (t *SQLTable) DescribeTable() (tblInfo map[string]Column, err error) {

	//log.Debug(fmt.Sprintf("DescribeTable with table [%+v] \n", t))
	tblInfo = make(map[string]Column)
	for cname, c := range t.columns {
		var cinfo Column
		cinfo.ColumnName = cname
		cinfo.IndexType = c.indexType
		cinfo.Primary = int(c.primary)
		cinfo.ColumnType = c.columnType
		// if _, ok := tblInfo[cname]; ok { // if ok, would mean for some reason there are two cols named the same thing
		// 	return tblInfo, & SQLError{Message: fmt.Sprintf("[table:DescribeTable] Duplicate column: [%s]", cname), ErrorCode: -1, ErrorMessage: "Table has Duplicate columns?"} //TODO: how would this occur?
		// }
		tblInfo[cname] = cinfo
	}
	//log.Debug(fmt.Sprintf("Returning from DescribeTable with table [%+v] \n", tblInfo))
	return tblInfo, nil
}

// Gets all rows of a table, in ascending or descending order
func (t *SQLTable) Scan(columnName string, ascending int) (rows []Row, err error) {

	if _, ok := t.columns[columnName]; !ok {
		return rows, &SQLError{Message: fmt.Sprintf("[table:Scan] columns array missing %s ", columnName), ErrorCode: 479, ErrorMessage: "Table Definition Missing Selected Column"}
	}
	if t.primaryColumnName != columnName { //TODO: this won't always be the case
		return rows, &SQLError{Message: fmt.Sprintf("[table:Scan] Skipping column %s", columnName), ErrorCode: -1, ErrorMessage: "Query Filters currently only supported on the primary key"}
	}

	c := t.columns[columnName].dbaccess

	if ascending == 1 {
		res, err := c.SeekFirst()
		if err == io.EOF {
			return rows, nil
		}
		if err != nil {
			return rows, fmt.Errorf("[table:Scan] SeekFirst %s ", err)
		}

		for k, v, errRes := res.Next(); errRes == nil; k, v, errRes = res.Next() {
			//fmt.Printf("\n *int*> %d: K: %s V: %v \n", records, KeyToString(column.columnType, k), v)
			rawRow, ok, err := t.Get(k)
			if err != nil {
				return rows, fmt.Errorf("[table:Scan] Get %s", err)
			}
			if !ok {
				//TODO: is this an error if nothing was gotten?
			}

			rowObj, err := t.byteArrayToRow(rawRow)
			if err != nil {
				return rows, fmt.Errorf("[table:Scan] byteArrayToRow [%s] bytearray to row: [%s]", v, err)
			}
			// fmt.Printf("table Scan, row set: %+v\n", row)
			rows = append(rows, rowObj)
		}
		return rows, nil
	}

	//if descending:
	res, err := c.SeekLast()
	if err != nil {
		return rows, fmt.Errorf("[table:Scan] SeekLast %s", err)
	}
	for k, _, errRes := res.Prev(); errRes == nil; k, _, errRes = res.Prev() {
		//fmt.Printf(" *int*> %d: K: %s V: %v\n", records, KeyToString( CT_STRING, k), KeyToString(t.columns[columnName].columnType, v))

		rawRow, ok, err := t.Get(k)
		if err != nil {
			return rows, fmt.Errorf("[table:Scan] Get %s", err)
		}
		if !ok {
			//TODO: is this an error if nothing was gotten?
		}

		rowObj, err := t.byteArrayToRow(rawRow)
		if err != nil {
			return rows, fmt.Errorf("[table:Scan] byteArrayToRow %s", err)
		}
		rows = append(rows, rowObj)
	}

	//log.Debug(fmt.Sprintf("table Scan, rows returned: %+v\n", rows))
	return rows, nil
}

//TODO: this is for WVM. Need to think about this.
func (t *SQLTable) OrderedPut(rawRow map[int]interface{}, isImmutable bool) (err error) {
	//log.Info(fmt.Sprintf("[table:OrderedPut] rawRow: %+v\n", rawRow))
	row := make(map[string]interface{})
	for _, c := range t.columns { //TODO:Map Review

		//log.Info(fmt.Sprintf("[table:OrderedPut] col name: %s, col id: %d \n", c.columnName, c.IDNum))
		if rawRow[int(c.IDNum)] != nil {
			row[c.columnName] = rawRow[int(c.IDNum)]
			//log.Info(fmt.Sprintf("[table:OrderedPut] assigned row[%s] = rawRow[%d] = %v\n", c.columnName, c.IDNum, rawRow[int(c.IDNum)]))
		}
	}
	//log.Info(fmt.Sprintf("[table:OrderedPut] row to Put: %+v\n", row))
	return t.Put(row, isImmutable)
}

// writes a Row into a Table chunk
func (t *SQLTable) Put(row Row, isImmutable bool) (err error) {

	// order the row by idnum, in case it isn't already
	// orderedColNames := make(map[int]string)
	// for _, col := range t.columns {
	// 	orderedColNames[col.IDNum] = col.columnName
	// }
	// fmt.Printf("[table:Put] orderedColNames: %+v\n", orderedColNames)
	// orderedRow :=  NewRow()
	// for _, colname := range orderedColNames {
	// 	orderedRow[colname] = row[colname]
	// }
	//fmt.Printf("[table:Put] orderedRow: %+v\n", orderedRow)
	//fmt.Printf("[table:Put] row: %+v\n", row)

	rawRow, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("[table:Put] %s", err)
	}

	for _, column := range t.columns {
		//fmt.Printf("\n[table:Put] Processing column %+v\n", column)
		if column.primary > 0 {
			//fmt.Printf("[table:Put] primary column: %+v, primary row value: %+v\n", t.primaryColumnName, row[t.primaryColumnName])
			value, ok := row[t.primaryColumnName]
			if !ok {
				return fmt.Errorf("[table:Put] primaryColumn (%s) missing in table columns description", t.primaryColumnName)
			}
			key := make([]byte, 32)
			key, err = convertValueToBytes(t.columns[t.primaryColumnName].columnType, value)
			if err != nil {
				return fmt.Errorf("[table:Put] %s", err)
			}

			//check for writing a duplicate chunk
			if isImmutable {
				_, ok, err := t.Get(key)
				if err != nil {
					return fmt.Errorf("[table:Put] %s", err)
				}
				if ok { // key found, cannot overwrite
					return fmt.Errorf("[table:Put] Chunk to put was found and chunk is immutable. Duplicate key (%x) Primary column (%s), primary row val (%v)\n", key, t.primaryColumnName, value)
				}
				// if !ok { // key not found
				// 	log.Info("[table:Put] Chunk not found. Good. Continuing", "key", key, "primary column", t.primaryColumnName, "row val", value)
				// }
			}

			chunkType := []byte("k") //TODO: get rid of legacy "K chunk" code
			sdata, err := t.BuildSdata(rawRow, chunkType)
			if err != nil {
				return fmt.Errorf("[table:Put] %s", err)
			}
			chunkKey, err := t.statedb.SetDBChunk(sdata, t.encrypted)
			if err != nil {
				return fmt.Errorf("[table:Put] %s", err)
			}
			//log.Info("[table:Put] calling dbaccess Put", "key", hex.EncodeToString(key), "chunkkey", hex.EncodeToString(chunkKey))
			_, err = column.dbaccess.Put(key, chunkKey)
			if err != nil {
				return fmt.Errorf("[table:Put] %s", err)
			}
			//log.Info("[table:Put] has", "col", column.columnName, "col roothash", column.roothash, "dbaccess", *column.dbaccess)

		} else { //not a primary column

			value, ok := row[column.columnName]
			if !ok { // this put is missing this column, OK b/c non-primary keys aren't required for rows
				log.Info("[table:Put] No value for this column. Continuing", "column", column.columnName)
				continue //go to next column
			}
			key := make([]byte, 32)
			key, err = convertValueToBytes(column.columnType, value)
			if err != nil {
				return fmt.Errorf("[table:Put] %s", err)
			}
			emptyChunkKey := make([]byte, 32)
			//log.Info("[table:Put] non primary col putting", "col", column.columnName, "val", value, "key", key, "chunkKey", emptyChunkKey)
			_, err = column.dbaccess.Put(key, emptyChunkKey) // we store empty value b/c row data is stored by primary key only
			if err != nil {
				return fmt.Errorf("[table:Put] key (%x) chunkKey (%x) %s", key, emptyChunkKey, err)
			}

		}
	}

	// do nothing until FlushBuffer called
	if !t.buffered {
		err = t.FlushBuffer() // TODO: THIS SHOULD BE DONE with a StateDB journal entry so that RevertToSnapshot can "revert"
		if err != nil {
			return fmt.Errorf("[table:Put] %s", err)
		}
	}

	return nil

}

//builds a table chunk with all table data
//TODO: should use chunkHeader struct
func (t *SQLTable) BuildSdata(value []byte, chunkType []byte) (mergedBodycontent []byte, err error) {
	// contentPrefix := t.BuildKChunkKey(key)
	// log.Debug(fmt.Sprintf("[table:buildSdata] contentPrefix is: %x", contentPrefix))
	//
	// var metadataBody []byte
	// metadataBody = make([]byte, CHUNK_START_CHUNKVAL)
	// copy(metadataBody[CHUNK_START_OWNER:CHUNK_END_OWNER], []byte(t.Owner))
	// copy(metadataBody[CHUNK_START_DB:CHUNK_END_DB], []byte(t.Database))
	// copy(metadataBody[CHUNK_START_TABLE:CHUNK_END_TABLE], []byte(t.Name))
	// copy(metadataBody[CHUNK_START_KEY:CHUNK_END_KEY], contentPrefix)
	// copy(metadataBody[CHUNK_START_PAYER:CHUNK_END_PAYER], u.Address)
	// copy(metadataBody[CHUNK_START_CHUNKTYPE:CHUNK_END_CHUNKTYPE], []byte("k")) //TODO: Define nodeType representation -- t.nodeType)
	// copy(metadataBody[CHUNK_START_RENEW:CHUNK_END_RENEW], IntToByte(u.AutoRenew))
	// copy(metadataBody[CHUNK_START_MINREP:CHUNK_END_MINREP], IntToByte(u.MinReplication))
	// copy(metadataBody[CHUNK_START_MAXREP:CHUNK_END_MAXREP], IntToByte(u.MaxReplication))
	// copy(metadataBody[CHUNK_START_ENCRYPTED:CHUNK_END_ENCRYPTED], IntToByte(t.encrypted))
	// copy(metadataBody[CHUNK_START_BIRTHTS:CHUNK_END_BIRTHTS], IntToByte(birthts))
	//
	// lastupdatets := int(time.Now().Unix())
	// copy(metadataBody[CHUNK_START_LASTUPDATETS:CHUNK_END_LASTUPDATETS], IntToByte(lastupdatets))
	// copy(metadataBody[CHUNK_START_VERSION:CHUNK_END_VERSION], IntToByte(version))
	//
	// unencryptedMetadata := metadataBody[CHUNK_END_MSGHASH:CHUNK_START_CHUNKVAL]
	// msg_hash := SignHash(unencryptedMetadata)
	//
	// //TODO: msg_hash --
	// copy(metadataBody[CHUNK_START_MSGHASH:CHUNK_END_MSGHASH], msg_hash)
	//
	// km := t.statedb.dbchunkstore.GetKeyManager()
	// sdataSig, errSign := km.SignMessage(msg_hash)
	// if errSign != nil {
	// 	return mergedBodycontent, & SQLError{Message: `[kademliadb:buildSdata] SignMessage ` + errSign.Error(), ErrorCode: 455, ErrorMessage: "Keymanager Unable to Sign Message"}
	// }
	//
	// //TODO: Sig -- document this
	// copy(metadataBody[CHUNK_START_SIG:CHUNK_END_SIG], sdataSig)
	// //log.Debug(fmt.Sprintf("Metadata is [%+v]", metadataBody))
	//

	mergedBodycontent = make([]byte, CHUNK_SIZE)
	metadataBody, err := t.statedb.BuildChunkHeader(t.Owner, t.Database, t.Name, chunkType, t.encrypted)
	if err != nil {
		return mergedBodycontent, fmt.Errorf("[table:BuildSdata] %s", err)
	}

	copy(mergedBodycontent[:], metadataBody)
	copy(mergedBodycontent[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL], value) // expected to be the encrypted body content

	//log.Debug(fmt.Sprintf("Merged Body Content: [%v]", mergedBodycontent))
	return mergedBodycontent, err

}

//TODO: difference between assignRowColumnTypes and just "checkRowColumnTypes" to err when there is wrong input? this forces the type, which may end up with weird data
func (t *SQLTable) assignRowColumnTypes(rows []Row) ([]Row, error) {
	//fmt.Printf("[table:assignRowColumnTypes]: cols %+v\n", t.columns)
	for _, row := range rows {
		for name, value := range row {
			c, ok := t.columns[name]
			if !ok {
				return rows, &SQLError{Message: fmt.Sprintf("[table:assignRowColumnTypes] Invalid column %s", name), ErrorCode: 404, ErrorMessage: fmt.Sprintf("Column Does Not Exist in table definition: [%s]", name)}
			}
			switch c.columnType {
			case CT_INTEGER:
				switch value.(type) {
				case int:
					row[name] = value.(int)
				case float64:
					row[name] = int(value.(float64))
				case string:
					f, err := strconv.ParseFloat(value.(string), 64)
					if err != nil {
						return rows, &SQLError{Message: fmt.Sprintf("[table:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, t.columns[name].columnType), ErrorCode: 427, ErrorMessage: fmt.Sprintf("The value passed in for [%s] cannot be converted to integer type", name)}
					}
					row[name] = int(f)
				default:
					return rows, &SQLError{Message: fmt.Sprintf("[table:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, t.columns[name].columnType), ErrorCode: 427, ErrorMessage: fmt.Sprintf("The value passed in for [%s] is of an unsupported type", name)}
				}
			case CT_STRING:
				switch value.(type) {
				case string:
					row[name] = value.(string)
				case int:
					row[name] = strconv.Itoa(value.(int))
				case float64:
					row[name] = strconv.FormatFloat(value.(float64), 'f', -1, 64)
					//TODO: handle err
					//log.Debug(fmt.Sprintf("Converting value[%s] from float64 to string => [%s]\n", value, row[name]))
				default:
					return rows, &SQLError{Message: fmt.Sprintf("[table:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, t.columns[name].columnType), ErrorCode: 427, ErrorMessage: fmt.Sprintf("The value passed in for [%s] is of an unsupported type", name)}
				}
			case CT_FLOAT:
				switch value.(type) {
				case float64:
					row[name] = value.(float64)
				case int:
					row[name] = float64(value.(int))
				case string:
					f, err := strconv.ParseFloat(value.(string), 64)
					if err != nil {
						return rows, &SQLError{Message: fmt.Sprintf("[table:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, t.columns[name].columnType), ErrorCode: 427, ErrorMessage: fmt.Sprintf("The value passed in for [%s] is of an unsupported type", name)}
					}
					row[name] = f
				default:
					return rows, &SQLError{Message: fmt.Sprintf("[table:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, t.columns[name].columnType), ErrorCode: 427, ErrorMessage: fmt.Sprintf("The value passed in for [%s] is of an unsupported type", name)}
				}
			//case  CT_BLOB:
			// TODO: add blob support
			default:
				return rows, &SQLError{Message: fmt.Sprintf("[table:assignRowColumnTypes] Coltype %s not found", t.columns[name].columnType), ErrorCode: 427, ErrorMessage: fmt.Sprintf("The value passed in for [%s] is of an unsupported type", name)}
			}

		}
	}
	return rows, nil
}

func (t *SQLTable) makeOrderedColumns() (orderedColumns []*ColumnInfo) {

	//order the cols by IDNum
	oCols := make(map[int]*ColumnInfo)
	for _, col := range t.columns {
		oCols[int(col.IDNum)] = col
	}

	//append them in the correct order
	orderedColumns = make([]*ColumnInfo, 0, len(t.columns))
	for i := 0; i < len(t.columns); i++ {
		orderedColumns = append(orderedColumns, oCols[i])
	}
	return orderedColumns

}

func (t *SQLTable) MakeCreateTableQuery() (sql string, err error) {

	//log.Info(fmt.Sprintf("[table:MakeCreateTableQuery] table used %#v", t))
	if _, ok := t.columns[t.primaryColumnName]; !ok {
		return sql, fmt.Errorf("[table:makeCreateTableQuery] no primary column %s", t.primaryColumnName)
	}
	primaryColumnType := strings.ToLower(string(t.columns[t.primaryColumnName].columnType))
	sql = "create table `" + t.Name + "` (`" + t.primaryColumnName + "` " + primaryColumnType + " primary key"

	orderedColumns := t.makeOrderedColumns()
	for _, colinfo := range orderedColumns {
		coltype := strings.ToLower(string(colinfo.columnType))
		if colinfo.columnName != t.primaryColumnName {
			sql = sql + ", `" + colinfo.columnName + "` " + coltype
		}
	}
	sql = sql + ")"
	sql = strings.Replace(sql, "integer", "int", -1) // if we don't do this, autoindexes don't show up. very odd.
	//log.Info("[table:MakeCreateTableQuery] sql made for sqlite", "sql", sql)
	return sql, nil
}

func MakeDropTableQuery(tableName string) (sql string) {
	return `DROP TABLE ` + tableName
}
