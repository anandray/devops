package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/wolkdb/cloudstore/client"
	"github.com/wolkdb/cloudstore/wolk"
)

const SQLPRINT = false

func SQLBench(wclient *client.WolkClient, options *wolk.RequestOptions, owner string, database string, reqs map[int]string, expected map[int]wolk.SQLResponse, user int) error {

	for i := 0; i <= 22; i++ {
		if _, ok := reqs[i]; !ok {
			continue
		}
		sqlprint("")
		sqlprint("(%d)(%d) req: (%s)", user, i, reqs[i])
		eventType, sqlReq, err := wolk.ParseRawRequest(reqs[i])
		if err != nil {
			sqlprint("(%d)(%d) FAIL: ParseRawRequest ERR: %s", user, i, err)
			continue
		}
		sqlReq.Owner = owner
		sqlReq.Database = database

		switch eventType {
		case "read":
			sqlprint("(%d)(%d) reading", user, i)
			actual, err := wclient.ReadSQL(owner, database, sqlReq, options)
			if err != nil {
				//return err
				sqlprint("(%d)(%d) FAIL: Read ERR: %s", user, i, err)
				continue
			}

			// compare
			sqlprint("(%d)(%d) actual: %+v", user, i, actual)
			if _, ok := expected[i]; !ok {
				// don't compare
			} else {
				if !reflect.DeepEqual(actual, expected[i]) {
					if wolk.MatchRowData(actual.Data, expected[i].Data) && actual.AffectedRowCount == expected[i].AffectedRowCount && actual.MatchedRowCount == expected[i].MatchedRowCount {
						sqlprint("(%d)(%d) real pass", user, i)
					} else {
						sqlprint("(%d)(%d) expected: %+v", user, i, expected[i])
						return fmt.Errorf("(%d) test(%d) failed! (%s)", user, i, reqs[i])
					}
				}
			}

		case "mutate":
			sqlprint("(%d)(%d) mutating", user, i)
			txhash, err := wclient.MutateSQL(owner, database, sqlReq)
			if err != nil {
				//return err
				sqlprint("(%d)(%d) FAIL: Mutate ERR: %s", user, i, err)
				continue
			}
			_, err = wclient.WaitForTransaction(txhash)
			if err != nil {
				//return err
				sqlprint("(%d)(%d) FAIL: Mutate WaitForTransaction ERR: %s", user, i, err)
				continue
			}
			// question: where are the sqlresponses for Mutate functions?
		default:
			return fmt.Errorf("(%d)(%d) sql eventType not mutate nor read", user, i)
		}
		sqlprint("(%d)(%d) SUCCESS: o(%s) db(%s) req(%s)", user, i, owner, database, reqs[i])
	}
	return nil
}

