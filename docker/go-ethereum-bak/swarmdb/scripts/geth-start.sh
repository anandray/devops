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
/var/www/vhosts/go-ethereum/build/bin/geth --datadir /usr/local/swarmdb/data --password /usr/local/swarmdb/data/.wolk account new
fi

sleep 5

# Making sure DB is clean
if [ -d /usr/local/swarmdb/data/geth ] && ! ps aux | grep "geth --datadir" | grep -v grep &> /dev/null; then
echo "
${bold}Making sure DB is clean
"
echo ${normal}

/var/www/vhosts/go-ethereum/build/bin/geth removedb --datadir /usr/local/swarmdb/data << EOF
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
 "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000$acct0000000
000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "nonce": "0x0",
  "timestamp": "0x5a73d033",
  "gasLimit": "0x47b760",
  "difficulty": "0x1",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "alloc": {
        "$acct": { "balance": "300000000000000000000" }
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

/var/www/vhosts/go-ethereum/build/bin/geth --datadir /usr/local/swarmdb/data init /usr/local/swarmdb/data/genesis.json

else
echo "
${bold}genesis.json was already initiated earlier...
"
echo ${normal}
fi

sleep 5

eth=`ifconfig eth0 | grep inet | awk '{print$2}'`
acct=`/var/www/vhosts/go-ethereum/build/bin/geth --datadir /usr/local/swarmdb/data account list 2> /dev/null | awk '{print$3}' | cut
 -d "{" -f2 | cut -d "}" -f1`

echo ${normal}

sleep 5

if [ -f /var/www/vhosts/go-ethereum/build/bin/geth ] && ! ps aux | grep "geth --datadir" | grep -v grep &> /dev/null; then
echo "
${bold}#### S T A R T I N G   G E T H ####
"

echo ${normal}

nohup /var/www/vhosts/go-ethereum/build/bin/geth \
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
/var/www/vhosts/go-ethereum/build/bin/geth --datadir /usr/local/swarmdb/data account list

sleep 2

echo "
${bold}eth.accounts:
"
/var/www/vhosts/go-ethereum/build/bin/geth attach /usr/local/swarmdb/data/geth.ipc --exec eth.accounts
echo ${normal}

sleep 2

echo "
${bold}nodeInfo:
"
/var/www/vhosts/go-ethereum/build/bin/geth attach /usr/local/swarmdb/data/geth.ipc --exec admin.nodeInfo.enode
fi

echo ${normal}
