// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"reflect"
	"strconv"
	"strings"

	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/wvmc"
)

const (
	OP_Trace        = "Trace"
	OP_Goto         = "Goto"
	OP_OpenRead     = "OpenRead"
	OP_ReopenIdx    = "ReopenIdx"
	OP_Rewind       = "Rewind"
	OP_Rowid        = "Rowid"
	OP_Column       = "Column"
	OP_ResultRow    = "ResultRow"
	OP_Next         = "Next"
	OP_Close        = "Close"
	OP_Halt         = "Halt"
	OP_Transaction  = "Transaction"
	OP_VerifyCookie = "VerifyCookie"
	OP_TableLock    = "TableLock"
	OP_OpenWrite    = "OpenWrite"
	OP_NewRowid     = "NewRowid"
	OP_Integer      = "Integer"
	OP_Multiply     = "Multiply"
	OP_Add          = "Add"
	OP_MakeRecord   = "MakeRecord"
	OP_Insert       = "Insert"
	OP_Delete       = "Delete"
	OP_Subtract     = "Subtract"
	OP_String8      = "String8"
	OP_Gosub        = "Gosub"
	OP_Return       = "Return"
	OP_Int64        = "Int64"
	OP_Real         = "Real"
	OP_Null         = "Null"
	OP_Move         = "Move"
	OP_Copy         = "Copy"
	OP_Divide       = "Divide"
	OP_Remainder    = "Remainder"
	OP_Eq           = "Eq"
	OP_Ne           = "Ne"
	OP_Lt           = "Lt"
	OP_Le           = "Le"
	OP_Gt           = "Gt"
	OP_Ge           = "Ge"
	OP_Compare      = "Compare"
	OP_Jump         = "Jump"
	OP_BitAnd       = "BitAnd"
	OP_BitOr        = "BitOr"
	OP_BitNot       = "BitNot"
	OP_And          = "And"
	OP_Or           = "Or"
	OP_Not          = "Not"
	OP_If           = "If"
	OP_IfNot        = "IfNot"
	OP_Count        = "Count"
	OP_SorterOpen   = "SorterOpen"
	OP_NotExists    = "NotExists"
	OP_Sequence     = "Sequence"
	OP_SorterData   = "SorterData"
	OP_Last         = "Last"
	OP_SorterSort   = "SorterSort"
	OP_Sort         = "Sort"
	OP_SorterNext   = "SorterNext"
	OP_PrevIfOpen   = "PrevIfOpen"
	OP_NextIfOpen   = "NextIfOpen"
	OP_Prev         = "Prev"
	OP_SorterInsert = "SorterInsert"
	OP_RowSetAdd    = "RowSetAdd"
	OP_RowSetRead   = "RowSetRead"
	OP_RowSetTest   = "RowSetTest"
	OP_IfPos        = "IfPos"
	OP_AggStep      = "AggStep"
	OP_AggStep0     = "AggStep0"
	OP_AggFinal     = "AggFinal"
	OP_Function     = "Function"
	OP_Function0    = "Function0"
	OP_Init         = "Init"
	OP_DecrJumpZero = "DecrJumpZero"
	OP_CollSeq      = "CollSeq"
	OP_Affinity     = "Affinity"
	OP_RealAffinity = "RealAffinity"
	OP_IsNull       = "IsNull"
	OP_Once         = "Once"
	OP_OpenPseudo   = "OpenPseudo"
	OP_SCopy        = "SCopy"
	OP_Found        = "Found"
	OP_NotFound     = "NotFound"
	OP_NoConflict   = "NoConflict"
	OP_IdxInsert    = "IdxInsert"
	OP_IdxDelete    = "IdxDelete"
	OP_Noop         = "Noop"
	OP_DeferredSeek = "DeferredSeek"
	OP_IdxRowid     = "IdxRowid"
	OP_SeekGE       = "SeekGE"
	// OP_SeekGe       = OP_SeekGE
	OP_SeekGT  = "SeekGT"
	OP_SeekLT  = "SeekLT"
	OP_SeekLE  = "SeekLE"
	OP_IdxGE   = "IdxGE"
	OP_IdxGT   = "IdxGT"
	OP_IdxLT   = "IdxLT"
	OP_IdxLE   = "IdxLE"
	OP_IntCopy = "IntCopy"
)

type OpDef struct {
	Opcode  string
	f       func(wvm *WVM, op Op) (err error)
	GasCost uint
}

type Op struct {
	N      int    `json:"n"`
	Opcode string `json:"opcode"`       /* What operation to perform */
	P1     int    `json:"p1,omitempty"` /* First operand */
	P2     int    `json:"p2,omitempty"` /* Second parameter (often the jump destination) */
	P3     int    `json:"p3,omitempty"` /* The third parameter */
	P4     string `json:"p4,omitempty"`
	P5     int    `json:"p5,omitempty"` /* Fifth parameter is an unsigned 16-bit integer */
}

type WVMSchema map[int]WVMIndex //int should be RootPage of each index
/*
CREATE TABLE sqlite_master(
  type text,
  name text,
  tbl_name text,
  rootpage integer,
  sql text
);*/
type WVMIndex struct {
	SchemaType string //table or index
	IndexName  string //if SchemaType is table, this is also the name of the table
	TableName  string
	RootPage   int
	Sql        string
}

type WVM struct {
	JumpTable map[string]OpDef

	// who is writing
	//user *SQLUser

	// useful to create cursors, interact with DBChunkstore
	sqlchain *StateDB

	// program counter
	pc     int
	jumped bool // keeps track of whether an opcode changes pc
	halted bool // false until we are done (or there is an error)

	// RootPageIndex map[int]

	// register
	register map[int]interface{}

	// TODO: use now
	activeDB *SQLDatabase

	// TODO: remove active tables
	activeTable *SQLTable

	// TODO: remove  active indexes
	index map[int]*Tree

	// active cursors
	cursors map[int]OrderedDatabaseCursor

	// once marker
	once_marker map[int]int

	// active key
	activeKey []byte

	// active row
	activeRow map[int]interface{}

	// rowsets
	rowSet map[int][]int64 //int is register num like index and cursors, []int64 is array of rowids

	//sorter object
	//sorter []SorterRow

	// output
	Rows             [][]interface{}
	AffectedRowCount int
	eof              bool // temp

	// flags
	OPFLAG_ISUPDATE bool
}

// type SorterRow struct {
// 	keyToSort []interface{}         //will sort by this
// 	cursor    OrderedDatabaseCursor //cursor to row in bplus tree
// }

func (wvm *WVM) Reset() {
	wvm.activeRow = nil
	wvm.eof = false
	wvm.AffectedRowCount = 0
	wvm.register = make(map[int]interface{}, 128)
	wvm.index = make(map[int]*Tree, 128)
	wvm.cursors = make(map[int]OrderedDatabaseCursor, 128)
	wvm.rowSet = make(map[int][]int64, 128)
	wvm.once_marker = make(map[int]int)
	//wvm.sorter = make([]SorterRow, 128)
	wvm.OPFLAG_ISUPDATE = false
}

// makes a list of sqlite opcodes from a single sql statement query
func CompileOpcodes(dbpath string, sql string) (ops []Op, err error) {

	wvmcOps, err := wvmc.GetOps(dbpath, sql)
	if err != nil {
		return ops, fmt.Errorf("[wvm:CompileOpcodes] %s", err)
	}
	if err = json.Unmarshal(wvmcOps, &ops); err != nil {
		return ops, fmt.Errorf("[wvm:CompileOpcodes] %s", err)
	}
	//log.Info("[wvm:CompileOpcodes] ops after unmarshal", "ops", ops)
	return ops, nil
}

// runs create/drop table to modify schema's .db file
func SetWVMSchema(dbpath string, sql string) (err error) {

	if len(dbpath) == 0 || len(sql) == 0 {
		return fmt.Errorf("[wvm:SetWVMSchema] no path or no sql")
	}
	var sqls []string
	sqls = append(sqls, sql)
	if err = wvmc.SetupSchema(dbpath, sqls); err != nil {
		return fmt.Errorf("[wvm:SetWVMSchema] SetupSchema for dbpath [%s] %s", dbpath, err)
	}

	//for testing only...
	// sb, err := ReadWVMSchemaBytes(dbpath)
	// if err != nil {
	// 	fmt.Printf("[wvm:SetWVMSchema] trying to read schema bytes... err %v\n", err)
	// }
	// fmt.Printf("[wvm:SetWVMSchema] new made schema: \n %s\n", sb)

	return nil
}

// gets the sqlite 'master table' and parses into sqlchain's WVMSchema
func GetWVMSchema(dbpath string) (schema WVMSchema, err error) {

	if len(dbpath) == 0 {
		return schema, fmt.Errorf("[wvm:GetWVMSchema] no path")
	}
	schema = make(WVMSchema)
	schemaJson, err := wvmc.GetMasterTable(dbpath)
	if err != nil {
		return schema, fmt.Errorf("[wvm:GetWVMSchema] GetMasterTable failed opening dbpath [%s] %s", dbpath, err)
	}
	//fmt.Printf("[wvm:GetWVMSchema] schemajson: %s\n", schemaJson)

	for _, sline := range schemaJson {
		var index WVMIndex
		if err = json.Unmarshal(sline, &index); err != nil {
			return schema, fmt.Errorf("[wvm:GetWVMSchema] json.Unmarshal %s", err)
		}
		schema[index.RootPage] = index
	}

	schema = updateWVMSchemaCustomIndexes(schema)

	return schema, nil

}

//gets the schema from the local file and converts it to chunk form
func ReadWVMSchemaBytes(dbpath string) (bytes []byte, err error) {
	bytes, err = ioutil.ReadFile(dbpath)
	if err != nil {
		return bytes, fmt.Errorf("[wvm:ReadWVMSchemaBytes] %s", err)
	}
	return bytes, nil
}

func WriteWVMSchemaBytes(dbpath string, bytes []byte) (err error) {
	err = ioutil.WriteFile(dbpath, bytes, 0644)
	if err != nil {
		return fmt.Errorf("[wvm:WriteWVMSchemaBytes] %s", err)
	}
	return nil
}

func DeleteLocalSchema(dbpath string) (err error) {
	err = os.Remove(dbpath)
	if err != nil {
		return fmt.Errorf("[wvm:DeleteLocalSchema] %s", err)
	}
	return nil
}

/*
index|first|human|7|CREATE INDEX first on human (first)
index|last|human|8|CREATE INDEX last on human (last)
index|age|human|9|CREATE INDEX age on human (age)
*/
func updateWVMSchemaCustomIndexes(schema WVMSchema) WVMSchema {
	for rootpage, index := range schema {
		if index.SchemaType == "index" && strings.Contains(index.Sql, "CREATE INDEX") {
			a := strings.Split(index.Sql, "(")
			b := strings.Split(a[1], ")")
			newIndexName := strings.TrimSpace(b[0])
			if newIndexName != index.IndexName {
				index.IndexName = newIndexName
				schema[rootpage] = index
			}
		}
	}
	return schema
}

func (wvm *WVM) getWVMSchemaRootPage(db *SQLDatabase, rootPage int) (schemaType string, tableName string, indexName string, err error) {

	for _, ti := range db.Schema {
		if ti.RootPage == rootPage {
			// db.wvmSchema[5] = WVMIndex{SchemaType: "index", IndexName: "first", TableName: "human", RootPage: 7, Sql: "CREATE INDEX first on human (first)"}
			return ti.SchemaType, ti.TableName, ti.IndexName, nil
		}
	}
	return schemaType, tableName, indexName, fmt.Errorf("[wvm:getWVMSchemaRootPage] Unknown rootPage [%d]", rootPage)
}

func getColumn(t *SQLTable, columnName string) (c *ColumnInfo, err error) {

	log.Info("[wvm:getColumn] tablename", "tblname", t.Name)
	log.Info(fmt.Sprintf("[wvm:getColumn] columns: %v", t.columns))
	log.Info("[wvm:getColumn] col to get", "col", t.columns[columnName])
	if _, ok := t.columns[columnName]; !ok {
		return c, fmt.Errorf("[wvm:getColumn] table definition missing selected column (%s); please check if table exists", columnName)
	}
	if t.columns[columnName] == nil { //actually shouldn't happen
		return c, fmt.Errorf("[wvm:getColumn] column %s exists but is empty ", columnName)
	}
	//fmt.Printf("[wvm:getColumn] column got: %+v\n", t.columns[columnName])

	return t.columns[columnName], nil
}

func getPrimaryColumn(t *SQLTable) (c *ColumnInfo, err error) {
	return getColumn(t, t.primaryColumnName)
}

func getIndex(c *ColumnInfo) (d *Tree) {
	return c.dbaccess
}

func setupIndex(wvm *WVM, cursorID int, rootPage int) (err error) {
	log.Info("[wvm:setupIndex] start", "cursorID", cursorID, "rootPage", rootPage)
	wvm.eof = false
	schemaType, tableName, indexName, err := wvm.getWVMSchemaRootPage(wvm.activeDB, rootPage)
	if err != nil {
		return fmt.Errorf("[wvm:setupIndex] %s", err)
	}
	if schemaType == "table" || schemaType == "index" {
		tbl, _ /*proof*/, _, err := wvm.sqlchain.GetTable(wvm.activeDB.Owner, wvm.activeDB.Name, tableName)
		if err != nil {
			return fmt.Errorf("[wvm:setupIndex] %s", err)
		}
		wvm.activeTable = tbl
		// this actually reads the T chunk, parses the columns and creates the indexes for the column_num
		if schemaType == "index" {
			if strings.Contains(indexName, "sqlite") {
				column, err := getPrimaryColumn(tbl)
				if err != nil {
					return fmt.Errorf("[wvm:setupIndex] %s", err)
				}
				wvm.index[cursorID] = getIndex(column)
			} else {
				column, err := getColumn(tbl, indexName)
				if err != nil {
					return fmt.Errorf("[wvm:setupIndex] %s", err)
				}
				wvm.index[cursorID] = getIndex(column)
			}
		} else {
			column, err := getPrimaryColumn(tbl)
			if err != nil {
				return fmt.Errorf("[wvm:setupIndex] %s", err)
			}
			wvm.index[cursorID] = getIndex(column)
		}
	}
	//wvm.dumpCursors()
	//fmt.Printf("[wvm:setupIndex] wvm.index[cursorID]: cursorID (%v), value (%+v)\n", cursorID, wvm.index[cursorID])

	return nil
}

func (wvm *WVM) setupRowId(key []byte) (rowid int64, ok bool, err error) {
	fmt.Printf("[wvm:setupRowId] key %+v\n", key)
	fres, ok, err := wvm.activeTable.Get(key)
	if err != nil {
		return rowid, false, fmt.Errorf("[wvm:setupRowId] %s", err)
	}
	if !ok {
		fmt.Printf("[wvm:setupRowId] chunk key [%x] not found!\n", key)
		return rowid, false, nil //chunk key not found
	}
	row, err := wvm.activeTable.ByteArrayToOrderedRow(fres)
	if err != nil {
		return rowid, false, fmt.Errorf("[wvm:setupRowId] %s", err)
	}
	fmt.Printf("[wvm:setupRowId] row output: %v\n", row) //opRowid?

	rowid = wvm.setRow(row)
	fmt.Printf("[wvm:setupRowId] rowid %+v", rowid)
	return rowid, true, nil
}

func (wvm *WVM) setupPrevRowId(cursorID int) (rowid int64, ok bool, err error) {

	//fmt.Printf("[wvm:setupPrevRowId] orig cursor: %+v\n", wvm.cursors[cursorID])
	cursor, ok := wvm.cursors[cursorID]
	if !ok {
		//wvm.eof = true
		return rowid, false, nil //fmt.Errorf("wvm:setupCurrentRowId] cursorID doesn't exist")
	}
	wvm.activeKey, _, err = cursor.Prev()
	if err != nil {
		//fmt.Printf("[wvm:setupPrevRowId] Throws err: %v\n", err)
		if err != io.EOF {
			return rowid, false, fmt.Errorf("[wvm:setupPrevRowId] %s", err)
		}
		wvm.eof = true
		//return rowid, false, nil
	}
	return wvm.setupRowId(wvm.activeKey)

}

func (wvm *WVM) setupNextRowId(cursorID int) (rowid int64, ok bool, err error) {

	fmt.Printf("[wvm:setupNextRowId] orig cursor: %+v\n", wvm.cursors[cursorID])

	cursor, ok := wvm.cursors[cursorID]
	if !ok {
		//wvm.eof = true
		return rowid, false, nil //fmt.Errorf("wvm:setupCurrentRowId] cursorID doesn't exist")
	}
	wvm.activeKey, _, err = cursor.Next()
	if err != nil {
		fmt.Printf("[wvm:setupNextRowId] Next throws err: %v\n", err)
		if err != io.EOF {
			return rowid, false, fmt.Errorf("[wvm:setupNextRowId] %s", err)
		}
		wvm.eof = true
		//return rowid, false, nil
	}
	fmt.Printf("[wvm:setupNextRowId] next cursor: %+v\n", cursor)
	return wvm.setupRowId(wvm.activeKey)

}

func (wvm *WVM) setupCurrentRowId(cursorID int) (rowid int64, ok bool, err error) {

	fmt.Printf("[wvm:setupCurrentRowId] cursorID is: %d\n", cursorID)
	fmt.Printf("[wvm:setupCurrentRowId] orig cursor: %+v\n", wvm.cursors[cursorID])
	cursor, ok := wvm.cursors[cursorID]
	if !ok {
		fmt.Printf(("wvm:setupCurrentRowId] cursorID doesn't exist\n"))
		//wvm.eof = true
		return rowid, false, nil //fmt.Errorf("wvm:setupCurrentRowId] cursorID doesn't exist")
	}

	//wvm.activeKey, _, err = cursor.Next()
	wvm.activeKey, err = cursor.GetCurrent()
	if err != nil { //todo
		fmt.Printf("[wvm:setupCurrentRowId] Current throws err: %v\n", err)
		if err != io.EOF {
			return rowid, false, fmt.Errorf("[wvm:setupCurrentRowId] %s", err) //fmt.Sprintf("[wvm:setupCurrentRowId] %s", err))
		}
		wvm.eof = true
		//return rowid, false, nil
	}
	fmt.Printf("[wvm:setupCurrentRowId] wvm.activeKey is now: [%+v]\n", wvm.activeKey)
	return wvm.setupRowId(wvm.activeKey)

}

//key []byte,
func (wvm *WVM) setRow(row map[int]interface{}) (result int64) {
	fmt.Printf("[wvm:setRow] ROW %v\n", row)

	// set the activerow`
	wvm.activeRow = row
	rowid := wvm.buildRowid()
	fmt.Printf("[wvm:setRow] built rowid: %v\n", rowid)
	return rowid

	//result = 43 //old stub
	//return result
}

func (wvm *WVM) buildRowid() (rowid int64) {
	rowid_string := string(wvm.activeKey) + wvm.activeTable.Name + wvm.activeDB.Name + wvm.activeDB.Owner
	rowid_hash := wolkcommon.Computehash([]byte(rowid_string))
	//var rowid_bytes [64]byte
	//copy(rowid_bytes[0:], rowid_hash)
	rowid = BytesToInt64(rowid_hash)

	return rowid

}

