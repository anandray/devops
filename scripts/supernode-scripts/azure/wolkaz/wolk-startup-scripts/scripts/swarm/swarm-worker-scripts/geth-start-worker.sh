#!/bin/bash

# fetch keystore and swarm/bzz-*/config.json from "demo-swarm-wolk-com-80kh"
if [ ! -d /var/www/vhosts/data/keystore/UTC--2017-12-19T00-43-03.321151789Z--2754c8db88cb5019c634f84618d0eaabdd99b1c6 ]; then
echo "`date +%T` - copying keystore account..."
sudo su - << EOF
mkdir -p /var/www/vhosts/data/keystore && scp demo-swarm-wolk-com-80kh:/var/www/vhosts/data/keystore/UTC--2017-12-19T00-43-03.321151789Z--2754c8db88cb5019c634f84618d0eaabdd99b1c6 /var/www/vhosts/data/keystore/

else
echo "`date +%T` - keystore account exists..."
EOF
fi

if [ ! -f /var/www/vhosts/data/swarm/bzz-2754c8db88cb5019c634f84618d0eaabdd99b1c6/config.json ]; then
echo "`date +%T` - copying bzz config.json..."
sudo su - << EOF
mkdir -p /var/www/vhosts/data/swarm/bzz-2754c8db88cb5019c634f84618d0eaabdd99b1c6 && scp demo-swarm-wolk-com-80kh:/var/www/vhosts/data/swarm/bzz-2754c8db88cb5019c634f84618d0eaabdd99b1c6/config.json /var/www/vhosts/data/swarm/bzz-2754c8db88cb5019c634f84618d0eaabdd99b1c6/

else
echo "`date +%T` - swarm bzz account config.json exists..."
EOF
fi

if [ ! -f /var/www/vhosts/data/genesis.json ]; then
echo "`date +%T` - copying /var/www/vhosts/data/genesis.json"
sudo su - << EOF
scp demo-swarm-wolk-com-80kh:/var/www/vhosts/data/genesis.json /var/www/vhosts/data/

else
echo "`date +%T` - /var/www/vhosts/data/genesis.json already exists..."
EOF
fi

# download geth and swarm binaries
if [ ! -f /usr/local/bin/geth ]; then
sudo su - << EOF
wget -O /usr/local/bin/geth https://github.com/wolktoken/swarm.wolk.com/raw/master/src/github.com/ethereum/go-ethereum/build/bin/geth
chmod +x /usr/local/bin/geth
EOF
fi

if [ ! -f /usr/local/bin/swarm ]; then
sudo su - << EOF
scp demo-swarm-wolk-com-80kh:/tmp/swarm /usr/local/bin/
chmod +x /usr/local/bin/swarm
EOF
fi

DATADIR=/var/www/vhosts/data
enode=enode://49f6146dde7af4776f74e5e053e68b00a89eba1fe3a74388e27f645b50c5f1f29c7b7550f9d7a2ac7c4a9a43eb1a6a8dd2a2ab5b40845317b05db8a2f51f3f54
ip=10.128.0.26
hostname=`hostname`

if [ ! -d /var/www/vhosts/data/geth ]; then
echo "running geth init..."
sudo su - << EOF
geth init --datadir $DATADIR genesis.json
EOF
fi

for i in {1..5}
do
if ! ps aux | grep 'geth --bootnodes' | grep -v grep &> /dev/null; then
echo "`date +%T` - geth is not running... starting geth using $enode and $ip..."
sudo su - << EOF
nohup geth --bootnodes $enode@$ip:30301 \
       --identity $hostname \
       --datadir /var/www/vhosts/data \
       --mine \
       --unlock 0 \
       --password <(echo -n "wolk") \
       --verbosity 6 \
       --networkid 55300 \
        2>> /var/www/vhosts/data/geth.log &
EOF
else
echo "`date +%T` - geth is already running..."
fi
sleep 10
done
