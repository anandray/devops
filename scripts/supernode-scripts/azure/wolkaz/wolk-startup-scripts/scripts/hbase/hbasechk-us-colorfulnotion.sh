#!/bin/bash
# installing GO
if [ ! -d /usr/local/go ]; then
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/go/go-install.sh /home/anand;
sudo sh /home/anand/go-install.sh;
fi

if [ ! -d /usr/local/hbase-1.1.2/ ]; then
# installing hbase if its not installed already OR Deleted in previous step because the script was unable to start it on first attempt
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/hbase/hbase-install-us-colorfulnotion.sh /root/scripts;
sudo sh /root/scripts/hbase-install-us-colorfulnotion.sh;
fi

if ! netstat -apn | grep -E '::8080|::8085' > /dev/null; then
echo hbase is NOT running
pkill -9 java;
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start
else
echo hbase is running
kill -9 $(ps aux | grep hbasechk | grep -v grep | awk '{print$2}')
ps aux | grep hbase > /var/log/hbasechk.log
fi

#if ! curl -s http://`hostname`/ads/systems/hbasehealthcheck.php | grep OK; then
#echo hbase is NOT running
#pkill -9 java;
#cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start
#else
#echo hbase is running
#kill -9 $(ps aux | grep hbasechk | grep -v grep | awk '{print$2}')
#ps aux | grep hbase > /var/log/hbasechk.log
#fi