// (!ok, nil) means at least one was nil, (!ok, error) means not found
func (wvm *WVM) registerCheck(opP1 int, opP2 int) (ok bool, err error) {

	rp1, ok1 := wvm.register[opP1]
	rp2, ok2 := wvm.register[opP2]
	if !ok1 || !ok2 {
		//log.Error("[wvm:registerCheck] keys not found in wvm.register map - item in query is not found")
		//return nil
		return false, fmt.Errorf("[wvm:registerCheck] one or both keys (%v) (%v) not found in wvm.register map (%v) - item in query is not found", opP1, opP2, wvm.register)
	}

	if rp1 == nil || rp2 == nil {
		// TODO: should this be halted? or = nil and continue?
		log.Info("[wvm:registerCheck] item in query is not found. one of these is nil", "regop", rp1, "regop", rp2)
		//wvm.halted = true
		return false, nil
	}
	return true, nil

}

// func (wvm *WVM) getRowId(key []byte) (rowid int, err error) {
// 	for id, entry := range wvm.rowsetMap {
// 		if entry.key == key {
// 			return id, nil
// 		}
// 	}
// 	return rowid, fmt.Errorf("[wvm:getRowid] no rowid for this key"))
// }

/* Opcode:  Goto * P2 * * *
**
** An unconditional jump to address P2.
** The next instruction executed will be
** the one at index P2 from the beginning of
** the program.
**
** The P1 parameter is not actually used by this opcode.  However, it
** is sometimes set to 1 instead of 0 as a hint to the command-line shell
** that this Goto is the bottom of a loop and that the lines from P2 down
** to the current line should be indented for EXPLAIN output.
 */
func opGoto(wvm *WVM, op Op) (err error) {
	wvm.pc = op.P2
	wvm.jumped = true
	return nil
}

/* Opcode:  Gosub P1 P2 * * *
**
** Write the current address onto register P1
** and then jump to address P2.
 */
func opGosub(wvm *WVM, op Op) (err error) {
	wvm.register[op.P1] = wvm.pc
	return nil
}

/* Opcode: DecrJumpZero P1 P2 * * *
** Synopsis: if (--r[P1])==0 goto P2
**
** Register P1 must hold an integer.  Decrement the value in register P1
** then jump to P2 if the new value is exactly zero.
 */
func opDecrJumpZero(wvm *WVM, op Op) (err error) {
	wvm.register[op.P1] = wvm.register[op.P1].(int) - 1
	if wvm.register[op.P1].(int) == 0 {
		wvm.pc = op.P2
		wvm.jumped = true
	}
	return nil
}

/* Opcode:  Return P1 * * * *
**
** Jump to the next instruction after the address in register P1.  After
** the jump, register P1 becomes undefined.
 */
func opReturn(wvm *WVM, op Op) (err error) {
	wvm.pc = wvm.register[op.P1].(int)
	wvm.jumped = true

	wvm.register[op.P1] = nil
	return nil
}

// The ReopenIdx opcode works exactly like ReadOpen except that it first checks to see if the cursor on P1 is already open with a root page number of P2 and if it is this opcode becomes a no-op. In other words, if the cursor is already open, do not reopen it.
// The ReopenIdx opcode may only be used with P5==0 and with P4 being a P4_KEYINFO object. Furthermore, the P3 value must be the same as every other ReopenIdx or OpenRead for the same cursor number.
func opReopenIdx(wvm *WVM, op Op) (err error) {
	err = opOpenRead(wvm, op)
	if err != nil {
		return fmt.Errorf("[wvm:opReopenIdx] %s", err)
	}
	return nil
}

/* Opcode: OpenRead P1 P2 P3 P4 P5
** Synopsis: root=P2 iDb=P3
**
** Open a read-only cursor for the database table whose root page is
** P2 in a database file.  The database file is determined by P3.
** P3==0 means the main database, P3==1 means the database used for
** temporary tables, and P3>1 means used the corresponding attached
** database.  Give the new cursor an identifier of P1.  The P1
** values need not be contiguous but all P1 values should be small integers.
** It is an error for P1 to be negative.
**
** If P5!=0 then use the content of register P2 as the root page, not
** the value of P2 itself.
**
** There will be a read lock on the database whenever there is an
** open cursor.  If the database was unlocked prior to this instruction
** then a read lock is acquired as part of this instruction.  A read
** lock allows other processes to read the database but prohibits
** any other process from modifying the database.  The read lock is
** released when all cursors are closed.  If this instruction attempts
** to get a read lock but fails, the script terminates with an
** SQLITE_BUSY error code.
**
** The P4 value may be either an integer (P4_INT32) or a pointer to
** a KeyInfo structure (P4_KEYINFO). If it is a pointer to a KeyInfo
** structure, then said structure defines the content and collating
** sequence of the index being opened. Otherwise, if P4 is an integer
** value, it is set to the number of columns in the table.
**
** See also: OpenWrite, ReopenIdx
 */
func opOpenRead(wvm *WVM, op Op) (err error) {
	err = setupIndex(wvm, op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opOpenRead] %s", err)
	}
	return nil
}

/* Opcode: OpenWrite P1 P2 P3 P4 P5
** Synopsis: root=P2 iDb=P3
**
** Open a read/write cursor named P1 on the table or index whose root
** page is P2.  Or if P5!=0 use the content of register P2 to find the
** root page.
**
** The P4 value may be either an integer (P4_INT32) or a pointer to
** a KeyInfo structure (P4_KEYINFO). If it is a pointer to a KeyInfo
** structure, then said structure defines the content and collating
** sequence of the index being opened. Otherwise, if P4 is an integer
** value, it is set to the number of columns in the table, or to the
** largest index of any column of the table that is actually used.
**
** This instruction works just like OpenRead except that it opens the cursor
** in read/write mode.  For a given table, there can be one or more read-only
** cursors or a single read/write cursor but not both.
**
** See also OpenRead.
 */
// {"n":1,"opcode":"OpenWrite","p2":2,"p4":"2"}
func opOpenWrite(wvm *WVM, op Op) (err error) {
	// same as OpenRead, but perhaps we could call StartBuffer, which would be Flushed upon "Close"?
	//fmt.Printf("[wvm:opOpenWrite] op: %+v\n", op)
	//fmt.Printf("[wvm:opOpenWrite] registers: %+v\n", wvm.register)
	//fmt.Printf("[wvm:opOpenWrite] activeRow: %+v\n", wvm.activeRow)
	err = setupIndex(wvm, op.P1, op.P2)
	if err != nil {
		return err
	}
	return nil
}

/* Opcode: Rewind P1 P2 * * *
**
** The next use of the Rowid or Column or Next instruction for P1
** will refer to the first entry in the database table or index.
** If the table or index is empty, jump immediately to P2.
** If the table or index is not empty, fall through to the following
** instruction.
**
** This opcode leaves the cursor configured to move in forward order,
** from the beginning toward the end.  In other words, the cursor is
** configured to use Next, not Prev.
 */
func opRewind(wvm *WVM, op Op) (err error) {
	// the active index specified by one of { OpenRead , P1 } will be used to call
	//  func (t *Tree) SeekFirst(u *SQLUser) (e OrderedDatabaseCursor, err error)
	index := wvm.index[op.P1]
	cursor, err := index.SeekFirst()
	if err != nil {
		fmt.Printf("[wvm:opRewind] SeekFirst %s", err)
		//log.Error(fmt.Sprintf("[wvm:opRewind] err: (%s). Continuing.", err))
		return nil
	}
	wvm.cursors[op.P1] = cursor
	// fmt.Printf("REWIND => SeekFirst %v\n", index)
	// index.Print()
	return nil
}

/* Opcode: Last P1 P2 * * *
**
** The next use of the Rowid or Column or Prev instruction for P1
** will refer to the last entry in the database table or index.
** If the table or index is empty and P2>0, then jump immediately to P2.
** If P2 is 0 or if the table or index is not empty, fall through
** to the following instruction.
**
** This opcode leaves the cursor configured to move in reverse order,
** from the end toward the beginning.  In other words, the cursor is
** configured to use Prev, not Next.
 */
func opLast(wvm *WVM, op Op) (err error) {
	// same as above except use SeekLast
	index := wvm.index[op.P1]
	cursor, err := index.SeekLast()
	if err != nil { //TODO: check this
		fmt.Printf("[wvm:opLast] SeekLast %s", err)
		//log.Error(fmt.Sprintf("[wvm:opLast] err: (%s). Continuing"))
		return nil
	}
	cursor.Prev()
	wvm.cursors[op.P1] = cursor
	//fmt.Printf("[wvm:opLast] SeekLast cursor: %v\n", cursor)
	return nil
}

/* Opcode: Rowid P1 P2 * * *
** Synopsis: r[P2]=rowid
**
** Store in register P2 an integer which is the key of the table entry that
** P1 is currently point to.
**
** P1 can be either an ordinary table or a virtual table.  There used to
** be a separate OP_VRowid opcode for use with virtual tables, but this
** one opcode now works for both table
 */
func opRowid(wvm *WVM, op Op) (err error) {
	// with the active cursor 'c' set up by Rewind (or Last)
	//  k, v, err := c.Next(u)
	// will hold the 32-byte key of what to read to get the record
	// but after getting the key, we need to get the K chunk for the record and use the Table to load up the register

	// rowid, ok, err := wvm.setupNextRowId(op.P1)
	// if err != nil {
	// 	return fmt.Errorf("[wvm:opRowid] %s", err))
	// }
	// if !ok {
	// 	fmt.Printf("[wvm:opRowid] chunk key not found. what to do?\n")
	// 	//chunk key not found, do what? //TODO
	// }

	rowid, ok, err := wvm.setupCurrentRowId(op.P1)
	if err != nil {
		return fmt.Errorf("[wvm:opRowid] %s", err)
	}
	if !ok {
		rowid, ok, err = wvm.setupNextRowId(op.P1)
		if err != nil {
			return fmt.Errorf("[wvm:opRowid] %s", err)
		}
		if !ok {
			return nil //chunk key not found. what TODO?
		}
	}

	if rowid != 0 {
		wvm.register[op.P2] = rowid
		wvm.rowSet[op.P2] = append(wvm.rowSet[op.P2], rowid)
		//fmt.Printf("[wvm:opRowid] rowid %v generated for op.P1 %v table entry\n", rowid, op.P1)
	} //else {
	//fmt.Printf("[wvm:opRowid] rowid is 0, generated for op.P1 %v table entry\n", op.P1)
	//}
	//wvm.register[op.P2] = 42 //old stub
	return nil
}

/* Opcode: Column P1 P2 P3 P4 P5
** Synopsis: r[P3]=PX
**
** Interpret the data that cursor P1 points to as a structure built using
** the MakeRecord instruction.  (See the MakeRecord opcode for additional
** information about the format of the data.)  Extract the P2-th column
** from this record.  If there are less that (P2+1)
** values in the record, extract a NULL.
**
** The value extracted is stored in register P3.
**
** If the record contains fewer than P2 fields, then extract a NULL.  Or,
** if the P4 argument is a P4_MEM use the value of the P4 argument as
** the result.
**
** If the OPFLAG_CLEARCACHE bit is set on P5 and P1 is a pseudo-table cursor,
** then the cache of the cursor is reset prior to extracting the column.
** The first OP_Column against a pseudo-table after the value of the content
** register has changed should have this bit set.
**
** If the OPFLAG_LENGTHARG and OPFLAG_TYPEOFARG bits are set on P5 then
** the result is guaranteed to only be used as the argument of a length()
** or typeof() function, respectively.  The loading of large blobs can be
** skipped for length() and all content loading can be skipped for typeof().
 */
func opColumn(wvm *WVM, op Op) (err error) {
	// use the JSON record from a "RowId" operation and the active Table columns
	// to read the P2th column and put it into register p3

	//fmt.Printf("[wvm:opColumn] op: %+v\n", op)
	if len(wvm.activeRow) > 0 {
		fmt.Printf("[wvm:opColumn] wvm.activeRow > 0, %+v\n", wvm.activeRow)
	} else {
		fmt.Printf("[wvm:opColumn] rowid to setup: %v\n", op.P1)
		_, ok, err := wvm.setupCurrentRowId(op.P1)
		if err != nil {
			return fmt.Errorf("[wvm:opColumn] %s", err)
		}
		if !ok {
			_, ok, err := wvm.setupNextRowId(op.P1)
			if err != nil {
				return fmt.Errorf("[wvm:opColumn] %s", err)
			}
			if !ok {
				log.Error("[wvm:opColumn] column not found. query item is not found")
				wvm.halted = true
				return nil
			}
			fmt.Printf("[wvm:opColumn] setupNextRowId \n")
		}

	}
	fmt.Printf("[wvm:opColumn] activeRow after: %+v\n", wvm.activeRow)
	wvm.register[op.P3] = wvm.activeRow[op.P2]
	fmt.Printf("[wvm:opColumn] Column %d (register %d) => %v\n", op.P2, op.P3, wvm.register[op.P3])
	return nil
}

/* Opcode: ResultRow P1 P2 * * *
** Synopsis: output=r[P1@P2]
**
** The register P1 through P1+P2-1 contain a single row of
** results. This opcode causes the sqlite3_step() call to terminate
** with an SQLITE_ROW return code and it sets up the sqlite3_stmt
** structure to provide access to the r(P1)..r(P1+P2-1) values as
** the result row.
 */
func opResultRow(wvm *WVM, op Op) (err error) {
	// Given a set of Column operations holding everything in a set of registers,
	// this will take all the register values and assemble them in self's SQLRow output
	//if wvm.eof {
	//	fmt.Printf("[wvm:opResultRow] wvm.eof\n")
	//	return nil
	//	}
	output := make([]interface{}, 0)
	for i := op.P1; i < op.P1+op.P2; i++ {
		if wvm.register[i] != nil {
			output = append(output, wvm.register[i])
		}
	}
	//fmt.Printf("[wvm:opResultRow] ---> OUTPUT: %v\n", output)
	if len(output) > 0 {
		wvm.Rows = append(wvm.Rows, output)
	}
	return nil
}

/* Opcode: Next P1 P2 P3 P4 P5
**
** Advance cursor P1 so that it points to the next key/data pair in its
** table or index.  If there are no more key/value pairs then fall through
** to the following instruction.  But if the cursor advance was successful,
** jump immediately to P2.
**
** The Next opcode is only valid following an SeekGT, SeekGE, or
** OP_Rewind opcode used to position the cursor.  Next is not allowed
** to follow SeekLT, SeekLE, or OP_Last.
**
** The P1 cursor must be for a real table, not a pseudo-table.  P1 must have
** been opened prior to this opcode or the program will segfault.
**
** The P3 value is a hint to the btree implementation. If P3==1, that
** means P1 is an SQL index and that this instruction could have been
** omitted if that index had been unique.  P3 is usually 0.  P3 is
** always either 0 or 1.
**
** P4 is always of type P4_ADVANCE. The function pointer points to
** sqlite3BtreeNext().
**
** If P5 is positive and the jump is taken, then event counter
** number P5-1 in the prepared statement is incremented.
**
** See also: Prev, NextIfOpen
 */
func opNext(wvm *WVM, op Op) (err error) {

	fmt.Printf("[wvm:opNext] start")
	_, ok, err := wvm.setupNextRowId(op.P1)
	if err != nil {
		return fmt.Errorf("[wvm:opNext] %s", err)
	}
	if !ok {
		fmt.Printf("[wvm:opNext] cursor doesn't exist.")
		return nil // fall through to the next instruction
	}
	if wvm.eof {
		return nil
	}
	wvm.pc = op.P2
	wvm.jumped = true
	return nil
}

// Because "RowId" actually advanced the pointer, it is not necessary to call
//  k, v, err := c.Next(u)
// Instead,
//  if err == nil jump to P2
//  if err != nil jump to next line

/* Opcode: Prev P1 P2 P3 P4 P5
**
** Back up cursor P1 so that it points to the previous key/data pair in its
** table or index.  If there is no previous key/value pairs then fall through
** to the following instruction.  But if the cursor backup was successful,
** jump immediately to P2.
**
**
** The Prev opcode is only valid following an SeekLT, SeekLE, or
** OP_Last opcode used to position the cursor.  Prev is not allowed
** to follow SeekGT, SeekGE, or OP_Rewind.
**
** The P1 cursor must be for a real table, not a pseudo-table.  If P1 is
** not open then the behavior is undefined.
**
** The P3 value is a hint to the btree implementation. If P3==1, that
** means P1 is an SQL index and that this instruction could have been
** omitted if that index had been unique.  P3 is usually 0.  P3 is
** always either 0 or 1.
**
** P4 is always of type P4_ADVANCE. The function pointer points to
** sqlite3BtreePrevious().
**
** If P5 is positive and the jump is taken, then event counter
** number P5-1 in the prepared statement is incremented.
 */
func opPrev(wvm *WVM, op Op) (err error) {
	_, _, err = wvm.setupPrevRowId(op.P1)
	if err != nil {
		return fmt.Errorf("[wvm:opPrev] %s", err)
	}
	if wvm.eof {
		return nil
	}
	wvm.pc = op.P2
	wvm.jumped = true
	return nil
}

/* Opcode: PrevIfOpen P1 P2 P3 P4 P5
**
** This opcode works just like Prev except that if cursor P1 is not
** open it behaves a no-op.
 */
func opPrevIfOpen(wvm *WVM, op Op) (err error) {
	return opPrev(wvm, op)
}
func opNextIfOpen(wvm *WVM, op Op) (err error) {
	return opNext(wvm, op)
}

/* Opcode: Close P1 * * * *
**
** Close a cursor previously opened as P1.  If P1 is not
** currently open, this instruction is a no-op.
 */
func opClose(wvm *WVM, op Op) (err error) {
	// call FlushBuffer, if there is some buffered output
	return nil
}

/* Opcode:  Halt P1 P2 * P4 P5
**
** Exit immediately.  All open cursors, etc are closed
** automatically.
**
** P1 is the result code returned by sqlite3_exec(), sqlite3_reset(),
** or sqlite3_finalize().  For a normal halt, this should be SQLITE_OK (0).
** For errors, it can be some other value.  If P1!=0 then P2 will determine
** whether or not to rollback the current transaction.  Do not rollback
** if P2==OE_Fail. Do the rollback if P2==OE_Rollback.  If P2==OE_Abort,
** then back out all changes that have occurred during this execution of the
** VDBE, but do not rollback the transaction.
**
** If P4 is not null then it is an error message string.
**
** P5 is a value between 0 and 4, inclusive, that modifies the P4 string.
**
**    0:  (no change)
**    1:  NOT NULL contraint failed: P4
**    2:  UNIQUE constraint failed: P4
**    3:  CHECK constraint failed: P4
**    4:  FOREIGN KEY constraint failed: P4
**
** If P5 is not zero and P4 is NULL, then everything after the ":" is
** omitted.
**
** There is an implied "Halt 0 0 0" instruction inserted at the very end of
** every program.  So a jump past the last instruction of the program
** is the same as executing Halt.
 */
