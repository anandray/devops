package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/client"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk"
)

const (
	UtilityName = "wcloud"
)

type WolkCloud struct {
	client  *client.WolkClient
	address common.Address
	verbose bool
}

func NewWolkCloud(_server string, _httpPort int, _name string, _verbose bool) (wcloud *WolkCloud, err error) {
	cl, err := client.NewWolkClient(_server, _httpPort, _name)
	if err != nil {
		return wcloud, err
	}
	if _verbose {
		cl.SetVerbose(_verbose)
	}

	wcloud = &WolkCloud{
		client:  cl,
		verbose: _verbose,
	}
	return wcloud, nil
}

func (self *WolkCloud) AddFile(args []string, options *wolk.RequestOptions) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s add localfile", UtilityName)
	}
	fn := args[1]
	if self.verbose {
		fmt.Printf("AddFile(%x) ...", fn)
	}
	start := time.Now()
	addr, sz, err := self.client.AddFile(fn, options)
	fmt.Printf("%x\n", addr)
	if self.verbose {
		fmt.Printf("Filesize: %d Time: %s\n", sz, time.Since(start))
	}
	return nil
}

func (self *WolkCloud) CatFile(args []string, options *wolk.RequestOptions) (err error) {
	start := time.Now()
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s cat filehash", UtilityName)
	}
	filehashString := args[1]
	if len(filehashString) != 64 && len(filehashString) != 66 {
		return fmt.Errorf("Invalid file hash %s", filehashString)
	}
	filehash := common.HexToHash(filehashString)
	if self.verbose {
		fmt.Printf("CatFile(%x)...", filehash)
	}
	output, err := self.client.CatFile(filehash, options)
	os.Stdout.Write(output)
	if self.verbose {
		fmt.Printf("\nSz: %d Time: %s\n", len(output), time.Since(start))
	}
	return nil
}

// FILE
func (self *WolkCloud) PutFile(args []string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	if len(args) < 3 {
		return txhash, fmt.Errorf("Usage: %s put localfile wolk://owner/bucket/file", UtilityName)
	}
	var localfile = args[1]
	var url = args[2]
	if self.verbose {
		fmt.Printf("PutFile(localfile:%s, wolkurl:%s) ...", localfile, url)
	}

	owner, collection, key, err := client.ParseWolkUrl(url)
	if err != nil {
		return txhash, err
	}
	path := path.Join(owner, collection, key)
	start := time.Now()
	txhash, err = self.client.PutFile(localfile, path, options)
	if self.verbose {
		fmt.Printf("Time: %s\n", time.Since(start))
	}
	return txhash, err
}

func (self *WolkCloud) GetName(args []string, options *wolk.RequestOptions) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s getname name", UtilityName)
	}
	var name = args[1]
	if self.verbose {
		fmt.Printf("name: %s\n", name)
	}
	addr, err := self.client.GetName(name, options)
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", addr)
	return nil
}
func (self *WolkCloud) SetDefaultAccount(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s setdefault name", UtilityName)
	}
	var name = args[1]
	if self.verbose {
		fmt.Printf("SetDefaultAccount(name: %s)\n", name)
	}
	err = self.client.SetDefaultAccount(name)
	if err != nil {
		return err
	}
	return nil
}

func (self *WolkCloud) CreateAccount(args []string) (txhash common.Hash, err error) {
	if len(args) < 2 {
		return txhash, fmt.Errorf("Usage: %s createaccount name", UtilityName)
	}
	var name = args[1]
	if self.verbose {
		fmt.Printf("CreateAccount(name: %s)\n", name)
	}
	txhash, err = self.client.CreateAccount(name)
	if err != nil {
		return txhash, err
	}
	//fmt.Printf("[wcloud:CreateAccount] txhash %x\n", txhash)
	return txhash, nil
}

func (self *WolkCloud) SetQuota(args []string) (txhash []byte, err error) {
	if len(args) < 2 {
		return txhash, fmt.Errorf("Usage: %s setquota amount", UtilityName)
	}
	amount, err := strconv.Atoi(args[1])
	if self.verbose {
		fmt.Printf("SetQuota(name: %d)\n", amount)
	}
	txhash, err = self.client.SetQuota(uint64(amount))
	if err != nil {
		return txhash, err
	}
	return txhash, nil
}

