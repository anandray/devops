// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
	//"github.com/wolkdb/cloudstore/log"
)

//for sql parsing
type QueryOption struct {
	Type      string //"Select" or "Insert" or "Update" probably should be an enum
	TableName string
	Columns   []Column // only for createtable
}

func (q *QueryOption) HasRequestColumns() (ok bool) {
	for _, c := range q.Columns {
		if c.ColumnName == "*" {
			return false
		}
	}
	return true
}

func IsQueryType(rawQuery string, querytype string) bool {
	return strings.HasPrefix(strings.ToLower(strings.Trim(rawQuery, " ")), strings.ToLower(querytype))
}

func getSuffix(rawQuery string) (suffix string) {
	words := strings.Split(rawQuery, " ")
	return words[len(words)-1]
}

// used for wcloud
func ParseRawRequest(rawQuery string) (eventType string, sqlReq *SQLRequest, err error) {

	sqlReq = new(SQLRequest)
	eventType = "mutate"
	switch {
	case IsQueryType(rawQuery, "CreateDatabase"):
		sqlReq.RequestType = RT_CREATE_DATABASE
		return eventType, sqlReq, nil
	case IsQueryType(rawQuery, "ListDatabases"):
		sqlReq.RequestType = RT_LIST_DATABASES
		eventType = "read"
		return eventType, sqlReq, nil
	case IsQueryType(rawQuery, "DescribeTable"):
		sqlReq.RequestType = RT_DESCRIBE_TABLE
		sqlReq.Table = getSuffix(rawQuery)
		eventType = "read"
		return eventType, sqlReq, nil
	case IsQueryType(rawQuery, "ListTables"):
		sqlReq.RequestType = RT_LIST_TABLES
		eventType = "read"
		return eventType, sqlReq, nil
	case IsQueryType(rawQuery, "DropDatabase"):
		sqlReq.RequestType = RT_DROP_DATABASE
		return eventType, sqlReq, nil
	case IsQueryType(rawQuery, "DropTable"):
		sqlReq.RequestType = RT_DROP_TABLE
		sqlReq.Table = getSuffix(rawQuery)
		return eventType, sqlReq, nil

		// case IsQueryType(rawQuery, "Get")
		// sqlReq.RequestType = RT_GET
		// need to parse for tablename
		// need to parse for key
		// case IsQueryType(rawQuery, "Put"):
		// sqlReq.RequestType = RT_PUT
		// need to parse for Columns
		// need to parse for tablename
	}

	rawQuery = strings.Replace(rawQuery, `string`, `text`, -1)
	queryOption, err := ParseQuery(rawQuery)
	if err != nil {
		return "", sqlReq, fmt.Errorf("[query:ParseRawRequest] %s", err)
	}
	sqlReq.Table = queryOption.TableName

	switch queryOption.Type {
	case "CreateTable":
		sqlReq.RequestType = RT_CREATE_TABLE
		sqlReq.Columns = queryOption.Columns
	case "Select":
		sqlReq.RequestType = RT_QUERY
		eventType = "read"
		sqlReq.Columns = queryOption.Columns
		sqlReq.RawQuery = rawQuery
	case "Insert":
		sqlReq.RequestType = RT_QUERY
		sqlReq.RawQuery = rawQuery
	case "Update":
		sqlReq.RequestType = RT_QUERY
		sqlReq.RawQuery = rawQuery
	case "Delete": // not sure this works
		sqlReq.RequestType = RT_QUERY
		sqlReq.RawQuery = rawQuery
	default:
		return "", sqlReq, fmt.Errorf("[query:ParseRawRequest] sql request type not supported")
	}

	return eventType, sqlReq, nil

}