func opHalt(wvm *WVM, op Op) (err error) {
	// Given a set of SQLRows output with ResultRow opcode, finish program execution
	wvm.halted = true

	// DEBUG only
	// fmt.Printf("\n[wvm:opHalt] op: state at opHALT:\n")
	// fmt.Printf("[wvm:opHalt] jumped: %+v\n", wvm.jumped)
	// fmt.Printf("[wvm:opHalt] pc: %+v\n", wvm.pc)
	// fmt.Printf("[wvm:opHalt] registers: %+v\n", wvm.register)
	// fmt.Printf("[wvm:opHalt] indexes: %+v\n", wvm.index)
	// wvm.dumpCursors()
	// fmt.Printf("[wvm:opHalt] rowSets: %+v\n", wvm.rowSet)
	// fmt.Printf("[wvm:opHalt] activeKey: %+v\n", wvm.activeKey)
	// fmt.Printf("[wvm:opHalt] activeRow: %+v\n", wvm.activeRow)
	// //fmt.Printf("[wvm:opHalt] sorter: %+v\n", wvm.sorter)
	// fmt.Printf("[wvm:opHalt] Rows(output): %+v\n", wvm.Rows)

	return nil
}

/* Opcode: NewRowid P1 P2 P3 * *
** Synopsis: r[P2]=rowid
**
** Get a new integer record number (a.k.a "rowid") used as the key to a table.
** The record number is not previously used as a key in the database
** table that cursor P1 points to.  The new record number is written
** written to register P2.
**
** If P3>0 then P3 is a register in the root frame of this VDBE that holds
** the largest previously generated record number. No new record numbers are
** allowed to be less than this value. When this value reaches its maximum,
** an SQLITE_FULL error is generated. The P3 register is updated with the '
** generated record number. This P3 mechanism is used to help implement the
** AUTOINCREMENT feature.
 */
func opNewRowid(wvm *WVM, op Op) (err error) {
	// Because there is no centralized auto incrementing primary key, skipping this
	return nil
}

/* Opcode: MakeRecord P1 P2 P3 P4 *
** Synopsis: r[P3]=mkrec(r[P1@P2])
**
** Convert P2 registers beginning with P1 into the [record format]
** use as a data record in a database table or as a key
** in an index.  The OP_Column opcode can decode the record later.
**
** P4 may be a string that is P2 characters long.  The N-th character of the
** string indicates the column affinity that should be used for the N-th
** field of the index key.
**
** The mapping from character to affinity is given by the SQLITE_AFF_
** macros defined in sqliteInt.h.
**
** If P4 is NULL then all index fields have the affinity BLOB.
 */
func opMakeRecord(wvm *WVM, op Op) (err error) {
	// use the Table and the reg p1 through p2 to
	// (a) build a record and put the data into r[P1@P2]

	r := make(map[int]interface{}, 0)
	for i := 0; i < op.P2; i++ {
		r[i] = wvm.register[op.P1+i]
	}
	wvm.register[op.P3] = wvm.setRow(r)

	return nil
}

/* Opcode: Insert P1 P2 P3 P4 P5
** Synopsis: intkey=r[P3] data=r[P2]
**
** Write an entry into the table of cursor P1.  A new entry is
** created if it doesn't already exist or the data for an existing
** entry is overwritten.  The data is the value MEM_Blob stored in register
** number P2. The key is stored in register P3. The key must
** be a MEM_Int.
**
** If the OPFLAG_NCHANGE flag of P5 is set, then the row change count is
** incremented (otherwise not).  If the OPFLAG_LASTROWID flag of P5 is set,
** then rowid is stored for subsequent return by the
** sqlite3_last_insert_rowid() function (otherwise it is unmodified).
**
** If the OPFLAG_USESEEKRESULT flag of P5 is set, the implementation might
** run faster by avoiding an unnecessary seek on cursor P1.  However,
** the OPFLAG_USESEEKRESULT flag must only be set if there have been no prior
** seeks on the cursor or if the most recent seek used a key equal to P3.
**
** If the OPFLAG_ISUPDATE flag is set, then this opcode is part of an
** UPDATE operation.  Otherwise (if the flag is clear) then this opcode
** is part of an INSERT operation.  The difference is only important to
** the update hook.
**
** Parameter P4 may point to a Table structure, or may be NULL. If it is
** not NULL, then the update-hook (sqlite3.xUpdateCallback) is invoked
** following a successful insert.
**
** (WARNING/TODO: If P1 is a pseudo-cursor and P2 is dynamically
** allocated, then ownership of P2 is transferred to the pseudo-cursor
** and register P2 becomes ephemeral.  If the cursor is changed, the
** value of register P2 will then change.  Make sure this does not
** cause any problems.)
**
** This instruction only works on tables.  The equivalent instruction
** for indices is OP_IdxInsert.
 */
func opInsert(wvm *WVM, op Op) (err error) {
	// Use the index to decide if:
	// (1) key exists in index, so increment version #
	// (2) key does not exist in index, so prepare to insert key with version # being 0
	// k = r[P3], data = r[P2]
	// (b) store the K chunk build in r[P2]
	// (c) call func (t *Tree) Put(u *SQLUser, key []byte /*K*/, v []byte /*V*/) (okresult bool, err error)

	log.Info("[wvm:opInsert] trying to insert row", "row", wvm.activeRow)
	isImmutable := true
	if wvm.OPFLAG_ISUPDATE {
		isImmutable = false
		log.Info(fmt.Sprintf("[wvm:opInsert] OPFLAG_ISUPDATE (%v), isImmutable (%v)", wvm.OPFLAG_ISUPDATE, isImmutable))
	}
	err = wvm.activeTable.OrderedPut(wvm.activeRow, isImmutable)
	if err != nil {
		return fmt.Errorf("[wvm:opInsert] %s", err)
	}
	wvm.AffectedRowCount++
	//fmt.Printf("[wvm:opInsert] inserted %v\n", wvm.activeRow)
	return nil
}

/* Opcode: Delete P1 P2 P3 P4 P5
**
** Delete the record at which the P1 cursor is currently pointing.
**
** If the OPFLAG_SAVEPOSITION bit of the P5 parameter is set, then
** the cursor will be left pointing at  either the next or the previous
** record in the table. If it is left pointing at the next record, then
** the next Next instruction will be a no-op. As a result, in this case
** it is ok to delete a record from within a Next loop. If
** OPFLAG_SAVEPOSITION bit of P5 is clear, then the cursor will be
** left in an undefined state.
**
** If the OPFLAG_AUXDELETE bit is set on P5, that indicates that this
** delete one of several associated with deleting a table row and all its
** associated index entries.  Exactly one of those deletes is the "primary"
** delete.  The others are all on OPFLAG_FORDELETE cursors or else are
** marked with the AUXDELETE flag.
**
** If the OPFLAG_NCHANGE flag of P2 (NB: P2 not P5) is set, then the row
** change count is incremented (otherwise not).
**
** P1 must not be pseudo-table.  It has to be a real table with
** multiple rows.
**
** If P4 is not NULL then it points to a Table object. In this case either
** the update or pre-update hook, or both, may be invoked. The P1 cursor must
** have been positioned using OP_NotFound prior to invoking this opcode in
** this case. Specifically, if one is configured, the pre-update hook is
** invoked if P4 is not NULL. The update-hook is invoked if one is configured,
** P4 is not NULL, and the OPFLAG_NCHANGE flag is set in P2.
**
** If the OPFLAG_ISUPDATE flag is set in P2, then P3 contains the address
** of the memory cell that contains the value that the rowid of the row will
** be set to by the update.
 */
func opDelete(wvm *WVM, op Op) (err error) {
	// whatever the active cursor is on, get the key
	ok, err := wvm.activeTable.Delete(wvm.activeKey)
	if err != nil {
		return fmt.Errorf("[wvm:opDelete] %s", err)
	}
	if ok {
		fmt.Printf("[wvm:opDelete] deleted key, incrementing affectedrowct\n")
		wvm.AffectedRowCount++
		wvm.cursors[op.P1], _, _ = wvm.index[op.P1].Seek(wvm.activeKey) //reset the cursor
	} else {
		log.Error("[wvm:opDelete] did not delete, key not found in table", "key", wvm.activeKey, "table", wvm.activeTable.Name)
		//return fmt.Errorf("[wvm:opDelete] did not delete, key %x/%v not found in table %+v", wvm.activeKey, wvm.activeKey, wvm.activeTable.Name)

	}

	return nil
}

/* Opcode: Integer P1 P2 * * *
** Synopsis: r[P2]=P1
**
** The 32-bit integer value P1 is written into register P2.
 */
func opInteger(wvm *WVM, op Op) (err error) {
	wvm.register[op.P2] = op.P1
	return nil
}

/* Opcode: Int64 * P2 * P4 *
** Synopsis: r[P2]=P4
**
** P4 is a pointer to a 64-bit integer value.
** Write that value into register P2.
 */
func opInt64(wvm *WVM, op Op) (err error) {
	p4int, err := strconv.ParseInt(op.P4, 10, 64)
	if err != nil {
		return fmt.Errorf("[wvm:opInt64] %s", err)
	}
	fmt.Printf("[wvm:opInt64] op.P4 (%v), p4int (%v)\n", op.P4, p4int)
	wvm.register[op.P2] = p4int
	return nil
}

/* Opcode: String8 * P2 * P4 *
** Synopsis: r[P2]='P4'
**
** P4 points to a nul terminated UTF-8 string. This opcode is transformed
** into a String opcode before it is executed for the first time.  During
** this transformation, the length of string P4 is computed and stored
** as the P1 parameter.
 */
func opString8(wvm *WVM, op Op) (err error) {
	wvm.register[op.P2] = op.P4 //p4 is a string
	op.P1 = len(op.P4)
	return nil
}

/* Opcode: Real * P2 * P4 *
** Synopsis: r[P2]=P4
**
** P4 is a pointer to a 64-bit floating point value.
** Write that value into register P2.
 */
func opReal(wvm *WVM, op Op) (err error) {
	// p4 is a string
	p4float, err := strconv.ParseFloat(op.P4, 64)
	if err != nil {
		return fmt.Errorf("[wvm:opReal] %s", err)
	}
	wvm.register[op.P2] = p4float
	return nil
}

/* Opcode: Null P1 P2 P3 * *
** Synopsis: r[P2..P3]=NULL
**
** Write a NULL into registers P2.  If P3 greater than P2, then also write
** NULL into register P3 and every register in between P2 and P3.  If P3
** is less than P2 (typically P3 is zero) then only register P2 is
** set to NULL.
**
** If the P1 value is non-zero, then also set the MEM_Cleared flag so that
** NULL values will not compare equal even if SQLITE_NULLEQ is set on
** OP_Ne or OP_Eq.
 */
func opNull(wvm *WVM, op Op) (err error) {
	if op.P3 > op.P2 {
		for i := op.P2; i <= op.P3; i++ {
			wvm.register[i] = nil
		}
	} else {
		wvm.register[op.P2] = nil
	}
	return nil
}

/* Opcode: Add P1 P2 P3 * *
** Synopsis: r[P3]=r[P1]+r[P2]
**
** Add the value in register P1 to the value in register P2
** and store the result in register P3.
** If either input is NULL, the result is NULL.
 */
func opAdd(wvm *WVM, op Op) (err error) {
	// TODO: check that a string in opAdd is always 0 in sqlite
	//wvm.register[op.P3] = wvm.register[op.P2].(int) + wvm.register[op.P1].(int)
	ok, err := wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opAdd] %s", err)
	}
	wvm.register[op.P3] = nil
	if !ok {
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opAdd] %s", err)
	}
	f2, s2, err := convertNumOrString(wvm.register[op.P2])
	if err != nil {
		return fmt.Errorf("[wvm:opAdd] %s", err)
	}
	// TODO: worry about adding ints and getting an int
	if f1 != nil && f2 != nil {
		wvm.register[op.P3] = *f1 + *f2
		return nil
	}
	if s1 != nil || s2 != nil {
		return fmt.Errorf("[wvm:opAdd] not implemented for strings")
	}

	// switch r1 := wvm.register[op.P1].(type) {
	// case int, int64:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64:
	// 		wvm.register[op.P3] = r1.(int) + r2.(int)
	// 		break
	// 	case float64:
	// 		wvm.register[op.P3] = float64(r1.(int)) + r2
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = r1 //TODO: make sure this is what SQLite does, strings seem to be 0
	// 		break
	// 	default:
	// 		return fmt.Errorf("[wvm:opAdd] type not supported (%v)", reflect.TypeOf(wvm.register[op.P2]))
	// 	}
	// 	break
	// case float64:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64:
	// 		wvm.register[op.P3] = r1 + float64(r2.(int))
	// 		break
	// 	case float64:
	// 		wvm.register[op.P3] = r1 + r2
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = r1 //TODO: make sure this is what SQLite does, strings seem to be 0
	// 		break
	// 	default:
	// 		return fmt.Errorf("[wvm:opAdd] type not supported (%v)", reflect.TypeOf(wvm.register[op.P2]))
	// 	}
	// 	break
	// case string:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64, float64:
	// 		wvm.register[op.P3] = r2
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = 0
	// 		break
	// 	}
	// 	break
	// default:
	// 	return fmt.Errorf("[wvm:opAdd] type not supported (%v)", reflect.TypeOf(wvm.register[op.P1]))
	// }

	return nil
}

/* Opcode: Subtract P1 P2 P3 * *
** Synopsis: r[P3]=r[P2]-r[P1]
**
** Subtract the value in register P1 from the value in register P2
** and store the result in register P3.
** If either input is NULL, the result is NULL.
 */
func opSubtract(wvm *WVM, op Op) (err error) {
	// TODO: check that a string in opSub is always 0 in sqlite
	ok, err := wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opSubtract] %s", err)
	}
	wvm.register[op.P3] = nil
	if !ok {
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opSubtract] %s", err)
	}
	f2, s2, err := convertNumOrString(wvm.register[op.P2])
	if err != nil {
		return fmt.Errorf("[wvm:opSubtract] %s", err)
	}
	// TODO: worry about subtracting ints and getting an int
	if f1 != nil && f2 != nil {
		wvm.register[op.P3] = *f2 - *f1
		return nil
	}
	if s1 != nil || s2 != nil {
		return fmt.Errorf("[wvm:opSubtract] not implemented for strings")
	}

	// switch r1 := wvm.register[op.P1].(type) {
	// case int, int64:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64:
	// 		wvm.register[op.P3] = r2.(int) - r1.(int)
	// 		break
	// 	case float64:
	// 		wvm.register[op.P3] = r2 - float64(r1.(int))
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = 0 - r1.(int)
	// 		break
	// 	default:
	// 		return fmt.Errorf("[wvm:opSubtract] type not supported (%v)", reflect.TypeOf(wvm.register[op.P2]))
	// 	}
	// case float64:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64:
	// 		wvm.register[op.P3] = float64(r2.(int)) - r1
	// 		break
	// 	case float64:
	// 		wvm.register[op.P3] = r2 - r1
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = 0 - r1
	// 		break
	// 	default:
	// 		return fmt.Errorf("[wvm:opSubtract] type not supported (%v)", reflect.TypeOf(wvm.register[op.P2]))
	// 	}
	// case string:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64, float64:
	// 		wvm.register[op.P3] = r2
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = 0
	// 		break
	// 	}
	// 	break
	// default:
	// 	return fmt.Errorf("[wvm:opSubtract] type not supported (%v)", reflect.TypeOf(wvm.register[op.P1]))
	// }

	return nil
}

/* Opcode: Multiply P1 P2 P3 * *
** Synopsis: r[P3]=r[P1]*r[P2]
**
**
** Multiply the value in register P1 by the value in register P2
** and store the result in register P3.
** If either input is NULL, the result is NULL.
 */
func opMultiply(wvm *WVM, op Op) (err error) {
	//wvm.register[op.P3] = wvm.register[op.P1].(int) * wvm.register[op.P2].(int)
	ok, err := wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opMultiply] %s", err)
	}
	wvm.register[op.P3] = nil
	if !ok {
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opMultiply] %s", err)
	}
	f2, s2, err := convertNumOrString(wvm.register[op.P2])
	if err != nil {
		return fmt.Errorf("[wvm:opMultiply] %s", err)
	}
	// TODO: worry about multiplying ints and getting an int
	if f1 != nil && f2 != nil {
		wvm.register[op.P3] = *f1 * *f2
		return nil
	}
	if s1 != nil || s2 != nil {
		return fmt.Errorf("[wvm:opMultiply] not implemented for strings")
	}
	// switch r1 := wvm.register[op.P1].(type) {
	// case int, int64:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64:
	// 		wvm.register[op.P3] = r1.(int) * r2.(int)
	// 		break
	// 	case float64:
	// 		wvm.register[op.P3] = float64(r1.(int)) * r2
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = 0
	// 		break
	// 	default:
	// 		return fmt.Errorf("[wvm:opMultiply] type not supported (%v)", reflect.TypeOf(wvm.register[op.P2]))
	// 	}
	// 	break
	// case float64:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64:
	// 		wvm.register[op.P3] = r1 * float64(r2.(int))
	// 		break
	// 	case float64:
	// 		wvm.register[op.P3] = r1 * r2
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = 0
	// 		break
	// 	default:
	// 		return fmt.Errorf("[wvm:opMultiply] type not supported (%v)", reflect.TypeOf(wvm.register[op.P2]))
	// 	}
	// 	break
	// case string:
	// 	wvm.register[op.P3] = 0
	// 	break
	// default:
	// 	return fmt.Errorf("[wvm:opMultiply] type not supported (%v)", reflect.TypeOf(wvm.register[op.P1]))
	// }

	return nil
}

/* Opcode: Divide P1 P2 P3 * *
** Synopsis: r[P3]=r[P2]/r[P1]
**
** Divide the value in register P1 by the value in register P2
** and store the result in register P3 (P3=P2/P1). If the value in
** register P1 is zero, then the result is NULL. If either input is
** NULL, the result is NULL.
 */
func opDivide(wvm *WVM, op Op) (err error) {
	//wvm.register[op.P3] = wvm.register[op.P1].(int) / wvm.register[op.P2].(int)
	ok, err := wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opDivide] %s", err)
	}
	wvm.register[op.P3] = nil
	if !ok {
		return nil
	}
	if wvm.register[op.P1] == 0 {
		wvm.register[op.P3] = nil
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opDivide] %s", err)
	}
	f2, s2, err := convertNumOrString(wvm.register[op.P2])
	if err != nil {
		return fmt.Errorf("[wvm:opDivide] %s", err)
	}
	if f1 != nil && *f1 == 0 {
		return nil
	}
	// TODO: worry about dividing ints and getting an int
	if f1 != nil && f2 != nil {
		wvm.register[op.P3] = *f1 / *f2
		return nil
	}
	if s1 != nil || s2 != nil {
		return fmt.Errorf("[wvm:opDivide] not implemented for strings")
	}

	return nil
	// switch r1 := wvm.register[op.P1].(type) {
	// case int, int64:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64:
	// 		wvm.register[op.P3] = r1.(int) / r2.(int)
	// 		break
	// 	case float64:
	// 		wvm.register[op.P3] = float64(r1.(int)) / r2
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = 0
	// 		break
	// 	default:
	// 		return fmt.Errorf("[wvm:opMultiply] type not supported (%v)", reflect.TypeOf(wvm.register[op.P2]))
	// 	}
	// 	break
	// case float64:
	// 	switch r2 := wvm.register[op.P2].(type) {
	// 	case int, int64:
	// 		wvm.register[op.P3] = r1 * float64(r2.(int))
	// 		break
	// 	case float64:
	// 		wvm.register[op.P3] = r1 * r2
	// 		break
	// 	case string:
	// 		wvm.register[op.P3] = 0
	// 		break
	// 	default:
	// 		return fmt.Errorf("[wvm:opMultiply] type not supported (%v)", reflect.TypeOf(wvm.register[op.P2]))
	// 	}
	// 	break
	// case string:
	// 	wvm.register[op.P3] = 0
	// 	break
	// default:
	// 	return fmt.Errorf("[wvm:opMultiply] type not supported (%v)", reflect.TypeOf(wvm.register[op.P1]))
	// }
	//
	// return nil
}

