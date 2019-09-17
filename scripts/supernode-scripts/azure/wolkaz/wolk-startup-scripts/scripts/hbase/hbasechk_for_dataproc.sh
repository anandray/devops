#!/bin/bash
if ps aux | grep java | grep hbase | grep -v grep > /dev/null; then
echo hbase is running
else
echo hbase is NOT running
#pkill -9 java;
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start&
ps aux | grep hbase > /var/log/hbasechk.log
cat /var/log/hbasechk.log | mail -s"HBase failed on $HOSTNAME - restart attempted" engineering@mdotm.com
uptime > /tmp/new_instance_`hostname`.log; cat /tmp/new_instance_`hostname`.log | mail -s"New instance created - $HOSTNAME - `date +%m%d%Y_%T`" engineering@mdotm.com
fi