func (self *WolkCloud) SetShimURL(args []string) (txhash []byte, err error) {
	if len(args) < 3 {
		return txhash, fmt.Errorf("Usage: %s setshimurl wolk://owner/collection shimurl", UtilityName)
	}

	var url = args[1]
	var shimURL = args[2]
	owner, collection, _, err := client.ParseWolkUrl(url)
	if err != nil {
		return txhash, err
	}
	if self.verbose {
		fmt.Printf("SetShimURL(owner:%s, collection:%s, shimURL: %s)\n", owner, collection, shimURL)
	}
	start := time.Now()
	txhash, err = self.client.SetShimURL(owner, collection, shimURL)
	if err != nil {
		fmt.Printf("[wcloud:SetShimURL] error %s", err)
		return txhash, err
	}
	if self.verbose {
		fmt.Printf("Time: %s\n", time.Since(start))
	}
	return txhash, nil
}

func (self *WolkCloud) getName() string {
	return self.client.GetSelfName()
}

func (self *WolkCloud) createBucket(bucketType uint8, url string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	owner, collection, _, err := client.ParseWolkUrl(url)
	if self.verbose {
		fmt.Printf("CreateBucket(bucket: %v)\n", collection)
	}
	if strings.Compare(self.getName(), owner) != 0 {
		return txhash, fmt.Errorf("Signed in as %s but trying to create bucket for %s\n", self.getName(), owner)
	}

	txhash, err = self.client.CreateBucket(bucketType, self.getName(), collection, options)
	if err != nil {
		return txhash, err
	}
	return txhash, nil
}

func (self *WolkCloud) Delete(args []string) (txhash common.Hash, err error) {
	if len(args) < 2 {
		return txhash, fmt.Errorf("Usage: %s delete wolk://owner/collection/key", UtilityName)
	}
	var url = args[1]
	owner, collection, key, err := client.ParseWolkUrl(url)
	if self.verbose {
		fmt.Printf("Delete(url %v)\n", url)
	}
	if len(key) > 0 {
		txhash, err = self.client.Delete(path.Join(owner, collection, key), nil)
		if err != nil {
			return txhash, err
		}
	}
	txhash, err = self.client.DeleteBucket(owner, collection)
	if err != nil {
		return txhash, err
	}
	return txhash, nil
}

func (self *WolkCloud) CreateBucket(args []string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	if len(args) < 2 {
		return txhash, fmt.Errorf("Usage: %s mkdir wolk://owner/bucket", UtilityName)
	}
	if len(args) > 2 {
		options.ShimURL = args[2]
	}
	return self.createBucket(wolk.BucketFile, args[1], options)
}

func (self *WolkCloud) CreateCollection(args []string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	if len(args) < 2 {
		return txhash, fmt.Errorf("Usage: %s createcoll wolk://owner/collection", UtilityName)
	}
	if len(args) > 2 {
		options.ShimURL = args[2]
	}
	return self.createBucket(wolk.BucketNoSQL, args[1], options)
}

func (self *WolkCloud) CreateDatabase(args []string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	if len(args) < 2 {
		return txhash, fmt.Errorf("Usage: %s createdb wolk://owner/database", UtilityName)
	}
	if len(args) > 2 {
		options.ShimURL = args[2]
	}

	return self.createBucket(wolk.BucketSQL, args[1], options)
}

// NoSQL
func (self *WolkCloud) SetKey(args []string, options *wolk.RequestOptions) (txhash common.Hash, err error) {
	if len(args) < 3 {
		return txhash, fmt.Errorf("Usage: %s setkey wolk://owner/collection/key val", UtilityName)
	}
	var url = args[1]
	var val = args[2]
	owner, collection, key, err := client.ParseWolkUrl(url)
	if err != nil {
		return txhash, err
	}
	if self.verbose {
		fmt.Printf("SetKey(owner:%s, collection:%s, key: %s, val: %s)\n", owner, collection, key, val)
	}
	start := time.Now()
	txhash, err = self.client.SetKey(owner, collection, key, []byte(val))
	if err != nil {
		return txhash, err
	}
	if self.verbose {
		fmt.Printf("Time: %s\n", time.Since(start))
	}
	return txhash, nil
}