/* Opcode: Remainder P1 P2 P3 * *
** Synopsis: r[P3]=r[P2]%r[P1]
**
** Compute the remainder after integer register P2 is divided by
** register P1 and store the result in register P3.
** If the value in register P1 is zero the result is NULL.
** If either operand is NULL, the result is NULL.
 */
func opRemainder(wvm *WVM, op Op) (err error) {

	ok, err := wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opRemainder] %s", err)
	}
	wvm.register[op.P3] = nil
	if !ok {
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opRemainder] %s", err)
	}
	f2, s2, err := convertNumOrString(wvm.register[op.P2])
	if err != nil {
		return fmt.Errorf("[wvm:opRemainder] %s", err)
	}
	if f1 != nil && *f1 == 0 { // b/c you can't mod by 0
		return nil
	}
	if f1 != nil && f2 != nil {
		//log.Info("[wvm:opRemainder]", "*f1", *f1, "*f2", *f2)
		wvm.register[op.P3] = int(*f2) % int(*f1)
	}
	if s1 != nil && s2 != nil {
		return fmt.Errorf("[wvm:opRemainder] not implemented for strings")
	}
	return nil
	// if wvm.register[op.P1] == 0 { // b/c you can't mod by 0
	// 	log.Info("[wvm:opRemainder] divisor is 0", "p1", op.P1)
	// 	wvm.register[op.P3] = nil
	// 	return nil
	// }
	// var p2 int
	// switch tp2 := wvm.register[op.P2].(type) {
	// case int:
	// 	p2 = tp2
	// 	break
	// case float64:
	// 	p2 = int(tp2)
	// 	break
	// case int64:
	// 	p2 = int(tp2)
	// 	break
	// case string:
	// 	intp2, err := strconv.Atoi(tp2)
	// 	if err != nil {
	// 		return fmt.Errorf("[wvm:opRemainder] %s", err)
	// 	}
	// 	p2 = intp2
	// 	break
	// default:
	// 	return fmt.Errorf("[wvm:opRemainder] type (%v) not supported for reg[p2] (%v)", reflect.TypeOf(tp2), tp2)
	// }
	//
	// switch p1 := wvm.register[op.P1].(type) {
	// case int:
	// 	wvm.register[op.P3] = p2 % p1
	// 	break
	// case float64:
	// 	wvm.register[op.P3] = p2 % int(p1)
	// 	break
	// case int64:
	// 	wvm.register[op.P3] = p2 % int(p1)
	// 	break
	// case string:
	// 	intp1, err := strconv.Atoi(p1)
	// 	if err != nil {
	// 		return fmt.Errorf("[wvm:opRemainder] %s", err)
	// 	}
	// 	wvm.register[op.P3] = p2 % intp1
	// default:
	// 	return fmt.Errorf("[wvm:opRemainder] type (%v) not supported for reg[p1] (%v)", reflect.TypeOf(p1), p1)
	// }
	// //wvm.register[op.P3] = wvm.register[op.P2].(int) % wvm.register[op.P1].(int)
	// return nil
}

/* Opcode: Move P1 P2 P3 * *
** Synopsis: r[P2@P3]=r[P1@P3]
**
** Move the P3 values in register P1..P1+P3-1 over into
** registers P2..P2+P3-1.  Registers P1..P1+P3-1 are
** left holding a NULL.  It is an error for register ranges
** P1..P1+P3-1 and P2..P2+P3-1 to overlap.  It is an error
** for P3 to be less than 1.
 */
func opMove(wvm *WVM, op Op) (err error) {
	for i := op.P2; i < op.P2+op.P3; i++ {
		wvm.register[i] = wvm.register[i-op.P2+op.P1]
		wvm.register[i-op.P2+op.P1] = nil
	}
	return nil
}

/* Opcode: Copy P1 P2 P3 * *
** Synopsis: r[P2@P3+1]=r[P1@P3+1]
**
** Make a copy of registers P1..P1+P3 into registers P2..P2+P3.
**
** This instruction makes a deep copy of the value.  A duplicate
** is made of any string or blob constant.  See also OP_SCopy.
 */
func opCopy(wvm *WVM, op Op) (err error) {
	for i := op.P1; i <= op.P1+op.P3; i++ {
		wvm.register[i-op.P1+op.P2] = wvm.register[i]
	}
	//	 copied: r5 => r1 (value <nil>)

	return nil
}

/* Opcode: Eq P1 P2 P3 P4 P5
** Synopsis: IF r[P3]==r[P1]
**
** Compare the values in register P1 and P3.  If reg(P3)==reg(P1) then
** jump to address P2.  Or if the SQLITE_STOREP2 flag is set in P5, then
** store the result of comparison in register P2.
**
** The SQLITE_AFF_MASK portion of P5 must be an affinity character -
** SQLITE_AFF_TEXT, SQLITE_AFF_INTEGER, and so forth. An attempt is made
** to coerce both inputs according to this affinity before the
** comparison is made. If the SQLITE_AFF_MASK is 0x00, then numeric
** affinity is used. Note that the affinity conversions are stored
** back into the input registers P1 and P3.  So this opcode can cause
** persistent changes to registers P1 and P3.
**
** Once any conversions have taken place, and neither value is NULL,
** the values are compared. If both values are blobs then memcmp() is
** used to determine the results of the comparison.  If both values
** are text, then the appropriate collating function specified in
** P4 is used to do the comparison.  If P4 is not specified then
** memcmp() is used to compare text string.  If both values are
** numeric, then a numeric comparison is used. If the two values
** are of different types, then numbers are considered less than
** strings and strings are considered less than blobs.
**
** If SQLITE_NULLEQ is set in P5 then the result of comparison is always either
** true or false and is never NULL.  If both operands are NULL then the result
** of comparison is true.  If either operand is NULL then the result is false.
** If neither operand is NULL the result is the same as it would be if
** the SQLITE_NULLEQ flag were omitted from P5.
**
** If both SQLITE_STOREP2 and SQLITE_KEEPNULL flags are set then the
** content of r[P2] is only changed if the new value is NULL or 0 (false).
** In other words, a prior r[P2] value will not be overwritten by 1 (true).
 */
func opEq(wvm *WVM, op Op) (err error) {
	if wvm.register[op.P1] == wvm.register[op.P3] {
		wvm.pc = op.P2
		wvm.jumped = true
	}
	return nil
}

/* Opcode: Ne P1 P2 P3 P4 P5
** Synopsis: IF r[P3]!=r[P1]
**
** This works just like the Eq opcode except that the jump is taken if
** the operands in registers P1 and P3 are not equal.  See the Eq opcode for
** additional information.
**
** If both SQLITE_STOREP2 and SQLITE_KEEPNULL flags are set then the
** content of r[P2] is only changed if the new value is NULL or 1 (true).
** In other words, a prior r[P2] value will not be overwritten by 0 (false).
 */

func opNe(wvm *WVM, op Op) (err error) {

	ok, err := wvm.registerCheck(op.P1, op.P3)
	if err != nil {
		return fmt.Errorf("[wvm:opNe] %s", err)
	}
	if !ok {
		log.Error("[wvm:opNe] op.P1 or op.P3 is nil - item in query is not found")
		wvm.halted = true
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opNe] op.P1 %s", err)
	}
	f3, s3, err := convertNumOrString(wvm.register[op.P3])
	if err != nil {
		return fmt.Errorf("[wvm:opNe] op.P3 %s", err)
	}

	ne := true
	if s1 != nil && s3 != nil {
		if *s1 == *s3 {
			ne = false
		}
	}
	if f1 != nil && f3 != nil {
		if *f1 == *f3 {
			ne = false
		}
	}
	if ne {
		wvm.pc = op.P2
		wvm.jumped = true
	}
	return nil

}

/* Opcode: Lt P1 P2 P3 P4 P5
** Synopsis: IF r[P3]<r[P1]
**
** Compare the values in register P1 and P3.  If reg(P3)<reg(P1) then
** jump to address P2.  Or if the SQLITE_STOREP2 flag is set in P5 store
** the result of comparison (0 or 1 or NULL) into register P2.
**
** If the SQLITE_JUMPIFNULL bit of P5 is set and either reg(P1) or
** reg(P3) is NULL then the take the jump.  If the SQLITE_JUMPIFNULL
** bit is clear then fall through if either operand is NULL.
**
** The SQLITE_AFF_MASK portion of P5 must be an affinity character -
** SQLITE_AFF_TEXT, SQLITE_AFF_INTEGER, and so forth. An attempt is made
** to coerce both inputs according to this affinity before the
** comparison is made. If the SQLITE_AFF_MASK is 0x00, then numeric
** affinity is used. Note that the affinity conversions are stored
** back into the input registers P1 and P3.  So this opcode can cause
** persistent changes to registers P1 and P3.
**
** Once any conversions have taken place, and neither value is NULL,
** the values are compared. If both values are blobs then memcmp() is
** used to determine the results of the comparison.  If both values
** are text, then the appropriate collating function specified in
** P4 is  used to do the comparison.  If P4 is not specified then
** memcmp() is used to compare text string.  If both values are
** numeric, then a numeric comparison is used. If the two values
** are of different types, then numbers are considered less than
** strings and strings are considered less than blobs.
 */
func opLt(wvm *WVM, op Op) (err error) {
	ok, err := wvm.registerCheck(op.P1, op.P3)
	if err != nil {
		return fmt.Errorf("[wvm:opLt] %s", err)
	}
	if !ok {
		log.Error("[wvm:opLt] op.P1 or op.P3 is nil - item in query is not found")
		wvm.halted = true // ?
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opLt] op.P1 %s", err)
	}
	f3, s3, err := convertNumOrString(wvm.register[op.P3])
	if err != nil {
		return fmt.Errorf("[wvm:opLt] op.P3 %s", err)
	}

	// TODO: not implemented
	if s1 != nil && s3 != nil {
		return fmt.Errorf("[wvm:opLt] not implemented for strings")
	}

	lt := false
	if f3 != nil && f1 != nil {
		if *f3 < *f1 {
			lt = true
		}
	}
	// numbers are considered less than strings
	if f3 != nil && s1 != nil { // f3 < s1
		lt = true
	}
	if lt {
		wvm.pc = op.P2
		wvm.jumped = true
	}

	// if wvm.register[op.P3] < wvm.register[op.P1] {
	// 	wvm.pc = op.P2
	// 	wvm.jumped = true
	// }

	return nil
}

// converts register interface to float or string
// uses pointers so can use nil pointer checks
func convertNumOrString(reg interface{}) (res *float64, str *string, err error) {

	if reg == nil {
		return res, str, nil
	}

	switch r := reg.(type) {
	case int:
		res_temp := float64(r)
		res = &res_temp
		break
	case int64:
		res_temp := float64(r)
		res = &res_temp
		break
	case float64:
		res = &r
		break
	case string:
		str = &r
		break
	default:
		return res, str, fmt.Errorf("[wvm:floatConvert] type (%v) not supported for reg op (%v)", reflect.TypeOf(reg), reg)
	}

	return res, str, nil
}

func (wvm *WVM) dumpRegisters() {
	for i := 0; i < 128; i++ {
		if wvm.register[i] != nil {
			fmt.Printf(" dump -  R%d: %v\n", i, wvm.register[i])
		}
	}
}

/* Opcode: Le P1 P2 P3 P4 P5
** Synopsis: IF r[P3]<=r[P1]
**
** This works just like the Lt opcode except that the jump is taken if
** the content of register P3 is less than or equal to the content of
** register P1.  See the Lt opcode for additional information.
 */
func opLe(wvm *WVM, op Op) (err error) {

	ok, err := wvm.registerCheck(op.P1, op.P3)
	if err != nil {
		return fmt.Errorf("[wvm:opLe] %s", err)
	}
	if !ok {
		log.Error("[wvm:opNe] op.P1 or op.P3 is nil - item in query is not found")
		wvm.halted = true // ?
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opLe] op.P1 %s", err)
	}
	f3, s3, err := convertNumOrString(wvm.register[op.P3])
	if err != nil {
		return fmt.Errorf("[wvm:opLe] op.P3 %s", err)
	}

	// TODO: not implemented
	if s1 != nil && s3 != nil {
		return fmt.Errorf("[wvm:opLe] not implemented for strings")
	}

	le := false
	if f3 != nil && f1 != nil {
		if *f3 <= *f1 {
			le = true
		}
	}
	// numbers are considered less than strings
	if f3 != nil && s1 != nil { // f3 < s1
		le = true
	}
	if le {
		wvm.pc = op.P2
		wvm.jumped = true
	}
	return nil

	// if wvm.register[op.P3].(float64) <= r1 {
	// 	wvm.pc = op.P2
	// 	wvm.jumped = true
	// }

}

/* Opcode: Gt P1 P2 P3 P4 P5
** Synopsis: IF r[P3]>r[P1]
**
** This works just like the Lt opcode except that the jump is taken if
** the content of register P3 is greater than the content of
** register P1.  See the Lt opcode for additional information.
 */
func opGt(wvm *WVM, op Op) (err error) {
	ok, err := wvm.registerCheck(op.P1, op.P3)
	if err != nil {
		return fmt.Errorf("[wvm:opGt] %s", err)
	}
	if !ok {
		log.Error("[wvm:opGt] op.P1 or op.P3 is nil - item in query is not found")
		wvm.halted = true // ?
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opGt] op.P1 %s", err)
	}
	f3, s3, err := convertNumOrString(wvm.register[op.P3])
	if err != nil {
		return fmt.Errorf("[wvm:opGt] op.P3 %s", err)
	}
	// TODO: not implemented
	if s1 != nil && s3 != nil {
		return fmt.Errorf("[wvm:opGt] not implemented for strings")
	}

	gt := false
	if f3 != nil && f1 != nil {
		if *f3 > *f1 {
			gt = true
		}
	}
	// numbers are considered less than strings
	if s3 != nil && f1 != nil { // s3 > f1
		gt = true
	}
	if gt {
		wvm.pc = op.P2
		wvm.jumped = true
	}

	// if wvm.register[op.P3].(int) > wvm.register[op.P1].(int) {
	// 	wvm.pc = op.P2
	// 	wvm.jumped = true
	// }
	return nil
}

/* Opcode: Ge P1 P2 P3 P4 P5
** Synopsis: IF r[P3]>=r[P1]
**
** This works just like the Lt opcode except that the jump is taken if
** the content of register P3 is greater than or equal to the content of
** register P1.  See the Lt opcode for additional information.
 */
func opGe(wvm *WVM, op Op) (err error) {

	ok, err := wvm.registerCheck(op.P1, op.P3)
	if err != nil {
		return fmt.Errorf("[wvm:opGe] %s", err)
	}
	if !ok {
		log.Error("[wvm:opGe] op.P1 or op.P3 is nil - item in query is not found")
		wvm.halted = true // ?
		return nil
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opGe] op.P1 %s", err)
	}
	f3, s3, err := convertNumOrString(wvm.register[op.P3])
	if err != nil {
		return fmt.Errorf("[wvm:opGe] op.P3 %s", err)
	}
	// TODO: not implemented
	if s1 != nil && s3 != nil {
		return fmt.Errorf("[wvm:opGe] not implemented for strings")
	}

	ge := false
	if f3 != nil && f1 != nil {
		if *f3 >= *f1 {
			ge = true
		}
	}
	// numbers are considered less than strings
	if s3 != nil && f1 != nil { // s3 > f1
		ge = true
	}
	if ge {
		wvm.pc = op.P2
		wvm.jumped = true
	}

	// if wvm.register[op.P3].(float64) >= wvm.register[op.P1].(float64) {
	// 	wvm.pc = op.P2
	// 	wvm.jumped = true
	// }
	return nil
}

/* Opcode: Compare P1 P2 P3 P4 P5
** Synopsis: r[P1@P3] <-> r[P2@P3]
**
** Compare two vectors of registers in reg(P1)..reg(P1+P3-1) (call this
** vector "A") and in reg(P2)..reg(P2+P3-1) ("B").  Save the result of
** the comparison for use by the next OP_Jump instruct.
**
** If P5 has the OPFLAG_PERMUTE bit set, then the order of comparison is
** determined by the most recent OP_Permutation operator.  If the
** OPFLAG_PERMUTE bit is clear, then register are compared in sequential
** order.
**
** P4 is a KeyInfo structure that defines collating sequences and sort
** orders for the comparison.  The permutation applies to registers
** only.  The KeyInfo elements are used sequentially.
**
** The comparison is a sort comparison, so NULLs compare equal,
** NULLs are less than numbers, numbers are less than strings,
** and strings are less than blobs.
 */
func opCompare(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("Compare not implemented")
}

/* Opcode: Jump P1 P2 P3 * *
**
** Jump to the instruction at address P1, P2, or P3 depending on whether
** in the most recent OP_Compare instruction the P1 vector was less than
** equal to, or greater than the P2 vector, respectively.
 */
func opJump(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("Jump not implemented")
}

/* Opcode: And P1 P2 P3 * *
** Synopsis: r[P3]=(r[P1] && r[P2])
**
** Take the logical AND of the values in registers P1 and P2 and
** write the result into register P3.
**
** If either P1 or P2 is 0 (false) then the result is 0 even if
** the other input is NULL.  A NULL and true or two NULLs give
** a NULL output.
 */
func opAnd(wvm *WVM, op Op) (err error) {

	_, err = wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opAnd] %s", err)
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opAnd] op.P1 %s", err)
	}
	f2, s2, err := convertNumOrString(wvm.register[op.P2])
	if err != nil {
		return fmt.Errorf("[wvm:opAnd] op.P2 %s", err)
	}
	wvm.register[op.P3] = nil
	if f1 != nil && f2 != nil {
		if *f1 != 0 && *f2 != 0 {
			wvm.register[op.P3] = 1
		} else {
			wvm.register[op.P3] = 0
		}
		return nil
	}
	if s1 != nil || s2 != nil {
		return fmt.Errorf("[wvm:opAnd] not implemented for strings")
	}
	if f1 != nil && *f1 == 0 {
		wvm.register[op.P3] = 0
	}
	if f2 != nil && *f2 == 0 {
		wvm.register[op.P3] = 0
	}

	// if wvm.register[op.P1].(int) > 0 && wvm.register[op.P2].(int) > 0 {
	// 	wvm.register[op.P3] = 1
	// } else {
	// 	wvm.register[op.P3] = 0
	// }
	return nil
}

