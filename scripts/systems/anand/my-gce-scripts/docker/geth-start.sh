#!/bin/bash

if [ ! -d /var/www/vhosts/data ]; then
mkdir -p /var/www/vhosts/data
fi

cd /var/www/vhosts/data
# generating boot.key using bootnode
if [ ! -f /var/www/vhosts/data/boot.key ]; then
bootnode --genkey=boot.key
fi

# run bootnode in the background
if ! ps aux | grep 'bootnode --nodekey=boot.key' | grep -v grep &> /dev/null; then
nohup bootnode --nodekey=boot.key 2>> /var/www/vhosts/data/bootnode.log &
fi

gethaccount=`grep -s Address /var/log/geth-account-new.log  | cut -d "{" -f2 | cut -d "}" -f1`
gethaccountcharlen=`echo -n $gethaccount | wc -m`
if [ $gethaccountcharlen -ne 40 ]; then
geth --datadir /var/www/vhosts/data account new &> /var/log/geth-account-new.log << EOF
mdotm
mdotm
EOF
fi

mainaccount=`cat /var/log/geth-account-new.log | grep Address | cut -d "{" -f2 | cut -d "}" -f1`

if [ ! -f /var/www/vhosts/data/genesis.json ]; then
sh /wolk/scripts/genesis.json.sh
geth --datadir /var/www/vhosts/data init /var/www/vhosts/data/genesis.json &> /dev/null
fi

enode=`grep -s "self=enode:" /var/www/vhosts/data/bootnode.log | head -n1 | awk '{print$6}' | cut -d "=" -f2 | cut -d "@" -f1`
ip=`ifconfig eth0 | grep inet | awk '{print$2}' | cut -d ":" -f2 | head -n1`

#if [ ! -f /var/www/vhosts/data/.mdotm ]; then
#sudo gsutil cp gs://startup_scripts_us/scripts/swarm/my-password /var/www/vhosts/data/.mdotm;
#else
#echo "password file already exists..."
#fi

if ! ps aux | grep 'geth --bootnodes' | grep -v grep &> /dev/null; then
echo "`date +%T` - geth is not running... starting geth using $enode and $ip..."
nohup geth --bootnodes $enode@$ip:30301 \
       --identity  "WolkMainNode" \
       --datadir /var/www/vhosts/data \
       --mine \
       --unlock 0 \
       --fast \
       --cache=1024 \
       --password <(echo -n "mdotm") \
       --verbosity 6 \
       --networkid 55333 \
        2>> /var/www/vhosts/data/geth.log &
fi

# DATADIR=/var/www/vhosts/data
#       --password $DATADIR/.mdotm \

sleep 5

if ! ps aux | grep "swarm --bzzaccount" | grep -v grep &> /dev/null; then
echo "`date +%T` - swarm is not running... starting swarm using $mainaccount.."
nohup swarm \
       --bzzaccount $mainaccount \
       --swap \
       --swap-api /var/www/vhosts/data/geth.ipc \
       --datadir /var/www/vhosts/data \
       --verbosity 6 \
       --ens-api /var/www/vhosts/data/geth.ipc \
       --bzznetworkid 55333 \
       2>> /var/www/vhosts/data/swarm.log < <(echo -n "mdotm") &
fi

sleep 5

if grep '127.0.0.1' /var/www/vhosts/data/swarm/bzz-*/config.json &> /dev/null; then
echo "`date +%T` - Change ListenAddr from 127.0.0.1 to 0.0.0.0"
pkill -9 swarm
sed -i 's/127.0.0.1/0.0.0.0/g' /var/www/vhosts/data/swarm/bzz-*/config.json
nohup swarm \
       --bzzaccount $mainaccount \
       --swap \
       --swap-api /var/www/vhosts/data/geth.ipc \
       --datadir /var/www/vhosts/data \
       --verbosity 6 \
       --ens-api /var/www/vhosts/data/geth.ipc \
       --bzznetworkid 55333 \
       2>> /var/www/vhosts/data/swarm.log < <(echo -n "mdotm") &
fi

if [ $? -ne 0 ]; then
echo "An error occurred..."
exit 3
fi

# Terminate shell script with success message
exit 0
