package wvmc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	//wvmc "github.com/wolkdb/plasma/sqlchain/wvmc"
)

const DBPATH = "../tests/testdb"

func TestSetupSchema(t *testing.T) {

	dbpath := filepath.Join(DBPATH, "test.db")
	os.MkdirAll(DBPATH, os.ModePerm) //make the testdb dir if it doesn't exist

	sqlsetups1 := []string{
		"create table person (person_id int primary key, name string)",
	}
	fmt.Printf("[wvmc_test] sqlsetups1: %v\n", sqlsetups1)
	SetupSchema(dbpath, sqlsetups1)
	bytes, _ := ioutil.ReadFile(dbpath)
	s := string(bytes)
	fmt.Printf("[wvmc_test] schema created: \n%s\n", bytes)
	if !strings.Contains(s, "indexsqlite_autoindex_person_1") || !strings.Contains(s, "tableperson") {
		t.Fatalf("sqlsetups1 no match")
	}

	sqlsetups2 := []string{
		"create table person_float (person_id int primary key, first string, last string, age float)",
	}
	fmt.Printf("[wvmc_test] sqlsetups2: %v\n", sqlsetups2)
	SetupSchema(dbpath, sqlsetups2)
	bytes, _ = ioutil.ReadFile(dbpath)
	s = string(bytes)
	fmt.Printf("[wvmc_test] schema created: \n%s\n", bytes)
	if !strings.Contains(s, "indexsqlite_autoindex_person_1") || !strings.Contains(s, "tableperson") || !strings.Contains(s, "indexsqlite_autoindex_person_float_1") || !strings.Contains(s, "tableperson_float") {
		t.Fatalf("sqlsetups2 no match")
	}

	sqlsetups3 := []string{
		"create table person (person_id int primary key, name string)",
		"create table person_float (person_id int primary key, first string, last string, age float)",
	}
	fmt.Printf("[wvmc_test] sqlsetups3: %v\n", sqlsetups3)
	SetupSchema(dbpath, sqlsetups3)
	bytes, _ = ioutil.ReadFile(dbpath)
	fmt.Printf("[wvmc_test] schema created: \n%s\n", bytes)
	if string(bytes) != s {
		t.Fatalf("sqlsetups3 does not match sqlsetups2")
	}
}

func TestGetMasterTable(t *testing.T) {

	dbpath := filepath.Join(DBPATH, "test.db")
	schemajson, err := GetMasterTable(dbpath)
	if err != nil {
		t.Fatalf("[wvmc_test] %s", err)
	}
	fmt.Printf("[wvmc_test] schemajson: %s\n", schemajson)

	for _, sline := range schemajson {
		if strings.Contains(string(sline), `"rootpage":5`) {
			return
		}
	}
	t.Fatalf("master table doesn't have enough rootpages")
}

func TestClearSchema(t *testing.T) {

	dbpath := filepath.Join(DBPATH, "test.db")
	err := os.Remove(dbpath)
	if err != nil {
		t.Fatalf("[wvmc_test] %s", err)
	}
	bytes, err := ioutil.ReadFile(dbpath)
	if err == nil || len(bytes) > 0 {
		t.Fatalf("[wvmc_test] schema was not removed")
	}
}

