#!/bin/bash
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        git config user.email "sourabh@crosschannel.com";
fi

if ! netstat -apn | egrep '::9900' > /dev/null; then
  echo bt is NOT running
  cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php 2> /var/log/git.err  > /var/log/git.log
else
  echo bt is running
fi

if ! php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/bthealthcheck.php | grep OK; then
echo bt is NOT running
pkill -9 goservice.php
cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php bt 2> /var/log/git.err  > /var/log/git.log
else
echo bt is running
kill -9 $(ps aux | grep go-btchk.sh | grep -v grep | awk '{print$2}')
ps aux | grep bt | grep -v grep > /var/log/go-btchk.log
fi
