#!/bin/bash

disk_size=`df -h | grep 'sda1' | awk '{print$5}' | cut -d "%" -f1`
disk_size1=`df -h | grep 'sda1' | awk '{print$5}'`

if [ $disk_size -ge 50 ];
  then
  echo "current disk usage $disk_size1 - `date +%T`"
> /var/log/ccdex.log
else
echo "current disk usage: $disk_size1 - `date +%T`"
fi

#if ! curl -s "http://127.0.0.1/data/healthcheck" | grep OK > /dev/null; then
#	cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh ccdex
#    else
#    echo "ccdex is running"
#fi
