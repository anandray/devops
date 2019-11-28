#!/bin/bash

bold=$(tput bold)
normal=$(tput sgr0)

source /root/.bashrc

if [ ! -d /usr/local/swarmdb/data ]; then
echo "
${bold}Creating DATADIR /usr/local/swarmdb/data
"
mkdir -p /usr/local/swarmdb/data
fi

echo wolk > /usr/local/swarmdb/data/.wolk

if [ -d /usr/local/swarmdb/data ] && [ ! -f /usr/local/swarmdb/data/keystore/UTC--* ]; then
echo "
${bold}Keystore account doesn't exist.. Adding new account
"
/var/www/vhosts/go-wolk/build/bin/geth --datadir /usr/local/swarmdb/data --password /usr/local/swarmdb/data/.wolk account new
fi

sleep 5

# Making sure DB is clean
if [ -d /usr/local/swarmdb/data/geth ] && ! ps aux | grep "geth --datadir" | grep -v grep &> /dev/null; then
echo "
${bold}Making sure DB is clean
"
echo ${normal}

/var/www/vhosts/go-wolk/build/bin/geth removedb --datadir /usr/local/swarmdb/data << EOF
y
y
EOF

else
echo "
${bold}No need to run \"geth removedb\".. geth is already running..
"
fi

echo ${normal}

sleep 5

if [ -d /usr/local/swarmdb/data ] && [ ! -f /usr/local/swarmdb/data/genesis.json ]; then
echo "
${bold}Creating GENESIS file
"

echo ${normal}

echo '{
  "config": {
    "chainId": 7272,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "clique": {
      "period": 15,
      "epoch": 30000
    }
  },
 "extraData": "0x00000000000000000000000000000000000000000000000000000000000000001d21f64b4048d91b8216209fb682c797d63f5dd10000000
000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "nonce": "0x0",
  "timestamp": "0x5a73d033",
  "gasLimit": "0x47b760",
  "difficulty": "0x1",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "alloc": {
        "08bc838cc67c4f35fd2ec7f5df26380643c11d87": { "balance": "6000000000000000000" },
        "69e03b7aa3eac087e80e8d42a6983720ce861c1b": { "balance": "5000000000000000000" },
        "1d21f64b4048d91b8216209fb682c797d63f5dd1": { "balance": "300000000000000000000" }
  },
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}' > /usr/local/swarmdb/data/genesis.json

else
echo "
${bold}GENESIS file already exists...
"
fi

echo ${normal}

sleep 5

if [ -f /usr/local/swarmdb/data/genesis.json ] && [ ! -d /usr/local/swarmdb/data/geth/chaindata ] && [ ! -d /usr/local/swarmdb/d
ata/geth/lightchaindata ]; then
echo "
${bold}Initializing genesis.json...
"

echo ${normal}

/var/www/vhosts/go-wolk/build/bin/geth --datadir /usr/local/swarmdb/data init /usr/local/swarmdb/data/genesis.json

else
echo "
${bold}genesis.json was already initiated earlier...
"
echo ${normal}
fi

sleep 5

eth=`ifconfig eth0 | grep inet | awk '{print$2}'`
acct=`/var/www/vhosts/go-wolk/build/bin/geth --datadir /usr/local/swarmdb/data account list 2> /dev/null | awk '{print$3}' | cut
 -d "{" -f2 | cut -d "}" -f1`

# create swarmdb.conf
if [ -d /usr/local/swarmdb/etc ] && [ ! -f /usr/local/swarmdb/etc/swarmdb.conf ]; then
echo "
${bold}Creating SWARMDB config file
"

