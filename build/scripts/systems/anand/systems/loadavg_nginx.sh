#!/bin/sh
LIMIT1=200
LIMIT2=5
#cat /proc/loadavg | awk '{print$1}' >$LIMIT #non-integer output
Load_AVG=`uptime | cut -d'l' -f2 | awk '{print $3}' | cut -d. -f1`
if [ $Load_AVG -gt $LIMIT1 ]; then
   #/sbin/service httpd restart && /sbin/service memcached restart >/dev/null 2>&1
   #/sbin/service httpd stop && /bin/kill -9 `pidof httpd` >/dev/null 2>&1
   /sbin/service nginx stop;
   /usr/bin/pkill -9 nginx;
#   /bin/sh /var/www/vhosts/mdotm.com/scripts/systems/ipcs_remove.sh;
   php /var/www/vhosts/mdotm.com/scripts/systems/datadog-event.php load
#fi
#   if [ $Load_AVG -lt $LIMIT2 ]; then
#   /sbin/service nginx start;
#   fi;
#echo "HIGH"
#else
#echo "LOW"
#fi
