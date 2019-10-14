#!/bin/bash
/usr/local/bin/wolk --datadir /usr/local --rpc --rpcaddr 0.0.0.0 --rpccorsdomain=* --rpcvhosts=* --config /usr/local/bin/wolk.toml >> /usr/local/bin/wolk.log 2>&1 &

sleep 5d
