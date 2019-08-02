// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/log"
)

// TODO: negative test
// - 0 txn in the pool, what happens
// - putting bad txn in
// - putting all bad txns in until pool is empty

func TestBackend_SQL(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Setup
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, privateKeys, _, _, _ := newWolkPeer(t, consensusSingleNode)
	time.Sleep(1 * time.Second)
	tprint("peered!")
	wolkStore := wolk[0]
	privateKey := privateKeys[0]

	mutatereqs, readreqs, readexpected, _, owner, _ := getRequests_backend_sql(t)
	tprint("setting owner %s", owner)
	setName(t, owner, privateKey, wolkStore)
	tprint("OWNER SET")

	//sql mutate requests
	txhashes := make([]common.Hash, 0)
	var i int
	for i = 0; i < len(mutatereqs)-1; i++ {
		req := mutatereqs[i]
		tprint("*** mutate req %d: %+v", i, req)
		txhash := mutate(t, &req, privateKey, wolkStore)
		txhashes = append(txhashes, txhash)
	}
	// mint new block
	tprint("waiting for mutate txns...")
	waitForTxsDone(t, txhashes, wolkStore)
	time.Sleep(5000 * time.Millisecond)
	tprint("mutate txns are done")

	// this one is an update / delete
	req := mutatereqs[i]
	tprint("*** mutate req %d: %+v", i, req)
	txhash := mutate(t, &req, privateKey, wolkStore)
	txhashes = append(txhashes, txhash)
	//mint new block
	tprint("waiting for mutate txn (%x)...", txhash)
	waitForTxDone(t, txhash, wolkStore)
	tprint("finished waiting for mutate txn")

	// sql read requests and check
	for i := 0; i < len(readreqs); i++ {

		req := readreqs[i]
		tprint("*** read req %d: %+v", i, req)
		options := NewRequestOptions()
		actual, err := wolkStore.Read(&req, options)
		if err != nil {
			t.Fatal(err)
		}
		tprint("(%d) result: %+v", i, actual)

		if _, ok := readexpected[i]; !ok { //skip tests w/o expected to match
			continue
		}
		if !reflect.DeepEqual(actual, readexpected[i]) {
			if matchRowData(actual.Data, readexpected[i].Data) && actual.AffectedRowCount == readexpected[i].AffectedRowCount && actual.MatchedRowCount == readexpected[i].MatchedRowCount {
				//passed
				tprint("(%d) passed.", i)
			} else {
				tprint("(%d) expected: %+v", i, readexpected[i])
				t.Fatalf("(%d) test failed", i)
			}
		}
	}

}

func mutate(t *testing.T, req *SQLRequest, privateKey *ecdsa.PrivateKey, wolkStore *WolkStore) (txhash common.Hash) {
	tx, err := NewTransaction(privateKey, http.MethodPost, path.Join(req.Owner, req.Database, "SQL"), req)
	if err != nil {
		t.Fatal(err)
	}
	txhash, err = wolkStore.SendRawTransaction(tx)
	if err != nil {
		t.Fatal(err)
	}
	tprint("tx hash: %x", txhash)
	return txhash

}

