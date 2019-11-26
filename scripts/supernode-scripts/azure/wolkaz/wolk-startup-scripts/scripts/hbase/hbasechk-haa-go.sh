#!/bin/bash
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        cd /var/www/vhosts
        git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        
        git config core.filemode false;
        git config user.email "sourabh@crosschannel.com";
        git config user.name "Sourabh Niyogi";
        git fetch upstream;
        git merge upstream/master;
        sudo chmod -R 0755 /var/www/vhosts/crosschannel.com/bidder/bin 
EOF
fi

if ! netstat -apn | grep bt | grep '::9900' | grep LISTEN > /dev/null; then
  echo bt is NOT running
  cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh bt 
else
  echo bt is running
fi

if ! php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/bthealthcheck.php | grep OK; then
  echo bt is NOT running
  cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh bt
else
  echo bt is running
fi

if ! netstat -apn | grep java | grep :8080 > /dev/null 2>&1; then
echo hbase is NOT running
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start&
else
echo hbase is running
fi

if ! php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/hbasehealthcheck.php | grep OK > /dev/null 2>&1; then
echo hbase is NOT running
cd /usr/local/hbase-1.1.2/ && ./bin/hbase rest start&

elif ! php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/hbasehealthcheck.php | grep OK > /dev/null 2>&1; then
echo "Installing hbase from scratch..."
sudo rm -rfv /usr/local/hbase-1.1.2*;
sudo gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-install-haa.sh /home/anand;
sudo sh /home/anand/hbase-install-haa.sh;

else
echo hbase is running
fi

if [ ! -d /usr/local/hbase-1.1.2/ ]; then
# installing hbase if its not installed already
sudo gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-install-haa.sh /home/anand;
sudo sh /home/anand/hbase-install-haa.sh;
fi
