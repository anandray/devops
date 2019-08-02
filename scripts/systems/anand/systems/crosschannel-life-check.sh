#!/bin/bash

crosschannel_life=`ps aux | grep crosschannel | grep -v grep | grep -E './crosschannel|/var/www/vhosts/crosschannel.com/bidder/bin/crosschannel' | awk '{print$10}' | cut -d":" -f1` > /dev/null

if [ $crosschannel_life -le 15 ];
  then
  echo "Crosschannel was restarted in the last 15 mins"
else
echo "Crosschannel has not been restarted in $crosschannel_life mins"
fi
