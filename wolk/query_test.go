// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestParseRawRequest(t *testing.T) {
	query := map[int]string{
		0: "select * from contacts",
		1: "select name, age from contacts",
		2: "select age, name from contacts",
		3: "listdatabases",
		4: "createdatabase",
		5: "create table `mycontacts` (`person_id` int primary key, `name` string)",
		6: "insert into  mycontacts (email, age) values ('test10@wolk.com', 6)",
		7: "update mycontacts set age = 23 where email = 'test07@wolk.com'",
		8: "delete from mycontacts where age = 5",
		9: "describetable contacts",
	}

	for i := 0; i < len(query); i++ {
		if _, ok := query[i]; !ok { //skip any commented out queries
			continue
		}
		fmt.Printf("[query_test] test %d: %s\n", i, query[i])
		eventType, actual, err := ParseRawRequest(query[i])
		actual.Owner = "fakeowner"
		actual.Database = "fakedb"
		fmt.Printf("[query_test] response: eventType(%s), (%+v)\n", eventType, actual)
		if err != nil {
			//fmt.Printf("[query_test] expected: %+v\n", expected[i])
			t.Fatal(err)
		}
		// if !reflect.DeepEqual(actual, expected[i]) {
		// 	if !matchResponse(actual.Columns, expected[i].Columns) {
		// 		fmt.Printf("[query_test] expected: %+v\n", expected[i])
		// 		t.Fatal(err)
		// 	}
		// }
	}

}

func TestParseCreateTable(t *testing.T) {
	//table := "testtable"
	//query := "create table `" + table + "` (`person_id` int primary key, `name` string)"
	query := "create table `2Bq1MGlQS` (`P2TGKQKk0iamWPIVwBV` int primary key, `9RC12xb5bn4FoTqvu2ffn` float, `eayPOg6FyjYc5MqrQ1` text, `0ZV2b7D` int, `JVmeqTTwqIyAXbZLpED` float)"
	//query := "select (`MLjZuNSshPXx4x` + 2569624501871852968*8088947552198212589 - -6760065269996489614%9076019266156152065) as w, `MLjZuNSshPXx4x`, case(`MLjZuNSshPXx4x` & 1) when 0 then `even` else `odd` end as parity from `v4AJ`"
	//query := "select (person_id + 3*4 - 7%6) as w, name, case(person_id & 1) when 0 then 'even' else 'odd' end as parity from person"
	fmt.Println(query)
	q, err := ParseQuery(query)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("\n%#v\n", q)

}

func TestParseSelect(t *testing.T) {
	query := map[int]string{
		0: "select * from contacts",
		1: "select name, age from contacts",
		2: "select age, name from contacts",
		//3: "describe table contacts",
	}

	expected := map[int]QueryOption{
		0: QueryOption{Type: "Select", TableName: "contacts", Columns: []Column{Column{ColumnName: "*", QueryID: 0}}},
		1: QueryOption{Type: "Select", TableName: "contacts", Columns: []Column{Column{ColumnName: "name", QueryID: 0}, Column{ColumnName: "age", QueryID: 1}}},
		2: QueryOption{Type: "Select", TableName: "contacts", Columns: []Column{Column{ColumnName: "age", QueryID: 0}, Column{ColumnName: "name", QueryID: 1}}},
	}

	for i := 0; i <= len(query); i++ {
		if _, ok := query[i]; !ok { //skip any commented out queries
			continue
		}
		fmt.Printf("[query_test] test %d: %s\n", i, query[i])
		actual, err := ParseQuery(query[i])
		fmt.Printf("[query_test] response: %+v\n", actual)
		if err != nil {
			fmt.Printf("[query_test] expected: %+v\n", expected[i])
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected[i]) {
			if !matchResponse(actual.Columns, expected[i].Columns) {
				fmt.Printf("[query_test] expected: %+v\n", expected[i])
				t.Fatal(err)
			}
		}
	}

}