func ParseQuery(rawQuery string) (query QueryOption, err error) {

	rawQuery = strings.Replace(rawQuery, `string`, `text`, -1)
	stmt, err := sqlparser.Parse(rawQuery)
	if err != nil {
		//fmt.Printf("[query:ParseQuery] err: %v\n", err)
		return query, fmt.Errorf("[query:ParseQuery] %s", err)
	}

	switch stmt := stmt.(type) {

	// DDL represents a CREATE, ALTER, DROP, RENAME or TRUNCATE statement.
	//https://github.com/xwb1989/sqlparser/blob/7569d6f10403c721d3f69781e1edd3b47d869161/ast.go#L648
	case *sqlparser.DDL:

		_, err = sqlparser.ParseStrictDDL(rawQuery)
		if err != nil {
			return query, fmt.Errorf("[query:ParseQuery] %s", err)
		}

		if stmt.Action == sqlparser.CreateStr && len(sqlparser.String(stmt.NewName.Name)) > 0 {
			query.Type = "CreateTable"
			query.TableName = trimQuotes(sqlparser.String(stmt.NewName.Name))

			for _, c := range stmt.TableSpec.Columns {
				col := new(Column)
				col.ColumnName = trimQuotes(sqlparser.String(c.Name))
				col.IndexType = IT_BPLUSTREE //TODO: hardcoded to default to this
				switch c.Type.Type {
				case "int":
					col.ColumnType = CT_INTEGER
					break
				case "text":
					col.ColumnType = CT_STRING
					break
				case "float":
					col.ColumnType = CT_FLOAT
					break
				default:
					return query, fmt.Errorf("[query:ParseQuery] column type %s not supported yet", c.Type.Type)
					//TODO: https://www.w3schools.com/sql/sql_dataasp
				}
				col.Primary = 0
				if c.Type.KeyOpt == 1 { //TODO: this should be colKeyPrimary from sqlparser
					col.Primary = 1
				}
				query.Columns = append(query.Columns, *col)
			}

		} else {
			return query, fmt.Errorf("[query:ParseQuery] DDL %s not supported", stmt.Action)
		}

		return query, nil

	//https://github.com/xwb1989/sqlparser/blob/7569d6f10403c721d3f69781e1edd3b47d869161/ast.go#L1405
	case *sqlparser.OtherRead: //DESCRIBE or EXPLAIN. Only an indicator
		q := strings.ToLower(rawQuery)
		if strings.Contains(q, "describe") {
			return query, fmt.Errorf("[query:ParseQuery] sqlite does not support DESCRIBE.")
			// or, could just parse out the tablename and put it through RT_DESCRIBE_TABLE
		}
	//note: sqlite does not have a DESCRIBE TABLE query

	//https://github.com/xwb1989/sqlparser/blob/7569d6f10403c721d3f69781e1edd3b47d869161/ast.go#L247
	case *sqlparser.Select:
		buf := sqlparser.NewTrackedBuffer(nil)
		stmt.Format(buf)
		//fmt.Printf("[query:ParseQuery] select: %v\n", buf.String())

		query.Type = "Select"
		for i, column := range stmt.SelectExprs { //TODO: what happens with *?
			//fmt.Printf("[query:ParseQuery] select %d: %+v\n", i, sqlparser.String(column)) // stmt.(*sqlparser.Select).SelectExprs)

			var newcolumn Column
			newcolumn.QueryID = i
			aliased, ok := column.(*sqlparser.AliasedExpr)
			if ok && sqlparser.String(aliased.As) != "" {
				//fmt.Printf("[query:ParseQuery] aliased.As (%s)\n", sqlparser.String(aliased.As))
				newcolumn.ColumnName = sqlparser.String(aliased.As)
			} else {
				newcolumn.ColumnName = trimQuotes(sqlparser.String(column))
			}
			query.Columns = append(query.Columns, newcolumn)

		}

		//From
		//fmt.Printf("from 0: %+v \n", sqlparser.String(stmt.From[0]))
		if len(stmt.From) == 0 {
			return query, fmt.Errorf("[query:ParseQuery] Invalid SQL - Missing FROM")
		}
		query.TableName = trimQuotes(sqlparser.String(stmt.From[0]))

		//Where & Having
		//fmt.Printf("where or having: %s \n", readable(stmt.Where.Expr))
		// if stmt.Where == nil {
		// 	log.Debug("NOT SUPPORTING SELECT WITH NO WHERE")
		// 	return query, fmt.Errorf("[query:ParseQuery] WHERE missing on Update query"), ErrorCode: 444, ErrorMessage: "SELECT & UPDATE query must have WHERE"}
		// }
		// if stmt.Where.Type == sqlparser.WhereStr { //Where
		// 	//fmt.Printf("type: %s\n", stmt.Where.Type)
		// 	query.Where, err = parseWhere(stmt.Where.Expr)
		// 	//this is where recursion for nested parentheses should take place
		// 	if err != nil {
		// 		return query, fmt.Errorf("[sqlchain:ParseQuery] parseWhere [%s]", rawQuery))
		// 	}
		// } else if stmt.Where.Type == sqlparser.HavingStr { //Having
		// 	fmt.Printf("type: %s\n", stmt.Where.Type)
		// 	//TODO: fill in having
		// 	return query, fmt.Errorf("[sqlchain:ParseQuery] Parse Having Clause Not currently supported"), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [HAVING clause not currently supported]"}
		// }

		//TODO: GroupBy ([]Expr)
		//for _, g := range stmt.GroupBy {
		//	fmt.Printf("groupby: %s \n", readable(g))
		//}

		//TODO: OrderBy
		// query.Ascending = 1 //default if nothing?

		//Limit
		return query, nil

	/* Other options inside Select:
	   type Select struct {
	   	Cache       string
	   	Comments    Comments
	   	Distinct    string
	   	Hints       string
	   	SelectExprs SelectExprs
	   	From        TableExprs
	   	Where       *Where
	   	GroupBy     GroupBy
	   	Having      *Where
	   	OrderBy     OrderBy
	   	Limit       *Limit
	   	Lock        string
	   }*/

	case *sqlparser.Insert:
		//for now, 1 row to insert only. still need to figure out multiple rows
		//i.e. INSERT INTO MyTable (id, name) VALUES (1, 'Bob'), (2, 'Peter'), (3, 'Joe')

		query.Type = "Insert"
		// query.Ascending = 1 //default
		//fmt.Printf("Action: %s \n", stmt.Action)
		//fmt.Printf("Comments: %+v \n", stmt.Comments)
		//fmt.Printf("Ignore: %s \n", stmt.Ignore)
		query.TableName = trimQuotes(sqlparser.String(stmt.Table.Name))
		// if len(stmt.Rows.(sqlparser.Values)) == 0 {
		// 	return query, fmt.Errorf("[query:ParseQuery] Insert has no values found"), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [INSERT query missing VALUES]"}
		// }
		// if len(stmt.Rows.(sqlparser.Values)[0]) != len(stmt.Columns) {
		// 	return query, fmt.Errorf("[query:ParseQuery] Insert has mismatch # of cols & vals"), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [Mismatch in number of columns and values]"}
		// }
		// insertCells := make(map[string]interface{})
		// for i, c := range stmt.Columns {
		// 	col := sqlparser.String(c)
		// 	if _, ok := insertCells[col]; ok {
		// 		return query, fmt.Errorf("[query:ParseQuery] Insert can't have duplicate col %s", col), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [INSERT cannot have duplicate columns]"}
		// 	}
		// 	//only detects string and float. how to do int? does it matter
		// 	value := sqlparser.String(stmt.Rows.(sqlparser.Values)[0][i])
		// 	if isQuoted(value) {
		// 		insertCells[col] = trimQuotes(value)
		// 	} else if isNumeric(value) {
		// 		insertCells[col], err = strconv.ParseFloat(value, 64)
		// 		if err != nil {
		// 			return query, fmt.Errorf("[query:ParseQuery] Insert can't have duplicate col %s", col), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [INSERT cannot have duplicate columns]"}
		// 		}
		// 	} else {
		// 		return query, fmt.Errorf("[query:ParseQuery] Insert value %s has unknown type", value), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [Invalid value type passed in.]"}
		// 		//TODO: more clear Message
		// 	}
		// 	//insertCells[col] = trimQuotes(sqlparser.String(stmt.Rows.(sqlparser.Values)[0][i]))
		// }
		// r :=  NewRow()
		// r = insertCells
		// query.Inserts = append(query.Inserts, r)
		//fmt.Printf("OnDup: %+v\n", stmt.OnDup)
		//fmt.Printf("Rows: %+v\n", stmt.Rows.(sqlparser.Values))
		//fmt.Printf("Rows: %+v\n", sqlparser.String(stmt.Rows.(sqlparser.Values)))
		//for i, v := range stmt.Rows.(sqlparser.Values)[0] {
		//	fmt.Printf("row: %v %+v\n", i, sqlparser.String(v))
		//}

	case *sqlparser.Update:

		query.Type = "Update"
		//fmt.Printf("Comments: %+v \n", stmt.Comments)
		query.TableName = trimQuotes(sqlparser.String(stmt.TableExprs[0]))
		// query.Update = make(map[string]interface{})
		// for _, expr := range stmt.Exprs {
		// 	col := sqlparser.String(expr.Name)
		// 	//fmt.Printf("col: %+v\n", col)
		// 	if _, ok := query.Update[col]; ok {
		// 		return query, fmt.Errorf("[query:ParseQuery] Update can't have duplicate col %s", col), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [INSERT cannot have duplicate columns]"}
		// 	}
		// 	value := readable(expr.Expr)
		// 	if isQuoted(value) {
		// 		query.Update[col] = trimQuotes(value)
		// 	} else if isNumeric(value) {
		// 		query.Update[col], err = strconv.ParseFloat(value, 64)
		// 		if err != nil {
		// 			return query, fmt.Errorf("[query:ParseQuery] ParseFloat %s", err.Error()), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [Float Value could not be parsed]"}
		// 		}
		// 	} else {
		// 		return query, fmt.Errorf("[query:ParseQuery] Update value %s has unknown type", value), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [Invalid value type passed in.]"}
		// 	}
		// 	//fmt.Printf("val: %v \n", query.Update[col])
		// }
		//
		// // Where
		// log.Debug(fmt.Sprintf("Statement: [%+v] | SqlParser: [%+v]", stmt, sqlparser.WhereStr))
		// if stmt.Where == nil {
		// 	log.Debug("NOT SUPPORTING UPDATES WITH NO WHERE")
		// 	return query, fmt.Errorf("[query:ParseQuery] WHERE missing on Update query"), ErrorCode: 444, ErrorMessage: "UPDATE query must have WHERE"}
		// }
		// if stmt.Where.Type == sqlparser.WhereStr {
		// 	query.Where, err = parseWhere(stmt.Where.Expr)
		// 	//TODO: this is where recursion for nested parentheses should probably take place
		// 	if err != nil {
		// 		return query, fmt.Errorf("[query:ParseQuery] parseWhere %s", err.Error()))
		// 	}
		// 	//fmt.Printf("Where: %+v\n", query.Where)
		// }
		//
		// //TODO: OrderBy
		// query.Ascending = 1 //default if nothing?

		//Limit
		//fmt.Printf("Limit: %v \n", stmt.Limit)
		return query, nil
	case *sqlparser.Delete:
		query.Type = "Delete"
		if len(stmt.TableExprs) == 0 {
			return query, fmt.Errorf("[query:ParseQuery] DELETE TableExprs empty")
		}
		query.TableName = trimQuotes(sqlparser.String(stmt.TableExprs[0])) // TODO: an OK around the array in case of panic
		//fmt.Printf("Comments: %+v \n", stmt.Comments)

		//Targets
		// for _, t := range stmt.Targets {
		// 	fmt.Printf("Targets: %s\n", t.Name)
		// }
		//
		// //Where
		// if stmt.Where == nil {
		// 	log.Debug("NOT SUPPORTING DELETES WITH NO WHERE")
		// 	return query, fmt.Errorf("[query:ParseQuery] WHERE missing on Delete query"), ErrorCode: 444, ErrorMessage: "DELETE query must have WHERE"}
		// }
		// if stmt.Where.Type == sqlparser.WhereStr { //Where
		// 	query.Where, err = parseWhere(stmt.Where.Expr)
		// 	//TODO: this is where recursion for nested parentheses should take place
		// 	if err != nil {
		// 		return query, fmt.Errorf("[query:ParseQuery] parseWhere %s", err.Error()))
		// 	}
		// 	//fmt.Printf("Where: %+v\n", query.Where)
		// }
		//
		// //TODO: OrderBy
		// query.Ascending = 1 //default if nothing?

		//Limit
		//fmt.Printf("Limit: %v \n", stmt.Limit)

		return query, nil

		/* Other Options for type of Query:
		   func (*Union) iStatement()      {}
		   func (*Select) iStatement()     {}
		   func (*Insert) iStatement()     {}
		   func (*Update) iStatement()     {}
		   func (*Delete) iStatement()     {}
		   func (*Set) iStatement()        {}
		   func (*DDL) iStatement()        {}
		   func (*Show) iStatement()       {}
		   func (*Use) iStatement()        {}
		   func (*OtherRead) iStatement()  {}
		   func (*OtherAdmin) iStatement() {}
		*/

	}

	return query, err
}

