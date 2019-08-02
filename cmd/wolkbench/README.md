# wolkbench

`wolkbench` submits a consistent stream of transactions to a blockchain from a small number of users (specified by `users`) in these ways:
 * key - parallel SetKey/GetKey -- short key and short value of length 30-32 bytes
 * file - parallel PutFile/GetFile -- short key but files are  length `filesize`
After setting up a random user and bucket,  `users * files` concurrent processes are run where the wolkbench process does not terminate.

A `check` parameter controls the flow of the test:
 * `-1` - no waiting for transaction to be includes
 * `0` [default] - no GetKey/GetFile, but waits for transaction to be included
 * `1` - does a GetKey/GetFile after transaction is included in block

### wolkbench key

```
$ wolkbench -server=c0.wolk.com -httpport=81 -users=5 -files=5 -run=key
```

This supports testing of:
* 4 tx/block:  `nohup wolkbench -server=c0.wolk.com -httpport=443 -users=2 -files=2 -run=key &> /var/log/wolkbench.log`
* 6 tx/block:  `nohup wolkbench -server=c0.wolk.com -httpport=81  -users=2 -files=3 -run=key &> /var/log/wolkbench1.log`
* 9 tx/block:  `nohup wolkbench -server=c0.wolk.com -httpport=82  -users=3 -files=3 -run=key &> /var/log/wolkbench2.log`
* 16 tx/block: `nohup wolkbench -server=c0.wolk.com -httpport=83  -users=4 -files=4 -run=key &> /var/log/wolkbench3.log`
* 20 tx/block: `nohup wolkbench -server=c1.wolk.com -httpport=85  -users=4 -files=5 -run=key &> /var/log/wolkbench5.log`


### wolkbench file

Same as the above run, but instead of short values, has files of size `filesize`

```
$ wolkbench -server=c0.wolk.com -httpport=81 -users=2 -files=2 -filesize=10000 -run=file
```

### GOALS:

Our Q2 goal is to have consistency of:
* 100 tx/block: `nohup wolkbench -server=c0.wolk.com -httpport=81  -users=1 -files=100 -filesize=100000 -run=file &> /var/log/wolkbench1.log`
* 100 tx/block: `nohup wolkbench -server=c0.wolk.com -httpport=82  -users=10 -files=10 -filesize=100000 -run=file &> /var/log/wolkbench2.log`
* 100 tx/block: `nohup wolkbench -server=c0.wolk.com -httpport=83  -users=100 -files=1 -filesize=100000 -run=file &> /var/log/wolkbench3.log`
