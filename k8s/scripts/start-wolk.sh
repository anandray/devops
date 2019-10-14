#!/bin/bash
'yes' | cp -rf /usr/local/bin/hosts /etc/hosts

/usr/local/bin/wolk --datadir /usr/local --rpc --rpcaddr 0.0.0.0 --rpccorsdomain=* --rpcvhosts=* --rpcport 9902 --port 32300 --httpport 82 --config /usr/local/bin/wolk.toml > /usr/local/bin/wolk.log &
