#!/bin/bash

# hbase-us

if netstat -apn | grep java | grep :8080 > /dev/null; then
echo hbase is running
else
echo hbase is NOT running
#pkill -9 java;
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start&
ps aux | grep hbase > /var/log/hbasechk.log
fi

# hbase-profile-check

if netstat -apn | grep java | grep :8081 > /dev/null; then
echo hbase-profile is running
else
echo hbase-profile is NOT running
cd /usr/local/hbase-profile/ && ./bin/hbase-daemon.sh start rest -p 8081 --infoport 8086 
ps aux | grep hbase > /var/log/hbasechk.log
fi

# hbase-eu-chk

if netstat -apn | grep java | grep :8082 > /dev/null; then
echo hbase-eu is running
else
echo hbase-eu is NOT running

cd /usr/local/hbase-eu/ && ./bin/hbase rest start -p 8082 --infoport 8087 &
ps aux | grep hbase > /var/log/hbasechk-eu.log
fi

# hbase-as-chk

if netstat -apn | grep java | grep :8083 > /dev/null; then
echo hbase-as is running
else
echo hbase-as is NOT running
cd /usr/local/hbase-as/ && ./bin/hbase rest start -p 8083 --infoport 8089 &
ps aux | grep hbase > /var/log/hbasechk-as.log
fi
