#!/bin/bash

log_size=`du -s /var/log/crosschannel.log | awk '{print$1}'`

if [ $log_size -ge 102400000 ]; then
echo "crosschannel log file size $size is larger than 102400000"
fi
