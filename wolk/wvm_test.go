// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wolkdb/cloudstore/log"
)

func TestSetWVMSchema(t *testing.T) {
	// config, err := testSQLConfig()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	//dbpath := filepath.Join(config.DataDir, "wvmtest.db")
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	dbpath := "./tests/wvmtest.db"
	var sqlsetup []string
	sqlsetup = make([]string, 3)
	sqlsetup[0] = "create table person (person_id int primary key, name string)"
	sqlsetup[1] = "create table person_float (person_id int primary key, first string, last string, age float)"
	sqlsetup[2] = "create table blockchain (blockchainid string primary key, blockchainname string, blockchaintype string, blockchainstatus string, owner string, tokenid string, rpcaddr string, rpcport string, genesis string)"

	for _, sql := range sqlsetup {
		err := SetWVMSchema(dbpath, sql)
		if err != nil {
			t.Fatal(err)
		}
	}

	bytes, err := ReadWVMSchemaBytes(dbpath)
	if err != nil {
		t.Fatalf("ReadWVMSchemaBytes ERR %s", err)
	}
	s := string(bytes)
	fmt.Printf("[wvm_test] saved schema: \n %s\nx", s)
	if !strings.Contains(s, "indexsqlite_autoindex_person_1") || !strings.Contains(s, "tableperson") || !strings.Contains(s, "indexsqlite_autoindex_person_float_1") || !strings.Contains(s, "tableperson_float") {
		t.Fatalf("sqlsetups2 no match")
	}

}