// func parseWhere(expr sqlparser.Expr) (where Where, err error) {
//
// 	switch expr := expr.(type) {
// 	case *sqlparser.OrExpr:
// 		where.Left = readable(expr.Left)
// 		where.Right = readable(expr.Right)
// 		where.Operator = "OR" //should be const
// 	case *sqlparser.AndExpr:
// 		where.Left = readable(expr.Left)
// 		where.Right = readable(expr.Right)
// 		where.Operator = "AND" //shoud be const
// 	case *sqlparser.IsExpr:
// 		where.Right = readable(expr.Expr)
// 		where.Operator = expr.Operator
// 	case *sqlparser.BinaryExpr:
// 		where.Left = readable(expr.Left)
// 		where.Right = readable(expr.Right)
// 		where.Operator = expr.Operator
// 	case *sqlparser.ComparisonExpr:
// 		where.Left = readable(expr.Left)
// 		where.Right = readable(expr.Right)
// 		where.Operator = expr.Operator
// 	default:
// 		return where, fmt.Errorf("[sqlchain:parseWhere] exp Type [%s] not supported", expr), ErrorCode: 401, ErrorMessage: fmt.Sprintf("SQL Parsing error: [Expression Type (%s) not currently supported]", expr)}
// 	}
// 	where.Right = trimQuotes(where.Right)
//
// 	return where, err
// }
//
func trimQuotes(s string) string {
	if len(s) > 0 && (s[0] == '\'' || s[0] == '`' || s[0] == '"') {
		s = s[1:]
	}
	if len(s) > 0 && (s[len(s)-1] == '\'' || s[len(s)-1] == '`' || s[len(s)-1] == '"') {
		s = s[:len(s)-1]
	}
	return s
}

