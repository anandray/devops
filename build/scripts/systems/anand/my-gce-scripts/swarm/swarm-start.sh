#!/bin/bash
SHELL=/bin/bash
BASH_ENV=/root/.bashrc
DATADIR=/var/www/vhosts/data
gethaddr=`geth --exec "eth.accounts" attach ipc:/var/www/vhosts/data/geth.ipc | cut -d "\"" -f2`

if ! ps aux | grep "swarm --bzzaccount" | grep -v grep &> /dev/null; then
echo "swarm is not running... starting swarm using $gethaddr..."
sudo su - << EOF
nohup swarm \
       --bzzaccount $gethaddr \
       --swap \
       --swap-api /var/www/vhosts/data/geth.ipc \
       --datadir /var/www/vhosts/data \
       --verbosity 6 \
       --ens-api /var/www/vhosts/data/geth.ipc \
       --bzznetworkid 1337 \
       --password $DATADIR/.mdotm \
       2>> /var/www/vhosts/data/swarm.log $DATADIR/.mdotm &
EOF
fi

if grep '127.0.0.1' /var/www/vhosts/data/swarm/bzz-*/config.json &> /dev/null; then
echo "Change ListenAddr from 127.0.0.1 to 0.0.0.0"

sudo pkill -9 swarm
sudo sed -i 's/127.0.0.1/0.0.0.0/g' /var/www/vhosts/data/swarm/bzz-*/config.json

sudo su - << EOF
nohup swarm \
       --bzzaccount $gethaddr \
       --swap \
       --swap-api /var/www/vhosts/data/geth.ipc \
       --datadir /var/www/vhosts/data \
       --verbosity 6 \
       --ens-api /var/www/vhosts/data/geth.ipc \
       --bzznetworkid 1337 \
       --password $DATADIR/.mdotm \
       2>> /var/www/vhosts/data/swarm.log $DATADIR/.mdotm &
EOF
fi
