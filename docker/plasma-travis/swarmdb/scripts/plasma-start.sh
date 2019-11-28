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
/usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb --datadir /usr/local/swarmdb/data --password /usr/local/swarmdb/data/.wolk account new
fi

sleep 5

# Making sure DB is clean
#if [ -d /usr/local/swarmdb/data/geth ] && ! ps aux | grep "swarmdb --datadir" | grep -v grep &> /dev/null; then
#echo "
#${bold}Making sure DB is clean
#"
#echo ${normal}
#
#/usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb removedb --datadir /usr/local/swarmdb/data << EOF
#y
#y
#EOF
#
#else
#echo "
#${bold}No need to run \"swarmdb removedb\".. swarmdb is already running..
#"
#fi

echo ${normal}

if [ -f /usr/sbin/ifconfig ]; then
eth=`ifconfig eth0 | grep inet | awk '{print$2}'`
fi
acct=`/usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb --datadir /usr/local/swarmdb/data account list 2> /dev/null | awk '{print$3}' | cut -d "{" -f2 | cut -d "}" -f1`

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

## create static-nodes.json
#if [ -d /usr/local/swarmdb/data ] && [ ! -f /usr/local/swarmdb/data/static-nodes.json ]; then
#echo "
#${bold}Creating static-nodes.json file
#"
#
#echo ${normal}
#
#echo '[
#	"enode://6629f2bfe15a97b057eea671ef6e2a8a463d2edd051c7bdf7fa123dc7935e75dd4be26300dfbbf5c8e686a153dd7e1cdded72a1b4501923516611323ba396dff@35.193.7.46:30303"
#]' > /usr/local/swarmdb/data/static-nodes.json
#
#else
#echo "
#${bold} static-nodes.json already exists...
#"
#fi

if [ -f /usr/local/go/src/github.com/wolkdb/plasma/build/bin/plasma ] && ! ps aux | grep "plasma --datadir" | grep -v grep &> /dev/null; then
echo "
${bold}#### S T A R T I N G   P L A S M A ####
"

echo ${normal}

#nohup /usr/local/go/src/github.com/wolkdb/plasma/build/bin/plasma \
#--datadir /usr/local/swarmdb/data \
#--networkid 7272 \
#--mine \
#--unlock 0x$acct \
#--etherbase 0x$acct \
#--verbosity 4 \
#--maxpeers 100 \
#--rpc \
#--rpcaddr $eth \
#--rpcport 8545 \
#--password /usr/local/swarmdb/data/.wolk \

nohup /usr/local/go/src/github.com/wolkdb/plasma/build/bin/plasma \
--datadir /usr/local/swarmdb/data \
--verbosity 4 \
2>> /usr/local/swarmdb/data/plasma.log &

else
echo "
${bold}PLASMA is already runing...
"
echo ${normal}
fi

sleep 5

if [ -f /usr/local/swarmdb/data/plasma.log ]; then
echo "
${bold}Check /usr/local/swarmdb/data/plasma.log
"

echo ${normal}
tail -n10 /usr/local/swarmdb/data/plasma.log
fi

sleep 2

if [ -d /usr/local/swarmdb/data/keystore ]; then
echo "
${bold}Listing SWARMDB account
"
/usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb --datadir /usr/local/swarmdb/data account list
fi

sleep 2

#echo "
#${bold}eth.accounts:
#"
#/usr/local/go/src/github.com/wolkdb/plasma/build/bin/plasma attach /usr/local/swarmdb/data/swarmdb.ipc --exec eth.accounts
#echo ${normal}
#
#sleep 2

#echo "
#${bold}nodeInfo:
#"
#/usr/local/go/src/github.com/wolkdb/plasma/build/bin/plasma attach /usr/local/swarmdb/data/plasma.ipc --exec admin.nodeInfo.enode
#fi

echo ${normal}
