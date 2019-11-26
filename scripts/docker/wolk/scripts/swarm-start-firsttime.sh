#!/bin/bash
SHELL=/bin/bash
BASH_ENV=/root/.bashrc
DATADIR=/var/www/vhosts/data
gethaddr=`geth --exec "eth.accounts" attach ipc:/var/www/vhosts/data/geth.ipc | cut -d "\"" -f2`

if ! ps aux | grep "swarm --bzzaccount" | grep -v grep &> /dev/null; then
echo "swarm is not running... starting swarm using $gethaddr..."
#sudo su - << EOF
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
#EOF
sed -i 's/\*\/1 \* \* \* \* sh \/root\/scripts\/swarm-start-firsttime.sh/#\*\/1 \* \* \* \* sh \/root\/scripts\/swarm-start-firsttime.sh/g' /var/spool/cron/root
echo "*/1 * * * * sh /root/scripts/swarm-start-with-changed-ip.sh &>> /var/log/swarm-start-with-changed-ip.log" >> /var/spool/cron/root
fi
sleep 5
pkill -9 swarm
