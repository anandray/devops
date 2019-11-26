#!/bin/bash

# generating boot.key using bootnode
if [ ! -d /var/www/vhosts/data ]; then
sudo su - << EOF
mkdir -p /var/www/vhosts/data
EOF
fi

if [ ! -f /var/www/vhosts/data/boot.key ]; then
sudo su - << EOF
cd /var/www/vhosts/data
bootnode --genkey=boot.key
#gsutil cp gs://startup_scripts_us/scripts/geth/boot.key /var/www/vhosts/data/
EOF
fi

# run bootnode in the background
if ! ps aux | grep 'bootnode --nodekey=boot.key' | grep -v grep &> /dev/null; then
sudo su - << EOF
cd /var/www/vhosts/data
nohup bootnode --nodekey=boot.key 2>> /var/www/vhosts/data/bootnode.log &
EOF
fi

gethaddr1=`ls /var/www/vhosts/data/keystore/ | cut -d "-" -f9`
if ! grep $gethaddr1 /var/log/geth-account-new.log &> /dev/null; then
sudo rm -rf /var/www/vhosts/data/keystore/UTC*
sudo geth --datadir /var/www/vhosts/data account new &> /var/log/geth-account-new.log << EOF
mdotm
mdotm
EOF
else
echo "geth account already exists..."
fi

gethaddr2=`cat /var/log/geth-account-new.log | grep Address | cut -d "{" -f2 | cut -d "}" -f1`

if ! grep $gethaddr2 /var/www/vhosts/data/genesis.json &> /dev/null; then
sh /root/scripts/generate-genesis.json.sh
else
echo "/var/www/vhosts/data/genesis.json already exists with the geth account number"
fi

if [ ! -d /var/www/vhosts/data/geth/chaindata ]; then
sudo su - << EOF
cd /var/www/vhosts/data
geth --datadir /var/www/vhosts/data init /var/www/vhosts/data/genesis.json
EOF
fi

enode=`grep "self=enode:" /var/www/vhosts/data/bootnode.log | head -n1 | awk '{print$6}' | cut -d "=" -f2 | cut -d "@" -f1`
ip=`ifconfig eth0 | grep inet | awk '{print$2}' | cut -d ":" -f2 | head -n1`
DATADIR=/var/www/vhosts/data

if [ ! -f /var/www/vhosts/data/.mdotm ]; then
sudo gsutil cp gs://startup_scripts_us/scripts/swarm/my-password /var/www/vhosts/data/.mdotm;
else
echo "password file already exists..."
fi

#nohup geth --bootnodes enode://067f1fdc793a5a5d3f1b98b2efa88622f8c55290efec6b76b59398063027b505c64d07e5d69bd962e4b993127f97bd7fdcd6bee5694ca16a05200c1d09859d05@10.128.0.21:30301
#RPCADDR=`ifconfig eth0 | grep 'inet addr:' | awk '{print$2}' | cut -d ":" -f2`

if ! ps aux | grep 'geth --bootnodes' | grep -v grep &> /dev/null; then
echo "`date +%T` - geth is not running... starting geth using $enode and $ip..."
sudo su - << EOF
nohup geth --bootnodes $enode@$ip:30301 \
       --identity  "WolkMain" \
       --datadir /var/www/vhosts/data \
       --mine \
       --unlock 0 \
       --password $DATADIR/.mdotm \
       --verbosity 6 \
       --networkid 55999 \
        2>> /var/www/vhosts/data/geth.log &
EOF
else
echo "`date +%T` - geth is already running..."
fi

# add swarm -start script to crontab
echo "*/1 * * * * sh /root/scripts/swarm-start-firsttime.sh &>> /var/log/swarm-start-firsttime.log" >> /var/spool/cron/root

# comment out the above line after adding to crontab and comment out the lines with sed
sed -i 's/echo \"\*\/1 \* \* \* \* sh/#echo \"\*\/1 \* \* \* \* sh/g' /root/scripts/geth-start.sh
sed -i 's/sed -i/#sed -i/g' /root/scripts/geth-start.sh