func (self *WolkCloud) Get(args []string, options *wolk.RequestOptions) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s get wolk://owner/collection/key [shows value of key]", UtilityName)
		return fmt.Errorf("Usage: %s get wolk://owner/collection     [shows all items in collection]", UtilityName)
		return fmt.Errorf("Usage: %s get wolk://owner                [shows all collections of owner]", UtilityName)
	}
	var url = args[1]
	if self.verbose {
		fmt.Printf("Get(%s, options: %s)\n", url, options)
	}
	owner, collection, key, err := client.ParseWolkUrl(url)
	if err != nil {
		return err
	}
	if len(key) == 0 && len(collection) == 0 {
		buckets, err := self.client.GetCollections(owner, options)
		if err != nil {
			return err
		}
		fmt.Printf("%20s\t%10s\t%8s\tWriters\n", "Name", "BucketType", "Req Pays")
		for _, b := range buckets {
			fmt.Printf("%20s\t%10s\t%8d\t%v\n",
				b.Name, wolkcommon.BucketTypeToString(b.BucketType), b.RequesterPays, b.Writers)
		}
		return err
	}
	if len(key) == 0 && len(collection) > 0 {
		txb, items, err := self.client.GetCollection(owner, collection, nil)
		if err != nil {
			return err
		}
		if txb.BucketType == wolk.BucketNoSQL {
			for _, item := range items {
				var ctx context.Context
				k := item.Key
				chunkHash := item.ValHash
				sz := item.Size
				val, ok, err := self.client.GetChunk(ctx, chunkHash.Bytes(), options)
				if err != nil {
					fmt.Printf(" %s\t%s [%s]\n", k, chunkHash, err)
				} else if !ok {
					fmt.Printf(" %s\t%s [NOT FOUND]\n", k, chunkHash)
				} else {
					fmt.Printf("{\"key\":\"%s\",\"val\":\"%s\",\"size\":%d}\n", k, val, sz)
				}
			}
		} else if txb.BucketType == wolk.BucketSQL {

		} else {
			fmt.Printf("%10s\t%32s\t%64s\t%42s\n", "Size", "Name", "Val Hash", "Last Updated By")
			for _, i := range items {
				addr := i.Writer
				fmt.Printf("%10d\t%32s\t%x\t0x%x\n", i.Size, i.Key, i.ValHash, addr)
			}
		}
	}

	start := time.Now()
	val, sz, err := self.client.GetKey(owner, collection, key, options)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(val))
	if self.verbose {
		fmt.Printf("Size: %d Time: %s\n", sz, time.Since(start))
	}
	return nil
}

// wcloud sql wolk://alina/db3 "select * from accounts"
// wcloud sql wolk://alina/db3 "insert into accounts"
func (self *WolkCloud) SQL(args []string, options *wolk.RequestOptions) (txhash common.Hash, checktx bool, err error) {
	if len(args) < 3 {
		return txhash, false, fmt.Errorf("Usage: %s sql wolk://owner/database \"select * from ...\"", UtilityName)
	}
	var wolkurl = args[1]
	var inputsql = args[2]
	if self.verbose {
		fmt.Printf("SQL(wolkurl:%s, sql:%s)", wolkurl, inputsql)
	}
	owner, database, _, err := client.ParseWolkUrl(wolkurl)
	if err != nil {
		return txhash, false, fmt.Errorf("[wcloud:SQL] %s", err)
	}
	start := time.Now()

	eventType, sqlReq, err := wolk.ParseRawRequest(inputsql)
	if err != nil {
		return txhash, false, fmt.Errorf("[wcloud:SQL] %s", err)
	}
	sqlReq.Owner = owner
	sqlReq.Database = database
	//log.Info("[wcloud:SQL] sqlReq is parsed", "sqlReq", sqlReq)

	switch eventType {
	case "read":
		resp, err := self.client.ReadSQL(owner, database, sqlReq, options)
		if err != nil {
			return txhash, false, fmt.Errorf("[wcloud:SQL] %s", err)
		}
		fmt.Printf("%+v\n", resp)
	case "mutate":
		txhash, err = self.client.MutateSQL(owner, database, sqlReq)
		if err != nil {
			return txhash, false, fmt.Errorf("[wcloud:SQL] %s", err)
		}
		//fmt.Printf("[wcloud:SQL] MutateSQL tx(%x)\n", txhash)
	default:
		return txhash, false, fmt.Errorf("[wcloud:SQL] sql eventType not mutate nor read")
	}
	if self.verbose {
		fmt.Printf("Time: %s\n", time.Since(start))
	}
	//fmt.Printf("[wcloud:SQL] SQL at the end, txhash(%x)\n", txhash)
	return txhash, true, nil
}

