
# wcloud Command Line Interface

wcloud is a command line interface

```
Wolk Cloudstore: (*-Not implemented/tested yet)

File Buckets:
 Store:        wcloud add   localfile
 Retrieve:     wcloud cat   filehash
 Put:          wcloud put   localfile wolk://owner/bucket/file
 Get:          wcloud get   wolk://owner/bucket/file localfile
 List Bucket:  wcloud ls    wolk://owner/bucket/
 List Buckets: wcloud ls    wolk://owner
 Make Bucket:  wcloud mkdir wolk://owner/bucket
Owner Names:
 GetName:      wcloud getname  name                      [wolk://wolk/names/name] (returns address)
 SetName:      wcloud setname  name                      [wolk://wolk/names/name] (maps  name to address)

NoSQL:
 SetKey:       wcloud setkey  wolk://owner/collection/key val
 GetKey:       wcloud getkey  wolk://owner/collection/key
 Scan:         wcloud scan    wolk://owner/collection/]

SQL:
*Execute:      wcloud sql     owner database sql         [wolk://owner/database/]

Chunk:
 GetChunk:     wcloud getchunk chunkhash                 [wolk://wolk/chunk/chunkhash]
 SetChunk:     wcloud setchunk file                      [wolk://wolk/chunk/]

Blockchain:
 GetLatestBlock:  wcloud blocknumber                     [wolk://wolk/chain/blocknumber]
 GetBalance:      wcloud balance address|name            [wolk://wolk/balance/address]
 GetBlock:        wcloud block blocknumber               [wolk://wolk/block/blocknumber]
 GetTransaction:  wcloud transaction  txhash             [wolk://wolk/transaction/txhash]
 GetNode:         wcloud node nodenumber                 [wolk://wolk/node/nodenumber]

 Transfer:        wcloud transfer recipientaddr amount
*RegisterNode:    wcloud registernode [nodenumber] [storageip] [consensusip] [region] [value]
*UpdateNode:      wcloud updatenode registernode [nodenumber] [storageip] [consensusip] [region] [value]
*BandwidthCheck:  wcloud bandwidthcheck [checkChunkHash]

Arguments: (optional)
  [all]
  -httpport=80
  -range=80-160    [cat]
  -pkey=privatekey [setkey, setname, sql, transfer, registernode, updatenode, bandwidthcheck]
  -proof=true      [getkey, get, getname, sql]
  -blocknumber=23  [balance, node, getkey, getname, sql]
```


Wolk node serves its users via HTTP and a command line interface.

## Wolk Accounts

Users reserve names and associate their address to a specific string with this [provided it has not yet been chosen]:

```
# wcloud createaccount sourabh
{"txhash":"535c72ce4c5a4b3a78db4a1fe8d432fc3d4bd7135711ae432b2899bf5b94f8e7"}
# wcloud getname sourabh
82a978b3f5962a5b0957d9ee9eef472ee55b42f1
```

## Wolk File Buckets and NoSQL Collections

With a name set up, a user can store files and set/get keys with familiar cloud provider interfaces
that interact with a cloudstore address's has a set of _buckets_, which have a storage quota.
Buckets are of different types: file buckets, NoSQL buckets, SQL buckets, and other bucket types.

For file buckets and NoSQL buckets, there are a set of keys mapped to fileHashes of content,
and the total amount of storage used by all the storage operations.
The same mechanism is used for storing files and NoSQL key-value pairs.

Generally, all operations that store content map to HTTP PUT operations,
and all operations that retrieve content map to HTTP GET operations, with the wcloud interface shown here.

A special collection "buckets" is used to store metadata about the users collection.