/* Opcode: BitAnd P1 P2 P3 * *
** Synopsis: r[P3]=r[P1]&r[P2]
**
** Take the bit-wise AND of the values in register P1 and P2 and
** store the result in register P3.
** If either input is NULL, the result is NULL.
 */
func opBitAnd(wvm *WVM, op Op) (err error) {
	ok, err := wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opBitAnd] %s", err)
	}
	wvm.register[op.P3] = nil
	if !ok {
		// one of the registers had nil
		return nil
	}
	r1, ok1 := wvm.register[op.P1].(int)
	r2, ok2 := wvm.register[op.P2].(int)
	if !ok1 || !ok2 {
		log.Error("[wvm:opBitAnd] cannot BitAnd non-ints. Continuing.", "r1", r1, "r2", r2)
		// should this be a halt? or real error?
		return nil
	}
	wvm.register[op.P3] = r1 & r2
	return nil
	//wvm.register[op.P3] = wvm.register[op.P1].(int) & wvm.register[op.P2].(int)
}

/* Opcode: Or P1 P2 P3 * *
** Synopsis: r[P3]=(r[P1] || r[P2])
**
** Take the logical OR of the values in register P1 and P2 and
** store the answer in register P3.
**
** If either P1 or P2 is nonzero (true) then the result is 1 (true)
** even if the other input is NULL.  A NULL and false or two NULLs
** give a NULL output.
 */
func opOr(wvm *WVM, op Op) (err error) {

	_, err = wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opOr] %s", err)
	}
	f1, s1, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opOr] op.P1 %s", err)
	}
	f2, s2, err := convertNumOrString(wvm.register[op.P2])
	if err != nil {
		return fmt.Errorf("[wvm:opOr] op.P2 %s", err)
	}
	wvm.register[op.P3] = nil
	if f1 != nil && f2 != nil {
		if *f1 != 0 || *f2 != 0 {
			wvm.register[op.P3] = 1
		} else {
			wvm.register[op.P3] = 0
		}
		return nil
	}
	if s1 != nil || s2 != nil {
		return fmt.Errorf("[wvm:opOr] not implemented for strings")
	}
	if f1 != nil && *f1 != 0 {
		wvm.register[op.P3] = 1
	}
	if f2 != nil && *f2 != 0 {
		wvm.register[op.P3] = 1
	}

	// if wvm.register[op.P1].(int) > 0 || wvm.register[op.P2].(int) > 0 {
	// 	wvm.register[op.P3] = 1
	// } else {
	// 	wvm.register[op.P3] = 0
	// }
	return nil
}

/* Opcode: BitOr P1 P2 P3 * *
** Synopsis: r[P3]=r[P1]|r[P2]
**
** Take the bit-wise OR of the values in register P1 and P2 and
** store the result in register P3.
** If either input is NULL, the result is NULL.
 */
func opBitOr(wvm *WVM, op Op) (err error) {
	ok, err := wvm.registerCheck(op.P1, op.P2)
	if err != nil {
		return fmt.Errorf("[wvm:opBitOr] %s", err)
	}
	wvm.register[op.P3] = nil
	if !ok {
		// one of the registers had nil
		return nil
	}
	r1, ok1 := wvm.register[op.P1].(int)
	r2, ok2 := wvm.register[op.P2].(int)
	if !ok1 || !ok2 {
		log.Error("[wvm:opBitOr] cannot BitOr non-ints. Continuing.", "r1", r1, "r2", r2)
		// should this be a halt? or real error?
		return nil
	}
	wvm.register[op.P3] = r1 | r2
	//wvm.register[op.P3] = wvm.register[op.P1].(int) | wvm.register[op.P2].(int)
	return nil
}

/* Opcode: Not P1 P2 * * *
** Synopsis: r[P2]= !r[P1]
**
** Interpret the value in register P1 as a boolean value.  Store the
** boolean complement in register P2.  If the value in register P1 is
** NULL, then a NULL is stored in P2.
 */
func opNot(wvm *WVM, op Op) (err error) {
	r1, ok := wvm.register[op.P1]
	if !ok {
		return fmt.Errorf("[wvm:opNot] key (%v) not found in registers (%v) - item in query is not found", op.P1, wvm.register)
	}
	if r1 == nil {
		wvm.register[op.P2] = nil
		return nil
	}
	wvm.register[op.P2] = ^(wvm.register[op.P1].(int))
	return nil
}

/* Opcode: BitNot P1 P2 * * *
** Synopsis: r[P1]= ~r[P1]
**
** Interpret the content of register P1 as an integer.  Store the
** ones-complement of the P1 value into register P2.  If P1 holds
** a NULL then store a NULL in P2.
 */
func opBitNot(wvm *WVM, op Op) (err error) {
	r1, ok := wvm.register[op.P1]
	if !ok {
		return fmt.Errorf("[wvm:opBitNot] key (%v) not found in registers (%v) - item in query is not found", op.P1, wvm.register)
	}
	if r1 == nil {
		wvm.register[op.P2] = nil
		return nil
	}
	wvm.register[op.P1] = ^(wvm.register[op.P1].(int))
	return nil
}

/* Opcode: If P1 P2 P3 * *
**
** Jump to P2 if the value in register P1 is true.  The value
** is considered true if it is numeric and non-zero.  If the value
** in P1 is NULL then take the jump if and only if P3 is non-zero.
 */
func opIf(wvm *WVM, op Op) (err error) {
	r1, ok := wvm.register[op.P1]
	if !ok {
		return fmt.Errorf("[wvm:opIf] op.P1 key (%v) not found in registers (%v) - item in query is not found", op.P1, wvm.register)
	}
	if r1 == nil {
		r3, ok := wvm.register[op.P3]
		if !ok {
			return fmt.Errorf("[wvm:opIf] op.P3 key (%v) not found in registers (%v) - item in query is not found", op.P3, wvm.register)
		}
		if r3.(int) != 0 { // this ok if r3 is a string?
			wvm.pc = op.P2
			wvm.jumped = true
			return nil
		}
	}
	f1, s1, err := convertNumOrString(r1)
	if err != nil {
		return fmt.Errorf("[wvm:opIf] %s", err)
	}
	if s1 != nil {
		return nil //needs to be numeric, not a string
	}
	if f1 != nil && *f1 != 0 {
		wvm.pc = op.P2
		wvm.jumped = true
	}
	return nil
}

/* Opcode: IfNot P1 P2 P3 * *
**
** Jump to P2 if the value in register P1 is False.  The value
** is considered false if it has a numeric value of zero.  If the value
** in P1 is NULL then take the jump if and only if P3 is non-zero.
 */
func opIfNot(wvm *WVM, op Op) (err error) {
	// if wvm.register[op.P1].(int) == 0 {
	// 	wvm.pc = op.P2
	// 	wvm.jumped = true
	// }

	r1, ok := wvm.register[op.P1]
	if !ok {
		return fmt.Errorf("[wvm:opIf] op.P1 key (%v) not found in wvm.register map - item in query is not found", op.P1)
	}
	if r1 == nil {
		r3, ok := wvm.register[op.P3]
		if !ok {
			return fmt.Errorf("[wvm:opIf] op.P3 key (%v) not found in wvm.register map - item in query is not found", op.P3)
		}
		if r3.(int) != 0 { // this ok if r3 is a string?
			wvm.pc = op.P2
			wvm.jumped = true
			return nil
		}
	}
	f1, s1, err := convertNumOrString(r1)
	if err != nil {
		return fmt.Errorf("[wvm:opIfNot] %s", err)
	}
	if s1 != nil {
		return nil //needs to be numeric, not a string
	}
	if f1 != nil && *f1 == 0 {
		wvm.pc = op.P2
		wvm.jumped = true
	}
	return nil
}

/* Opcode: IfPos P1 P2 P3 * *
** Synopsis: if r[P1]>0 then r[P1]-=P3, goto P2
**
** Register P1 must contain an integer.
** If the value of register P1 is 1 or greater, subtract P3 from the
** value in P1 and jump to P2.
**
** If the initial value of register P1 is less than 1, then the
** value is unchanged and control passes through to the next instruction.
 */
func opIfPos(wvm *WVM, op Op) (err error) {
	_, ok := wvm.register[op.P1]
	if !ok {
		return fmt.Errorf("[wvm:opIfPos] key op.P1 (%v) not found in wvm.register map - item in query not found", op.P1)
	}
	r1, ok := wvm.register[op.P1].(int)
	if !ok { // what should happen if r1 doesn't have an integer?
		return fmt.Errorf("[wvm:opIfPos] r1 didn't have an integer (%v)", r1)
	}
	if r1 >= 1 {
		r3, ok := wvm.register[op.P3].(int)
		if !ok { // what should happen if r3 isn't an integer?
			return fmt.Errorf("[wvm:opIfPos] r3 is not an integer (%v)", r3)
		}
		wvm.register[op.P1] = r1 - r3
		wvm.pc = op.P2
		wvm.jumped = true
		return nil
	}

	return nil

	// if wvm.register[op.P1].(int) > 0 {
	// 	wvm.register[op.P1] = wvm.register[op.P1].(int) - op.P3
	// }
}

/* Opcode: Count P1 P2 * * *
** Synopsis: r[P2]=count()
**
** Store the number of entries (an integer value) in the table or index
** opened by cursor P1 in register P2
 */
func opCount(wvm *WVM, op Op) (err error) {

	index := wvm.index[op.P1]

	cursor, err := index.SeekFirst()
	if err != nil { //or should this rtn an err? TODO check this on empty Btree
		wvm.register[op.P2] = 0
		return nil
	}
	key, _, _ := cursor.Next()
	//fmt.Printf("[wvm:opCount] cursor: %+v\n", cursor)
	count := 1

	cursorEnd, err := index.SeekLast()
	if err != nil {
		return fmt.Errorf("[wvm:opCount] %s", err)
	}
	endkey, _, _ := cursorEnd.Next()
	//fmt.Printf("[wvm:opCount] cursorEnd: %+v\n", cursorEnd)

	for !bytes.Equal(key, endkey) {
		//fmt.Printf("[wvm:opCount] next cursor: %+v\n", cursor)
		key, _, _ = cursor.Next()
		count++
		if count > 500 { //TODO: should be a const somewhere, just a safeguard
			//wvm.eof = true
			//wvm.halted = true
			return fmt.Errorf("[wvm:opCount] over 500 ct!")
		}
	}
	wvm.register[op.P2] = count
	return nil
}

/* Opcode: NotExists P1 P2 P3 * *
** Synopsis: intkey=r[P3]
**
** P1 is the index of a cursor open on an SQL table btree (with integer
** keys).  P3 is an integer rowid.  If P1 does not contain a record with
** rowid P3 then jump immediately to P2.  Or, if P2 is 0, raise an
** SQLITE_CORRUPT error. If P1 does contain a record with rowid P3 then
** leave the cursor pointing at that record and fall through to the next
** instruction.
**
** The OP_SeekRowid opcode performs the same operation but also allows the
** P3 register to contain a non-integer value, in which case the jump is
** always taken.  This opcode requires that P3 always contain an integer.
**
** The OP_NotFound opcode performs the same operation on index btrees
** (with arbitrary multi-value keys).
**
** This opcode leaves the cursor in a state where it cannot be advanced
** in either direction.  In other words, the Next and Prev opcodes will
** not work following this opcode.
**
** See also: Found, NotFound, NoConflict, SeekRowid
 */
func opNotExists(wvm *WVM, op Op) (err error) {
	//return fmt.Errorf("NotExists not implemented")
	//log.Info("[wvm:opNotExists]", "activeKey", hex.EncodeToString(wvm.activeKey), "activeRow", wvm.activeRow, "op.P3", op.P3, "cursor of op.P3", wvm.cursors[op.P3])
	cursor, ok, err := wvm.index[op.P1].Seek(wvm.activeKey) //used activeKey instead of Rowid...
	if err != nil {
		return fmt.Errorf("[wvm:opNotExists] %s", err)
	}
	if !ok {
		if op.P2 == 0 {
			return fmt.Errorf("[wvm:opNotExists] op P2 was nil, SQLITE_CORRUPT")
		}
		wvm.pc = op.P2
		wvm.jumped = true
		return nil
	}
	wvm.cursors[op.P1] = cursor
	return nil
}

/* Opcode: Sequence P1 P2 * * *
** Synopsis: r[P2]=cursor[P1].ctr++
**
** Find the next available sequence number for cursor P1.
** Write the sequence number into register P2.
** The sequence number on the cursor is incremented after this
** instruction.
 */
func opSequence(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("[wvm:opSequence] not implemented")
}

/* Opcode: OP_OpenEphemeral
Open a new cursor P1 to a transient table. The cursor is always opened read/write even if the main database is read-only. The ephemeral table is deleted automatically when the cursor is closed.
P2 is the number of columns in the ephemeral table. The cursor points to a BTree table if P4==0 and to a BTree index if P4 is not 0. If P4 is not NULL, it points to a KeyInfo structure that defines the format of keys in the index.

The P5 parameter can be a mask of the BTREE_* flags defined in btree.h. These flags control aspects of the operation of the btree. The BTREE_OMIT_JOURNAL and BTREE_SINGLE flags are added automatically.
*/
func opOpenEphemeral(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("[wvm:opOpenEphemeral] not implemented")
}

/* Opcode: SorterOpen P1 P2 P3 P4 *
**
** This opcode works like OP_OpenEphemeral except that it opens
** a transient index that is specifically designed to sort large
** tables using an external merge-sort algorithm.
**
** If argument P3 is non-zero, then it indicates that the sorter may
** assume that a stable sort considering the first P3 fields of each
** key is sufficient to produce the required results.
 */

func opSorterOpen(wvm *WVM, op Op) (err error) {
	//fmt.Printf("[wvm:SorterOpen] not implemented\n")
	return fmt.Errorf("[wvm:opSorterOpen] not implemented")

	//initalize sorterobj - already done
	//does active cursor need to be set? or first item in the sorter obj?
	//return nil
}

/* Opcode: SorterData P1 P2 P3 * *
** Synopsis: r[P2]=data
**
** Write into register P2 the current sorter data for sorter cursor P1.
** Then clear the column header cache on cursor P3.
**
** This opcode is normally use to move a record out of the sorter and into
** a register that is the source for a pseudo-table cursor created using
** OpenPseudo.  That pseudo-table cursor is the one that is identified by
** parameter P3.  Clearing the P3 column cache as part of this opcode saves
** us from having to issue a separate NullRow instruction to clear that cache.
 */
func opSorterData(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("SorterData not implemented")
}

/* Opcode: SorterSort P1 P2 * * *
**
** After all records have been inserted into the Sorter object
** identified by P1, invoke this opcode to actually do the sorting.
** Jump to P2 if there are no records to be sorted.
**
** This opcode is an alias for OP_Sort and OP_Rewind that is used
** for Sorter objects.
 */
func opSorterSort(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("SorterSort not implemented")
}

/* Opcode: Sort P1 P2 * * *
**
** This opcode does exactly the same thing as OP_Rewind except that
** it increments an undocumented global variable used for testing.
**
** Sorting is accomplished by writing records into a sorting index,
** then rewinding that index and playing it back from beginning to
** end.  We use the OP_Sort opcode instead of OP_Rewind to do the
** rewinding so that the global variable will be incremented and
** regression tests can determine whether or not the optimizer is
** correctly optimizing out sorts.
 */
func opSort(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("[wvm:opSort] not implemented")
}

/* Opcode: SorterNext P1 P2 * * P5
**
** This opcode works just like OP_Next except that P1 must be a
** sorter object for which the OP_SorterSort opcode has been
** invoked.  This opcode advances the cursor to the next sorted
** record, or jumps to P2 if there are no more sorted records.
 */
func opSorterNext(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("SorterNext not implemented")
}

/* Opcode: SorterInsert P1 P2 * * *
** Synopsis: key=r[P2]
**
** Register P2 holds an SQL index key made using the
** MakeRecord instructions.  This opcode writes that key
** into the sorter P1.  Data for the entry is nil.
 */
func opSorterInsert(wvm *WVM, op Op) (err error) {
	//var srow SorterRow

	//srow.keyToSort = wvm.activeKey
	//append(wvm.sorter, srow)
	//return fmt.Errorf("SorterInsert not implemented")
	return nil
}

/* Opcode: RowSetAdd P1 P2 * * *
** Synopsis: rowset(P1)=r[P2]
**
** Insert the integer value held by register P2 into a RowSet object
** held in register P1.
**
** An assertion fails if P2 is not an integer.
 */
func opRowSetAdd(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("RowSetAdd not implemented")
}

/* Opcode: RowSetRead P1 P2 P3 * *
** Synopsis: r[P3]=rowset(P1)
**
** Extract the smallest value from the RowSet object in P1
** and put that value into register P3.
** Or, if RowSet object P1 is initially empty, leave P3
** unchanged and jump to instruction P2.
 */
func opRowSetRead(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("RowSetRead not implemented")
}

/* Opcode: RowSetTest P1 P2 P3 P4
** Synopsis: if r[P3] in rowset(P1) goto P2
**
** Register P3 is assumed to hold a 64-bit integer value. If register P1
** contains a RowSet object and that RowSet object contains
** the value held in P3, jump to register P2. Otherwise, insert the
** integer in P3 into the RowSet and continue on to the
** next opcode.
**
** The RowSet object is optimized for the case where sets of integers
** are inserted in distinct phases, which each set contains no duplicates.
** Each set is identified by a unique P4 value. The first set
** must have P4==0, the final set must have P4==-1, and for all other sets
** must have P4>0.
**
** This allows optimizations: (a) when P4==0 there is no need to test
** the RowSet object for P3, as it is guaranteed not to contain it,
** (b) when P4==-1 there is no need to insert the value, as it will
** never be tested for, and (c) when a value that is part of set X is
** inserted, there is no need to search to see if the same value was
** previously inserted as part of set X (only if it was previously
** inserted as part of some other set).
 */
func opRowSetTest(wvm *WVM, op Op) (err error) {

	rowid := wvm.register[op.P3].(int64) //assumes that value of P3 has a 64 bit int

	for _, r := range wvm.rowSet[op.P3] { //shouldn't this be P1's rowset?? not P3?
		if rowid == r {

			//set the row
			var newrow []interface{}
			for i := 0; i < len(wvm.activeRow); i++ {
				newrow = append(newrow, wvm.activeRow[i])
			}
			wvm.Rows = append(wvm.Rows, newrow)
			//fmt.Printf("[wvm:opRowSetTest] rowid %v found in rowset %v, jumping to op.P2 %v\n", rowid, op.P3, op.P2)

			wvm.pc = op.P2
			wvm.jumped = true
			return nil

		}
	}

	if rowid != 0 {
		wvm.rowSet[op.P1] = append(wvm.rowSet[op.P1], rowid)
		//fmt.Printf("[wvm.opRowSetTest] appending rowid %v to rowset %v\n", rowid, op.P1)
	}

	return nil
}

