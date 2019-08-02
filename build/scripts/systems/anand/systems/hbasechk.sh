#!/bin/bash
if ps aux | grep java | grep -v grep > /dev/null; then
echo hbase is running
else
echo hbase is NOT running
pkill -9 java;
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start&
ps aux | grep hbase > /var/log/hbasechk.log
cat /var/log/hbasechk.log | mail -s"apache failed on $HOSTNAME - restart attempted" engineering@mdotm.com
fi