func matchResponse(actual []Column, expected []Column) (ok bool) {
	if len(actual) != len(expected) {
		return false
	}
	if len(actual) == 0 {
		return true
	}
	foundrows := 0
	for _, erow := range expected {
		er_marshalled, _ := json.Marshal(erow)
		//var e sqlcom.Column
		var e interface{}
		json.Unmarshal(er_marshalled, &e)
		//fmt.Printf("--considering expected row: %+v\n", e)
		for _, arow := range actual {
			ar_marshalled, _ := json.Marshal(arow)
			//var a sqlcom.Column
			var a interface{}
			json.Unmarshal(ar_marshalled, &a)
			//fmt.Printf("--considering actual row: %+v\n", a)
			if reflect.DeepEqual(a, e) {
				//fmt.Printf("--found a match\n")
				foundrows++
			}
		}
	}
	if foundrows == len(expected) {
		return true
	}
	return false

}

/*
func TestParseQuery(t *testing.T) {

		rawqueries := map[string]string{
			`get1`:         `select name from contacts where age >= 35`,
			`get2`:         `select name, age from contacts where email = 'rodney@wolk.com'`,
			`doublequotes`: `select name, age from contacts where email = "rodney@wolk.com"`,
			`not`:          `select name, age from contacts where email != 'rodney@wolk.com'`,
			`insert`:       `insert into contacts(email, name, age) values("bertie@gmail.com","Bertie Basset", 7)`,
			`update`:       `UPDATE contacts set age = 8, name = "Bertie B" where email = "bertie@gmail.com"`,
			`delete`:       `delete from contacts where age >= 25`,
			//`precedence`:   `select * from a where a=b and c=d or e=f`,
			//`like`:         `select name, age from contacts where email like '%wolk%'`,
			//`is`:           `select name, age from contacts where age is not null`,
			//`and`:          `select name, age from contacts where email = 'rodney@wolk.com' and age = 38`,
			//`or`:           `select name, age from contacts where email = 'rodney@wolk.com' or age = 35`,
			//`groupby`:      `select name, age from contacts where age >= 35 group by email`,
		}

		expected := make(map[string] QueryOption)
		expected[`get1`] =  QueryOption{
			Type:  "Select",
			Table: "contacts",
			RequestColumns: [] Column{
				 Column{ColumnName: "name"},
			},
			Where:      Where{Left: "age", Right: "35", Operator: ">="},
			Ascending: 1,
		}
		expected[`get2`] =  QueryOption{
			Type:  "Select",
			Table: "contacts",
			RequestColumns: [] Column{
				 Column{ColumnName: "name"},
				 Column{ColumnName: "age"},
			},
			Where:      Where{Left: "email", Right: "rodney@wolk.com", Operator: "="},
			Ascending: 1,
		}
		expected[`doublequotes`] =  QueryOption{
			Type:  "Select",
			Table: "contacts",
			RequestColumns: [] Column{
				 Column{ColumnName: "name"},
				 Column{ColumnName: "age"},
			},
			Where:      Where{Left: "email", Right: "rodney@wolk.com", Operator: "="},
			Ascending: 1,
		}
		expected[`not`] =  QueryOption{
			Type:  "Select",
			Table: "contacts",
			RequestColumns: [] Column{
				 Column{ColumnName: "name"},
				 Column{ColumnName: "age"},
			},
			Where:      Where{Left: "email", Right: "rodney@wolk.com", Operator: "!="},
			Ascending: 1,
		}
		expected[`insert`] =  QueryOption{
			Type:  "Insert",
			Table: "contacts",
			Inserts: [] Row{
				 Row{
					"name":  "Bertie Basset",
					"age":   float64(7),
					"email": "bertie@gmail.com"},
			},
			Ascending: 1,
		}
		expected[`update`] =  QueryOption{
			Type:  "Update",
			Table: "contacts",
			Update: map[string]interface{}{
				"age":  float64(8),
				"name": "Bertie B",
			},
			Where:      Where{Left: "email", Right: "bertie@gmail.com", Operator: "="},
			Ascending: 1,
		}
		expected[`delete`] =  QueryOption{
			Type:      "Delete",
			Table:     "contacts",
			Where:      Where{Left: "age", Right: "25", Operator: ">="},
			Ascending: 1,
		}

		var fail []string
		for testid, raw := range rawqueries {

			clean, err :=  ParseQuery(raw)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(clean, expected[testid]) {
				fmt.Printf("\n[%s] raw: %s\n", testid, raw)
				fmt.Printf("clean: %+v\n", clean)
				fmt.Printf("expected: %+v\n\n", expected[testid])
				fail = append(fail, testid)
			}

		}
		if len(fail) > 0 {
			t.Fatal(fmt.Errorf("tests [%s] failed", strings.Join(fail, ",")))
		}

}
*/