```
# wcloud mkdir wolk://sourabh/planets
{"txhash":"db8d40ae0ec6fda0310120c99381d8aa782f9ceea04430f84552fc387d502a40"}

# wcloud get wolk://sourabh/
 33	planets	d083f26804b3b0740cb918cdf3a30a5bbb82d84c8dab35a1cf6ecd47808d72c7
 34	photos	8f17ac69be0a35b67fe25ed45c2dbb65c06fa19deea99bf6ba4f2a4c27730a51
 100000	videos	85e3b06bfb718ad3183fde82d0e1aeae50258354fc10e007b2a986df9c50cedc
# wcloud set wolk://sourabh/planets/pluto small
{"txhash":"8f92d2e796ee0158e0bed61c390b225584244e2831d7a12e589bd3d5e88797ac"}

# wcloud get wolk://sourabh/planets/pluto
small
# wcloud set wolk://sourabh/planets/jupiter big
{"txhash":"550b110c56ee7a28a5ffc4cb9c764f8dc08550eddeaa479a4910309bd7ece90b"}

# wcloud ls wolk://sourabh/planets
5	  pluto	42faffb02208ec37aa24a44ff1b1f8d6d4ae17eab58c5f82f33992cc8e1b545c
3	  jupiter	9c559b43ff30ffc56fc3b4808fdc33eeccf2f09141b4642fb19ca393cb83b142
10	saturn	01f1c01a94cca3c8adf7f2cf00bf5d1da2aafff41d7b6fc9277ab81e8a7cfbfb

# wcloud put wcloud/content/banana.gif wolk://sourabh/photos/banana.gif
{"txhash":"9e15486eae8566e1d53fe3c27b9991ecb48d7363a509edc74f923fc4849c16fe"}

# wcloud get wolk://sourabh/photos/
 73623	banana.gif	6f439c8124e896af214b3cceda4f249d94df3c72a350358a904a38cf8d341cea

# wcloud  account sourabh
{"balance":200003,"quota":516880,"usage":8}
```

### NoSQL Proofs

```
wcloud -proof=1 -v setname sourabh
wcloud -proof=1 -pkey=c89efdaa54c0f20c7adf612882df0950f5a951637e0307cdcb4c672f298b8bc6 -v setname amit
wcloud -proof=1 -pkey=ad7c5bef027816a800da1736444fb58a807ef4c9603b7848673f7e3a68eb14a5 -v setname anjuli

wcloud -proof=1 -v getname amit
wcloud -proof=1 -v getname anjuli
```

## SQL operations