func TestWVMC(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}
	dbpath := filepath.Join(DBPATH, "test.db")
	os.MkdirAll(DBPATH, os.ModePerm)

	sqlsetups := []string{
		"create table person (person_id int primary key, name string)",
		//"create table person_noprimary (person_id int, name string)",
		"create table person_float (person_id int primary key, first string, last string, age float)",
		"create table blockchain (blockchainid string primary key, blockchainname string, blockchaintype string, blockchainstatus string, owner string, tokenid string, rpcaddr string, rpcport string, genesis string)",
	}
	SetupSchema(dbpath, sqlsetups)

	sqltests := map[int]string{
		0: "insert into person (person_id, name) values (11, 'Happy')",
		1: "select * from person where person_id = 11",
		2: "insert into person (person_id, name) values (12, 'Bertie')",
		3: "insert into person (person_id, name) values (13, 'Sammy')",
		4: "insert into person (person_id, name) values (14, 'Minnie')",
		5: "select * from person",
		6: "select * from person where person_id > 5 and person_id < 8",
		7: "select * from person where person_id < 3 or person_id > 12",
		8: "select * from person where person_id <= 2 or person_id >= 9 limit 4",
		9: "select (person_id + 3*4 - 7%6) as w, name, case(person_id & 1) when 0 then 'even' else 'odd' end as parity from person",

		// function
		10: "select * from person where length(name) = 5",
		11: "select upper(name) as u from person where length(name) = 6",
		12: "select substr(name, 3) as x, name from person where length(name) = 5",
		13: "select lower(name) as x from person where name like '%ie'",

		// update
		14: "update person set name = 'Beagle' where person_id = 13",
		15: "update person set name = 'Basset' where name = 'Bertie'",
		16: "update person set name = 'Basset' where name = 'Puppy'",

		// delete
		17: "delete from person where name = 'Happy'",
		18: "select * from person where person_id = 11",
		19: "delete from person where name = 'Happy'",

		// primary key
		20: "insert into person (person_id, name) values (11, 'Happy')",
		21: "select * from person where person_id = 11",
		22: "select * from person order by person_id",
		23: "select * from person order by person_id desc",
		//24: "select * from person where person_id >= 5 and person_id <= 8",
		24: "select * from person where person_id > 11 and person_id <= 14",

		// aggregates
		25: "select sum(person_id) as s from person",
		26: "select count(*) as c from person",
		27: "select min(person_id) as m0, max(person_id) as m1 from person",
		28: "select name, count(person_id) as c from person group by name",
		29: "select name, count(person_id) as c from person group by name having count(person_id) > 1",
		30: "select name, sum(person_id) s, min(name) m from person group by name having name >= 'Minnie' limit 5",

		// floats
		31: "insert into person_float (person_id, first, last, age) values (3, 'Fred', 'Flintstone', '48.222')",
		32: "insert into person_float (person_id, first, last, age) values (4, 'Wilma', 'Flintstone', 39.333)",
		33: "insert into person_float (person_id, first, last, age) values (5, 'Betty', 'Rubble', 36.777)",
		34: "insert into person_float (person_id, first, last, age) values (6, 'Barney', 'Rubble', 35.888)",
		35: "insert into person_float (person_id, first, last, age) values (7, 'Pebbles', 'Flintstone', 3.232)",
		36: "insert into person_float (person_id, first, last, age) values (8, 'Bambam', 'Rubble', 2.321)",
		37: "select count(*) from person_float where age > 34.14159 and age < 37.14159",
		38: "select last, count(*) c, min(age) a0, max(age) a1, avg(age) m from person_float group by last",
		39: "delete from person_float where age > 39",
		40: "select count(*) from person_float where age > 34.14159 and age < 37.14159",
		41: "select * from person_float",

		//blockchain
		// 42: "insert into blockchain(blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport,  genesis) values('0xf680ef8555114eb0', 'testsqlchain', 'sql', 'inactive', '0x2aaabf70a9a8752756dbe9fe650221be5525011d', '0x6a660c8e5e5a9e43', 'http://localhost', '22004', 'tbd')",
		// 43: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain",
		// 44: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain where owner = '0x2aaabf70a9a8752756dbe9fe650221be5525011d'",
		// 45: "update blockchain set blockchainstatus='active' where blockchainid = '0xf680ef8555114eb0'",
		// 46: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain where owner = '0x2aaabf70a9a8752756dbe9fe650221be5525011d'",
		// 47: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain where blockchainid = '0xf680ef8555114eb0'",
		// 48: "delete from blockchain where blockchainid = '0xf680ef8555114eb0'",
		// 49: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain where blockchainid = '0xf680ef8555114eb0'",
		// 50: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain where owner = '0x2aaabf70a9a8752756dbe9fe650221be5525011d'",
		// 51: "select blockchainid, blockchainname, blockchaintype, blockchainstatus, owner, tokenid, rpcaddr, rpcport, genesis from blockchain",
	}
	//for i, sql := range sqltests {
	for i := 0; i < len(sqltests); i++ {
		sql := sqltests[i]
		ops_byte, err := GetOps(dbpath, sql)
		if err != nil {
			t.Fatalf("Compile error: dbpath: (%s) %s => %v\n", dbpath, sql, err)
		}
		var ops []Op
		if err = json.Unmarshal(ops_byte, &ops); err != nil {
			t.Fatalf("Unmarshal error: %s => %v\n", string(ops_byte), err)
		}

		/*
			var ops [] Op
			for _, op_i := range ops_interface {
				var op  Op
				if err = json.Unmarshal(op_i.(byte), op)
				ops = append(ops, )
			}
		*/
		name := fmt.Sprintf("wvmc-%d.json", i)
		fn := filepath.Join(DBPATH, name)

		f, err := os.Create(fn)
		if err != nil {
			t.Fatalf("Could not create %s", fn)
		}
		defer f.Close()
		for _, op := range ops {
			op_out, err := json.Marshal(op)
			if err != nil {
				t.Fatalf("Marshal error: %s => %v\n", sql, err)
			}
			fmt.Printf("%s\n", string(op_out))
			f.WriteString(fmt.Sprintf("%s\n", op_out))
		}
		f.Sync()
	}

	err := os.Remove(dbpath)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := ioutil.ReadFile(dbpath)
	if err == nil || len(bytes) > 0 {
		t.Fatalf("schema was not removed")
	}
	fmt.Printf("[wvmc_test] %s schema removed\n", dbpath)

}
