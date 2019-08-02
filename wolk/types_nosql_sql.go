// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

// TODO: move the versions
const (
	CHUNK_SIZE    = 4096
	VERSION_MAJOR = 0
	VERSION_MINOR = 1
	VERSION_PATCH = 2
	VERSION_META  = "poc"
)

var NoSQLChainVersion = func() string {
	v := fmt.Sprintf("%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
	if VERSION_META != "" {
		v += "-" + VERSION_META
	}
	return v
}()

var SQLChainVersion = func() string {
	v := fmt.Sprintf("%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
	if VERSION_META != "" {
		v += "-" + VERSION_META
	}
	return v
}()

func EmptyBytes(hashid []byte) (valid bool) {
	valid = true
	for i := 0; i < len(hashid); i++ {
		if hashid[i] != 0 {
			return false
		}
	}
	return valid
}

func IsHash(hashid []byte) (valid bool) {
	cnt := 0
	for i := 0; i < len(hashid); i++ {
		if hashid[i] == 0 {
			cnt++
		}
	}
	if cnt > 3 {
		return false
	} else {
		return true
	}
}

func IntToByte(i int) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, uint64(i))
	return k
}

func Int64ToByte(i int64) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, uint64(i))
	return k
}

func Uint64ToByte(i uint64) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, i)
	return k
}

func FloatToByte(f float64) (k []byte) {
	bits := math.Float64bits(f)
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, bits)
	return k
}

func BytesToInt64(b []byte) (i int64) {
	if len(bytes.Trim(b, "\x00")) == 0 {
		return 0
	}
	i = int64(binary.BigEndian.Uint64(b))
	return i
}

func BytesToInt(b []byte) (i int) {
	if len(bytes.Trim(b, "\x00")) == 0 {
		return 0
	}
	i = int(binary.BigEndian.Uint64(b))
	return i
}

func BytesToFloat(b []byte) (f float64) {
	bits := binary.BigEndian.Uint64(b)
	f = math.Float64frombits(bits)
	return f
}

func SHA256(inp string) (k []byte) {
	h := sha256.New()
	h.Write([]byte(inp))
	k = h.Sum(nil)
	return k
}

func isNil(a interface{}) bool {
	if a == nil { // || reflect.ValueOf(a).IsNil()  {
		return true
	}
	return false
}

type Shim struct {
	Name    string
	RootURL string
}

//gets data (Row) out of a slice of Rows, and rtns as one json.
func rowDataToJson(rows []Row) (string, error) {
	var resRows []map[string]interface{}
	for _, row := range rows {
		resRows = append(resRows, row)
	}
	resBytes, err := json.Marshal(resRows)
	if err != nil {
		return "", err
	}
	return string(resBytes), nil
}

//json input string should be []map[string]interface{} format
func JsonDataToRow(in string) (rows []Row, err error) {

	var jsonRows []map[string]interface{}
	if err = json.Unmarshal([]byte(in), &jsonRows); err != nil {
		return rows, err
	}
	for _, jRow := range jsonRows {
		row := NewRow()
		row = jRow
		rows = append(rows, row)
	}
	return rows, nil
}

func stringToColumnType(in string, columnType ColumnType) (out interface{}, err error) {
	switch columnType {
	case CT_INTEGER:
		out, err = strconv.Atoi(in)
	case CT_STRING:
		out = in
	case CT_FLOAT:
		out, err = strconv.ParseFloat(in, 64)
	//case:  CT_BLOB:
	//?
	default:
		err = fmt.Errorf("[types|stringToColumnType] columnType not found")
	}
	return out, err
}

//gets only the specified Columns (column name and value) out of a single Row, returns as a Row with only the relevant data
func filterRowByColumns(row Row, columns []Column) (filteredRow Row) {
	filteredRow = make(map[string]interface{})
	for _, col := range columns {
		if _, ok := row[col.ColumnName]; ok {
			filteredRow[col.ColumnName] = row[col.ColumnName]
		}
	}
	return filteredRow
}

func CheckColumnType(colType ColumnType) bool {
	/*
		var ct uint8
		switch colType.(type) {
		case int:
			ct = uint8(colType.(int))
		case uint8:
			ct = colType.(uint8)
		case float64:
			ct = uint8(colType.(float64))
		case string:
			cttemp, _ := strconv.ParseUint(colType.(string), 10, 8)
			ct = uint8(cttemp)
		case ColumnType:
			ct = colType.(ColumnType)
		default:
			fmt.Printf("CheckColumnType not a type I can work with\n")
			return false
		}
	*/
	ct := colType
	if ct == CT_INTEGER || ct == CT_STRING || ct == CT_FLOAT { //|| ct ==  CT_BLOB {
		return true
	}
	return false
}

