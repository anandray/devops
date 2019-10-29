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
if ! ps aux | grep 'geth --networkid' | grep -v grep &> /dev/null; then
echo "`date +%T` - geth is not running... starting geth..."
sudo su - << EOF
nohup geth --networkid 1 \
           --fast \
           --datadir /var/www/vhosts/data \
           --unlock 0 \
           --cache=1024 \
           --password <(echo -n "14all41") \
           --verbosity 6 \
            2>> /var/www/vhosts/data/geth.log &
EOF
else
echo "`date +%T` - geth is already running..."
fi
sleep 5;
done
