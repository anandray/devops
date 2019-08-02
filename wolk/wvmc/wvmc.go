package wvmc

/*
#cgo LDFLAGS: -L/usr/local/lib
#cgo LDFLAGS: -lsqlite3
#include "sqlite3.h"
#include "stdio.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unsafe"
)

type SQLiteError struct {
	Code         int
	ExtendedCode int
	err          string
}

func (err SQLiteError) Error() string {
	if err.err != "" {
		return err.err
	}
	return errorString(err)
}

func errorString(err SQLiteError) string {
	return C.GoString(C.sqlite3_errstr(C.int(err.Code)))
}

func sqlliteError(db *C.sqlite3) error {
	rv := C.sqlite3_errcode(db)
	if rv == C.SQLITE_OK {
		return nil
	}
	return SQLiteError{
		Code:         int(rv),
		ExtendedCode: int(C.sqlite3_extended_errcode(db)),
		err:          C.GoString(C.sqlite3_errmsg(db)),
	}
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

type Index struct {
	SchemaType string `json:"schematype"` //table or index
	IndexName  string `json:"indexname"`  //if schemaType is table, this is also the name of the table
	TableName  string `json:"tablename"`
	RootPage   int    `json:"rootpage"`
	Sql        string `json:"sql,omitempty"`
}

// runs the sql to modify schema (create/drop table)
func SetupSchema(dbpath string, sqlsetups []string) (err error) {

	var db *C.sqlite3
	var res *C.sqlite3_stmt
	var tail *C.char

	rc := C.sqlite3_open(C.CString(dbpath), &db)
	if rc != C.SQLITE_OK {
		return fmt.Errorf("could not open sqlite file")
	}

	for _, sql := range sqlsetups {
		sql_cstr := C.CString(sql)
		rc = C.sqlite3_prepare_v2(db, sql_cstr, -1, &res, &tail)
		if rc != C.SQLITE_OK {
			return sqlliteError(db)
		}
		for C.sqlite3_step(res) == C.SQLITE_ROW {
		}
	}

	C.sqlite3_finalize(res)
	C.sqlite3_close(db)
	return nil

}

// uses schema to run 'explain' to get list of opcodes, parses into Json format
func GetOps(dbpath string, sql string) (opsJson []byte, err error) {
	var db *C.sqlite3
	var res *C.sqlite3_stmt
	var tail *C.char
	var ops []Op

	rc := C.sqlite3_open(C.CString(dbpath), &db)
	if rc != C.SQLITE_OK {
		return opsJson, fmt.Errorf("could not open sqlite file")
	}

	sql_cstr := C.CString(fmt.Sprintf("explain %s", sql))
	rc = C.sqlite3_prepare_v2(db, sql_cstr, -1, &res, &tail)
	if rc != C.SQLITE_OK {
		return opsJson, sqlliteError(db)
	}

	for C.sqlite3_step(res) == C.SQLITE_ROW {
		var op Op
		// line #
		n := int(C.sqlite3_column_bytes(res, C.int(0)))
		c1 := C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 0))), C.int(n))
		op.N, _ = strconv.Atoi(c1)

		// op code
		n = int(C.sqlite3_column_bytes(res, C.int(1)))
		c1 = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 1))), C.int(n))
		op.Opcode = c1

		// p1
		n = int(C.sqlite3_column_bytes(res, C.int(2)))
		c1 = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 2))), C.int(n))
		op.P1, _ = strconv.Atoi(c1)

		// p2
		n = int(C.sqlite3_column_bytes(res, C.int(3)))
		c1 = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 3))), C.int(n))
		op.P2, _ = strconv.Atoi(c1)

		// p3
		n = int(C.sqlite3_column_bytes(res, C.int(4)))
		c1 = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 4))), C.int(n))
		op.P3, _ = strconv.Atoi(c1)

		// p4
		n = int(C.sqlite3_column_bytes(res, C.int(5)))
		op.P4 = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 5))), C.int(n))
		if op.N == 0 && op.Opcode == "Init" {
			op.P4 = sql
		}
		// p5
		n = int(C.sqlite3_column_bytes(res, C.int(6)))
		c1 = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 6))), C.int(n))
		op.P5, _ = strconv.Atoi(c1)

		// fmt.Printf("%v\n", op)
		ops = append(ops, op)
	}

	C.sqlite3_finalize(res)
	C.sqlite3_close(db)

	opsJson, err = json.Marshal(ops)
	if err != nil {
		return opsJson, err
	}

	return opsJson, nil
}

func GetMasterTable(dbpath string) (schemaJson [][]byte, err error) {
	var db *C.sqlite3
	var res *C.sqlite3_stmt
	var tail *C.char

	rc := C.sqlite3_open(C.CString(dbpath), &db)
	if rc != C.SQLITE_OK {
		return schemaJson, fmt.Errorf("could not open sqlite file")
	}
	sql := "select * from main.sqlite_master;"
	sql_cstr := C.CString(sql)
	rc = C.sqlite3_prepare_v2(db, sql_cstr, -1, &res, &tail)
	if rc != C.SQLITE_OK {
		return schemaJson, sqlliteError(db)
	}

	var schema [][]byte

	for C.sqlite3_step(res) == C.SQLITE_ROW {

		var index Index

		// schemaType
		n := int(C.sqlite3_column_bytes(res, C.int(0)))
		index.SchemaType = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 0))), C.int(n))

		// indexName
		n = int(C.sqlite3_column_bytes(res, C.int(1)))
		index.IndexName = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 1))), C.int(n))

		// tableName
		n = int(C.sqlite3_column_bytes(res, C.int(2)))
		index.TableName = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 2))), C.int(n))

		// rootPage
		n = int(C.sqlite3_column_bytes(res, C.int(3)))
		rp := C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 3))), C.int(n))
		index.RootPage, _ = strconv.Atoi(rp)

		// sql
		n = int(C.sqlite3_column_bytes(res, C.int(4)))
		index.Sql = C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 4))), C.int(n))

		//fmt.Printf("appending index %+v\n", index)
		indexJson, err := json.Marshal(index)
		if err != nil {
			return schemaJson, err
		}
		//fmt.Printf("appending index %s\n\n", string(indexJson))
		schema = append(schema, indexJson)

	}

	C.sqlite3_finalize(res)
	C.sqlite3_close(db)

	return schema, nil
}
