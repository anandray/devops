#!/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/usr/lib64/google-cloud-sdk/bin:/root/bin:/root/sql/bin

datadir=/root/sql/qdata/dd
blockChainId=0x3b6a2ac8b193b705
#datadir=$1
#blockChainId=$2

mkdir -p $datadir

echo "{
    \"blockChainId\": \"$blockChainId\"
}" > $datadir/genesis.json

/root/sql/bin/sql \
--datadir $datadir \
--nodiscover \
--rpc \
--rpcaddr 0.0.0.0 \
--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,sql \
--rpcport 24000 \
--port 21000 \
--rpccorsdomain=* \
--rpcvhosts=* \
--plasmaaddr "sanmateo.wolk.com" \
--plasmaport 32003 \
--unlock 0 \
--verbosity 6 \
2>> $datadir/sql.log &