func (self *WolkCloud) GetChunk(args []string, options *wolk.RequestOptions) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s getchunk chunkhash", UtilityName)
	}
	var chunkHashString = args[1]
	chunkHash := common.HexToHash(chunkHashString)
	if self.verbose {
		fmt.Printf("GetChunk(chunkhash:%x)\n", chunkHash)
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	v, ok, err := self.client.GetChunk(ctx, chunkHash.Bytes(), options)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("No chunk hash %x", chunkHash)
	}
	fmt.Printf("%s", v)
	if self.verbose {
		fmt.Printf("Time: %s\n", time.Since(start))
	}
	return nil
}

func (self *WolkCloud) SetChunk(args []string, options *wolk.RequestOptions) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s setchunk filename", UtilityName)
	}
	var filename = args[1]
	if self.verbose {
		fmt.Printf("SetChunk(filename:%s)", filename)
	}

	start := time.Now()
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}
	var ctx context.Context
	chunkHash, err := self.client.SetChunk(ctx, content, options)
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", chunkHash)
	if self.verbose {
		fmt.Printf("Time: %s\n", time.Since(start))
	}
	return nil
}

func (self *WolkCloud) GetShare(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s getshare chunkhash", UtilityName)
	}
	var chunkHashString = args[1]
	chunkHash := common.HexToHash(chunkHashString)
	if self.verbose {
		fmt.Printf("GetShare(chunkhash:%x)\n", chunkHash)
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()
	v, ok, err := self.client.GetShare(ctx, chunkHash.Bytes())
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("No chunk hash %x", chunkHash)
	}
	fmt.Printf("%s\n", v)
	if self.verbose {
		fmt.Printf("Time: %s\n", time.Since(start))
	}
	return nil
}

func (self *WolkCloud) SetShare(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s setshare filename", UtilityName)
	}
	var filename = args[1]
	if self.verbose {
		fmt.Printf("SetShare(filename:%s)", filename)
	}

	start := time.Now()
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var ctx context.Context
	chunkHash, err := self.client.SetShare(ctx, content)
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", chunkHash)
	if self.verbose {
		fmt.Printf("Time: %s\n", time.Since(start))
	}
	return nil
}

func (self *WolkCloud) Transfer(args []string) (txhash common.Hash, err error) {
	if len(args) < 3 {
		return txhash, fmt.Errorf("Usage: %s transfer [recipient] [amount]", UtilityName)
	}
	var recipient = args[1]
	var amountString = args[2]
	amount, err := strconv.Atoi(amountString)
	if self.verbose {
		fmt.Printf("Transfer(recipient: %x, amount: %d)", recipient, amount)
	}
	txhash, err = self.client.Transfer(recipient, uint64(amount))
	if err != nil {
		return txhash, err
	}
	return txhash, nil
}

func (self *WolkCloud) RegisterNode(args []string) (txhash common.Hash, err error) {
	if len(args) < 6 {
		return txhash, fmt.Errorf("Usage: %s registernode [nodenumber] [storageip] [consensusip] [region] [value]", UtilityName)
	}
	nodenumber, err := strconv.Atoi(args[1])
	if err != nil {
		return txhash, err
	}
	node := uint64(nodenumber)
	var ip = args[2]
	var consensusip = args[3]
	regionnumber, err := strconv.Atoi(args[4])
	if err != nil {
		return txhash, err
	}
	region := uint8(regionnumber)
	value, err := strconv.Atoi(args[5])
	if err != nil {
		return txhash, err
	}
	if self.verbose {
		fmt.Printf("RegisterNode(nodenumber: %d, ip: %s, consensusip: %s, region: %d, value: %d)", nodenumber, ip, consensusip, region, value)
	}
	txhash, err = self.client.RegisterNode(node, ip, consensusip, region, uint64(value))
	if err != nil {
		return txhash, err
	}
	return txhash, nil
}

