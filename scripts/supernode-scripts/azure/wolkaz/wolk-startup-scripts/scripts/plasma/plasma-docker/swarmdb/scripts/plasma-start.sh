#!/bin/bash

#bold=$(tput bold)
#normal=$(tput sgr0)

#if -f /root/go/src/github.com/wolkdb/plasma/build/bin/plasma &> /dev/null; then
#echo "
#${bold}#### S T A R T I N G   P L A S M A ####
#"
# configure and start syslog-ng
if ! md5sum /etc/syslog-ng/syslog-ng.conf | grep 977f1630eda1dac56854af17f864df9b &> /dev/null; then
cp /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-orig
wget -O /etc/syslog-ng/syslog-ng.conf http://www6001.wolk.com/.start/syslog-ng.conf
fi

if ! grep log0 /etc/hosts; then
echo '
104.154.36.231    log6
35.193.168.171    log0' >> /etc/hosts
fi

if ! ps aux | grep "/usr/sbin/syslog-ng -F -p /var/run/syslogd.pid" | grep -v grep; then
/usr/sbin/syslog-ng -F -p /var/run/syslogd.pid &
else
echo "syslog-ng is already running..."
fi

#echo ${normal}
if [ ! -d /root/data ]; then
mkdir -p /root/data
fi

#if ! /usr/bin/md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | grep 3314661b9d1d5e30b236c8102fb684ee &> /dev/null; then
#pkill -9 plasma
#wget -O /root/go/src/github.com/wolkdb/plasma/build/bin/plasma http://www6001.wolk.com/.start/plasma
#chmod +x /root/go/src/github.com/wolkdb/plasma/build/bin/plasma*
#fi

#if ! md5sum /root/.bashrc | grep cbe45e5b9e467651bf3f6ef33da6afda &> /dev/null; then
cp -rf /usr/local/swarmdb/scripts/bashrc /root/.bashrc
#source /root/.bashrc
#exec -l $SHELL
#fi

nohup /root/go/src/github.com/wolkdb/plasma/build/bin/plasma \
--bootnodes enode://827192dd0616dc4f4ae9676b7cb2c56f8fdb478afb04a8f1bc74471806379cf27eaa5d6c1262005180cb516bb171060aaac5d287ddbaddde0ca498eb7fab1f3d@35.193.142.191:30303 \
--datadir /root/data \
--verbosity 4 \
--maxpeers 25 \
2>> /root/data/plasma.log &

#echo ${normal}
#fi

#sleep 5

#if [ -f /root/data/plasma.log ]; then
#echo "
#${bold}Check /root/data/plasma.log
#"

#echo ${normal}
#tail -n10 /root/data/plasma.log
#fi

#echo ${normal}