func getRequests_backend_sql(t *testing.T) (mutaterequests map[int]SQLRequest, readrequests map[int]SQLRequest, readexpected map[int]SQLResponse, cleanuprequests map[int]SQLRequest, owner string, dbname string) {
	owner = makeName("testowner-debug")
	dbname = makeName("testdb")
	tablename := makeName("newtable")
	encrypted := 0

	columns := []Column{
		Column{
			ColumnName: "email",
			ColumnType: CT_STRING,
			IndexType:  IT_BPLUSTREE,
			Primary:    1,
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

	mutaterequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_CREATE_DATABASE},
		1: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, Columns: columns, RequestType: RT_CREATE_TABLE},
		2: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: RT_PUT, Rows: putrows},
		3: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (email, age) values ('test10@wolk.com', 6)"},
		4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "update " + tablename + " set age = 23 where email = 'test07@wolk.com'"},
	}

	readrequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Encrypted: encrypted, RequestType: RT_LIST_DATABASES},
		1: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: RT_DESCRIBE_TABLE},
		2: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, RequestType: RT_GET, Key: "test06@wolk.com"},
		3: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename},
		4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where age = 6"},
	}

	readexpected = map[int]SQLResponse{
		//0: {Data: []Row{Row{"database": dbname}}, MatchedRowCount: 1},
		1: {Data: []Row{Row{"ColumnName": "age", "IndexType": "BPLUS", "Primary": 0, "ColumnType": "INTEGER"}, Row{"ColumnName": "email", "IndexType": "BPLUS", "Primary": 1, "ColumnType": "STRING"}}, AffectedRowCount: 0, MatchedRowCount: 0},
		2: {Data: []Row{Row{"age": 2, "email": "test06@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		3: {Data: []Row{
			Row{"age": 1, "email": "test05@wolk.com"},
			Row{"age": 2, "email": "test06@wolk.com"},
			Row{"age": 23, "email": "test07@wolk.com"},
			Row{"age": 4, "email": "test08@wolk.com"},
			Row{"age": 5, "email": "test09@wolk.com"},
			Row{"age": 6, "email": "test10@wolk.com"}},
			AffectedRowCount: 0,
			MatchedRowCount:  6},
		4: {Data: []Row{Row{"age": 6, "email": "test10@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
	}

	cleanuprequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_DROP_DATABASE},
	}

	return mutaterequests, readrequests, readexpected, cleanuprequests, owner, dbname
}

func getRequests_backend_sql2(t *testing.T) (mutaterequests map[int]SQLRequest, readrequests map[int]SQLRequest, readexpected map[int]SQLResponse, cleanuprequests map[int]SQLRequest, owner string, dbname string) {
	owner = MakeName("testowner-debug")
	dbname = MakeName("testdb")
	tablename := "testtable"
	encrypted := 0

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

	mutaterequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_CREATE_DATABASE},
		1: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, Columns: columns, RequestType: RT_CREATE_TABLE},
		2: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (6, 'dinozzo')"},
		3: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (23, 'ziva')"},
		4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "delete from " + tablename + " where name = 'dinozzo'"},

		// 4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (2, 'mcgee')"},
		// 5: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (101, 'gibbs')"},
	}

	readrequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename},
		// 1: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id = 6"},
		// 2: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id > 6"},
		// 3: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id >= 6"},
		// 4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id > 7"},
		// 5: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id >= 7"},
		// 6: SQLRequest{Owner: owner, Encrypted: encrypted, RequestType: RT_LIST_DATABASES, RawQuery: "select * from " + tablename},
	}

	readexpected = map[int]SQLResponse{
		0: {Data: []Row{
			Row{"name": "ziva", "person_id": 23}},
			MatchedRowCount: 1},
		// 0: {Data: []Row{
		// 	Row{"name": "dinozzo", "person_id": 6},
		// 	Row{"name": "ziva", "person_id": 23},
		// 	Row{"name": "gibbs", "person_id": 101}},
		// 	MatchedRowCount: 3},

		// 	//0: {Data: [] Row{ Row{"database": dbname}}},
		// 	1: {Data: []wolk.Row{wolk.Row{"ColumnName": "age", "IndexType": "BPLUS", "Primary": 0, "ColumnType": "INTEGER"}, Row{"ColumnName": "email", "IndexType": "BPLUS", "Primary": 1, "ColumnType": "STRING"}}, AffectedRowCount: 0, MatchedRowCount: 0},
		// 	2: {Data: []wolk.Row{wolk.Row{"age": 2, "email": "test06@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		// 	3: {Data: []wolk.Row{
		// 		wolk.Row{"age": 1, "email": "test05@wolk.com"},
		// 		wolk.Row{"age": 2, "email": "test06@wolk.com"},
		// 		wolk.Row{"age": 23, "email": "test07@wolk.com"},
		// 		wolk.Row{"age": 4, "email": "test08@wolk.com"},
		// 		wolk.Row{"age": 5, "email": "test09@wolk.com"},
		// 		wolk.Row{"age": 6, "email": "test10@wolk.com"}},
		// 		AffectedRowCount: 0,
		// 		MatchedRowCount:  6},
		// 	4: {Data: []wolk.Row{wolk.Row{"age": 6, "email": "test10@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		// 	5: {Data: []wolk.Row{
		// 		wolk.Row{"age": 5,"email": "test09@wolk.com"},
		// 		wolk.Row{"age": 6, "email": "test10@wolk.com"},
		// 		wolk.Row{"age": 23,"email": "test07@wolk.com"}},
		// 		AffectedRowCount: 0,
		// 		MatchedRowCount: 3},
		// 	6: {Data: []wolk.Row{
		// 		wolk.Row{"age": 5,"email": "test09@wolk.com"},
		// 		wolk.Row{"age": 6, "email": "test10@wolk.com"},
		// 		wolk.Row{"age": 23,"email": "test07@wolk.com"}},
		// 		AffectedRowCount: 0,
		// 		MatchedRowCount: 3},
		// 	7: {Data: []wolk.Row{
		// 		wolk.Row{"age":4,"email":"test08@wolk.com"},
		// 		wolk.Row{"age": 5,"email": "test09@wolk.com"},
		// 		wolk.Row{"age": 6, "email": "test10@wolk.com"},
		// 		wolk.Row{"age": 23,"email": "test07@wolk.com"}},
		// 		AffectedRowCount: 0,
		// 		MatchedRowCount: 4},
	}

	cleanuprequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_DROP_DATABASE},
	}

	return mutaterequests, readrequests, readexpected, cleanuprequests, owner, dbname
}

