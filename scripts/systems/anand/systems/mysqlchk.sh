#!/bin/sh
if ps -A -U mysql | grep mysqld > /dev/null; then
echo mysqld is running
else
echo mysqld is NOT running
kill -9 `pidof mysqld`
killall -9 mysqld
/sbin/service mysqld stop
/sbin/service mysqld start
/bin/ps aux | grep mysqld > /root/scripts/mysqld.log
#cat /root/scripts/mysqld.log | mail -s"mysqld failed on $HOSTNAME - restart attempted" engineering@mdotm.com
php /var/www/vhosts/mdotm.com/scripts/systems/datadog-event.php mysql
fi
