#!/bin/bash


mkdir -p "$1"/sql

echo "{
    \"blockChainId\": \"$2\"
}" > $1/genesis.json

#sed -i 's/ID/\"$2\"/g' "$1"/genesis.json

/root/sql/bin/sql \
--datadir $1 \
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
2>> $1/sql.log &