/* Opcode: AggStep0 * P2 P3 P4 P5
** Synopsis: accum=r[P3] step(r[P2@P5])
**
** Execute the step function for an aggregate.  The
** function has P5 arguments.   P4 is a pointer to the FuncDef
** structure that specifies the function.  Register P3 is the
** accumulator.
**
** The P5 arguments are taken from register P2 and its
** successors.
 */

//25: "select sum(person_id) as s from person",
//27: "select min(person_id) as m0, max(person_id) as m1 from person"
func opAggStep0(wvm *WVM, op Op) (err error) {
	//TODO: the function has P5 arguments
	//TODO: other agg cmds besides sum, min, max, count
	aggcmd := op.P4

	if strings.Contains(aggcmd, "sum") {
		if wvm.register[op.P3] == nil {
			wvm.register[op.P3] = 0
		}
		switch wvm.register[op.P2].(type) {
		case int:
			wvm.register[op.P3] = wvm.register[op.P2].(int) + wvm.register[op.P3].(int)
			break
		case int64:
			wvm.register[op.P3] = wvm.register[op.P2].(int64) + wvm.register[op.P3].(int64)
			break
		case uint64:
			wvm.register[op.P3] = wvm.register[op.P2].(uint64) + wvm.register[op.P3].(uint64)
			break
		case float64:
			wvm.register[op.P3] = wvm.register[op.P2].(float64) + wvm.register[op.P3].(float64)
			break
		default:
			return fmt.Errorf("[wvm:opAggStep0] unknown type for Aggregate SUM")
		}

	} else if strings.Contains(aggcmd, "min") {
		if wvm.register[op.P3] == nil {
			wvm.register[op.P3] = wvm.register[op.P2]
			return nil
		}
		switch wvm.register[op.P2].(type) {
		case int:
			if wvm.register[op.P3].(int) > wvm.register[op.P2].(int) {
				wvm.register[op.P3] = wvm.register[op.P2]
			}
			break
		case int64:
			if wvm.register[op.P3].(int64) > wvm.register[op.P2].(int64) {
				wvm.register[op.P3] = wvm.register[op.P2]
			}
			break
		case uint64:
			if wvm.register[op.P3].(uint64) > wvm.register[op.P2].(uint64) {
				wvm.register[op.P3] = wvm.register[op.P2]
			}
			break
		case float64:
			if wvm.register[op.P3].(float64) > wvm.register[op.P2].(float64) {
				wvm.register[op.P3] = wvm.register[op.P2]
			}
			break
		default:
			return fmt.Errorf("[wvm:opAggStep0] unknown type for Aggregate MIN")
		}

	} else if strings.Contains(aggcmd, "max") {
		if wvm.register[op.P3] == nil {
			wvm.register[op.P3] = wvm.register[op.P2]
			return nil
		}
		switch wvm.register[op.P2].(type) {
		case int:
			if wvm.register[op.P3].(int) < wvm.register[op.P2].(int) {
				wvm.register[op.P3] = wvm.register[op.P2]
			}
			break
		case int64:
			if wvm.register[op.P3].(int64) < wvm.register[op.P2].(int64) {
				wvm.register[op.P3] = wvm.register[op.P2]
			}
			break
		case uint64:
			if wvm.register[op.P3].(uint64) < wvm.register[op.P2].(uint64) {
				wvm.register[op.P3] = wvm.register[op.P2]
			}
			break
		case float64:
			if wvm.register[op.P3].(float64) < wvm.register[op.P2].(float64) {
				wvm.register[op.P3] = wvm.register[op.P2]
			}
			break
		default:
			return fmt.Errorf("[wvm:opAggStep0] unknown type for Aggregate MIN")
		}

	} else if strings.Contains(aggcmd, "count") {
		if wvm.register[op.P3] == nil {
			wvm.register[op.P3] = 0
		}
		wvm.register[op.P3] = wvm.register[op.P3].(int) + 1

	} else {
		return fmt.Errorf("[wvm:opAggStep0] Unimplemented Aggregate %v", aggcmd)
	}
	return nil
}

/* Opcode: AggStep * P2 P3 P4 P5
** Synopsis: accum=r[P3] step(r[P2@P5])
**
** Execute the step function for an aggregate.  The
** function has P5 arguments.   P4 is a pointer to an sqlite3_context
** object that is used to run the function.  Register P3 is
** as the accumulator.
**
** The P5 arguments are taken from register P2 and its
** successors.
**
** This opcode is initially coded as OP_AggStep0.  On first evaluation,
** the FuncDef stored in P4 is converted into an sqlite3_context and
** the opcode is changed.  In this way, the initialization of the
** sqlite3_context only happens once, instead of on each call to the
** step function.
 */
func opAggStep(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("AggStep not implemented")
}

/* Opcode: AggFinal P1 P2 * P4 *
** Synopsis: accum=r[P1] N=P2
**
** Execute the finalizer function for an aggregate.  P1 is
** the memory location that is the accumulator for the aggregate.
**
** P2 is the number of arguments that the step function takes and
** P4 is a pointer to the FuncDef for this function.  The P2
** argument is not used by this opcode.  It is only there to disambiguate
** functions that can take varying numbers of arguments.  The
** P4 argument is only needed for the degenerate case where
** the step function was not previously called.
 */
func opAggFinal(wvm *WVM, op Op) (err error) {
	wvm.eof = true //is this needed?
	return nil
	//return fmt.Errorf("AggFinal not implemented")
}

/* Opcode: Function0 P1 P2 P3 P4 P5
** Synopsis: r[P3]=func(r[P2@P5])
**
** Invoke a user function (P4 is a pointer to a FuncDef object that
** defines the function) with P5 arguments taken from register P2 and
** successors.  The result of the function is stored in register P3.
** Register P3 must not be one of the function inputs.
**
** P1 is a 32-bit bitmask indicating whether or not each argument to the
** function was determined to be constant at compile time. If the first
** argument was constant then bit 0 of P1 is set. This is used to determine
** whether meta data associated with a user function argument using the
** sqlite3_set_auxdata() API may be safely retained until the next
** invocation of this opcode.
**
** See also: Function, AggStep, AggFinal
 */
func opFunction0(wvm *WVM, op Op) (err error) {
	//fmt.Printf("[wvm:opFunction0] op: %+v\n", op)
	_, ok := wvm.register[op.P2]
	if !ok {
		// ? maybe should just be halt
		return nil
	}
	_, ok = wvm.register[op.P2].(string)
	if !ok {
		return fmt.Errorf("[wvm:opFunction0] for non-strings not implemented")
	}

	cmd := op.P4
	if strings.Contains(cmd, "length") {
		wvm.register[op.P3] = len(wvm.register[op.P2].(string))
	} else if strings.Contains(cmd, "substr") {
		//The substr(X,Y,Z) function returns a substring of input string X that begins with the Y-th character and which is Z characters long. If Z is omitted then substr(X,Y) returns all characters through the end of the string X beginning with the Y-th. The left-most character of X is number 1. If Y is negative then the first character of the substring is found by counting from the right rather than the left. If Z is negative then the abs(Z) characters preceding the Y-th character are returned. If X is a string then characters indices refer to actual UTF-8 characters. If X is a BLOB then the indices refer to bytes.
		splitleft := strings.Split(cmd, `(`)
		if len(splitleft) < 2 {
			return fmt.Errorf("[wvm:opFunction0] this form of substr, op: %s not implemented", cmd)
		}
		splitright := strings.Split(splitleft[1], `)`)
		if strings.Contains(splitright[0], `,`) {
			return fmt.Errorf("[wvm:opFunction0] substr(X,Y,Z) not implemented")
		}
		pos, err := strconv.Atoi(splitright[0])
		if err != nil {
			return fmt.Errorf("[wvm:opFunction0] %s", err)
		}
		str := wvm.register[op.P2].(string)
		wvm.register[op.P3] = str[pos:]

	} else if strings.Contains(cmd, "lower") {
		str := wvm.register[op.P2].(string)
		wvm.register[op.P3] = strings.ToLower(str)

	} else if strings.Contains(cmd, "upper") {
		str := wvm.register[op.P2].(string)
		wvm.register[op.P3] = strings.ToUpper(str)

	} else if strings.Contains(cmd, "like") {
		str := wvm.register[op.P2].(string) //"%ie"
		temp_str := strings.Replace(str, "%", "*", -1)
		temp_str = strings.Replace(temp_str, "_", "?", -1)
		ok, err := fp.Match(temp_str, wvm.activeRow[op.P3].(string))
		if err != nil {
			return fmt.Errorf("[wvm:opFunction0] %s", err)
		}
		if ok {
			wvm.register[op.P3] = 1
		} else {
			wvm.register[op.P3] = 0
		}

	} else {
		//fmt.Printf("[wvm:opFunction0] TODO: function %s needs to be implemented\n", cmd)
		log.Error(fmt.Sprintf("[wvm:opFunction0] TODO: function %s needs to be implemented\n", cmd))
		//wvm.register[op.P3] = "TODO"
		wvm.halted = true
	}

	//fmt.Printf("[wvm:opFunction0] register op.p2 has (%v), op.p5 = %v, => register[op.P3] has (%v)\n", wvm.register[op.P2], op.P5, wvm.register[op.P3])
	return nil
}

/* Opcode: Function P1 P2 P3 P4 P5
** Synopsis: r[P3]=func(r[P2@P5])
**
** Invoke a user function (P4 is a pointer to an sqlite3_context object that
** contains a pointer to the function to be run) with P5 arguments taken
** from register P2 and successors.  The result of the function is stored
** in register P3.  Register P3 must not be one of the function inputs.
**
** P1 is a 32-bit bitmask indicating whether or not each argument to the
** function was determined to be constant at compile time. If the first
** argument was constant then bit 0 of P1 is set. This is used to determine
** whether meta data associated with a user function argument using the
** sqlite3_set_auxdata() API may be safely retained until the next
** invocation of this opcode.
**
** SQL functions are initially coded as OP_Function0 with P4 pointing
** to a FuncDef object.  But on first evaluation, the P4 operand is
** automatically converted into an sqlite3_context object and the operation
** changed to this OP_Function opcode.  In this way, the initialization of
** the sqlite3_context object occurs only once, rather than once for each
** evaluation of the function.
**
** See also: Function0, AggStep, AggFinal
 */
func opFunction(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("Function not implemented")
}

/* Opcode: CollSeq P1 * * P4
**
** P4 is a pointer to a CollSeq object. If the next call to a user function
** or aggregate calls sqlite3GetFuncCollSeq(), this collation sequence will
** be returned. This is used by the built-in min(), max() and nullif()
** functions.
**
** If P1 is not zero, then it is a register that a subsequent min() or
** max() aggregate will set to 1 if the current row is not the minimum or
** maximum.  The P1 register is initialized to 0 by this instruction.
**
** The interface used by the implementation of the aforementioned functions
** to retrieve the collation sequence set by this opcode is not available
** publicly.  Only built-in functions have access to this feature.
 */
func opCollSeq(wvm *WVM, op Op) (err error) {
	//return fmt.Errorf("[wvm:opCollSeq] not implemented")
	//fmt.Printf("[wvm:opCollSeq] not implemented - AggStep0 following seems to take care of it\n") //TODO check this
	return nil
}

/* Opcode: Affinity P1 P2 * P4 *
** Synopsis: affinity(r[P1@P2])
**
** Apply affinities to a range of P2 registers starting with P1.
**
** P4 is a string that is P2 characters long. The N-th character of the
** string indicates the column affinity that should be used for the N-th
** memory cell in the range.
 */
func opAffinity(wvm *WVM, op Op) (err error) {
	// TODO see https://github.com/r-dbi/RSQLite/blob/master/src/affinity.c#L63
	return nil
}

/* Opcode: RealAffinity P1 * * * *
 *
 * If register P1 holds an integer convert it to a real value.
 *
 * This opcode is used when extracting information from a column that
 * has REAL affinity.  Such column values may still be stored as
 * integers, for space efficiency, but after extraction we want them
 * to have only a real value.
 */
func opRealAffinity(wvm *WVM, op Op) (err error) {
	r1, ok := wvm.register[op.P1]
	if !ok {
		return fmt.Errorf("[wvm:opRealAffinity] key (%v) not found in registers (%v) - item in query is not found", op.P1, wvm.register)
	}
	if r1 == nil {
		log.Error("[wvm:opRealAffinity] register[op.P1] is nil - query item is not found")
		wvm.halted = true
		return nil
	}
	f1, s1, err := convertNumOrString(r1)
	if err != nil {
		return fmt.Errorf("[wvm:opRealAffinity] %s", err)
	}

	if s1 != nil {
		return nil //not an integer
	}
	if f1 != nil {
		wvm.register[op.P1] = *f1
	}

	// switch s := wvm.register[op.P1].(type) {
	// case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
	// 	wvm.register[op.P1] = float64(s.(int))
	// 	break
	// case float64:
	// 	//do nothing, it's right
	// 	break
	// default:
	// 	log.Error(fmt.Sprintf("[wvm:opRealAffinity] Register P1 [%+v] contains unknown type", wvm.register[op.P1]))
	// 	wvm.halted = true
	// 	return nil
	// }
	return nil
}

/* Opcode: IdxDelete P1 P2 P3 * *
 * Synopsis: key=r[P2@P3]
 *
 * The content of P3 registers starting at register P2 form
 * an unpacked index key. This opcode removes that entry from the
 * index opened by cursor P1.
 */
func opIdxDelete(wvm *WVM, op Op) (err error) {
	//fmt.Printf("[wvm:opIdxDelete] not implemented, opDelete seems to do this next\n")
	return nil
}

/* Opcode: IsNull P1 P2 * * *
** Synopsis: if r[P1]==NULL goto P2
**
** Jump to P2 if the value in register P1 is NULL.
 */
func opIsNull(wvm *WVM, op Op) (err error) {
	r1, ok := wvm.register[op.P1]
	if !ok {
		return fmt.Errorf("[wvm:opIsNull] key (%v) not found in registers (%v) - item in query is not found", op.P1, wvm.register)
	}
	if r1 == nil {
		_, ok := wvm.register[op.P2]
		if !ok {
			wvm.halted = true
			return fmt.Errorf("[wvm:opIsNull] key (%v) not found in registers (%v) - item in query is not found", op.P2, wvm.register)
		}
		wvm.pc = wvm.register[op.P2].(int)
		wvm.jumped = true
	}
	return nil
}

/* Opcode: Once P1 P2 * * *
**
** Fall through to the next instruction the first time this opcode is
** encountered on each invocation of the byte-code program.  Jump to P2
** on the second and all subsequent encounters during the same invocation.
**
** Top-level programs determine first invocation by comparing the P1
** operand against the P1 operand on the OP_Init opcode at the beginning
** of the program.  If the P1 values differ, then fall through and make
** the P1 of this opcode equal to the P1 of OP_Init.  If P1 values are
** the same then take the jump.
**
** For subprograms, there is a bitmask in the VdbeFrame that determines
** whether or not the jump should be taken.  The bitmask is necessary
** because the self-altering code trick does not work for recursive
** triggers.
 */
func opOnce(wvm *WVM, op Op) (err error) {
	//fmt.Printf("[wvm:opOnce] wvm.once_marker: %+v\n", wvm.once_marker)
	if wvm.once_marker[wvm.pc] > 0 {
		//fmt.Printf("[wvm:opOnce] once marker is > 0\n")
		wvm.pc = op.P2
		wvm.jumped = true
	} else {
		wvm.once_marker[wvm.pc] = 1
	}
	return nil
}

/* Opcode: OpenPseudo P1 P2 P3 * *
** Synopsis: P3 columns in r[P2]
**
** Open a new cursor that points to a fake table that contains a single
** row of data.  The content of that one row is the content of memory
** register P2.  In other words, cursor P1 becomes an alias for the
** MEM_Blob content contained in register P2.
**
** A pseudo-table created by this opcode is used to hold a single
** row output from the sorter so that the row can be decomposed into
** individual columns using the OP_Column opcode.  The OP_Column opcode
** is the only cursor opcode that works with a pseudo-table.
**
** P3 is the number of fields in the records that will be stored by
** the pseudo-table.
 */
func opOpenPseudo(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("OpenPseudo not implemented")
}

/* Opcode: SCopy P1 P2 * * *
** Synopsis: r[P2]=r[P1]
**
** Make a shallow copy of register P1 into register P2.
**
** This instruction makes a shallow copy of the value.  If the value
** is a string or blob, then the copy is only a pointer to the
** original and hence if the original changes so will the copy.
** Worse, if the original is deallocated, the copy becomes invalid.
** Thus the program must guarantee that the original will not change
** during the lifetime of the copy.  Use OP_Copy to make a complete
** copy.
 */
func opSCopy(wvm *WVM, op Op) (err error) {
	wvm.register[op.P2] = wvm.register[op.P1] // TODO: check if "shallow" copy
	return nil
}

/* Opcode: IntCopy P1 P2 * * *
** Synopsis: r[P2]=r[P1]
**
** Transfer the integer value held in register P1 into register P2.
**
** This is an optimized version of SCopy that works only for integer
** values.
 */
func opIntCopy(wvm *WVM, op Op) (err error) {
	wvm.register[op.P2] = wvm.register[op.P1] // check if int
	return nil
}

/* Opcode: Found P1 P2 P3 P4 *
** Synopsis: key=r[P3@P4]
**
** If P4==0 then register P3 holds a blob constructed by MakeRecord.  If
** P4>0 then register P3 is the first of P4 registers that form an unpacked
** record.
**
** Cursor P1 is on an index btree.  If the record identified by P3 and P4
** is a prefix of any entry in P1 then a jump is made to P2 and
** P1 is left pointing at the matching entry.
**
** This operation leaves the cursor in a state where it can be
** advanced in the forward direction.  The Next instruction will work,
** but not the Prev instruction.
**
** See also: NotFound, NoConflict, NotExists. SeekGe
 */
func opFound(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("Found not implemented")
}

/* Opcode: NotFound P1 P2 P3 P4 *
** Synopsis: key=r[P3@P4]
**
** If P4==0 then register P3 holds a blob constructed by MakeRecord.  If
** P4>0 then register P3 is the first of P4 registers that form an unpacked
** record.
**
** Cursor P1 is on an index btree.  If the record identified by P3 and P4
** is not the prefix of any entry in P1 then a jump is made to P2.  If P1
** does contain an entry whose prefix matches the P3/P4 record then control
** falls through to the next instruction and P1 is left pointing at the
** matching entry.
**
** This operation leaves the cursor in a state where it cannot be
** advanced in either direction.  In other words, the Next and Prev
** opcodes do not work after this operation.
**
** See also: Found, NotExists, NoConflict
 */
func opNotFound(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("NotFound not implemented")
}

/* Opcode: NoConflict P1 P2 P3 P4 *
** Synopsis: key=r[P3@P4]
**
** If P4==0 then register P3 holds a blob constructed by MakeRecord.  If
** P4>0 then register P3 is the first of P4 registers that form an unpacked
** record.
**
** Cursor P1 is on an index btree.  If the record identified by P3 and P4
** contains any NULL value, jump immediately to P2.  If all terms of the
** record are not-NULL then a check is done to determine if any row in the
** P1 index btree has a matching key prefix.  If there are no matches, jump
** immediately to P2.  If there is a match, fall through and leave the P1
** cursor pointing to the matching row.
**
** This opcode is similar to OP_NotFound with the exceptions that the
** branch is always taken if any part of the search key input is NULL.
**
** This operation leaves the cursor in a state where it cannot be
** advanced in either direction.  In other words, the Next and Prev
** opcodes do not work after this operation.
**
** See also: NotFound, Found, NotExists
 */
