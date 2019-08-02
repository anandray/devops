#!/bin/bash

mkdir -p "$1"

echo "{
    \"blockChainId\": "$2"
}" > "$1"/genesis.json

/root/sql/bin/sql \
--datadir "$1" \
--nodiscover \
--rpc \
--rpcaddr 0.0.0.0 \
--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,sql,raft \
--emitcheckpoints \
--raftport 50400 \
--rpcport 22000 \
--port 21000 \
--rpccorsdomain=* \
--rpcvhosts=* \
--plasmaaddr 'cloudstore.wolk.com' \
--plasmaport 80 \
--unlock 0 \
--verbosity 4 \
--raft
#2>> /sql/qdata/sql.log &