func CheckIndexType(it IndexType) bool {
	if it == IT_BPLUSTREE || it == IT_NONE {
		return true
	}
	return false
}

func StringToBytes(columnType ColumnType, key string) (k []byte) {
	k = make([]byte, 32)
	switch columnType {
	case CT_INTEGER:
		// convert using atoi to int
		i, _ := strconv.Atoi(key)
		k8 := IntToByte(i) // 8 byte
		copy(k, k8)        // 32 byte
	case CT_STRING:
		copy(k, []byte(key))
	case CT_FLOAT:
		f, _ := strconv.ParseFloat(key, 64)
		k8 := FloatToByte(f) // 8 byte
		copy(k, k8)          // 32 byte
	case CT_BLOB:
		// TODO: do this correctly with JSON treatment of binary
		copy(k, []byte(key))
	}
	return k
}

func BytesToString(columnType ColumnType, k []byte) (out string) {
	switch columnType {
	case CT_BLOB:
		return fmt.Sprintf("%v", k)
	case CT_STRING:
		return fmt.Sprintf("%s", string(k))
	case CT_INTEGER:
		a := binary.BigEndian.Uint64(k)
		return fmt.Sprintf("%d [%x]", a, k)
	case CT_FLOAT:
		bits := binary.BigEndian.Uint64(k)
		f := math.Float64frombits(bits)
		return fmt.Sprintf("%f", f)
	}
	return "unknown key type"

}

func ValueToString(v []byte) (out string) {
	if IsHash(v) {
		return fmt.Sprintf("%x", v)
	} else {
		return fmt.Sprintf("%v", string(v))
	}
}

func padBytesForComparison(a []byte, b []byte) (ares []byte, bres []byte) {
	if len(a) == len(b) {
		return a, b
	}

	if len(a) > len(b) {
		padded_b := make([]byte, len(a))
		copy(padded_b[0:len(b)], b)
		return a, padded_b
	}

	//len(b) > len(a)
	padded_a := make([]byte, len(b))
	copy(padded_a[0:len(a)], a)
	return padded_a, b
}

func ByteToColumnType(b byte) (ct ColumnType, err error) {
	switch b {
	case 1:
		return CT_INTEGER, err
	case 2:
		return CT_STRING, err
	case 3:
		return CT_FLOAT, err
	case 4:
		return CT_BLOB, err
	default:
		return CT_INTEGER, fmt.Errorf("Invalid Column Type")
	}
}

func ByteToIndexType(b byte) (it IndexType) {
	switch b {
	case 1:
		return IT_HASHTREE
	case 2:
		return IT_BPLUSTREE
	case 3:
		return IT_FULLTEXT
	default:
		return IT_NONE
	}
}

func ColumnTypeToInt(ct ColumnType) (v int, err error) {
	switch ct {
	case CT_INTEGER:
		return 1, err
	case CT_STRING:
		return 2, err
	case CT_FLOAT:
		return 3, err
	case CT_BLOB:
		return 4, err
	default:
		return -1, fmt.Errorf("[types|ColumnTypeToInt] columnType not supported")
	}
}

func IndexTypeToInt(it IndexType) (v int) {
	switch it {
	case IT_HASHTREE:
		return 1
	case IT_BPLUSTREE:
		return 2
	case IT_FULLTEXT:
		return 3
	/*
		case "FRACTAL":
			//return  IT_FRACTALTREE
	*/
	case IT_NONE:
		return 0
	default:
		return 0
	}
}

// converts a Row value to bytes, using appropriate ColumnType, for chunk storage
func convertValueToBytes(columnType ColumnType, value interface{}) (k []byte, err error) {
	//fmt.Printf("[types:convertValueToBytes]: CONVERT %v (columnType %v)\n", value, columnType)
	switch svalue := value.(type) {
	case int, int64:
		i := fmt.Sprintf("%d", svalue)
		k = StringToBytes(columnType, i)
	case (float64):
		f := ""
		switch columnType {
		case CT_INTEGER:
			f = fmt.Sprintf("%d", int(svalue))
		case CT_FLOAT:
			f = fmt.Sprintf("%f", svalue)
		case CT_STRING:
			f = fmt.Sprintf("%f", svalue)
		}
		k = StringToBytes(columnType, f)
	case (string):
		k = StringToBytes(columnType, svalue)
	case ([]byte):
		k = svalue
	default:
		return k, fmt.Errorf("[types:convertValueToBytes] unknown type (%v) for value (%v)", reflect.TypeOf(svalue), svalue)
	}
	return k, nil
}