func opNoConflict(wvm *WVM, op Op) (err error) {
	if wvm.register[op.P3] == nil {
		wvm.jumped = true
		wvm.pc = op.P2
	}
	noconflict := true // TODO: check in P1 Btree for matching key previx
	if noconflict {
		wvm.jumped = true
		wvm.pc = op.P2
	}

	return nil // fmt.Errorf("NoConflict not implemented")
}

/* Opcode: IdxInsert P1 P2 P3 P4 P5
** Synopsis: key=r[P2]
**
** Register P2 holds an SQL index key made using the
** MakeRecord instructions.  This opcode writes that key
** into the index P1.  Data for the entry is nil.
**
** If P4 is not zero, then it is the number of values in the unpacked
** key of reg(P2).  In that case, P3 is the index of the first register
** for the unpacked key.  The availability of the unpacked key can sometimes
** be an optimization.
**
** If P5 has the OPFLAG_APPEND bit set, that is a hint to the b-tree layer
** that this insert is likely to be an append.
**
** If P5 has the OPFLAG_NCHANGE bit set, then the change counter is
** incremented by this instruction.  If the OPFLAG_NCHANGE bit is clear,
** then the change counter is unchanged.
**
** If the OPFLAG_USESEEKRESULT flag of P5 is set, the implementation might
** run faster by avoiding an unnecessary seek on cursor P1.  However,
** the OPFLAG_USESEEKRESULT flag must only be set if there have been no prior
** seeks on the cursor or if the most recent seek used a key equivalent
** to P2.
**
** This instruction only works for indices.  The equivalent instruction
** for tables is OP_Insert.
 */
func opIdxInsert(wvm *WVM, op Op) (err error) {
	//key := wvm.register[op.P2]
	//fmt.Printf("[wvm:opIdxInsert]: registers: %v\n", wvm.register)
	//fmt.Printf("[wvm:opIdxInsert]: INSERT KEY: %v\n", key)
	//fmt.Printf("[wvm:opIdxInsert]: TODO: something with above KEY\n")
	return nil // TODO: fmt.Errorf("IdxInsert not implemented")
}

/* Opcode: Noop * * * * *
**
** Do nothing.  This instruction is often useful as a jump
** destination.
 */
func opNoop(wvm *WVM, op Op) (err error) {
	return nil
}

/* Opcode: IdxRowid P1 P2 * * *
** Synopsis: r[P2]=rowid
**
** Write into register P2 an integer which is the last entry in the record at
** the end of the index key pointed to by cursor P1.  This integer should be
** the rowid of the table entry to which this index entry points.
**
** See also: Rowid, MakeRecord.
 */
func opIdxRowid(wvm *WVM, op Op) (err error) {
	rowid, ok, err := wvm.setupCurrentRowId(op.P1)
	if err != nil {
		return fmt.Errorf("[wvm:opIdxRowid] %s", err)
	}
	if !ok {
		log.Error("[wvm:opIdxRowid] chunk key not found. Has it been stored? Continuing.")
		//wvm.halted = true
		//return fmt.Errorf("[wvm:opIdxRowid] record seeked not found. has it been stored?")
	}
	if rowid != 0 {
		wvm.register[op.P2] = rowid
		wvm.rowSet[op.P2] = append(wvm.rowSet[op.P2], rowid)
		log.Info(fmt.Sprintf("[wvm:opIdxRowid] rowid (%v) generated for op.P1 (%v) table entry", rowid, op.P1))
	} //else {
	//log.Error("[wvm:opIdxRowid] rowid is 0, generated for op.P1 table entry", "op.P1", op.P1)
	//}

	return nil
}

/* Opcode: DeferredSeek P1 * P3 P4 *
** Synopsis: Move P3 to P1.rowid if needed
**
** P1 is an open index cursor and P3 is a cursor on the corresponding
** table.  This opcode does a deferred seek of the P3 table cursor
** to the row that corresponds to the current row of P1.
**
** This is a deferred seek.  Nothing actually happens until
** the cursor is used to read a record.  That way, if no reads
** occur, no unnecessary I/O happens.
**
** P4 may be an array of integers (type P4_INTARRAY) containing
** one entry for each column in the P3 table.  If array entry a(i)
** is non-zero, then reading column a(i)-1 from cursor P3 is
** equivalent to performing the deferred seek and then reading column i
** from P1.  This information is stored in P3 and used to redirect
** reads against P3 over to P1, thus possibly avoiding the need to
** seek and read cursor P3.
 */
func opDeferredSeek(wvm *WVM, op Op) (err error) {
	//P3 table cursor moves to the row (rowid) in P1
	wvm.cursors[op.P3] = wvm.cursors[op.P1]
	//fmt.Printf("[wvm:opDeferredSeek] Not implemented - not sure why this is needed\n")
	return nil
}

/* Opcode: SeekGE P1 P2 P3 P4 *
 ** Synopsis: key=r[P3@P4]
 **
 ** If cursor P1 refers to an SQL table (B-Tree that uses integer keys),
 ** use the value in register P3 as the key.  If cursor P1 refers
 ** to an SQL index, then P3 is the first in an array of P4 registers
 ** that are used as an unpacked index key.
 **
 ** Reposition cursor P1 so that  it points to the smallest entry that
 ** is greater than or equal to the key value. If there are no records
 ** greater than or equal to the key and P2 is not zero, then jump to P2.
 **
 ** If the cursor P1 was opened using the OPFLAG_SEEKEQ flag, then this
 ** opcode will always land on a record that equally equals the key, or
 ** else jump immediately to P2.  When the cursor is OPFLAG_SEEKEQ, this
 ** opcode must be followed by an IdxLE opcode with the same arguments.
 ** The IdxLE opcode will be skipped if this opcode succeeds, but the
 ** IdxLE opcode will be used on subsequent loop iterations.
 **
 ** This opcode leaves the cursor configured to move in forward order,
 ** from the beginning toward the end.  In other words, the cursor is
 ** configured to use Next, not Prev.
 **
 ** See also: Found, NotFound, SeekLt, SeekGt, SeekLe
 */
func opSeekGE(wvm *WVM, op Op) (err error) {

	if _, ok := wvm.register[op.P3]; !ok {
		return fmt.Errorf("[wvm:opSeekGE] no key (%v) to look up in registers (%+v)", op.P3, wvm.register)
	}

	var k []byte
	if wvm.register[op.P3] != nil { // if it is nil, it needs to Seek and fail
		switch wvm.register[op.P3].(type) {
		case int:
			k = IntToByte(wvm.register[op.P3].(int))
			break
		case int64:
			k = Int64ToByte(wvm.register[op.P3].(int64))
			break
		case uint64:
			k = Uint64ToByte(wvm.register[op.P3].(uint64))
			break
		case float64:
			k = FloatToByte(wvm.register[op.P3].(float64))
			break
		case string:
			k = []byte(wvm.register[op.P3].(string))
			break
		default:
			return fmt.Errorf("[wvm:opSeekGE] type (%v) not supported.", reflect.TypeOf(wvm.register[op.P3]))
		}
	}

	cursor, ok, err := wvm.index[op.P1].Seek(k)
	fmt.Printf("[wvm:opSeekGE] cursor gotten: %+v\n", cursor)
	if err != nil {
		return fmt.Errorf("[wvm:opSeekGE] %s", err)
	}
	if ok { // found the key
		fmt.Printf("[wvm:opSeekGE] found the key(%v)\n", k)
		//fmt.Printf("[wvm:opSeekGE] HMMMMM reposition the cursor to thte next?")
		//_, _, err = cursor.Next()
		//if err != nil {
		//	return fmt.Errorf("[wvm:opSeekGE] %s", err)
		//}
		wvm.cursors[op.P1] = cursor
		fmt.Printf("[wvm:opSeekGE] cursors[op.P1] = (%+v)\n", cursor)
		wvm.dumpCursors()
		return nil
	} else {
		fmt.Printf("[wvm.opSeekGE] key is > or eof\n")
	}
	_, _, err = cursor.Next()
	fmt.Printf("[wvm:opSeekGE] cursor after a Next is: %+v\n", cursor)
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("[wvm:opSeekGE] %s", err)
		}
		if op.P2 > 0 {
			// wvm.eof = true //TODO: is this supposed to be here?
			wvm.pc = op.P2
			wvm.jumped = true
			fmt.Printf("[wvm.opSeekGE] wvm.pc is P2: %d\n", wvm.pc)
		} else {
			fmt.Printf("[wvm.opSeekGE] wvm.eof\n")
			wvm.eof = true
		}
	}
	wvm.cursors[op.P1] = cursor
	fmt.Printf("[wvm:opSeekGE] ")
	wvm.dumpCursors()

	return nil
}

/* Opcode: SeekGT P1 P2 P3 P4 *
 ** Synopsis: key=r[P3@P4]
 **
 ** If cursor P1 refers to an SQL table (B-Tree that uses integer keys),
 ** use the value in register P3 as a key. If cursor P1 refers
 ** to an SQL index, then P3 is the first in an array of P4 registers
 ** that are used as an unpacked index key.
 **
 ** Reposition cursor P1 so that  it points to the smallest entry that
 ** is greater than the key value. If there are no records greater than
 ** the key and P2 is not zero, then jump to P2.
 **
 ** This opcode leaves the cursor configured to move in forward order,
 ** from the beginning toward the end.  In other words, the cursor is
 ** configured to use Next, not Prev.
 **
 ** See also: Found, NotFound, SeekLt, SeekGe, SeekLe
 */
func opSeekGT(wvm *WVM, op Op) (err error) {

	if _, ok := wvm.register[op.P3]; !ok {
		return fmt.Errorf("[wvm:opSeekGT] no key (%v) to look up in registers (%+v)", op.P3, wvm.register)
	}

	var k []byte
	if wvm.register[op.P3] != nil { // if it is nil, then it still needs to reposition to the next cursor

		switch wvm.register[op.P3].(type) {
		case int:
			k = IntToByte(wvm.register[op.P3].(int))
			break
		case int64:
			k = Int64ToByte(wvm.register[op.P3].(int64))
			break
		case uint64:
			k = Uint64ToByte(wvm.register[op.P3].(uint64))
			break
		case float64:
			k = FloatToByte(wvm.register[op.P3].(float64))
			break
		case string:
			k = []byte(wvm.register[op.P3].(string))
			break
		default:
			return fmt.Errorf("[wvm:opSeekGT] type (%v) not supported.", reflect.TypeOf(wvm.register[op.P3]))
		}
	}

	// use P1's btree to find the key
	cursor, ok, err := wvm.index[op.P1].Seek(k)
	fmt.Printf("[wvm:opSeekGT] cursor gotten: %+v\n", cursor)
	if err != nil {
		return fmt.Errorf("[wvm:opSeekGT] %s", err)
	}
	_, _, err = cursor.Next() //check current value seek'ed
	if !ok {                  //reports key is > item.key, or at the end
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("[wvm:opSeekGT] %s", err)
			}
			if op.P2 > 0 {
				//wvm.eof = true //TODO: is this supposed to be here?
				wvm.cursors[op.P1] = cursor
				wvm.pc = op.P2
				wvm.jumped = true
				return nil
			}
		}
		wvm.cursors[op.P1] = cursor
		//wvm.pc = op.P2    //why does this work here?
		//wvm.jumped = true //why does this work here?
		return nil
	}

	//found the key, need to get >
	//fmt.Printf("[wvm:opSeekGT] is == key, found it\n")
	_, _, err = cursor.Next()
	wvm.cursors[op.P1] = cursor
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("[wvm:opSeekGE] %s", err)
		}
		if op.P2 > 0 {
			//wvm.eof = true ??
			wvm.pc = op.P2
			wvm.jumped = true
		} else {
			wvm.eof = true
		}
	}

	fmt.Printf("[wvm:opSeekGT] nextcursor: %+v\n", wvm.cursors[op.P1])
	return nil
}

/* Opcode: SeekLT P1 P2 P3 P4 *
 ** Synopsis: key=r[P3@P4]
 **
 ** If cursor P1 refers to an SQL table (B-Tree that uses integer keys),
 ** use the value in register P3 as a key. If cursor P1 refers
 ** to an SQL index, then P3 is the first in an array of P4 registers
 ** that are used as an unpacked index key.
 **
 ** Reposition cursor P1 so that  it points to the largest entry that
 ** is less than the key value. If there are no records less than
 ** the key and P2 is not zero, then jump to P2.
 **
 ** This opcode leaves the cursor configured to move in reverse order,
 ** from the end toward the beginning.  In other words, the cursor is
 ** configured to use Prev, not Next.
 **
 ** See also: Found, NotFound, SeekGt, SeekGe, SeekLe
 */
func opSeekLT(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("SeekLT not implemented")
}

/* Opcode: SeekLE P1 P2 P3 P4 *
 ** Synopsis: key=r[P3@P4]
 **
 ** If cursor P1 refers to an SQL table (B-Tree that uses integer keys),
 ** use the value in register P3 as a key. If cursor P1 refers
 ** to an SQL index, then P3 is the first in an array of P4 registers
 ** that are used as an unpacked index key.
 **
 ** Reposition cursor P1 so that it points to the largest entry that
 ** is less than or equal to the key value. If there are no records
 ** less than or equal to the key and P2 is not zero, then jump to P2.
 **
 ** This opcode leaves the cursor configured to move in reverse order,
 ** from the end toward the beginning.  In other words, the cursor is
 ** configured to use Prev, not Next.
 **
 ** If the cursor P1 was opened using the OPFLAG_SEEKEQ flag, then this
 ** opcode will always land on a record that equally equals the key, or
 ** else jump immediately to P2.  When the cursor is OPFLAG_SEEKEQ, this
 ** opcode must be followed by an IdxGE opcode with the same arguments.
 ** The IdxGE opcode will be skipped if this opcode succeeds, but the
 ** IdxGE opcode will be used on subsequent loop iterations.
 **
 ** See also: Found, NotFound, SeekGt, SeekGe, SeekLt
 */
func opSeekLE(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("SeekLE not implemented")
}

/* Opcode: IdxGE P1 P2 P3 P4 P5
 ** Synopsis: key=r[P3@P4]
 **
 ** The P4 register values beginning with P3 form an unpacked index
 ** key that omits the PRIMARY KEY.  Compare this key value against the index
 ** that P1 is currently pointing to, ignoring the PRIMARY KEY or ROWID
 ** fields at the end.
 **
 ** If the P1 index entry is greater than or equal to the key value
 ** then jump to P2.  Otherwise fall through to the next instruction.
 */
func opIdxGE(wvm *WVM, op Op) (err error) {
	ok, err := wvm.registerCheck(op.P1, op.P3)
	if err != nil {
		return fmt.Errorf("[wvm:opIdxGE] %s", err)
	}
	if !ok {
		log.Info(fmt.Sprintf("[wvm:opIdxGE] one or both keys (%v) (%v) have nil registers. Continuing.", op.P1, op.P3)) // ?
		return nil
	}
	f1, _, err := convertNumOrString(wvm.register[op.P1])
	if err != nil {
		return fmt.Errorf("[wvm:opIdxGE] %s", err)
	}
	f3, _, err := convertNumOrString(wvm.register[op.P3])
	if err != nil {
		return fmt.Errorf("[wvm:opIdxGE] %s", err)
	}
	if f1 != nil && f3 != nil {
		if *f3 <= *f1 {
			wvm.pc = op.P2
			wvm.jumped = true
			return nil
		}
	}
	// if wvm.register[op.P3].(int) <= wvm.register[op.P1].(int) {
	// 	//fmt.Printf("[wvm:opIdxGE] reg p3 found to be > reg p1 : %+v >= %+v\n", wvm.register[op.P3], wvm.register[op.P1])
	// 	wvm.pc = op.P2
	// 	wvm.jumped = true
	//
	// } //else {
	//fmt.Printf("[wvm:opIdxGE] reg p3 found to be NOT > reg p1 : %+v < %+v\n", wvm.register[op.P3], wvm.register[op.P1])
	//}
	return nil
}

/* Opcode: IdxGT P1 P2 P3 P4 P5
 ** Synopsis: key=r[P3@P4]
 **
 ** The P4 register values beginning with P3 form an unpacked index
 ** key that omits the PRIMARY KEY.  Compare this key value against the index
 ** that P1 is currently pointing to, ignoring the PRIMARY KEY or ROWID
 ** fields at the end.
 **
 ** If the P1 index entry is greater than the key value
 ** then jump to P2.  Otherwise fall through to the next instruction.
 */
func opIdxGT(wvm *WVM, op Op) (err error) {
	// if wvm.register[op.P1] == nil || wvm.register[op.P3] == nil {
	// 	fmt.Printf("[wvm:opIdxGT] someone is nil: register(op.P3): %+v, register(op.P1): %+v\n", wvm.register[op.P3], wvm.register[op.P1])
	// 	log.Info("[wvm:opIdxGT] register is nil. Continuing.", "reg(P3)", wvm.register[op.P3], "reg(P1)", wvm.register[op.P1])
	// 	return nil //TODO should return err?
	// }
	ok, err := wvm.registerCheck(op.P1, op.P3)
	if err != nil {
		return fmt.Errorf("[wvm:opIdxGT] %s", err)
	}
	if !ok {
		fmt.Printf("[wvm:opIdxGT] one or both keys (%v) (%v) have nil registers. Continuing.", op.P1, op.P3)
		log.Info(fmt.Sprintf("[wvm:opIdxGT] one or both keys (%v) (%v) have nil registers. Continuing.", op.P1, op.P3)) // ?
		return nil
	}
	f3, s3, err := convertNumOrString(wvm.register[op.P3])
	if err != nil {
		return fmt.Errorf("[wvm:opIdxGT] %s", err)
	}

	fmt.Printf("[wvm:opIdxGT] op.P3 (%v), op.P1 (%v), registers (%v)\n", op.P3, op.P1, wvm.register)
	if _, ok := wvm.cursors[op.P1]; !ok {
		fmt.Printf("[wvm:opIdxGT] no cursor for op.P1!")
		wvm.dumpCursors()
	}
	kp1_byte, err := wvm.cursors[op.P1].GetCurrent()
	if err != nil { //TODO: check this
		if err != io.EOF {
			return fmt.Errorf("[wvm:opIdxGT] %s", err)
		}
		//wvm.eof = true ??
		return nil
	}

	jump := false
	if f3 != nil {
		kp1 := BytesToInt(kp1_byte)
		fmt.Printf("[wvm:opIdxGT] kp1 (%v) *f3 (%v)\n", kp1, *f3)
		if kp1 > int(*f3) {
			jump = true
			fmt.Printf("[wvm:opIdxGT] jump is true")
		}
	} else if s3 != nil {
		log.Error("[wvm:opIdxGT] not sure we should be doing this for strings?")
		k1, k3 := padBytesForComparison(kp1_byte, []byte(*s3))
		if string(k1) > string(k3) {
			jump = true
		}
	}

	// switch wvm.register[op.P3].(type) {
	// case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
	// 	kp1 := BytesToInt(kp1_byte)
	// 	fmt.Printf("[wvm:opIdxGT] kp1 (%v) wvm.register[op.P3].(int) (%v)\n", kp1, wvm.register[op.P3].(int))
	// 	if kp1 > wvm.register[op.P3].(int) {
	// 		jump = true
	// 	}
	// case float64:
	// 	kp1 := BytesToFloat(kp1_byte)
	// 	fmt.Printf("[wvm:opIdxGT] kp1 (%v) wvm.register[op.P3].(float64) (%v)\n", kp1, wvm.register[op.P3].(float64))
	// 	if kp1 > wvm.register[op.P3].(float64) {
	// 		jump = true
	// 	}
	// case string:
	// 	k1, k3 := padBytesForComparison(kp1_byte, []byte(wvm.register[op.P3].(string)))
	// 	//fmt.Printf("[wvm:opIdxGT] k1: %v\n", k1)
	// 	//fmt.Printf("[wvm:opIdxGT] k3: %v\n", k3)
	// 	if string(k1) > string(k3) {
	// 		//fmt.Printf("[wvm:opIdxGT] k1 > k3, jump\n")
	// 		jump = true
	// 	} else {
	// 		//fmt.Printf("[wvm:opIdxGT] k1 < k3, no jump \n")
	// 	}
	// default:
	// 	return fmt.Errorf("[wvm:opIdxGT] type of register[op.P3] %v not supported.", reflect.TypeOf(wvm.register[op.P3]))
	// }

	if jump {
		//fmt.Printf("[wvm:opIdxGT] reg op.p1 > op.p3 \n")
		wvm.pc = op.P2
		wvm.jumped = true
	}
	return nil
}

