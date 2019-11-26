#!/bin/bash
#if ps aux | grep java | grep hbase | grep -v grep > /dev/null; then
if netstat -apn | grep java | grep :8080 > /dev/null; then
echo hbase is running
else
echo hbase is NOT running
#pkill -9 java;
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start&
ps aux | grep hbase > /var/log/hbasechk.log
#cat /var/log/hbasechk.log | mail -s"HBase failed on $HOSTNAME - restart attempted" engineering@crosschannel.com
#uptime > /tmp/new_instance_`hostname`.log; cat /tmp/new_instance_`hostname`.log | mail -s"New instance created - $HOSTNAME - `date +%m%d%Y_%T`" engineering@crosschannel.com
fi

if [ ! -d /usr/local/hbase-1.1.2/ ]; then
# installing hbase if its not installed already
sudo gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-install-centOS-us.sh /root/scripts;
sudo sh /root/scripts/hbase-install-centOS-us.sh;
fi
