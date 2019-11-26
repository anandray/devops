#!/bin/bash

percent=`df -h | grep dev | head -n1 | awk '{print$5}' | cut -d "%" -f1`
#percent=`df -h | grep overlay | awk '{print$5}' | cut -d "%" -f1`

if [[ $percent -ge 60 || $percent -eq 60 ]]; then
echo "
`date +%m/%d-%T` - Disk usage is $percent% >= 60% !! Clearing plasma.log file...
"
> /root/plasma/qdata/plasma.log
> /root/sql/qdata/sql.log
> /root/nosql/qdata/nosql.log

unset percent
echo "
`date +%m/%d-%T` - $percent"

else
echo "
`date +%m/%d-%T` - Disk usage is $percent% < 60%
"
fi
