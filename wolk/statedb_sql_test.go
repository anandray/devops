package wolk

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
)

func TestStateDB_StateObject(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, _, _, _, nodelist := newWolkPeer(t, defaultConsensus)
	defer ReleaseNodes(t, nodelist)
	time.Sleep(1 * time.Second)
	tprint("peered!")
	ws := wolk[0]
	ctx := context.TODO()
	sdb, err := NewStateDB(ctx, ws.Storage, ws.LastKnownBlock().Hash())
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	rawKey := []byte("randomkey")
	rawVal := []byte("randomval")
	keyHash := common.BytesToHash(wolkcommon.Computehash(rawKey))
	valHash := common.BytesToHash(wolkcommon.Computehash(rawVal))
	expected_so := NewStateObject(sdb, keyHash, valHash)

	tprint("after NewStateObject: wolkstore state objects: %+v", sdb.sql.stateObjects)

	err = sdb.updateStateObject(ctx, expected_so)
	if err != nil {
		t.Fatal(err)
	}

	tprint("after updateStateObject: wolkstore state objects: %+v", sdb.sql.stateObjects)

	// make sure we aren't reading the cached state object
	sdb.sql.stateObjects = make(map[common.Hash]*stateObject)

	tprint("after cleanup: wolkstore state objects: %+v", sdb.sql.stateObjects)

	actual_so, _, ok, err := sdb.getStateObject(ctx, keyHash)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("NOT OK, didn't find this stateobj in keystorage")
	}

	tprint("after getStateObject: wolkstore state objects: %+v", sdb.sql.stateObjects)

	if expected_so.val != actual_so.val {
		t.Fatalf("expected state obj val(%v) != actual state obj val(%v)", expected_so.val, actual_so.val)
	}
	tprint("passed")

}

func TestDBChunk(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, _, _, _, nodelist := newWolkPeer(t, defaultConsensus)
	defer ReleaseNodes(t, nodelist)
	time.Sleep(1 * time.Second)
	tprint("peered!")
	ws := wolk[0]

	//currentTS := time.Now().Unix()
	blk := ws.LastKnownBlock()
	tprint("current block (%+v)", blk)
	lbn, _ := ws.LatestBlockNumber()
	if blk.BlockNumber != lbn {
		t.Fatalf("LastKnownBlock.Number(%v) != LatestBLockNumber(%v)", blk.BlockNumber, lbn)
	}

	ctx := context.TODO()
	sdb, err := NewStateDB(ctx, ws.Storage, blk.Hash())
	if err != nil {
		t.Fatalf("newStateDB %v\n", err)
	}
	//_, expected := generateRandomData(int(64))
	expected := []byte("randomdata")
	tprint("chunkData (%s) (%x)", expected, expected)
	encrypted := 0
	expectedhash, err := sdb.SetDBChunk(expected, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	tprint("done sleeping. current block (%v)", ws.LastKnownBlock())
	tprint("expectedhash: %v", expectedhash)
	actual, ok, err := sdb.GetDBChunk(expectedhash)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		tprint("getdbchunk did not get (%v) chunk!", expectedhash)
		t.Fatalf("NO chunk gotten!")
	}
	tprint("chunk gotten (%s) (%x)", actual, actual)
	if string(actual) != string(expected) {
		t.Fatalf("actual (%s) != expected (%s), fail.", actual, expected)
	}
	tprint("passed")

}

func TestStateDB_SQL(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, privateKeys, _, _, _ := newWolkPeer(t, defaultConsensus)
	time.Sleep(1 * time.Second)
	tprint("peered!")
	wolkStore := wolk[0]
	privateKey := privateKeys[0]

	requests, expected, owner := getRequests_statedb_sql(t)

	tprint("setting owner %s", owner)
	setName(t, owner, privateKey, wolkStore)
	tprint("OWNER SET")

	ctx := context.TODO()
	sdb, err := NewStateDB(ctx, wolkStore.Storage, wolkStore.LastKnownBlock().Hash())
	if err != nil {
		t.Fatal(err)
	}

	// run tests
	for i := 0; i < len(requests); i++ {
		if _, ok := requests[i]; !ok {
			continue
		}
		req := requests[i]
		tprint("** req %v: %+v", i, requests[i])
		actual, _, err := sdb.SelectHandler(&req, false)
		if err != nil {
			if i == 16 && strings.Contains(err.Error(), "Table does not exist") {
				tprint("resp: %+v", err.Error())
				continue
			}
			tprint("SelectHandler req(%d) err!", i)
			t.Fatal(err)
		}
		tprint("resp: %+v", actual)
		// err = sdb.Storage.Flush()
		// if err != nil {
		// 	t.Fatal(err)
		// }
		time.Sleep(5 * time.Second) // just in case

		if _, ok := expected[i]; !ok {
			continue
		}
		if !reflect.DeepEqual(actual, expected[i]) {
			if matchRowData(actual.Data, expected[i].Data) && actual.AffectedRowCount == expected[i].AffectedRowCount && actual.MatchedRowCount == expected[i].MatchedRowCount {
				//passed
			} else {
				tprint("expected: %+v", expected[i])
				t.Fatalf("test %v failed", i)
			}
		}
	}
}

