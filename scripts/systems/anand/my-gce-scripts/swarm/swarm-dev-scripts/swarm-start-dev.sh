#!/bin/bash
SHELL=/bin/bash
BASH_ENV=/root/.bashrc
DATADIR=/var/www/vhosts/data
gethaddr=`geth --exec "eth.accounts" attach ipc:/var/www/vhosts/data/geth.ipc | cut -d "\"" -f2`

if grep '127.0.0.1' /var/www/vhosts/data/swarm/bzz-*/config.json &> /dev/null; then
echo "Change ListenAddr from 127.0.0.1 to 0.0.0.0"
#sudo su - << EOF
pkill -9 swarm
sed -i 's/127.0.0.1/0.0.0.0/g' /var/www/vhosts/data/swarm/bzz-*/config.json
#EOF
fi

for i in {1..5}
do
if ! ps aux | grep "swarm --bzzaccount" | grep -v grep &> /dev/null; then
echo "`date +%T` - swarm is not running... starting swarm using $gethaddr..."
#sudo su - << EOF
nohup /usr/local/bin/swarm \
       --bzzaccount $gethaddr \
       --swap \
       --swap-api /var/www/vhosts/data/geth.ipc \
       --datadir /var/www/vhosts/data \
       --verbosity 3 \
       --ens-api /var/www/vhosts/data/geth.ipc \
       --bzznetworkid 55300 \
       2>> /var/www/vhosts/data/swarm.log < <(echo -n "wolk") &
else
echo "`date +%T` - swarm is already running using $gethaddr..."
#EOF
fi
sleep 10
done
