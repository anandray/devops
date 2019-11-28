#!/bin/bash

if [ ! -d /root/sql/qdata/dd ]; then
mkdir -p /root/sql/qdata/dd
fi

if [ ! -d /root/sql/bin ]; then
mkdir -p /root/sql/bin
fi

if [ ! -f /root/sql/bin/sql ]; then
echo "
Downloading the sql binary...
"
wget -O /root/sql/bin/sql www6001.wolk.com/.start/sql
chmod +x /root/sql/bin/sql
fi

MD5=`ssh -q nginx.wolk.com md5sum /root/sql/bin/sql | awk '{print$1}'`

if ! md5sum /root/sql/bin/sql | grep $MD5 &> /dev/null; then
kill -9 `ps aux | grep "sql --datadir" | grep -v grep | awk '{print$2}'`
wget -O /root/sql/bin/sql www6001.wolk.com/.start/sql
chmod +x /root/sql/bin/sql;
#sed -i 's/#nohup/nohup/g' /root/scripts/sql-start.sh;
sleep 10
fi

#if [ -d /root/sql/qdata/dd/quorum-raft-state ]; then
#rm -rf /root/sql/qdata/dd/quorum-raft-state
#fi

#if [ -d /root/sql/qdata/dd/raft-snap ]; then
#rm -rf /root/sql/qdata/dd/raft-snap
#fi

#if [ -d /root/sql/qdata/dd/raft-wal ]; then
#rm -rf /root/sql/qdata/dd/raft-wal
#fi

#if [ -f /root/sql/qdata/dd/sql.ipc ]; then
#rm -rf /root/sql/qdata/dd/sql.ipc
#fi

nohup /root/sql/bin/sql \
--datadir /root/sql/qdata/dd \
--nodiscover \
--rpc \
--rpcaddr 0.0.0.0 \
--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,sql,raft \
--emitcheckpoints \
--raftport 50404 \
--rpcport 22003 \
--port 21003 \
--rpccorsdomain=* \
--rpcvhosts=* \
--plasmaaddr 'cloudstore.wolk.com' \
--plasmaport 80 \
--unlock 0 \
--verbosity 4 \
--raft \
--blockchainid 77 \
2>> /root/sql/qdata/sql.log &
