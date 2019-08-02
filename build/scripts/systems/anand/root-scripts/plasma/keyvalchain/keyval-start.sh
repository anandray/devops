#!/bin/bash

if [ ! -d /root/keyvalchain/qdata/dd ]; then
mkdir -p /root/keyvalchain/qdata/dd
fi

nohup /root/keyvalchain/bin/keyvalchain \
--datadir /root/keyvalchain/qdata/dd \
--nodiscover \
--rpc \
--rpcaddr 0.0.0.0 \
--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,swarmdb \
--emitcheckpoints \
--raftport 50403 \
--rpcport 22002 \
--port 21002 \
--unlock 0 \
--verbosity 6 \
2>> /root/keyvalchain/qdata/keyvalchain.log &
