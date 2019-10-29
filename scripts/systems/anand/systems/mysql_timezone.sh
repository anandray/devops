#!/bin/bash

#mysql_current_time=$(date -d `mysql --login-path=/usr/local/nagios -hdb03 mdotm -e "select CURTIME();" | grep -v CUR` +"%Y%m%d%H%M")
mysql_current_time=$(date -d `mysql -udb -p1wasb0rn2 -hdb03 mysql -e "select CURTIME();" | grep -v CUR` +"%Y%m%d%H%M")
system_time=$(date -d `date +%T` +"%Y%m%d%H%M")

if [[ ! "$mysql_current_time" = "$system_time" ]];
  then
  echo "$mysql_current_time not same as the  $system_time"
else
echo "MYSQL TIME_ZONE is OK - $mysql_current_time IS EQUAL TO $system_time"
fi
