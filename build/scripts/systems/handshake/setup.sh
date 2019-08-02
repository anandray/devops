#!/bin/sh
wget https://rpm.nodesource.com/pub_8.x/el/6/x86_64/nodejs-8.9.4-1nodesource.x86_64.rpm
rpm -Uvh nodejs-8.9.4-1nodesource.x86_64.rpm
npm install -g npm
git clone git://github.com/handshake-org/hs-miner.git
npm install
make testnet
/root/src/hs-miner/bin/hs-miner --rpc-host localhost --rpc-port 13037 --rpc-pass 14all41 &> /var/log/hs-miner.log &
git clone git://github.com/handshake-org/hsd.git
npm install --production
/root/src/hsd/bin/hsd --daemon --rs-host 0.0.0.0 --rs-port 53 --listen --max-inbound=20 &> /var/log/hsd-recursive.log