func isQuoted(s string) bool { //string
	if (len(s) > 0) && (s[0] == '\'' || s[0] == '`' || s[0] == '"') && (s[len(s)-1] == '\'' || s[len(s)-1] == '`' || s[len(s)-1] == '"') {
		return true
	}
	return false
}

//
// func isNumeric(s string) bool { //float or int
// 	if _, err := strconv.ParseFloat(s, 64); err == nil {
// 		return true
// 	}
// 	return false
// }
//
// func readable(expr sqlparser.Expr) string {
// 	switch expr := expr.(type) {
// 	case *sqlparser.OrExpr:
// 		return fmt.Sprintf("(%s or %s)", readable(expr.Left), readable(expr.Right))
// 	case *sqlparser.AndExpr:
// 		return fmt.Sprintf("(%s and %s)", readable(expr.Left), readable(expr.Right))
// 	case *sqlparser.BinaryExpr:
// 		return fmt.Sprintf("(%s %s %s)", readable(expr.Left), expr.Operator, readable(expr.Right))
// 	case *sqlparser.IsExpr:
// 		return fmt.Sprintf("(%s %s)", readable(expr.Expr), expr.Operator)
// 	case *sqlparser.ComparisonExpr:
// 		return fmt.Sprintf("(%s %s %s)", readable(expr.Left), expr.Operator, readable(expr.Right))
// 	default:
// 		return sqlparser.String(expr)
// 	}
// }
