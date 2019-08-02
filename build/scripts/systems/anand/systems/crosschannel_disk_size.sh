#!/bin/bash

disk_size=`df -h | grep 'sda1' | awk '{print$5}' | cut -d "%" -f1`
disk_size1=`df -h | grep 'sda1' | awk '{print$5}'`

if [ $disk_size -ge 75 ];
  then
  echo "current disk usage $disk_size1 - `date +%T`"
  rm -rf /var/log/crosschannel.log && touch /var/log/crosschannel.log && kill -USR1 `cat /var/run/syslog-ng.pid` && sh /var/www/vhosts/mdotm.com/scripts/utils/crosschannelchk.sh
else
echo "current disk usage: $disk_size1 - `date +%T`"
fi

if ! curl -s "http://127.0.0.1/ads/systems/bthealthcheck.php" | grep "Hi there:" > /dev/null; then
	cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh crosschannel
    else
    echo "crosschannel is running"
fi
