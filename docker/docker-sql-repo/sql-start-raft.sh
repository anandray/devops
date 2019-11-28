#!/bin/bash

mkdir -p "$1"/sql
#ip=`ifconfig eth0 | grep inet | awk '{print$2}' | cut -d ":" -f2`
ip=172.17.0.2

echo "{
    \"blockChainId\": "$2"
}" > "$1"/genesis.json

/root/sql/bin/bootnode --genkey="$1"/sql/nodekey && /root/sql/bin/bootnode -writeaddress --nodekey="$1"/sql/nodekey | awk '{print"[","\n","\t","\"enode://"$1"@""'$ip'"":22000?discport=0&raftport=50400\"","\n]"}' > "$1"/static-nodes.json
#/root/sql/bin/bootnode --genkey="$1"/sql/nodekey && /root/sql/bin/bootnode -writeaddress --nodekey="$1"/sql/nodekey | awk '{print"[\"enode://"$1"@""'$ip'"":22000?discport=0&raftport=50400\"]"}' > "$1"/static-nodes.json

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
--raft \
2>> "$1"/sql.log
