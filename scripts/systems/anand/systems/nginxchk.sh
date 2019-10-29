#!/bin/bash
if ps -A -U nginx | grep nginx > /dev/null; then
echo nginx is running
else
echo nginx is NOT running
pkill -9 nginx;
/etc/init.d/nginx stop;
/etc/init.d/nginx start;
ps aux | grep nginx > /root/scripts/nginxchk.log
#cat /root/scripts/nginxchk.log | mail -s"nginx failed on $HOSTNAME - restart attempted" engineering@mdotm.com
php /var/www/vhosts/mdotm.com/scripts/systems/datadog-event.php nginx
fi
