package main

/*
#cgo LDFLAGS: -L/usr/local/lib
#cgo LDFLAGS: -lsqlite3
#include "sqlite3.h"
#include "stdio.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// uses schema to run 'explain' to get list of opcodes, parses into Json format
func main() {
	var db *C.sqlite3
	var res *C.sqlite3_stmt
	var tail *C.char

	dbpath := "versioncheck.db"
	rc := C.sqlite3_open(C.CString(dbpath), &db)
	if rc != C.SQLITE_OK {
		fmt.Errorf("could not open sqlite file")

	}

	sql_cstr := C.CString("select sqlite_version();")
	rc = C.sqlite3_prepare_v2(db, sql_cstr, -1, &res, &tail)
	if rc != C.SQLITE_OK {
		fmt.Printf("ERROR\n") // sqlliteError(db)

	}

	for C.sqlite3_step(res) == C.SQLITE_ROW {
		// line #
		n := int(C.sqlite3_column_bytes(res, C.int(0)))
		c1 := C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(res, 0))), C.int(n))
		fmt.Printf("%s\n", c1)
	}
}