func TestGetWVMSchema(t *testing.T) {
	// config, err := testSQLConfig()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// dbpath := filepath.Join(config.DataDir, "wvmtest.db")
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	dbpath := "./tests/wvmtest.db"
	schema, err := GetWVMSchema(dbpath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("[wvm_test] parsed schema: %+v\n", schema)
	if _, ok := schema[5]; !ok {
		t.Fatal("not enough rootpages", "num", len(schema))
	}

}

func TestDeleteLocalSchema(t *testing.T) {
	// config, err := testSQLConfig()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// dbpath := filepath.Join(config.DataDir, "wvmtest.db")
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	dbpath := "./tests/wvmtest.db"
	err := DeleteLocalSchema(dbpath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSVM(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// need to run wvmc/wvmc_test.go first

	//logprint()
	//tests to skip (for future implementation)
	skiptests := map[int]string{
		28: "GROUPBY",
		29: "GROUPBY",
		30: "GROUPBY",
		38: "GROUPBY-FLOAT",
	}

	//expected output
	expected := map[int]Expected{
		// 0: "insert into person (person_id, name) values (11, 'Happy')",
		0: Expected{AffectedRowCount: 1},
		// 1: "select * from person were person_id=11",
		1: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{11, "Happy"}}},
		// 2: "insert into person (person_id, name) values (12, 'Bertie')",
		2: Expected{AffectedRowCount: 1},
		// 3: "insert into person (person_id, name) values (13, 'Sammy')",
		3: Expected{AffectedRowCount: 1},
		// 4: "insert into person (person_id, name) values (14, 'Minnie')",
		4: Expected{AffectedRowCount: 1},
		// 5: "select * from person",
		5: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{11, "Happy"}, []interface{}{12, "Bertie"}, []interface{}{13, "Sammy"}, []interface{}{14, "Minnie"}}},
		// 6: "select * from person where person_id > 5 and person_id < 8",
		// 7: "select * from person where person_id < 3 or person_id > 12",
		7: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{13, "Sammy"}, []interface{}{14, "Minnie"}}},
		// 8: "select * from person where person_id <= 2 or person_id >= 9 limit 4",
		8: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{11, "Happy"}, []interface{}{12, "Bertie"}, []interface{}{13, "Sammy"}, []interface{}{14, "Minnie"}}},
		// 9: "select (person_id + 3*4 - 7%6) as w, name, case(person_id & 1) when 0 then 'even' else 'odd' end as parity from person",
		9: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{22, "Happy", "odd"}, []interface{}{23, "Bertie", "even"}, []interface{}{24, "Sammy", "odd"}, []interface{}{25, "Minnie", "even"}}},
		// 10: "select * from person where length(name) = 5",
		10: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{11, "Happy"}, []interface{}{13, "Sammy"}}},
		// 11: "select upper(name) as u from person where length(name) = 6",
		11: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{"BERTIE"}, []interface{}{"MINNIE"}}},
		// 12: "select substr(name, 3) as x, name from person where length(name) = 5",
		12: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{"ppy", "Happy"}, []interface{}{"mmy", "Sammy"}}},
		// 13: "select lower(name) as x from person where name like '%ie'",
		13: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{"bertie"}, []interface{}{"minnie"}}},
		// 14: "update person set name = 'Beagle' where person_id = 13",
		14: Expected{AffectedRowCount: 1},
		// 15: "update person set name = 'Basset' where name = 'Bertie'",
		15: Expected{AffectedRowCount: 1},
		// 16: "update person set name = 'Basset' where name = 'Puppy'",
		16: Expected{AffectedRowCount: 0},
		// 17: "delete from person where name = 'Happy'",
		17: Expected{AffectedRowCount: 1},
		// 18: "select * from person where person_id = 11",
		// 19: "delete from person where name = 'Happy'",
		19: Expected{AffectedRowCount: 0},
		// 20: "insert into person (person_id, name) values (11, 'Happy')",
		20: Expected{AffectedRowCount: 1},
		// 21: "select * from person where person_id = 11",
		21: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{11, "Happy"}}},
		// 22: "select * from person order by person_id",
		22: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{11, "Happy"}, []interface{}{12, "Basset"}, []interface{}{13, "Beagle"}, []interface{}{14, "Minnie"}}},
		// 23: "select * from person order by person_id desc",
		23: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{14, "Minnie"}, []interface{}{13, "Beagle"}, []interface{}{12, "Basset"}, []interface{}{11, "Happy"}}},
		// 24: "select * from person where person_id > 11 and person_id <= 14",
		24: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{12, "Basset"}, []interface{}{13, "Beagle"}, []interface{}{14, "Minnie"}}},
		// 25: "select sum(person_id) as s from person",
		25: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{50}}},
		// 26: "select count(*) as c from person",
		26: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{4}}},
		// 27: "select min(person_id) as m0, max(person_id) as m1 from person",
		27: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{11, 14}}},
		// 28: "select name, count(person_id) as c from person group by name",
		28: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{"Basset", 1}, []interface{}{"Beagle", 1}, []interface{}{"Happy", 1}, []interface{}{"Minnie", 1}}},
		// 29: "select name, count(person_id) as c from person group by name having count(person_id) > 1",
		// 30: "select name, sum(person_id) s, min(name) m from person group by name having name >= 'Minnie' limit 5",
		30: Expected{AffectedRowCount: 3, Rows: []interface{}{[]interface{}{"Minnie", 14, "Minnie"}}},
		// 31: "insert into person_float (person_id, first, last, age) values (3, 'Fred', 'Flintstone', '48.222')",
		31: Expected{AffectedRowCount: 1},
		// 32: "insert into person_float (person_id, first, last, age) values (4, 'Wilma', 'Flintstone', 39.333)",
		32: Expected{AffectedRowCount: 1},
		// 33: "insert into person_float (person_id, first, last, age) values (5, 'Betty', 'Rubble', 36.777)",
		33: Expected{AffectedRowCount: 1},
		// 34: "insert into person_float (person_id, first, last, age) values (6, 'Barney', 'Rubble', 35.888)",
		34: Expected{AffectedRowCount: 1},
		// 35: "insert into person_float (person_id, first, last, age) values (7, 'Pebbles', 'Flintstone', 3.232)",
		35: Expected{AffectedRowCount: 1},
		// 36: "insert into person_float (person_id, first, last, age) values (8, 'Bambam', 'Rubble', 2.321)",
		36: Expected{AffectedRowCount: 1},
		// 37: "select count(*) from person_float where age > 34.14159 and age < 37.14159",
		37: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{2}}},
		// 38: "select last, count(*) c, min(age) a0, max(age) a1, avg(age) m from person_float group by last",
		// Flintstone|2|39.333|48.222|43.7775
		// Rubble|3|2.321|36.777|24.9953333333333
		38: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{"Flintstone", 2, 39.333, 48.222, 43.7775}, []interface{}{"Rubble", 3, 2.321, 36.777, 24.9953333333333}}},
		// 39: "delete from person_float where age > 39",
		39: Expected{AffectedRowCount: 2},
		// 40: "select count(*) from person_float where age > 34.14159 and age < 37.14159",
		40: Expected{AffectedRowCount: 0, Rows: []interface{}{[]interface{}{2}}},
		// 41: test select *
		41: Expected{AffectedRowCount: 0, Rows: []interface{}{
			[]interface{}{5, "Betty", "Rubble", 36.777},
			[]interface{}{6, "Barney", "Rubble", 35.888},
			[]interface{}{7, "Pebbles", "Flintstone", 3.232},
			[]interface{}{8, "Bambam", "Rubble", 2.321}}},
		//42: "insert into blockchain(blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport,  genesis) values('0xf680ef8555114eb0', 'testsqlchain', 'sql', 'inactive', '0x2aaabf70a9a8752756dbe9fe650221be5525011d', '0x6a660c8e5e5a9e43', '', '', '')",
		42: Expected{AffectedRowCount: 1},

		//43: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain",
		43: Expected{AffectedRowCount: 0, Rows: []interface{}{
			[]interface{}{"0xf680ef8555114eb0", "testsqlchain", "sql", "inactive", "0x2aaabf70a9a8752756dbe9fe650221be5525011d", "0x6a660c8e5e5a9e43", "http://localhost", 22004, "tbd"}},
		},
		//44: "select * from blockchain where owner = '0x2aaabf70a9a8752756dbe9fe650221be5525011d'",
		44: Expected{AffectedRowCount: 0, Rows: []interface{}{
			[]interface{}{"0xf680ef8555114eb0", "testsqlchain", "sql", "inactive", "0x2aaabf70a9a8752756dbe9fe650221be5525011d", "0x6a660c8e5e5a9e43", "http://localhost", 22004, "tbd"}},
		},

		//45: "update blockchain set status='active' where blockchainid = '0xf680ef8555114eb0'",
		45: Expected{AffectedRowCount: 1},

		//46: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain where owner = '0x2aaabf70a9a8752756dbe9fe650221be5525011d'",
		46: Expected{AffectedRowCount: 0, Rows: []interface{}{
			[]interface{}{"0xf680ef8555114eb0", "testsqlchain", "sql", "active", "0x2aaabf70a9a8752756dbe9fe650221be5525011d", "0x6a660c8e5e5a9e43", "http://localhost", 22004, "tbd"}},
		},

		//47: "select * from blockchain where blockchainid = '0xf680ef8555114eb0'",
		47: Expected{AffectedRowCount: 0, Rows: []interface{}{
			[]interface{}{"0xf680ef8555114eb0", "testsqlchain", "sql", "active", "0x2aaabf70a9a8752756dbe9fe650221be5525011d", "0x6a660c8e5e5a9e43", "http://localhost", 22004, "tbd"}},
		},

		//48: "delete from blockchain where blockchainid = '0xf680ef8555114eb0'",
		48: Expected{AffectedRowCount: 1},

		//49: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain where blockchainid = '0xf680ef8555114eb0'",
		49: Expected{AffectedRowCount: 0},
		//50: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain where owner = '0x2aaabf70a9a8752756dbe9fe650221be5525011d'",
		50: Expected{AffectedRowCount: 0},
		//51: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain",
		51: Expected{AffectedRowCount: 0},
	}

	// setup
	// _, err := testSQLConfig()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	owner := "testowner.eth"
	database := make_tmp_name("testdb")
	encrypted := int(0)
	// sdb, err := getTestStateDB_RemoteStorage()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, privateKeys, _, _, nodelist := newWolkPeer(t, defaultConsensus)
	defer ReleaseNodes(t, nodelist)
	time.Sleep(1 * time.Second)
	//currentTS := time.Now().Unix()
	tprint("peered!")
	wolkStore := wolk[0]
	privateKey := privateKeys[0]
	tx, err := NewTransaction(privateKey, http.MethodPost, path.Join(owner, database), &TxBucket{Quota: MinimumQuota})
	if err != nil {
		t.Fatal(err)
	}
	doTransaction(t, wolkStore, tx)
	ctx := context.TODO()
	sdb, err := NewStateDB(ctx, wolkStore.Storage, wolkStore.LastKnownBlock().Hash())
	if err != nil {
		t.Fatalf("newStateDB %v", err)
	}
	// create database
	tReq := new(SQLRequest)
	tReq.RequestType = RT_CREATE_DATABASE
	tReq.Owner = owner
	tReq.Database = database
	tReq.Encrypted = encrypted
	fmt.Printf("[wvm_test] Create Database req: %+v\n", tReq)
	//mReq, _ := json.Marshal(tReq)
	res, _, err := sdb.SelectHandler(tReq, false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("[wvm_test] Create Database resp: %+v\n", res)
	//TODO: check if it's working
	var wg *sync.WaitGroup
	err = sdb.Flush(ctx, wg, true)
	if err != nil {
		log.Error("Error flushing chunks", "error", err)
		t.Fatalf(fmt.Sprintf("[backend:ApplyTransaction] Error flushing chunks %s", err))
	}

	db, ok, err := sdb.GetDatabase(owner, database)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("Database doesn't exist\n")
	}
	fmt.Printf("[wvm_test] Database gotten: %s\n", db.Name)

	var sqlsetup []string
	sqlsetup = make([]string, 3)
	sqlsetup[0] = "create table person (person_id int primary key, name string)"
	sqlsetup[1] = "create table person_float (person_id int primary key, first string, last string, age float)"
	sqlsetup[2] = "create table blockchain (blockchainid string primary key, blockchainname string, blockchaintype string, blockchainstatus string, owner string, tokenid string, rpcaddr string, rpcport string, genesis string)"

	dbpath := db.GetSchemaPath()
	// create table
	for _, sql := range sqlsetup {
		fmt.Printf("[wvm_test] %s\n", sql)
		rows, affectedrowct, _ /*proof*/, err := sdb.Query(owner, database, sql)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) == 0 && affectedrowct == 1 {
			fmt.Printf("[wvm_test] %s. resp: %v, affectedrowct: %v\n\n", sql, rows, affectedrowct)
		} else {
			t.Fatalf("Failed : %s", sql)
		}
	}

	// setup schema
	//dbpath := "./tmp/testdb/wvm.db" // if using custom wvmc/wvmc_test.go schema
	fmt.Printf("[wvm_test] using schema path: %s\n", dbpath)
	fmt.Printf("[wvm_test] schema used: %+v\n", db.Schema)
	vm := NewWVM(sdb, db)
	process := true

	// execute and test results
	for i := 0; i <= 41; i++ {

		// tests to skip (for future implementation)
		if reason, ok := skiptests[i]; ok {
			fmt.Printf("\n[wvm_test] SKIPPING test %d. reason: %s\n", i, reason)
			continue
		}

		testFileLocation := "./tests/testdb"
		filename := fmt.Sprintf("wvmc-%d.json", i)
		file := filepath.Join(testFileLocation, filename)
		fmt.Printf("\n[wvm_test] *****Executing %s\n", file)
		ops, err := loadOps(file)
		if err != nil {
			t.Fatalf("Could not load ops %s", file)
		}

		vm.Reset()
		if !process { //looking for unimplemented opcodes
			err = vm.CheckOps(ops)
			if err != nil {
				t.Fatalf("[wvm_test] vm.CheckOps err: %s", err)
			}
			continue
		}
		fmt.Printf("[wvm_test] ops[0].P4 (%s)\n", ops[0].P4)
		if strings.Contains(strings.ToLower(ops[0].P4), "update") {
			vm.SetFlags("update")
		}
		err = vm.ProcessOps(ops)
		if err != nil {
			//t.Fatalf("ERROR test %v: %s \nOPS: %+v\n", i, err, ops)
			t.Fatalf("ERROR test %v: %s", i, err)
		}
		//nrows := dumpRows(vm.Rows)
		fmt.Printf("\n[wvm_test] **test %v actual: {Rows:%+v AffectedRowCount:%+v}\n", i, vm.Rows, vm.AffectedRowCount)
		fmt.Printf("[wvm_test] **test %v expected: %+v\n", i, expected[i])
		if !matchRows(vm.Rows, expected[i]) || vm.AffectedRowCount != expected[i].AffectedRowCount {
			fmt.Printf("\n")
			t.Fatalf("Failed test %v", i)
		}
		//TODO: check if it's working
		err = sdb.Flush(ctx, wg, true)
		if err != nil {
			log.Error("Error flushing chunks", "error", err)
			t.Fatalf(fmt.Sprintf("[backend:ApplyTransaction] Error flushing chunks %s", err))
		}
	}

	// clean up
	fmt.Printf("\n[wvm_test] *****CLEAN UP\n")
	o, ok, err := sdb.GetOwner(owner)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("owner %s doesn't exist", owner)
	}
	for db, _ := range o.DatabaseNames {
		ok, err := sdb.DropDatabase(owner, db)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("Database %s to drop doesn't exist!", db)
		} else {
			fmt.Printf("[wvm_test] database %s is deleted\n", db)
		}
	}

}

