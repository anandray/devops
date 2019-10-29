#!/bin/bash

percent=`df -h | grep sda1 | awk '{print$5}' | cut -d "%" -f1`

if [[ $percent -ge 80 || $percent -eq 80 ]]; then
echo "
`date +%m/%d-%T` - Disk usage is $percent% >= 80% !! Clearing geth.log file...
"

wc1=`wc -l /var/log/wolkcronlog/nosql_sim_sh.log`
wc2=`wc -l /var/log/wolkcronlog/nosql_sim.log`

diff1=`awk -v var1="$wc1" -v var2=2 'BEGIN { print  ( var1 / var2 ) }'`
diff2=`awk -v var3="$wc2" -v var4=2 'BEGIN { print  ( var3 / var4 ) }'`

diff3=`echo $diff1 | cut -d "." -f1`
diff4=`echo $diff2 | cut -d "." -f1`

#echo $diff3
#echo $diff4

tail -n $diff3 /var/log/wolkcronlog/nosql_sim_sh.log > /var/log/wolkcronlog/nosql_sim_sh.log1
tail -n $diff4 /var/log/wolkcronlog/nosql_sim.log > /var/log/wolkcronlog/nosql_sim.log1

mv -f /var/log/wolkcronlog/nosql_sim_sh.log1 /var/log/wolkcronlog/nosql_sim_sh.log
mv -f /var/log/wolkcronlog/nosql_sim.log1 /var/log/wolkcronlog/nosql_sim.log

#> /var/log/wolkcronlog/nosql_sim_sh.log
#> /var/log/wolkcronlog/nosql_sim.log

unset percent wc1 wc2 diff1 diff2 diff3 diff4
echo "
`date +%m/%d-%T` - $percent"

else
echo "
`date +%m/%d-%T` - Disk usage is $percent% < 80%
"
fi
