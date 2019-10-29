#!/bin/bash

mkdir -p "$1"

/root/cloudstore/bin/cloudstore \
--datadir "$1" \
--rpc \
--rpcaddr 0.0.0.0 \
--rpcport "$2" \
--rpccorsdomain=* \
--rpcvhosts=* \
--rpcapi personal,db,eth,net,web3,swarmdb,plasma,admin,cloudstore \
--verbosity 4
#2>> /root/cloudstore/cloudstore.log &