```
# wcloud createaccount alinatest
{"txhash":"0b15841975550ac6214e74017eee7d7644e37d89e3784e8a3966139250695707"}

# wcloud sql wolk://alinatest/ncis createdatabase
{"txhash":"850beeff8801f9bce17d65dc553b19305d205d667c91749a20a3eb4f5e259097"}

# wcloud sql wolk://alinatest/ncistest "create table personnel (person_id int primary key, name string)"
{"txhash":"3891882b0403b7092cd46dbb44c367b12ba92e2b833fc9aac1c0b4c8ada92848"}

# wcloud sql wolk://alinatest/ncis "insert into personnel (person_id, name) values (6, 'dinozzo')"
{"txhash":"af560ee314240c9ed6f3542843a7b702ffe8f3f5b70fe943450a373e0d4d35e6"}

# wcloud sql wolk://alinatest/ncis "insert into personnel (person_id, name) values (23, 'ziva')"
{"txhash":"1752e47167bea218b73c9e8b96b0640f6f2e7bbc025371cd61d9fae5e0e13a8a"}

# wcloud sql wolk://alinatest/ncis "insert into personnel (person_id, name) values (2, 'mcgee')"
{"txhash":"b63f503ee235e44aa64fba34afd01237ec7449d0220e06bb59bdc94ad36f9fc7"}

# wcloud sql wolk://alinatest/ncis "insert into personnel (person_id, name) values (101, 'gibbs')"
{"txhash":"6ab3f77fcbc6a8b808a5b17be7dfd0509fba0085ed179c4b698e5ecb870e627b"}

# wcloud sql wolk://alinatest/ncis "select * from personnel"
{"data":[{"name":"mcgee","person_id":2},{"name":"dinozzo","person_id":6},{"name":"ziva","person_id":23},{"name":"gibbs","person_id":101}],"matchedrowcount":4}
{"txhash":"0000000000000000000000000000000000000000000000000000000000000000"}

# wcloud sql wolk://alinatest/ncis "delete from personnel where name = 'ziva'"
{"txhash":"1feeae04389a67645dcc9ca6c913ff44d1dcb5154d5f865cde8766adde5ebd5f"}

# wcloud sql wolk://alinatest/ncis "select * from personnel"
{"data":[{"name":"mcgee","person_id":2},{"name":"dinozzo","person_id":6},{"name":"gibbs","person_id":101}],"matchedrowcount":3}
{"txhash":"0000000000000000000000000000000000000000000000000000000000000000"}

# wcloud sql wolk://alinatest/ncis "delete from personnel where person_id = 6"
{"txhash":"a80f4eb237c48ed564dab22c2ea127603dcee96551ad37bb2feee17c09b97f7e"}

# wcloud sql wolk://alinatest/ncis "select * from personnel"
{"data":[{"name":"mcgee","person_id":2},{"name":"gibbs","person_id":101}],"matchedrowcount":2}
{"txhash":"0000000000000000000000000000000000000000000000000000000000000000"}

# wcloud sql wolk://alinatest/ncis "describetable personnel"
{"data":[{"ColumnName":"person_id","ColumnType":"INTEGER","IndexType":"BPLUS","Primary":1},{"ColumnName":"name","ColumnType":"STRING","IndexType":"BPLUS","Primary":0}]}
{"txhash":"0000000000000000000000000000000000000000000000000000000000000000"}

# wcloud sql wolk://alinatest/ncis "create table awards (person_id int primary key, name string)"
{"txhash":"c7517b770692658f07331641e55492d8969f7d7b2c68de78d85d791c5657283e"}

# wcloud sql wolk://alinatest/ncis "listtables"
{"data":[{"table":"personnel"},{"table":"awards"}],"matchedrowcount":2}
{"txhash":"0000000000000000000000000000000000000000000000000000000000000000"}

# wcloud sql wolk://alinatest/ncis "droptable awards"
{"txhash":"ab9ec7533d342f42619bf76882637f3a9c257b084842058aca680d0b2b3981b8"}

# wcloud sql wolk://alinatest/ncis "listtables"
{"data":[{"table":"personnel"}],"matchedrowcount":1}
{"txhash":"0000000000000000000000000000000000000000000000000000000000000000"}

# wcloud sql wolk://alinatest/ncistest createdatabase
{"txhash":"1797301b0527e5a4452bf8194c6ee5e72b6f632033314822956cd4a539b20d76"}

# wcloud sql wolk://alinatest/ncis "listdatabases"
{"data":[{"database":"ncis"},{"database":"ncistest"}],"matchedrowcount":2}
{"txhash":"0000000000000000000000000000000000000000000000000000000000000000"}

# wcloud sql wolk://alinatest/ncis "dropdatabase"
{"txhash":"3f754ec09bcc02ca80c363e49af969f60169fff58649ea62d21c2c8f2d9aea00"}

# wcloud sql wolk://alinatest/ncis "listdatabases"
{"data":[{"database":"ncistest"}],"matchedrowcount":1}
{"txhash":"0000000000000000000000000000000000000000000000000000000000000000"}```
```

## Blockchain Operations

### Get Latest Block:
```
# wcloud blocknumber
507
```

### Get Block:
```
# wcloud block 383
{"parentHash":"0xd519a8095d1ef23b52d72e2871795a3d85a94661c797293663aaee79020a9f57","blockNumber":383,"seed":"0x642383cea99d80118390c342e371eca0aced55b3dcbc0d57efd497e88f8ee2c5","accountRoot":"0x1b222b61e3e975c2f25aa3cdc9086665d7038617a15f18bb2a841313a142baf1","registryRoot":"0x0fca18a9a2c1b93446aa206ed8a4cd75d47975f8c2399642ea7e7b65f03e9fb0","checkRoot":"0x0000000000000000000000000000000000000000000000000000000000000000","keyRoot":"0x5bb4d8b25ed387aafa89776f7574ad235121711e844034e9609ec1d49f05d2f5","productsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000","nameRoot":"0x15fa6ec968b705b2d341d81be59f39d87fd329b0ef2a3e9b45aabb5631a41206","transactions":[{"transactionType":6,"recipient":"0x82a978b3f5962a5b0957d9ee9eef472ee55b42f1","hash":"0x011b4d03dd8c01f1049143cf9c4c817e4b167f1d1b83e5c6f0f10d89ba1e7bce","collection":"cGhvdG9z","key":"YmFuYW5hLmdpZg=="}],"deltaPos":1,"deltaNeg":2,"storageBeta":30,"bandwidthBeta":40,"phaseVotes":0,"q":3,"gamma":100,"sig":""}
```

### Get Transaction:
```
# wcloud tx 109a252be54ba2780950ba1dfbc6b6eea72123a6759ebd53dd0eab8bb8424070
{"transactionType":6,"recipient":"0x82a978b3f5962a5b0957d9ee9eef472ee55b42f1","hash":"0x011b4d03dd8c01f1049143cf9c4c817e4b167f1d1b83e5c6f0f10d89ba1e7bce","collection":"cGhvdG9z","key":"YmFuYW5hLmdpZg=="}
```

### Get Balance:
```
# wcloud balance 0x82A978B3f5962A5b0957d9ee9eEf472EE55B42F1
{"address":"82a978b3f5962a5b0957d9ee9eef472ee55b42f1", "balance":100000}
```