func main() {
	//LogInit("", true)
	server := flag.String("server", "c0.wolk.com", "transactions being submitted to here")
	httpport := flag.Uint("httpport", 84, "cloudstore is listening on this port")
	users := flag.Uint("users", 10, "number of users querying in parallel")
	sqlprint("[server=%s:%d, users=%d]", *server, *httpport, *users)
	flag.Parse()

	var wg sync.WaitGroup
	for i := 0; i < int(*users); i++ {
		sqlprint("user (%d)", i)
		wclient, err := client.NewWolkClient(*server, int(*httpport), "")
		if err != nil {
			sqlprint("%s\n", err)
			os.Exit(0)
		}

		wg.Add(1)
		go func(i int) {
			reqs, expected, owner, database := getRequests()
			sqlprint("creating account for owner(%s)", owner)
			txhash, err := wclient.CreateAccount(owner)
			if err != nil {
				if strings.Contains(err.Error(), "Account exists already on blockchain") {
					// use existing account
				} else {
					sqlprint("(%d) CreateAccount %s", i, err)
					//log.Fatal(err)
				}
			}
			sqlprint("waiting for create acct txn (%x)...", txhash)
			_, err = wclient.WaitForTransaction(txhash)
			if err != nil {
				sqlprint("(%d) WaitForTransaction %s", i, err)
				//log.Fatal(err)
			}
			options := wolk.NewRequestOptions()

			sqlprint("done waiting, running queries.")
			err = SQLBench(wclient, options, owner, database, reqs, expected, i)
			if err != nil {
				sqlprint("(%d) SQLBench %s", i, err)
				log.Fatal(err)
			}
			sqlprint("(%d) done with user", i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	sqlprint("all done.")

}

func getRequests() (requests map[int]string, expected map[int]wolk.SQLResponse, owner string, dbname string) {

	owner = makeName("testowner-debug")
	dbname = makeName("newdb")
	tablename := "testtable"

	sqlprint("o: %s, db %s, tbl: %s", owner, dbname, tablename)

	requests = map[int]string{

		0: "createdatabase",
		1: "listdatabases",
		2: "create table " + tablename + " (email string primary key, age int)",
		3: "describetable " + tablename,

		4:  "insert into " + tablename + " (email, age) values ('test05@wolk.com', 1)",
		5:  "insert into " + tablename + " (email, age) values ('test06@wolk.com', 2)",
		6:  "insert into " + tablename + " (email, age) values ('test07@wolk.com', 23)",
		7:  "insert into " + tablename + " (email, age) values ('test08@wolk.com', 45)",
		8:  "insert into " + tablename + " (email, age) values ('test09@wolk.com', 50)",
		9:  "select * from " + tablename,
		10: "select * from " + tablename + " where email = 'test06@wolk.com'",
		11: "delete from " + tablename + " where age = 23",
		12: "select * from " + tablename + " where age = 23",
		13: "select * from " + tablename + " where age >= 30",
		14: "select * from " + tablename + " where age > 30",
		15: "select * from " + tablename + " where age >= 45",
		16: "select * from " + tablename + " where age > 45",
		17: "listtables",
		18: "droptable " + tablename,
		19: "listtables",
		20: "dropdatabase",
		21: "listdatabases",
	}

	expected = map[int]wolk.SQLResponse{
		0: {AffectedRowCount: 1, MatchedRowCount: 0},
		//1: {Data: []wolk.Row{wolk.Row{"database": dbname}}, AffectedRowCount: 0, MatchedRowCount: 1},
		2: {AffectedRowCount: 1, MatchedRowCount: 0},

		3: {Data: []wolk.Row{wolk.Row{"ColumnName": "age", "IndexType": "BPLUS", "Primary": 0, "ColumnType": "INTEGER"}, wolk.Row{"ColumnName": "email", "IndexType": "BPLUS", "Primary": 1, "ColumnType": "STRING"}}, AffectedRowCount: 0, MatchedRowCount: 0},

		4: {AffectedRowCount: 1, MatchedRowCount: 0},
		5: {AffectedRowCount: 1, MatchedRowCount: 0},
		6: {AffectedRowCount: 1, MatchedRowCount: 0},
		7: {AffectedRowCount: 1, MatchedRowCount: 0},
		8: {AffectedRowCount: 1, MatchedRowCount: 0},

		9:  {Data: []wolk.Row{wolk.Row{"email": "test05@wolk.com", "age": 1}, wolk.Row{"age": 2, "email": "test06@wolk.com"}, wolk.Row{"email": "test07@wolk.com", "age": 23}, wolk.Row{"email": "test08@wolk.com", "age": 45}, wolk.Row{"email": "test09@wolk.com", "age": 50}}, AffectedRowCount: 0, MatchedRowCount: 5},
		10: {Data: []wolk.Row{wolk.Row{"age": 2, "email": "test06@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		11: {AffectedRowCount: 1, MatchedRowCount: 0},
		12: {},
		13: {Data: []wolk.Row{wolk.Row{"age": 45, "email": "test08@wolk.com"},
			wolk.Row{"age": 50, "email": "test09@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 2},
		14: {Data: []wolk.Row{wolk.Row{"age": 45, "email": "test08@wolk.com"},
			wolk.Row{"age": 50, "email": "test09@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 2},
		15: {Data: []wolk.Row{wolk.Row{"age": 45, "email": "test08@wolk.com"},
			wolk.Row{"age": 50, "email": "test09@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 2},
		16: {Data: []wolk.Row{wolk.Row{"age": 50, "email": "test09@wolk.com"}}, AffectedRowCount: 0, MatchedRowCount: 1},
		17: {Data: []wolk.Row{wolk.Row{"table": tablename}}, MatchedRowCount: 1},
		18: {AffectedRowCount: 1, MatchedRowCount: 0},
		19: {},
		20: {AffectedRowCount: 1, MatchedRowCount: 0},
		//21: {},
	}

	return requests, expected, owner, dbname
}

func makeName(prefix string) (nm string) {
	return fmt.Sprintf("%s%d", prefix, int32(time.Now().Unix()))
}

func sqlprint(in string, args ...interface{}) {
	if SQLPRINT {
		if in == "" {
			fmt.Println()
		} else {
			fmt.Printf("[sqlbench] "+in+"\n", args...)
		}
	}
}