func (self *WolkCloud) GetBlock(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s block [latest|blocknumber]", UtilityName)
	}
	var block = args[1]
	var blocknumber int
	if block == "latest" {
		blocknumber, err = self.client.LatestBlockNumber()
		if err != nil {
			return err
		}
		fmt.Printf("LATEST BLOCK: %d\n", blocknumber)
	} else {
		blocknumber, err = strconv.Atoi(block)
		if err != nil {
			return err
		}
	}
	if self.verbose {
		fmt.Printf("GetBlock(blocknumber: %d)", blocknumber)
	}
	b, err := self.client.GetBlock(blocknumber)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", b.String())
	return nil
}

func (self *WolkCloud) GetAccount(args []string, options *wolk.RequestOptions) (err error) {
	name := args[1]
	account, err := self.client.GetAccount(name, options)
	if self.verbose {
		fmt.Printf("GetAccount(name: %s, options: %s)", name, options)
	}
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", account.String())
	return nil
}

func (self *WolkCloud) GetTransaction(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s tx [txhash]", UtilityName)
	}
	var txhash = args[1]
	if self.verbose {
		fmt.Printf("GetTransaction(txhash: %s)", txhash)
	}
	txh := common.HexToHash(txhash)
	tx, err := self.client.GetTransaction(txh)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", tx.String())
	return nil
}

func (self *WolkCloud) GetNode(args []string, options *wolk.RequestOptions) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s node [nodenumber]", UtilityName)
	}
	nodenumber, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}
	if self.verbose {
		fmt.Printf("GetNode(nodenumber: %d)", nodenumber)
	}
	n, err := self.client.GetNode(nodenumber, options)
	fmt.Printf("%s\n", n.String())
	return nil
}

func (self *WolkCloud) GetPeers() (err error) {
	if self.verbose {
		fmt.Printf("GetPeers()")
	}
	n, err := self.client.GetPeers()
	fmt.Printf("%d\n", n)
	return nil
}

func (self *WolkCloud) getWriters(writers string) (validWriters []common.Address) {
	sa := strings.Split(writers, ",")
	validWriters = make([]common.Address, 0)
	for _, s := range sa {
		if len(s) == 40 || len(s) == 42 {
			addr := common.HexToAddress(s)
			validWriters = append(validWriters, addr)
		}
	}
	return validWriters
}

func (self *WolkCloud) CheckTxHash(txhash common.Hash, waitfortx bool) {
	// TODO: report on tx inclusion in blockchain
	if waitfortx {
		stx, err := self.client.WaitForTransaction(txhash)
		if err != nil {
			fmt.Printf("%v", err)
			return
		}
		fmt.Printf("%s", stx.String())
	} else {
		fmt.Printf("{\"txhash\":\"%x\"}\n", txhash)
	}
}

func getenvInt(key string, def int) int {
	res := def
	str := os.Getenv(key)
	if len(str) > 0 {
		alt, err := strconv.Atoi(str)
		if err == nil {
			res = alt
		}
	}
	return res
}

func getenvStr(key string, def string) string {
	res := def
	str := os.Getenv(key)
	if len(str) > 0 {
		return str
	}
	return res
}

