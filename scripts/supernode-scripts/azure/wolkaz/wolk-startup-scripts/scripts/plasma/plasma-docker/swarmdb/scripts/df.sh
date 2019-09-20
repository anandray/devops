#!/bin/bash

percent=`df -h | grep sda1 | awk '{print$5}' | cut -d "%" -f1`

if [[ $percent -ge 40 || $percent -eq 40 ]]; then
echo "
`date +%m/%d-%T` - Disk usage is $percent% >= 40% !! Clearing geth.log file...
"
> /usr/local/swarmdb/data/geth.log

unset percent
echo "
`date +%m/%d-%T` - $percent"

else
echo "
`date +%m/%d-%T` - Disk usage is $percent% < 40%
"
fi