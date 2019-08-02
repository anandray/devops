#!/bin/bash

percent=`df -h | grep dev | head -n1 | awk '{print$5}' | cut -d "%" -f1`
#percent=`df -h | grep overlay | awk '{print$5}' | cut -d "%" -f1`

if [[ $percent -ge 85 || $percent -eq 85 ]]; then
echo "
`date +%m/%d-%T` - Disk usage is $percent% >= 85% !! Clearing plasma.log file...
"
> /var/log/wolk.log
> /var/log/wolk1.log
> /var/log/wolk2.log
> /var/log/wolk3.log
> /var/log/wolk4.log
> /var/log/wolk5.log
> /var/log/wolk6.log
> /var/log/messages
> /var/log/wolkbench1.log
> /var/log/wolkbench2.log
> /var/log/wolkbench3.log
> /var/log/wolkbench4.log

service syslog-ng restart

unset percent
echo "
`date +%m/%d-%T` - $percent"

else
echo "
`date +%m/%d-%T` - Disk usage is $percent% < 85%
"
fi

if find /var/log -name "*$YEAR*" &> /dev/null; then
YEAR=`date +%Y`
find /var/log -name "*$YEAR*" -exec rm -rfv {} \;
unset YEAR
fi