func main() {
	var verbose = flag.Bool("v", false, "verbose")
	defaultEndpoint := getenvStr("CLIENTENDPOINT", "cloud.wolk.com")
	server := flag.String("server", defaultEndpoint, "Cloudstore Endpoint")
	defaultHTTPPort := getenvInt("HTTPPORT", 443)
	httpport := flag.Int("httpport", defaultHTTPPort, "HTTP Port")
	var frange = flag.String("range", "all", "Range")
	var requesterpays = flag.Int("requesterpays", 0, "Requester Pays")
	var writers = flag.String("writers", "", "Writers")
	var blocknum = flag.Int("blocknumber", 0, "Block Number")
	var proof = flag.Bool("proof", false, "Show Proof")
	var history = flag.Bool("history", false, "Get Key history")
	var waitfortx = flag.Bool("waitfortx", false, "Wait For Tx Inclusion")
	//var du = flag.Bool("du", false, "Show Storage Usage")
	var schema = flag.String("schema", "wolk://wolk/schema/Thing", "Schema URL (wolk://wolk/schema/Thing)")
	var encryption = flag.String("encryption", "none", "Encryption (none, rsa, symmetric)")
	var maxcontentlength = flag.Int("maxcontentlength", 10000000, "Max Content Length (bytes)")
	var name = flag.String("name", "", "Account name")

	flag.Parse()
	offset := *httpport - 80
	if offset > 9 {
		offset = 9
	}
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace%d", offset))

	if *verbose {
		fmt.Printf("Wclient Setting:[server=%s] [httpport=%v]\n", *server, *httpport)
	}

	if len(flag.Args()) < 1 {
		fmt.Printf("Wolk Cloudstore: (*-Not implemented/tested yet)\n\n")
		fmt.Printf("Names/Accounts:\n")
		fmt.Printf("-GetName:      %s getname  name                     (returns address)\n", UtilityName)
		fmt.Printf(" GetAccount:   %s account  name                     (returns account)\n", UtilityName)
		fmt.Printf(" CreateAccount:%s createaccount name                (creates keys in ~/.wolk/name directory)\n", UtilityName)
		fmt.Printf(" SetDefault:   %s setdefault name                   (sets default name in ~/.wolk/name directory)\n\n", UtilityName)

		fmt.Printf("File Buckets:\n")
		fmt.Printf(" Make Bucket:  %s mkdir  wolk://owner/bucket\n", UtilityName)
		fmt.Printf(" List Buckets: %s get    wolk://owner                (same as ls wolk://owners/buckets/SCAN)\n", UtilityName)
		fmt.Printf(" Delete Bucket:%s delete wolk://owner/bucket         (same as rm wolk://owners/buckets/bucket)\n", UtilityName)
		fmt.Printf(" Put File:     %s put    localfile wolk://owner/bucket/file  (same as setkey wolk://owner/bucket/file @localfile)\n", UtilityName)
		fmt.Printf(" Get File:     %s get    wolk://owner/bucket/file localfile  (same as getkey wolk://owner/bucket/file)\n", UtilityName)
		fmt.Printf(" Delete File:  %s delete wolk://owner/bucket/file\n", UtilityName)
		fmt.Printf(" List Bucket:  %s get    wolk://owner/bucket/\n", UtilityName)

		fmt.Printf("NoSQL Collections:\n")
		fmt.Printf(" + Collection: %s createcoll wolk://owner/collection  (same as setkey wolk://owner/buckets/collection)\n", UtilityName)
		fmt.Printf(" Scan Colls:   %s get        wolk://owner/            (same as ls wolk://owner/buckets/SCAN)\n", UtilityName)
		fmt.Printf(" - Collection: %s delete     wolk://owner/collection  (same as rm wolk://owner/buckets/collection)\n", UtilityName)
		fmt.Printf(" Set Key:      %s put        wolk://owner/collection/key val\n", UtilityName)
		fmt.Printf(" Get Key:      %s get        wolk://owner/collection/key\n", UtilityName)
		fmt.Printf(" Delete Key:   %s delete     wolk://owner/collection/key\n", UtilityName)
		fmt.Printf(" Scan:         %s ls         wolk://owner/collection/ (same as ls wolk://owner/collection/SCAN)\n\n", UtilityName)

		fmt.Printf("SQL Databases:\n")
		fmt.Printf(" Create DB:    %s createdb wolk://owner/database    (same as mkdir wolk://owner/database)\n", UtilityName)
		fmt.Printf("*Show DBs:     %s ls       wolk://owner/            (same as ls    wolk://owner/buckets/SCAN)\n", UtilityName)
		fmt.Printf("*Drop DB:      %s delete   wolk://owner/database    (same as rm    wolk://owner/buckets/database)\n", UtilityName)
		fmt.Printf("*Show Tables:  %s sql      wolk://owner/database  \"SHOW TABLES\"\n", UtilityName)
		fmt.Printf(" Create Table: %s sql      wolk://owner/database  \"CREATE TABLE tablename ...\"\n", UtilityName)
		fmt.Printf("*Drop Table:   %s sql      wolk://owner/database  \"DROP TABLE tablename\"\n", UtilityName)
		fmt.Printf(" Mutate:       %s sql      wolk://owner/database  \"INSERT into tablename ...\"\n", UtilityName)
		fmt.Printf(" Read:         %s sql      wolk://owner/database  \"SELECT * from tablename ...\"\n", UtilityName)

		fmt.Printf("Blockchain:\n")
		fmt.Printf(" GetLatestBlock:  %s blocknumber                     (curl /wolk/latest/blocknumber]\n", UtilityName)
		fmt.Printf(" GetBalance:      %s balance address|name            (curl /wolk/balance/address]\n", UtilityName)
		fmt.Printf(" GetBlock:        %s block blocknumber               [curl /wolk/block/blocknumber]\n", UtilityName)
		fmt.Printf(" GetTransaction:  %s tx  txhash                      [curl /wolk/tx/txhash]\n", UtilityName)
		fmt.Printf(" GetNode:         %s node nodenumber                 [curl /wolk/node/nodenumber]\n", UtilityName)
		fmt.Printf(" GetNumPeers:     %s peers                           [curl /wolk/peers]\n", UtilityName)
		fmt.Printf(" Transfer:        %s transfer recipient amount \n", UtilityName)
		fmt.Printf("*RegisterNode:    %s registernode [nodenumber] [storageip] [consensusip] [region] [value]\n", UtilityName)

		fmt.Printf("File: (for debugging only)\n")
		fmt.Printf(" Store:        %s add   localfile                    (returns filehash)\n", UtilityName)
		fmt.Printf(" Retrieve:     %s cat   filehash\n", UtilityName)

		fmt.Printf("Chunk: (for debugging only)\n")
		fmt.Printf(" GetChunk:     %s getchunk chunkhash                 [curl /wolk/chunk/chunkhash]\n", UtilityName)
		fmt.Printf(" SetChunk:     %s setchunk file                      [curl /wolk/chunk/]\n\n", UtilityName)

		fmt.Printf("Share: (for debugging only)\n")
		fmt.Printf(" GetShare:     %s getshare chunkhash                 [curl /wolk/share/chunkhash]\n", UtilityName)
		fmt.Printf(" SetShare:     %s setshare file                      [curl /wolk/share/]\n\n", UtilityName)

		fmt.Printf("Arguments: (optional; must be suppied first)\n")
		fmt.Printf("  server=%s [all], export CLIENTENDPOINT=server\n", *server)
		fmt.Printf("  httpport=%v,  export HTTPPORT=port\n", *httpport)
		fmt.Printf("  -range=80-160    [cat]\n")
		fmt.Printf("  -maxcontentlength=1000000    [createcoll, createdb, mkdir, getkey, get, sql]\n")
		fmt.Printf("  -requesterpays=0 [createcoll, createdb, mkdir]\n")
		fmt.Printf("  -writers=0x1234...,0xabcd... [createcoll, createdb, mkdir]\n")
		fmt.Printf("  -name=someone\n")
		fmt.Printf("  -proof=true      [getkey, get, getname, sql]\n")
		fmt.Printf("  -blocknumber=23  [balance, node, getkey, getname, sql]\n")
		fmt.Printf("  -history=true    [getkey]\n")
		os.Exit(0)
	}

	flagArgs := flag.Args()
	var op = flag.Arg(0)
	if *verbose {
		fmt.Printf("Wolk File System op: %s\nArgs=%s\n", op, flagArgs)
	}

	// -stage=1 wil set httpport=80+stage
	stageenv := os.Getenv("STAGE")
	if len(stageenv) > 0 {
		stage, err := strconv.Atoi(stageenv)
		if err != nil {
			fmt.Printf("invalid stage [%s]", err)
			os.Exit(0)
		} else if *verbose {
			*httpport = stage + 80
			fmt.Printf("STAGE=[%d]\n Setting httpport=%d \n", stage, *httpport)
		}
	}

	wcloud, err := NewWolkCloud(*server, int(*httpport), *name, *verbose)
	if err != nil {
		fmt.Printf("Client: %s\n", err)
		os.Exit(0)
	}

	if strings.Compare(op, "createaccount") == 0 {
		// exception!
	} else if !wcloud.client.KeysLoaded() {

		fmt.Printf("Need to createaccount first!\n")
		os.Exit(0)
	}
	var txhash common.Hash
	blockNumber := *blocknum
	if strings.Compare(op, "get") == 0 ||
		strings.Compare(op, "ls") == 0 ||
		strings.Compare(op, "getkey") == 0 ||
		strings.Compare(op, "scan") == 0 ||
		strings.Compare(op, "sql") == 0 ||
		strings.Compare(op, "getname") == 0 ||
		strings.Compare(op, "balance") == 0 ||
		strings.Compare(op, "node") == 0 {
		if blockNumber == 0 {
			latestBlockNumber, err := wcloud.client.LatestBlockNumber()
			if err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(0)
			}
			blockNumber = latestBlockNumber
			if *verbose {
				fmt.Printf("Latest Block Number: %d\n", latestBlockNumber)
			}
		}
	}
	options := new(wolk.RequestOptions)
	if *proof {
		options.Proof = true
	}
	if *history {
		options.History = "1"
	}
	options.Encryption = *encryption
	options.BlockNumber = int(*blocknum)
	options.MaxContentLength = uint64(*maxcontentlength)
	options.Range = *frange
	options.RequesterPays = uint8(*requesterpays)
	options.ValidWriters = wcloud.getWriters(*writers)
	options.Schema = *schema

	checktx := false
	if strings.Compare(op, "add") == 0 {
		err = wcloud.AddFile(flag.Args(), options)
	} else if strings.Compare(op, "cat") == 0 {
		err = wcloud.CatFile(flag.Args(), options)
	} else if strings.Compare(op, "put") == 0 {
		txhash, err = wcloud.PutFile(flag.Args(), options)
		checktx = true
	} else if strings.Compare(op, "get") == 0 {
		err = wcloud.Get(flag.Args(), options)
	} else if strings.Compare(op, "mkdir") == 0 {
		txhash, err = wcloud.CreateBucket(flag.Args(), options)
		checktx = true
	} else if strings.Compare(op, "set") == 0 {
		txhash, err = wcloud.SetKey(flag.Args(), options)
		checktx = true
	} else if strings.Compare(op, "delete") == 0 {
		txhash, err = wcloud.Delete(flag.Args())
	} else if strings.Compare(op, "createcoll") == 0 {
		txhash, err = wcloud.CreateCollection(flag.Args(), options)
		checktx = true
	} else if strings.Compare(op, "sql") == 0 {
		txhash, checktx, err = wcloud.SQL(flag.Args(), options)
		//fmt.Printf("[wcloud:main] txhash output: %x\n", txhash)
	} else if strings.Compare(op, "createdb") == 0 {
		txhash, err = wcloud.CreateDatabase(flag.Args(), options)
		checktx = true
	} else if strings.Compare(op, "getchunk") == 0 {
		err = wcloud.GetChunk(flag.Args(), options)
	} else if strings.Compare(op, "setchunk") == 0 {
		err = wcloud.SetChunk(flag.Args(), options)
	} else if strings.Compare(op, "getshare") == 0 {
		err = wcloud.GetShare(flag.Args())
	} else if strings.Compare(op, "setshare") == 0 {
		err = wcloud.SetShare(flag.Args())
	} else if strings.Compare(op, "createaccount") == 0 {
		txhash, err = wcloud.CreateAccount(flag.Args())
		checktx = true
	} else if strings.Compare(op, "getname") == 0 {
		err = wcloud.GetName(flag.Args(), options)
	} else if strings.Compare(op, "setdefault") == 0 {
		err = wcloud.SetDefaultAccount(flag.Args())
	} else if strings.Compare(op, "setquota") == 0 {
		_, err = wcloud.SetQuota(flag.Args())
	} else if strings.Compare(op, "setshimurl") == 0 {
		_, err = wcloud.SetShimURL(flag.Args())
	} else if strings.Compare(op, "account") == 0 {
		err = wcloud.GetAccount(flag.Args(), options)
	} else if strings.Compare(op, "block") == 0 {
		err = wcloud.GetBlock(flag.Args())
	} else if strings.Compare(op, "tx") == 0 {
		err = wcloud.GetTransaction(flag.Args())
	} else if strings.Compare(op, "peers") == 0 {
		err = wcloud.GetPeers()
	} else if strings.Compare(op, "node") == 0 {
		err = wcloud.GetNode(flag.Args(), options)
	} else if strings.Compare(op, "transfer") == 0 {
		txhash, err = wcloud.Transfer(flag.Args())
		checktx = true
	} else if strings.Compare(op, "registernode") == 0 {
		txhash, err = wcloud.RegisterNode(flag.Args())
		checktx = true
	} else {
		err = fmt.Errorf("Unknown operation %s", op)
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else if checktx {
		wcloud.CheckTxHash(txhash, *waitfortx)
	}
}
