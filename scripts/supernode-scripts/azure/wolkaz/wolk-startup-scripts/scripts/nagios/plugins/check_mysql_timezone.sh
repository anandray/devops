#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#mysql_current_time=`mysql --login-path=/usr/local/nagios -hdb03 mdotm -e "select CURTIME();" | grep -v CUR`
mysql_current_time=`mysql -udb -p1wasb0rn2 -hdb03 mdotm -e "select CURTIME();" | grep -v CUR`
system_time=`date +%T`

#mysql_current_time=`mysql --login-path=/usr/local/nagios -hdb03 mdotm -e "select CURTIME();" | grep -v CUR | grep -c $(date +%H:%M)`

#MYSQL_TIMEZONE_CHECK=`mysql --login-path=~/ -hdb03 mdotm -e "show variables like 'time_zone'" | grep time_zone | awk '{print$2}' | grep "\-07:00" | wc -l`
#MYSQL_TIMEZONE_CHECK=`mysql --login-path=/usr/local/nagios -hdb03 mdotm -e "select CURTIME();" | grep -v CUR | grep -c \$(date +%H:%M)`
#MYSQL_TIMEZONE_CHECK=`mysql --login-path=/usr/local/nagios -hdb03 mdotm -e "select CURTIME();" | grep -v CUR | grep \$(date +%H:%M) | wc -l`
#MYSQL_TIMEZONE_CHECK=`sh /root/scripts/check_mysql_timezone.sh  | grep 'MYSQL TIME_ZONE is OK' | wc -l`

MYSQL_TIMEZONE_CHECK=`sh /var/www/vhosts/mdotm.com/scripts/systems/mysql_timezone.sh  | grep 'MYSQL TIME_ZONE is OK' | wc -l`

case "${MYSQL_TIMEZONE_CHECK}" in
        0)  echo "MYSQL TIME_ZONE NOT OK - $mysql_current_time - change \"default_time_zone\" in SQL"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "MYSQL TIME_ZONE is OK - $mysql_current_time vs $system_time"; exit ${STATE_OK}
        ;;
#        *)  echo "MYSQL TIME_ZONE is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
