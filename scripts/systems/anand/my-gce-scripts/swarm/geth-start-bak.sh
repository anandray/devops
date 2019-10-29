#!/bin/bash

sudo geth --datadir /var/www/vhosts/data account new &>> /var/log/geth-account-new.log << EOF
mdotm
mdotm
EOF

sudo su - << EOF
mkdir -p /var/www/vhosts/data
cd /var/www/vhosts/data
gethaddr=`cat /var/log/geth-account-new.log | grep Address | cut -d "{" -f2 | cut -d "}" -f1`
echo '
{
  "config": {
        "chainId": 15,
        "homesteadBlock": 0,
        "eip155Block": 0,
        "eip158Block": 0
  },
  "nonce": "0x0000000000000042",
  "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "difficulty": "0x4000",
  "alloc": {
    "0x'$gethaddr'": {
    "balance": "1000000000000000000000"
    }
  },
  "coinbase": "0x0000000000000000000000000000000000000000",
  "timestamp": "0x00",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "extraData": "0x00",
  "gasLimit": "0xffffffff"
}' > /var/www/vhosts/data/genesis1.json

geth --datadir /var/www/vhosts/data init /var/www/vhosts/data/genesis.json
EOF

enode=`grep "self=enode:" /var/www/vhosts/data/bootnode.log | head -n1 | awk '{print$6}' | cut -d "=" -f2 | cut -d "@" -f1`
ip=`ifconfig eth0 | grep 'inet addr' | awk '{print$2}' | cut -d ":" -f2`
DATADIR=/var/www/vhosts/data
sudo gsutil cp gs://startup_scripts_us/scripts/swarm/my-password /var/www/vhosts/data/.mdotm;

#nohup geth --bootnodes enode://067f1fdc793a5a5d3f1b98b2efa88622f8c55290efec6b76b59398063027b505c64d07e5d69bd962e4b993127f97bd7fdcd6bee5694ca16a05200c1d09859d05@10.128.0.21:30301

if ! ps aux | grep 'geth --bootnodes' | grep -v grep &> /dev/null; then
echo "geth is not running... starting geth using $enode and $ip..."
sudo su - << EOF
nohup geth --bootnodes $enode@$ip:30301 \
       --identity  "WolkMainNode" \
       --datadir /var/www/vhosts/data \
       --mine \
       --unlock 0 \
       --password $DATADIR/.mdotm \
       --verbosity 6 \
       --networkid 55300 \
        2>> /var/www/vhosts/data/geth.log &
EOF
else
echo "geth is already running..."
fi
