#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#mysql_time_zone=`mysql --login-path=~/ -hdb03 mdotm -e "show variables like 'time_zone'" | grep time_zone | awk '{print$1,"|",$2}'`
mysql_current_time=`mysql -udb -p1wasb0rn2 -hdb03 mdotm -e "select CURTIME();" | grep -v CUR`
#MYSQL_TIMEZONE_CHECK=`mysql --login-path=~/ -hdb03 mdotm -e "show variables like 'time_zone'" | grep time_zone | awk '{print$2}' | grep "\-07:00" | wc -l`

#MYSQL_TIMEZONE_CHECK=`mysql --login-path=/usr/local/nagios -hdb03 mdotm -e "select CURTIME();" | grep -v CUR | grep -c \$(date +%H:%M)`
MYSQL_TIMEZONE_CHECK=`mysql -udb -p1wasb0rn2 -hdb03 mdotm -e "select CURTIME();" | grep -v CUR | grep -c \$(date +%H:%M)`

case "${MYSQL_TIMEZONE_CHECK}" in
        0)  echo "MYSQL TIME_ZONE NOT OK - $mysql_current_time - change \"default_time_zone\" in SQL"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "MYSQL TIME_ZONE is OK - $mysql_current_time"; exit ${STATE_OK}
        ;;
#        *)  echo "MYSQL TIME_ZONE is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
