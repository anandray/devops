#!/bin/bash

# generating boot.key using bootnode
if [ ! -d /var/www/vhosts/data ]; then
sudo su - << EOF
mkdir -p /var/www/vhosts/data
EOF
fi

gethaddr=`ls /var/www/vhosts/data/keystore/ | cut -d "-" -f9`
if ! grep $gethaddr /var/log/geth-account-new.log &> /dev/null; then
sudo rm -rf /var/www/vhosts/data/keystore/UTC*
sudo geth --datadir /var/www/vhosts/data account new &> /var/log/geth-account-new.log << EOF
14all41
14all41
EOF
else
echo "geth account already exists..."
fi

for i in {1..11}
do
if ! ps aux | grep 'geth --testnet' | grep -v grep &> /dev/null; then
echo "geth is not running... starting geth..."
sudo su - << EOF
nohup geth --testnet \
           --networkid 3 \
           --fast \
           --datadir /var/www/vhosts/data \
           --bootnodes "enode://20c9ad97c081d63397d7b685a412227a40e23c8bdc6688c6f37e97cfbc22d2b4d1db1510d8f61e6a8866ad7f0e17c02b14182d37ea7c3c8b9c2683aeb6b733a1@52.169.14.227:30303,enode://6ce05930c72abc632c58e2e4324f7c7ea478cec0ed4fa2528982cf34483094e9cbc9216e7aa349691242576d552a2a56aaeae426c5303ded677ce455ba1acd9d@13.84.180.240:30303" \
       --unlock 0 \
       --cache=1024 \
       --password <(echo -n "14all41") \
       --verbosity 6 \
        2>> /var/www/vhosts/data/geth.log &
EOF
else
echo "geth is already running..."
fi
sleep 5;
done