/* Opcode: IdxLT P1 P2 P3 P4 P5
 ** Synopsis: key=r[P3@P4]
 **
 ** The P4 register values beginning with P3 form an unpacked index
 ** key that omits the PRIMARY KEY or ROWID.  Compare this key value against
 ** the index that P1 is currently pointing to, ignoring the PRIMARY KEY or
 ** ROWID on the P1 index.
 **
 ** If the P1 index entry is less than the key value then jump to P2.
 ** Otherwise fall through to the next instruction.
 */
func opIdxLT(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("IdxLT not implemented")
}

/* Opcode: IdxLE P1 P2 P3 P4 P5
 ** Synopsis: key=r[P3@P4]
 **
 ** The P4 register values beginning with P3 form an unpacked index
 ** key that omits the PRIMARY KEY or ROWID.  Compare this key value against
 ** the index that P1 is currently pointing to, ignoring the PRIMARY KEY or
 ** ROWID on the P1 index.
 **
 ** If the P1 index entry is less than or equal to the key value then jump
 ** to P2. Otherwise fall through to the next instruction.
 */
func opIdxLE(wvm *WVM, op Op) (err error) {
	return fmt.Errorf("IdxLE not implemented")
}

// ------------------------ SKIPPABLE OPCODES (for now)
/* Opcode: Transaction P1 P2 P3 P4 P5
**
** Begin a transaction on database P1 if a transaction is not already
** active.
** If P2 is non-zero, then a write-transaction is started, or if a
** read-transaction is already active, it is upgraded to a write-transaction.
** If P2 is zero, then a read-transaction is started.
**
** P1 is the index of the database file on which the transaction is
** started.  Index 0 is the main database file and index 1 is the
** file used for temporary tables.  Indices of 2 or more are used for
** attached databases.
**
** If a write-transaction is started and the Vdbe.usesStmtJournal flag is
** true (this flag is set if the Vdbe may modify more than one row and may
** throw an ABORT exception), a statement transaction may also be opened.
** More specifically, a statement transaction is opened iff the database
** connection is currently not in autocommit mode, or if there are other
** active statements. A statement transaction allows the changes made by this
** VDBE to be rolled back after an error without having to roll back the
** entire transaction. If no error is encountered, the statement transaction
** will automatically commit when the VDBE halts.
**
** If P5!=0 then this opcode also checks the schema cookie against P3
** and the schema generation counter against P4.
** The cookie changes its value whenever the database schema changes.
** This operation is used to detect when that the cookie has changed
** and that the current process needs to reread the schema.  If the schema
** cookie in P3 differs from the schema cookie in the database header or
** if the schema generation counter in P4 differs from the current
** generation counter, then an SQLITE_SCHEMA error is raised and execution
** halts.  The sqlite3_step() wrapper function might then reprepare the
** statement and rerun it from the beginning.
 */
func opTransaction(wvm *WVM, op Op) (err error) {
	return nil
}

/* Opcode: Trace P1 P2 * P4 *
**
** Write P4 on the statement trace output if statement tracing is
** enabled.
**
** Operand P1 must be 0x7fffffff and P2 must positive.
 */
func opTrace(wvm *WVM, op Op) (err error) {
	return nil
}

/* Opcode: Init P1 P2 P3 P4 *
** Synopsis: Start at P2
**
** Programs contain a single instance of this opcode as the very first
** opcode.
**
** If tracing is enabled (by the sqlite3_trace()) interface, then
** the UTF-8 string contained in P4 is emitted on the trace callback.
** Or if P4 is blank, use the string returned by sqlite3_sql().
**
** If P2 is not zero, jump to instruction P2.
**
** Increment the value of P1 so that OP_Once opcodes will jump the
** first time they are evaluated for this run.
**
** If P3 is not zero, then it is an address to jump to if an SQLITE_CORRUPT
** error is encountered.
 */
func opInit(wvm *WVM, op Op) (err error) {
	if op.P2 > 0 {
		wvm.pc = op.P2
		wvm.jumped = true
	}
	return nil
}

func opVerifyCookie(wvm *WVM, op Op) (err error) {
	return nil
}

/* Opcode: TableLock P1 P2 P3 P4 *
** Synopsis: iDb=P1 root=P2 write=P3
**
** Obtain a lock on a particular table. This instruction is only used when
** the shared-cache feature is enabled.
**
** P1 is the index of the database in sqlite3.aDb[] of the database
** on which the lock is acquired.  A readlock is obtained if P3==0 or
** a write lock if P3==1.
**
** P2 contains the root-page of the table to lock.
**
** P4 contains a pointer to the name of the table being locked. This is only
** used to generate an error message if the lock cannot be obtained.
 */
func opTableLock(wvm *WVM, op Op) (err error) {
	return nil
}
func NewWVM(sdb *StateDB, db *SQLDatabase) *WVM {

	// sqlite> select * from main.sqlite_master;
	// table|person|person|2|CREATE TABLE person (person_id int, name string)
	// table|people|people|3|CREATE TABLE people (person_id int primary key, name string)
	// index|sqlite_autoindex_people_1|people|4|
	// table|human|human|5|CREATE TABLE human (person_id int primary key, first string, last string, age float)
	// index|sqlite_autoindex_human_1|human|6|
	// index|first|human|7|CREATE INDEX first on human (first)
	// index|last|human|8|CREATE INDEX last on human (last)
	// index|age|human|9|CREATE INDEX age on human (age)
	//
	// db.wvmSchema[0] = WVMIndex{SchemaType: "table", IndexName: "person", TableName: "person", RootPage: 2, Sql: "CREATE TABLE person (person_id int, name string)"}
	// db.wvmSchema[1] = WVMIndex{SchemaType: "table", IndexName: "people", TableName: "people", RootPage: 3, Sql: "CREATE TABLE people (person_id int primary key, name string)"}
	// db.wvmSchema[2] = WVMIndex{SchemaType: "index", IndexName: "sqlite_autoindex_people_1", TableName: "people", RootPage: 4, Sql: ""}
	// db.wvmSchema[3] = WVMIndex{SchemaType: "table", IndexName: "human", TableName: "", RootPage: 5, Sql: "CREATE TABLE human (person_id int primary key, first string, last string, age float)"}
	// db.wvmSchema[4] = WVMIndex{SchemaType: "index", IndexName: "sqlite_autoindex_human_1", TableName: "human", RootPage: 6, Sql: ""}
	// db.wvmSchema[5] = WVMIndex{SchemaType: "index", IndexName: "first", TableName: "human", RootPage: 7, Sql: "CREATE INDEX first on human (first)"}
	// db.wvmSchema[6] = WVMIndex{SchemaType: "index", IndexName: "last", TableName: "human", RootPage: 8, Sql: "CREATE INDEX last on human (last)"}
	// db.wvmSchema[7] = WVMIndex{SchemaType: "index", IndexName: "age", TableName: "human", RootPage: 9, Sql: "CREATE INDEX age on human (age)"}

	vm := &WVM{
		//user:     user,
		sqlchain: sdb,
		activeDB: db,
		// activeTable: tbl,
		register:        make(map[int]interface{}, 128),
		once_marker:     make(map[int]int, 128),
		index:           make(map[int]*Tree, 128),
		cursors:         make(map[int]OrderedDatabaseCursor, 128),
		rowSet:          make(map[int][]int64, 128),
		JumpTable:       SVMVenusInstructionSet(),
		OPFLAG_ISUPDATE: false,
	}
	return vm
}

func SVMVenusInstructionSet() (jt map[string]OpDef) {
	jt = make(map[string]OpDef, 256)
	jt[OP_Trace] = OpDef{f: opTrace}
	jt[OP_Goto] = OpDef{f: opGoto}
	jt[OP_ReopenIdx] = OpDef{f: opReopenIdx}
	jt[OP_OpenRead] = OpDef{f: opOpenRead}
	jt[OP_Rewind] = OpDef{f: opRewind}
	jt[OP_Rowid] = OpDef{f: opRowid}
	jt[OP_Column] = OpDef{f: opColumn}
	jt[OP_ResultRow] = OpDef{f: opResultRow}
	jt[OP_Next] = OpDef{f: opNext}
	jt[OP_Close] = OpDef{f: opClose}
	jt[OP_Halt] = OpDef{f: opHalt}
	jt[OP_Transaction] = OpDef{f: opTransaction}
	jt[OP_VerifyCookie] = OpDef{f: opVerifyCookie}
	jt[OP_TableLock] = OpDef{f: opTableLock}
	jt[OP_OpenWrite] = OpDef{f: opOpenWrite}
	jt[OP_NewRowid] = OpDef{f: opNewRowid}
	jt[OP_Integer] = OpDef{f: opInteger}
	jt[OP_Multiply] = OpDef{f: opMultiply}
	jt[OP_Add] = OpDef{f: opAdd}
	jt[OP_MakeRecord] = OpDef{f: opMakeRecord}
	jt[OP_Insert] = OpDef{f: opInsert}
	jt[OP_Delete] = OpDef{f: opDelete}
	jt[OP_Subtract] = OpDef{f: opSubtract}
	jt[OP_String8] = OpDef{f: opString8}
	jt[OP_Gosub] = OpDef{f: opGosub}
	jt[OP_Return] = OpDef{f: opReturn}
	jt[OP_Int64] = OpDef{f: opInt64}
	jt[OP_Real] = OpDef{f: opReal}
	jt[OP_Null] = OpDef{f: opNull}
	jt[OP_Move] = OpDef{f: opMove}
	jt[OP_Copy] = OpDef{f: opCopy}
	jt[OP_Divide] = OpDef{f: opDivide}
	jt[OP_Remainder] = OpDef{f: opRemainder}
	jt[OP_Eq] = OpDef{f: opEq}
	jt[OP_Ne] = OpDef{f: opNe}
	jt[OP_Lt] = OpDef{f: opLt}
	jt[OP_Le] = OpDef{f: opLe}
	jt[OP_Gt] = OpDef{f: opGt}
	jt[OP_Ge] = OpDef{f: opGe}
	jt[OP_Compare] = OpDef{f: opCompare}
	jt[OP_Jump] = OpDef{f: opJump}
	jt[OP_And] = OpDef{f: opAnd}
	jt[OP_Or] = OpDef{f: opOr}
	jt[OP_Not] = OpDef{f: opNot}
	jt[OP_BitAnd] = OpDef{f: opBitAnd}
	jt[OP_BitOr] = OpDef{f: opBitOr}
	jt[OP_BitNot] = OpDef{f: opBitNot}
	jt[OP_If] = OpDef{f: opIf}
	jt[OP_IfNot] = OpDef{f: opIfNot}
	jt[OP_Count] = OpDef{f: opCount}
	jt[OP_SorterOpen] = OpDef{f: opSorterOpen}
	jt[OP_NotExists] = OpDef{f: opNotExists}
	jt[OP_Sequence] = OpDef{f: opSequence}
	jt[OP_SorterData] = OpDef{f: opSorterData}
	jt[OP_Last] = OpDef{f: opLast}
	jt[OP_SorterSort] = OpDef{f: opSorterSort}
	jt[OP_Sort] = OpDef{f: opSort}
	jt[OP_SorterNext] = OpDef{f: opSorterNext}
	jt[OP_PrevIfOpen] = OpDef{f: opPrevIfOpen}
	jt[OP_NextIfOpen] = OpDef{f: opNextIfOpen}
	jt[OP_Prev] = OpDef{f: opPrev}
	jt[OP_SorterInsert] = OpDef{f: opSorterInsert}
	jt[OP_RowSetAdd] = OpDef{f: opRowSetAdd}
	jt[OP_RowSetRead] = OpDef{f: opRowSetRead}
	jt[OP_RowSetTest] = OpDef{f: opRowSetTest}
	jt[OP_IfPos] = OpDef{f: opIfPos}
	jt[OP_AggStep] = OpDef{f: opAggStep}
	jt[OP_AggStep0] = OpDef{f: opAggStep0}
	jt[OP_AggFinal] = OpDef{f: opAggFinal}
	jt[OP_Function] = OpDef{f: opFunction}
	jt[OP_Function0] = OpDef{f: opFunction0}
	jt[OP_DecrJumpZero] = OpDef{f: opDecrJumpZero}
	jt[OP_Init] = OpDef{f: opInit}
	jt[OP_CollSeq] = OpDef{f: opCollSeq}
	jt[OP_Affinity] = OpDef{f: opAffinity}
	jt[OP_RealAffinity] = OpDef{f: opRealAffinity}
	jt[OP_SCopy] = OpDef{f: opSCopy}
	jt[OP_Found] = OpDef{f: opFound}
	jt[OP_NotFound] = OpDef{f: opNotFound}
	jt[OP_NoConflict] = OpDef{f: opNoConflict}
	jt[OP_IdxInsert] = OpDef{f: opIdxInsert}
	jt[OP_Noop] = OpDef{f: opNoop}
	jt[OP_DeferredSeek] = OpDef{f: opDeferredSeek}
	jt[OP_IdxRowid] = OpDef{f: opIdxRowid}
	jt[OP_IdxDelete] = OpDef{f: opIdxDelete}
	jt[OP_IsNull] = OpDef{f: opIsNull}
	jt[OP_Once] = OpDef{f: opOnce}
	jt[OP_OpenPseudo] = OpDef{f: opOpenPseudo}
	jt[OP_SeekGE] = OpDef{f: opSeekGE}
	jt[OP_SeekGT] = OpDef{f: opSeekGT}
	jt[OP_SeekLT] = OpDef{f: opSeekLT}
	jt[OP_SeekLE] = OpDef{f: opSeekLE}
	jt[OP_IdxGE] = OpDef{f: opIdxGE}
	jt[OP_IdxGT] = OpDef{f: opIdxGT}
	jt[OP_IdxLT] = OpDef{f: opIdxLT}
	jt[OP_IdxLE] = OpDef{f: opIdxLE}
	jt[OP_IntCopy] = OpDef{f: opIntCopy}
	return jt
}

func (wvm *WVM) CheckOps(ops []Op) (err error) {
	missing_opcodes := make(map[string]int)
	cnt := 0
	for i, op := range ops {
		var ok bool
		if _, ok = wvm.JumpTable[op.Opcode]; !ok {
			fmt.Printf("MISSING %d | %s\n", i, op.Opcode)
			missing_opcodes[op.Opcode]++
			cnt++
		}
	}
	if cnt > 0 {
		cnt = 0
		for opcode, _ := range missing_opcodes {
			fmt.Printf("%s\n", opcode)
			cnt++
		}
		fmt.Printf("%d missing opcodes\n", cnt)
	}
	return nil
}

// TODO: remove when debugging is done
func (wvm *WVM) dumpCursors() {
	fmt.Printf("[wvm:dumpCursors] cursors:\n")
	for i, c := range wvm.cursors {
		fmt.Printf("    %v:%+v\n", i, c)
	}
}

func (wvm *WVM) SetFlags(flag string) {
	f := strings.ToLower(flag)
	switch f {
	case "update":
		wvm.OPFLAG_ISUPDATE = true
		break
	default:
		log.Error("[wvm:SetFlags] Flag type not supported", "type", flag)
	}
}

func (wvm *WVM) ProcessOps(ops []Op) (err error) {

	// DEBUG only
	fmt.Printf("[wvm:ProcessOps] ops: \n")
	for _, o := range ops {
		fmt.Printf("%+v\n", o)
	}

	wvm.pc = 0
	// keep going until there is an error or we have halted
	wvm.Rows = make([][]interface{}, 0)
	count := 0
	for wvm.halted = false; wvm.halted == false; {
		op := ops[wvm.pc]
		var d OpDef
		var ok bool

		// DEBUG
		fmt.Printf("\n[wvm:ProcessOps] op: %+v\n", op)
		fmt.Printf("[wvm:ProcessOps] jumped: %+v\n", wvm.jumped)
		fmt.Printf("[wvm:ProcessOps] pc: %+v\n", wvm.pc)
		fmt.Printf("[wvm:ProcessOps] registers: %+v\n", wvm.register)
		fmt.Printf("[wvm:ProcessOps] indexes: %+v\n", wvm.index)
		wvm.dumpCursors()
		fmt.Printf("[wvm:ProcessOps] rowSets: %+v\n", wvm.rowSet)
		fmt.Printf("[wvm:ProcessOps] activeKey: %+v\n", wvm.activeKey)
		fmt.Printf("[wvm:ProcessOps] activeRow: %+v\n", wvm.activeRow)
		//fmt.Printf("[wvm:ProcessOps] sorter: %+v\n", wvm.sorter)
		fmt.Printf("[wvm:ProcessOps] Rows(output): %+v\n", wvm.Rows)

		//log.Info("[wvm:ProcessOps] processing", "pc", wvm.pc, "count", count, "op", op)

		if d, ok = wvm.JumpTable[op.Opcode]; !ok {
			return fmt.Errorf("[wvm:ProcessOps] Unknown opcode %s", op.Opcode)
		}
		wvm.jumped = false
		err = d.f(wvm, op)
		if err != nil {
			wvm.halted = true
			return err
		} else if wvm.halted {
			// return back rows ... or something
		} else if wvm.jumped {
			// we have updated the program counter
		} else {
			wvm.pc++
		}

		if count > 500 {
			log.Info("[wvm:ProcessOps] 500 opcodes processed. halting.")
			wvm.halted = true
		}
		count++
	}

	log.Info("[wvm:ProcessOps] output", "rows", wvm.Rows)
	return nil
}