echo '{
        "listenAddrTCP": "0.0.0.0",
        "portTCP": 2001,
        "listenAddrHTTP": "0.0.0.0",
        "portHTTP": 8501,
        "address": "",
        "privateKey": "bc677a6518c85a61606547e670751829ffe36c3e8827f0c1780ff96f031fcb50",
        "decryptionKey": "3d51e84f0270019e9238f6946bd35a8f",
        "chunkDBPath": "/usr/local/swarmdb/data",
        "usersKeysPath": "/usr/local/swarmdb/data/keystore",
        "ensKeyPath": "/usr/local/swarmdb/data/keystore",
        "authentication": 1,
        "users": [
                {
                        "address": "",
                        "passphrase": "wolk",
                        "minReplication": 3,
                        "maxReplication": 5,
                        "autoRenew": 1
                }
        ],
        "currency": "WLK",
        "targetCostStorage": 2.71828,
        "targetCostBandwidth": 3.14159,
        "isLeader": false
}' > /usr/local/swarmdb/etc/swarmdb.conf

#DELETE "address" from /usr/local/swarmdb/apptxn/index.js
sed -i '/address/d' /usr/local/swarmdb/etc/swarmdb.conf

#Add "address" to the 5th and 15th lines of swarmdb.conf:
sed -i "5 i\        \"address\": \"0x$acct\"," /usr/local/swarmdb/etc/swarmdb.conf
sed -i "15 i\                        \"address\": \"0x$acct\"," /usr/local/swarmdb/etc/swarmdb.conf
fi

echo ${normal}

sleep 5

# create static-nodes.json
if [ -d /usr/local/swarmdb/data ] && [ ! -f /usr/local/swarmdb/data/static-nodes.json ]; then
echo "
${bold}Creating static-nodes.json file
"

echo ${normal}

echo '[
	"enode://6629f2bfe15a97b057eea671ef6e2a8a463d2edd051c7bdf7fa123dc7935e75dd4be26300dfbbf5c8e686a153dd7e1cdded72a1b4501923516611323ba396dff@35.193.7.46:30303"
]' > /usr/local/swarmdb/data/static-nodes.json

else
echo "
${bold} static-nodes.json already exists...
"
fi

if [ -f /var/www/vhosts/go-wolk/build/bin/geth ] && ! ps aux | grep "geth --datadir" | grep -v grep &> /dev/null; then
echo "
${bold}#### S T A R T I N G   G E T H ####
"

echo ${normal}

nohup /var/www/vhosts/go-wolk/build/bin/geth \
--datadir /usr/local/swarmdb/data \
--networkid 7272 \
--mine \
--unlock 0x$acct \
--etherbase 0x$acct \
--verbosity 4 \
--maxpeers 100 \
--rpc \
--rpcaddr $eth \
--rpcport 8545 \
--password /usr/local/swarmdb/data/.wolk \
2>> /usr/local/swarmdb/data/geth.log &

else
echo "
${bold}GETH is already runing...
"
echo ${normal}
fi

sleep 5

if [ -f /usr/local/swarmdb/data/geth.log ]; then
echo "
${bold}Check /usr/local/swarmdb/data/geth.log
"

echo ${normal}
tail -n10 /usr/local/swarmdb/data/geth.log
fi

sleep 2

if [ -d /usr/local/swarmdb/data/keystore ]; then
echo "
${bold}Listing GETH account
"
/var/www/vhosts/go-wolk/build/bin/geth --datadir /usr/local/swarmdb/data account list

sleep 2

echo "
${bold}eth.accounts:
"
/var/www/vhosts/go-wolk/build/bin/geth attach /usr/local/swarmdb/data/geth.ipc --exec eth.accounts
echo ${normal}

sleep 2

echo "
${bold}nodeInfo:
"
/var/www/vhosts/go-wolk/build/bin/geth attach /usr/local/swarmdb/data/geth.ipc --exec admin.nodeInfo.enode
fi

echo ${normal}

#DELETE "const from" from /usr/local/swarmdb/apptxn/index.js
sed -i '/const from/d' /usr/local/swarmdb/apptxn/index.js

#Add "const from" to the 3rd line of index.js:
sed -i "3 i\const from = \"0x$acct\";" /usr/local/swarmdb/apptxn/index.js