func Rng() *mathutil.FC32 {
	x, err := mathutil.NewFC32(math.MinInt32/4, math.MaxInt32/4, false)
	if err != nil {
		panic(err)
	}
	return x
}
func SignHash(unencrypted []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(unencrypted), unencrypted)
	return wolkcommon.Computehash([]byte(msg))
}

const (
	Hash64Length = 8
)

var (
	hash64T = reflect.TypeOf(Hash64{})
)

type Hash64 [Hash64Length]byte

func BytesToHash64(b []byte) Hash64 {
	var h Hash64
	h.SetBytes(b)
	return h
}

func BigToHash64(b *big.Int) Hash64 { return BytesToHash64(b.Bytes()) }

func HexToHash64(s string) Hash64 { return BytesToHash64(common.FromHex(s)) }

func Uint64ToHash64(i uint64) Hash64 { return HexToHash64(hexutil.EncodeUint64(i)) }

func (h Hash64) Bytes() []byte { return h[:] }

func (h Hash64) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }

func (h Hash64) Hex() string { return hexutil.Encode(h[:]) }

func (h Hash64) String() string {
	return fmt.Sprintf("0x%s", h.Hex())
}

func (h *Hash64) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Hash64", input, h[:])
}

func (h *Hash64) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(hash64T, input, h[:])
}

func (h Hash64) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

func (h *Hash64) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-Hash64Length:]
	}

	copy(h[Hash64Length-len(b):], b)
}

// old "sqlcommon"
type Column struct {
	ColumnName string     `json:"columnname,omitempty"` // e.g. "accountID"
	IndexType  IndexType  `json:"indextype,omitempty"`  // IT_BTREE
	ColumnType ColumnType `json:"columntype,omitempty"`
	Primary    int        `json:"primary,omitempty"`
	QueryID    int        `json:"queryid,omitempty"`
}

type SQLRequest struct {
	RequestType string      `json:"requesttype"` //"OpenConnection, Insert, Get, Put, etc"
	Owner       string      `json:"owner,omitempty"`
	Database    string      `json:"database,omitempty"`
	Table       string      `json:"table,omitempty"` //"contacts"
	Encrypted   int         `json:"encrypted,omitempty"`
	Key         interface{} `json:"key,omitempty"`  //value of the key, like "rodney@wolk.com"
	Rows        []Row       `json:"rows,omitempty"` //value of val, usually the whole json record
	Columns     []Column    `json:"columns,omitempty"`
	RawQuery    string      `json:"query,omitempty"` //"Select name, age from contacts where email = 'blah'"
	//BlockNumber int64       `json:"blocknumber,omitempty"`
}

type SQLResponse struct {
	Error            *SQLError `json:"error,omitempty"`
	ErrorCode        int       `json:"errorcode,omitempty"`
	ErrorMessage     string    `json:"errormessage,omitempty"`
	Data             []Row     `json:"data"`
	AffectedRowCount int       `json:"affectedrowcount,omitempty"`
	MatchedRowCount  int       `json:"matchedrowcount,omitempty"`
}

func (resp *SQLResponse) String() string {
	return resp.Stringify()
}

func (resp *SQLResponse) Stringify() string {
	/*
	   wolkErr, ok := resp.Error.(*SQL.SQLError)
	   if !ok {
	           return (`{ "errorcode":-1, "errormessage":"UNKNOWN ERROR"}`) //TODO: Make Default Error Handling
	   }
	   if wolkErr.ErrorCode == 0 { //FYI: default empty int is 0. maybe should be a pointer.  //TODO this is a hack with what errors are being returned right now
	           //fmt.Printf("wolkErr.ErrorCode doesn't exist\n")
	           respObj.ErrorCode = 474
	           respObj.ErrorMessage = resp.Error.Error()
	   } else {
	           respObj.ErrorCode = wolkErr.ErrorCode
	           respObj.ErrorMessage = wolkErr.ErrorMessage
	   }
	*/
	jbyte, jErr := json.Marshal(resp)
	if jErr != nil {
		//fmt.Printf("Error: [%s] [%+v]", jErr.Error(), resp)
		return `{ "errorcode":474, "errormessage":"ERROR Encountered Generating Response"}` //TODO: Make Default Error Handling
	}
	jstr := string(jbyte)
	return jstr
}

