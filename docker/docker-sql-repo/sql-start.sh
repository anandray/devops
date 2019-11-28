#!/bin/bash

# activate sqlite3-3.22
echo 'export CGO_CFLAGS="-I/usr/local/sqlite3"' >> /root/.bashrc
source /root/.bashrc
ldconfig -v
exec -l $SHELL

# repo + golang
mkdir -p /root/go/src/github.com/wolkdb
cd /root/go/src/github.com/wolkdb
git clone git@github.com:wolkdb/plasma.git

# GOLANG v 1.10.2
cd /usr/local && \
wget https://dl.google.com/go/go1.10.2.linux-amd64.tar.gz && \
tar zxvpf go1.10.2.linux-amd64.tar.gz && \
ln -s /usr/local/go/bin/go /usr/bin/go && \
exec -l $SHELL && \
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc && \
source /root/.bashrc


echo "
35.224.4.165    plasma.wolk.com cloudstore.wolk.com
35.193.168.171  log0 log0.wolk.com" >> /etc/hosts

if ! ps aux | grep syslog-ng | grep -v grep; then
/usr/sbin/syslog-ng -F &
fi

mkdir -p "$1"/sql
ip=172.17.0.2

echo "{
   \"blockChainId\": \"$2\"
}" > "$1"/genesis.json

/root/sql/bin/sql \
--datadir "$1" \
--nodiscover \
--rpc \
--rpcaddr 0.0.0.0 \
--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,sql \
--rpcport 24000 \
--port 21000 \
--rpccorsdomain=* \
--rpcvhosts=* \
--plasmaaddr 'plasma.wolk.com' \
--plasmaport 80 \
--unlock 0 \
--verbosity 6 \
2>> "$1"/sql.log
