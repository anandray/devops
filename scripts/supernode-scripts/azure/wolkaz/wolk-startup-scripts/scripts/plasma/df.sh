#!/bin/bash

percent=`df -h | grep dev | head -n1 | awk '{print$5}' | cut -d "%" -f1`
#percent=`df -h | grep overlay | awk '{print$5}' | cut -d "%" -f1`

if [[ $percent -ge 75 || $percent -eq 75 ]]; then
echo "
`date +%m/%d-%T` - Disk usage is $percent% >= 75% !! Clearing plasma.log file...
"
> /var/log/wolk.log
> /var/log/cloudstore.log
> /var/log/messages

unset percent
echo "
`date +%m/%d-%T` - $percent"

else
echo "
`date +%m/%d-%T` - Disk usage is $percent% < 75%
"
fi