// testing owner chunk
func getRequests_backend_sql3(t *testing.T) (mutaterequests map[int]SQLRequest, readrequests map[int]SQLRequest, readexpected map[int]SQLResponse, cleanuprequests map[int]SQLRequest, owner string, dbname string) {

	owner = MakeName("testowner-debug")
	dbname0 := MakeName("testdb0")
	dbname1 := MakeName("testdb1")
	dbname2 := MakeName("testdb2")
	//tablename0 := "testdb0-testtable"
	//tablename1 := "testdb1-testtable"
	//tablename2 := "testdb2-testtable"

	mutaterequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname0, RequestType: RT_CREATE_DATABASE},
		//1: SQLRequest{Owner: owner, Database: dbname, RequestType: RT_QUERY, RawQuery: "create table personnel (person_id int primary key, name string)"},
		//2: SQLRequest{Owner: owner, Database: dbname, RequestType: RT_QUERY, RawQuery: "insert into personnel (person_id, name) values (6, 'dinozzo')"},
		3: SQLRequest{Owner: owner, Database: dbname1, RequestType: RT_CREATE_DATABASE},
		4: SQLRequest{Owner: owner, Database: dbname2, RequestType: RT_CREATE_DATABASE},

		// 1: SQLRequest{Owner: owner, Database: dbname, Table: tablename, Encrypted: encrypted, Columns: columns, RequestType: RT_CREATE_TABLE},
		// 2: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (6, 'dinozzo')"},
		// 3: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (23, 'ziva')"},
		// 4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (2, 'mcgee')"},
		// 5: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "insert into " + tablename + " (person_id, name) values (101, 'gibbs')"},
		//4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "update " + tablename + " set age = 23 where email = 'test07@wolk.com'"},
	}

	readrequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, RequestType: RT_LIST_DATABASES},
		3: SQLRequest{Owner: owner, RequestType: RT_LIST_DATABASES},
		4: SQLRequest{Owner: owner, RequestType: RT_LIST_DATABASES},
		// 0: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename},
		// 1: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id = 6"},
		// 2: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id > 6"},
		// 3: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id >= 6"},
		// 4: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id > 7"},
		// 5: SQLRequest{Owner: owner, Database: dbname, Encrypted: encrypted, RequestType: RT_QUERY, RawQuery: "select * from " + tablename + " where person_id >= 7"},
		// 6: SQLRequest{Owner: owner, Encrypted: encrypted, RequestType: RT_LIST_DATABASES, RawQuery: "select * from " + tablename},
	}

	readexpected = map[int]SQLResponse{
		0: {Data: []Row{Row{"database": dbname0}}, MatchedRowCount: 1},
		3: {Data: []Row{
			Row{"database": dbname0},
			Row{"database": dbname1}},
			MatchedRowCount: 2,
		},
		4: {Data: []Row{
			Row{"database": dbname0},
			Row{"database": dbname1},
			Row{"database": dbname2}},
			MatchedRowCount: 3,
		},
		// 3: {Data: []Row{
		// 	Row{"name": "dinozzo", "person_id": 6},
		// 	Row{"name": "ziva", "person_id": 23},
		// 	Row{"name": "gibbs", "person_id": 101}},
		// 	MatchedRowCount: 3},

	}

	cleanuprequests = map[int]SQLRequest{
		0: SQLRequest{Owner: owner, Database: dbname, RequestType: RT_DROP_DATABASE},
	}

	return mutaterequests, readrequests, readexpected, cleanuprequests, owner, dbname
}