func getRequests_statedb_sql2(t *testing.T) (requests map[int]SQLRequest, expected map[int]SQLResponse, owner string) {

	owner = makeName("testowner-debug")
	dbname := makeName("newdb")
	tablename := "testtable"
	encrypted := 0
	tprint("o: %s, db %s, tbl: %s", owner, dbname, tablename)

	columns := []Column{
		Column{
			ColumnName: "person_id",
			ColumnType: CT_INTEGER,
			IndexType:  IT_BPLUSTREE,
			Primary:    0,
		},
		Column{
			ColumnName: "name",
			ColumnType: CT_STRING,
			IndexType:  IT_BPLUSTREE,
			Primary:    1,
		},
	}

	requests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_CREATE_DATABASE},
		1: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, Columns: columns, RequestType: RT_CREATE_TABLE},
		2: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (6, 'dinozzo')"},
		3: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (23, 'ziva')"},
		4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename},
		5: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "delete from " + tablename + " where name = 'dinozzo'"},
		6: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename},
	}

	expected = map[int]SQLResponse{
		4: {Data: []Row{
			Row{"name": "dinozzo", "person_id": 6},
			Row{"name": "ziva", "person_id": 23}},
			MatchedRowCount: 2},
		6: {Data: []Row{
			Row{"name": "ziva", "person_id": 23}},
			MatchedRowCount: 1},
	}

	return requests, expected, owner
}

