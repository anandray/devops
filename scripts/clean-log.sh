#!/bin/bash

percent=`df -h | grep sda | head -n1 | awk '{print$5}' | cut -d "%" -f1`
#percent=`df -h | grep overlay | awk '{print$5}' | cut -d "%" -f1`

#if [[ $percent -ge 85 || $percent -eq 85 ]]; then
if [ $percent -ge 85 ]; then
echo "
`date +%m/%d-%T` - Disk usage is $percent% >= 85% !! Clearing .log file...
"
truncate -s 0 /var/log/messages
truncate -s 0 /var/log/wolk*.log
truncate -s 0 /var/log/wolkbench*.log

service syslog-ng restart

unset percent
echo "
`date +%m/%d-%T` - $percent"

else
echo "
`date +%m/%d-%T` - Disk usage is $percent% < 85%
"
fi

YEAR=$(date +%Y)
if find /var/log -name "*$YEAR*" &> /dev/null; then
find /var/log -name "*$YEAR*" -exec rm -rfv {} \;
unset YEAR
fi