func make_tmp_name(prefix string) (nm string) {
	return fmt.Sprintf("%s%d", prefix, int32(time.Now().Unix()))
}

func loadOps(fn string) (ops []Op, err error) {
	f, err := os.Open(fn)
	if err != nil {
		return ops, fmt.Errorf("Could not open %s", fn)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		op_in := scanner.Text()
		var op Op
		err := json.Unmarshal([]byte(op_in), &op)
		if err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

//if actual contains the fields in expected, and the num of expected rows are found, then this passes.  Actual could have more fields per row and more rows.
//TODO: this should probably preserve order?
func matchRows(actual [][]interface{}, expected Expected) (ok bool) {
	if len(actual) != len(expected.Rows) {
		log.Info(fmt.Sprintf("MatchRows Fail: lengths mismatch - actual length = [%d] expected length = [%d]", len(actual), len(expected.Rows)))
		return false
	}
	if len(actual) == 0 {
		return true
	}

	foundrows := 0
	er_marshalled, _ := json.Marshal(expected.Rows)
	//fmt.Printf("--er marshalled %+v\n", string(er_marshalled))
	var expRows [][]interface{}
	json.Unmarshal(er_marshalled, &expRows)
	//fmt.Printf("--expRows %+v\n", expRows)
	for _, exprow := range expRows {
		for _, actrow := range actual {
			actbyte, _ := json.Marshal(actrow)
			actstr := string(actbyte)
			//fmt.Printf("--considering actual row: %s\n", actstr)
			founditems := 0
			for _, item := range exprow {
				expbyte, _ := json.Marshal(item)
				expstr := string(expbyte)
				//fmt.Printf("--considering expected item: %s vs actual item: %s Found: %d\n", expstr, actstr, founditems)
				if strings.Contains(actstr, expstr) {
					founditems++
					//fmt.Printf("--item found\n")
				}
			}
			//fmt.Printf("\nRowItems: Found: %d vs Expected: %d", founditems, len(exprow))
			if founditems == len(exprow) {
				//fmt.Printf("--row found\n")
				foundrows++
			}
		}
	}
	//fmt.Printf("Rows: Found: %d vs Expected: %d", foundrows, len(expected.Rows))
	if foundrows == len(expected.Rows) {
		return true
	}

	return false
}

// func dumpRows(rows [][]interface{}) (nrows int) {
// 	fmt.Printf("dumping rows: \n")
// 	nrows = 0
// 	for i, r := range rows {
// 		fmt.Printf("ROW %d: %v\n", i, r)
// 		nrows++
// 	}
// 	return nrows
// }
