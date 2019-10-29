#!/bin/bash

if netstat -apn | grep java | grep :8080 > /dev/null; then
echo hbase is running
else
echo hbase is NOT running
#pkill -9 java;
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start&
fi

if [ ! -d /usr/local/hbase-1.1.2/ ]; then
# installing hbase if its not installed already
sudo gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-install-centOS-us.sh /root/scripts;
sudo sh /root/scripts/hbase-install-centOS-us.sh;
fi