func getRequests_statedb_sql(t *testing.T) (requests map[int]SQLRequest, expected map[int]SQLResponse, owner string) {

	owner = makeName("testowner-debug")
	dbname := makeName("newdb")
	tablename := "testtable"
	encrypted := 0

	tprint("o: %s, db %s, tbl: %s", owner, dbname, tablename)

	columns := []Column{
		Column{
			Primary:    1,
			ColumnName: "email",
			ColumnType: CT_STRING,
			IndexType:  IT_BPLUSTREE,
		},
		Column{
			ColumnName: "age",
			ColumnType: CT_INTEGER,
			IndexType:  IT_BPLUSTREE,
			Primary:    0,
		},
	}

	putrows := []Row{
		Row{"email": "test05@wolk.com", "age": 1},
		Row{"email": "test06@wolk.com", "age": 2},
		Row{"email": "test07@wolk.com", "age": 3},
		Row{"email": "test08@wolk.com", "age": 4},
		Row{"email": "test09@wolk.com", "age": 5},
	}

	requests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_CREATE_DATABASE},
		1: SQLRequest{Owner: owner, Encrypted: encrypted, RequestType: RT_LIST_DATABASES},
		2: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, Columns: columns, RequestType: RT_CREATE_TABLE},

		3: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: RT_DESCRIBE_TABLE},
		4: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: RT_PUT, Rows: putrows},
		5: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: RT_GET, Key: "test06@wolk.com"},
		6: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename},
		7: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (email, age) values ('test10@wolk.com', 6)"},
		8: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "update " + tablename + " set age = 23 where email = 'test07@wolk.com'"},
		9: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where age = 6"},
		//10: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "delete from " + tablename + " where age = 5"},
		10: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "delete from " + tablename + " where email = 'test09@wolk.com'"},
		11: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where age = 5"},
		12: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where email = 'test09@wolk.com'"},
		13: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where email = 'test08@wolk.com'"},
		14: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename},
		15: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: RT_DROP_TABLE},
		16: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename},
		// test for insert nils:   SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType:  RT_QUERY, RawQuery: "insert into " + tablename + " (email) values ('test22@wolk.com')"},

		17: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "create table testtableone (email string, person_id int primary key,  name string)"},
		18: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into testtableone (person_id, email, name) values (1001, '1001@wolk.com', 'Gibbs')"},
		19: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into testtableone (email, name, person_id) values ('1002@wolk.com', 'DiNozzo', 1002)"},
		20: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from testtableone"},
		//20:   SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType:  RT_QUERY, RawQuery: "select name, person_id from testtableone where email = '1001@wolk.com'"},

		21: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "create table testtabletwo (person_id int, name string primary key, email string)"},
		22: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into testtabletwo (person_id, email, name) values (1001, '1001@wolk.com', 'Gibbs')"},
		23: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into testtabletwo (email, name, person_id) values ('1002@wolk.com', 'DiNozzo', 1002)"},
		24: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select name, person_id from testtabletwo where email = '1001@wolk.com'"},
		25: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select email, name from testtabletwo where name = 'DiNozzo'"},

		// clean up
		26: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_LIST_TABLES},
		27: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_DROP_DATABASE},
		28: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_LIST_DATABASES},
	}

	expected = map[int]SQLResponse{
		0: {AffectedRowCount: 1, MatchedRowCount: 0},
		//1: {Data: []Row{Row{"database": dbname}}, AffectedRowCount: 0, MatchedRowCount: 1},
		2: {AffectedRowCount: 1, MatchedRowCount: 0},

		3:  {Data: []Row{Row{"ColumnName": "age", "IndexType": "BPLUS", "Primary": 0, "ColumnType": "INTEGER"}, Row{"ColumnName": "email", "IndexType": "BPLUS", "Primary": 1, "ColumnType": "STRING"}}, AffectedRowCount: 0, MatchedRowCount: 0},
		4:  {AffectedRowCount: 5, MatchedRowCount: 0},
		5:  {Data: []Row{Row{"age": 2, "email": "test06@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		6:  {Data: []Row{Row{"email": "test05@wolk.com", "age": 1}, Row{"age": 2, "email": "test06@wolk.com"}, Row{"email": "test07@wolk.com", "age": 3}, Row{"email": "test08@wolk.com", "age": 4}, Row{"email": "test09@wolk.com", "age": 5}}, AffectedRowCount: 0, MatchedRowCount: 5},
		7:  {AffectedRowCount: 1, MatchedRowCount: 0},
		8:  {AffectedRowCount: 1, MatchedRowCount: 0},
		9:  {Data: []Row{Row{"email": "test10@wolk.com", "age": 6}}, AffectedRowCount: 0, MatchedRowCount: 1},
		10: {AffectedRowCount: 1, MatchedRowCount: 0},
		11: {AffectedRowCount: 0, MatchedRowCount: 0},
		12: {AffectedRowCount: 0, MatchedRowCount: 0},
		13: {Data: []Row{Row{"age": 4, "email": "test08@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		14: {Data: []Row{Row{"email": "test05@wolk.com", "age": 1}, Row{"email": "test06@wolk.com", "age": 2}, Row{"email": "test07@wolk.com", "age": 23}, Row{"email": "test08@wolk.com", "age": 4}, Row{"email": "test10@wolk.com", "age": 6}}, AffectedRowCount: 0, MatchedRowCount: 5},
		15: {AffectedRowCount: 1, MatchedRowCount: 0},
		// 16: should be an error

		17: {AffectedRowCount: 1, MatchedRowCount: 0},
		18: {AffectedRowCount: 1, MatchedRowCount: 0},
		19: {AffectedRowCount: 1, MatchedRowCount: 0},
		20: {Data: []Row{Row{"person_id": 1001, "name": "Gibbs", "email": "1001@wolk.com"}, Row{"person_id": 1002, "name": "DiNozzo", "email": "1002@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 2},

		21: {AffectedRowCount: 1, MatchedRowCount: 0},
		22: {AffectedRowCount: 1, MatchedRowCount: 0},
		23: {AffectedRowCount: 1, MatchedRowCount: 0},
		24: {Data: []Row{Row{"name": "Gibbs", "person_id": 1001}}, AffectedRowCount: 0, MatchedRowCount: 1},
		25: {Data: []Row{Row{"name": "DiNozzo", "email": "1002@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},

		26: {Data: []Row{Row{"table": "testtableone"}, Row{"table": "testtabletwo"}}, AffectedRowCount: 0, MatchedRowCount: 2},
		27: {AffectedRowCount: 1, MatchedRowCount: 0},
		//28: {AffectedRowCount: 0, MatchedRowCount: 0},
	}
	return requests, expected, owner
}
