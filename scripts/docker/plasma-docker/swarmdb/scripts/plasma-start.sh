#!/bin/bash

#bold=$(tput bold)
#normal=$(tput sgr0)

#if -f /root/go/src/github.com/wolkdb/plasma/build/bin/plasma &> /dev/null; then
#echo "
#${bold}#### S T A R T I N G   P L A S M A ####
#"
# configure and start syslog-ng
md5=`ssh -q www6001 md5sum /var/www/vhosts/mdotm.com/httpdocs/.start/syslog-ng.conf | cut -d " " -f1`
md5local=`md5sum /etc/syslog-ng/syslog-ng.conf | cut -d " " -f1`

#if ! md5sum /etc/syslog-ng/syslog-ng.conf | grep $md5 &> /dev/null; then
#'yes' | cp /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-orig
#wget -O /etc/syslog-ng/syslog-ng.conf http://www6001.wolk.com/.start/syslog-ng.conf
#fi

#if ! grep log0 /etc/hosts; then
#echo '
#35.193.168.171    log0' >> /etc/hosts
#fi

#if ! ps aux | grep "/usr/sbin/syslog-ng -F -p /var/run/syslogd.pid" | grep -v grep; then
#/usr/sbin/syslog-ng -F -p /var/run/syslogd.pid &
#else
#echo "syslog-ng is already running..."
#fi

#echo ${normal}
if [ ! -d /root/newdata ]; then
mkdir -p /root/newdata
fi

md5plasma=`ssh -q 35.193.7.46 md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | cut -d " " -f1`
md5www6001=`ssh -q 35.202.46.8 md5sum /var/www/vhosts/mdotm.com/httpdocs/.start/plasma | cut -d " " -f1`
md5localplasma=`md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | cut -d " " -f1`

if ! md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | grep $md5plasma &> /dev/null; then
echo "
MD5 of \"PLASMA\" does not match with \"PLASMA\" on \"anand-docker\"
"
  if ps aux | grep "plasma \--datadir" | grep -v grep; then
    kill -9 $(ps aux | grep "plasma \--datadir" | grep -v grep | awk '{print$2}')
    sleep 3
  fi
mkdir -p /root/go/src/github.com/wolkdb/plasma/build/bin
wget -O /root/go/src/github.com/wolkdb/plasma/build/bin/plasma http://www6001.wolk.com/.start/plasma
sleep 3

else
echo "
MD5 of \"PLASMA\" matches with \"PLASMA\" on \"anand-docker\"

MD5 --> $md5local
"
sleep 2
fi

#if ! md5sum /root/.bashrc | grep cbe45e5b9e467651bf3f6ef33da6afda &> /dev/null; then
cp -rf /usr/local/swarmdb/scripts/bashrc /root/.bashrc
#source /root/.bashrc
#exec -l $SHELL
#fi

if ! ps aux | grep "\--datadir /root/newdata" | grep -v grep; then
echo "
plasma on newdata dir is not running.. Starting plasma...
"
rm -rf /tmp/plasma/*
rm -rf /tmp/plasmachain/*
rm -rf /root/newdata/*

nohup /root/go/src/github.com/wolkdb/plasma/build/bin/plasma \
--datadir /root/newdata \
--rpc \
--rpcaddr 0.0.0.0 \
--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,plasma,swarmdb \
--rpcport 8545 \
--rpccorsdomain=* \
--rpcvhosts=* \
--verbosity 4 \
--maxpeers  25 \
2>> /root/newdata/plasma.log &

else
echo "
Plasma is already running on newdata dir...
"
fi
