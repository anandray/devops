#!/bin/bash
#if ps aux | grep java | grep hbase | grep -v grep > /dev/null; then
if netstat -apn | grep java | grep :8082 > /dev/null; then
echo hbase is running
else
echo hbase is NOT running

cd /usr/local/hbase-eu/ && ./bin/hbase rest start -p 8082 --infoport 8087 &
ps aux | grep hbase > /var/log/hbasechk-eu.log
fi
