#!/bin/bash
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        git config user.email "sourabh@crosschannel.com";
fi

# installing GO
if [ ! -d /usr/local/go ]; then
sudo gsutil cp gs://startup_scripts_us/scripts/go/go-install.sh /home/anand;
sudo sh /home/anand/go-install.sh;
fi

#> /tmp/bthealthcheck.log
#for ((n=0;n<10;n++))
#do
#/usr/bin/php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/bthealthcheck.php | /bin/grep OK >> /tmp/bthealthcheck.log
#done
if ! php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/bthealthcheck.php | grep OK; then
  cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php bt;
else
  echo bt is running
fi

if netstat -apn | egrep '::9900' > /dev/null; then
  echo bt is running
else
  echo bt is NOT running
  cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php bt 2> /var/log/git-bt.err  > /var/log/git-bt.log
fi

if netstat -apn | egrep '::8080|::8085' > /dev/null; then
echo hbase is running
fi

if ! curl -s http://`hostname`/ads/systems/hbasehealthcheck.php | grep OK > /dev/null; then
echo hbase is NOT running
pkill -9 java;
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start&
else
echo hbase is running
kill -9 $(ps aux | grep hbasechk | grep -v grep | awk '{print$2}')
ps aux | grep hbase > /var/log/hbasechk.log
fi

if [ ! -d /usr/local/hbase-1.1.2/ ]; then
# installing hbase if its not installed already OR Deleted in previous step because the script was unable to start it on first attempt
sudo gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-install-centOS-us-east.sh /root/scripts;
sudo sh /root/scripts/hbase-install-centOS-us-east.sh;
fi
