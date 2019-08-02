#!/bin/bash
#if ps aux | grep java | grep hbase | grep -v grep > /dev/null; then
if netstat -apn | grep java | grep :8083 > /dev/null; then
echo hbase is running
else
echo hbase is NOT running
cd /usr/local/hbase-as/ && ./bin/hbase rest start -p 8083 --infoport 8088 &
ps aux | grep hbase > /var/log/hbasechk-as.log
fi
