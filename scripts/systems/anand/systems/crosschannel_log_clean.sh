#!/bin/bash

count=`cat /var/log/crosschannel.log | wc -l` > /dev/null

if [ $count -ge 100000 ]; then
echo "/var/log/crosschannel.log has > 100000 lines - $count lines"
tail -n 100000 /var/log/crosschannel.log > /var/log/crosschannel.log;
else
echo "/var/log/crosschannel.log has < 100000 lines - $count lines"
fi
