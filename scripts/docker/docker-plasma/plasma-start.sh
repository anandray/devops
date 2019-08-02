#!/bin/bash

# activate sqlite3-3.22
echo 'export CGO_CFLAGS="-I/usr/local/sqlite3"' >> /root/.bashrc
source /root/.bashrc
ldconfig
#exec -l $SHELL

echo "
35.224.4.165    plasma.wolk.com cloudstore.wolk.com
35.193.168.171  log0 log0.wolk.com" >> /etc/hosts

if ! ps aux | grep syslog-ng | grep -v grep; then
/usr/sbin/syslog-ng -F &
fi

mkdir -p "$1"/plasma

/root/plasma/bin/plasma \
--datadir "$1" \
--nodiscover \
--rpc \
--rpcaddr 0.0.0.0 \
--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,plasma \
--rpcport 32003 \
--port 31003 \
--rpccorsdomain=* \
--rpcvhosts=* -unlock 0 \
--verbosity 6 \
2>> "$1"/plasma.log