type ColumnType string
type IndexType string
type RequestType string

// type Chunk struct {
// 	ChunkID []byte `json:"chunkID"`
// 	Value   []byte `json:"val"`
// 	OK      bool   `json:"ok"`
// }

//note: these are the only consts needed for client, SQL has a much larger list
const (
	CT_INTEGER = "INTEGER"
	CT_STRING  = "STRING"
	CT_FLOAT   = "FLOAT"
	CT_BLOB    = "BLOB"

	IT_NONE      = "NONE"
	IT_HASHTREE  = "HASH"
	IT_BPLUSTREE = "BPLUS"
	IT_FULLTEXT  = "FULLTEXT"

	RT_CREATE_DATABASE = "CreateDatabase"
	RT_LIST_DATABASES  = "ListDatabases"
	RT_DROP_DATABASE   = "DropDatabase"

	RT_CREATE_TABLE   = "CreateTable"
	RT_DESCRIBE_TABLE = "DescribeTable"
	RT_LIST_TABLES    = "ListTables"
	RT_DROP_TABLE     = "DropTable"
	RT_CLOSE_TABLE    = "CloseTable" //moon branch only

	RT_START_BUFFER = "StartBuffer"
	RT_FLUSH_BUFFER = "FlushBuffer"

	RT_PUT    = "Put"
	RT_GET    = "Get"
	RT_DELETE = "Delete"
	RT_QUERY  = "Query"
	RT_SCAN   = "Scan"
)

type Row map[string]interface{}

func NewRow() (r Row) {
	r = make(map[string]interface{})
	return r
}

func (r Row) Set(columnName string, val interface{}) {
	r[columnName] = val
}

//assignRowColumnTypes' new version:
func assignRowColumnTypes(columns map[string]Column, rows []Row) ([]Row, error) {

	for _, row := range rows {
		for name, value := range row {
			if c, ok := columns[name]; !ok {
				return rows, fmt.Errorf("[SQLChainlib:assignRowColumnTypes] Invalid column %s", name)
			} else {
				switch c.ColumnType {
				case CT_INTEGER:
					switch value.(type) {
					case int:
						row[name] = value.(int)
					case float64:
						row[name] = int(value.(float64))
					case string:
						f, err := strconv.ParseFloat(value.(string), 64)
						if err != nil {
							return rows, fmt.Errorf("[SQLChainlib:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, columns[name].ColumnType)
						}
						row[name] = int(f)
					default:
						return rows, fmt.Errorf("[SQLChainlib:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, columns[name].ColumnType)
					}
				case CT_STRING:
					switch value.(type) {
					case string:
						row[name] = value.(string)
					case int:
						row[name] = strconv.Itoa(value.(int))
					case float64:
						row[name] = strconv.FormatFloat(value.(float64), 'f', -1, 64)
						//TODO: handle err
					default:
						return rows, fmt.Errorf("[SQLChainlib:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, columns[name].ColumnType)
					}
				case CT_FLOAT:
					switch value.(type) {
					case float64:
						row[name] = value.(float64)
					case int:
						row[name] = float64(value.(int))
					case string:
						f, err := strconv.ParseFloat(value.(string), 64)
						if err != nil {
							return rows, fmt.Errorf("[SQLChainlib:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, columns[name].ColumnType)
						}
						row[name] = f
					default:
						return rows, fmt.Errorf("[SQLChainlib:assignRowColumnTypes] TypeConversion Error: value [%v] does not match column type [%v]", value, columns[name].ColumnType)
					}
				//case CT_BLOB:
				// TODO: add blob support
				default:
					return rows, fmt.Errorf("[SQLChainlib:assignRowColumnTypes] Coltype not found Value [%v] [%v]", value, columns[name].ColumnType)
				}
			}
		}
	}
	return rows, nil
}

func IsReadQuery(rawQuery string) bool {
	return strings.HasPrefix(strings.ToLower(strings.Trim(rawQuery, " ")), "select")
}

type SQLError struct {
	Message      string
	ErrorCode    int
	ErrorMessage string
}

func (e *SQLError) Error() string {
	return e.Message
}

func (e *SQLError) SetError(m string) {
	e.Message = m
}
