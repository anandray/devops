#!/bin/sh
LIMIT1=8
LIMIT2=2
#cat /proc/loadavg | awk '{print$1}' >$LIMIT #non-integer output
Load_AVG=`uptime | cut -d'l' -f2 | awk '{print $3}' | cut -d. -f1`
if [ $Load_AVG -gt $LIMIT1 ]; then
#   /sbin/service nginx reload;
   /usr/sbin/nginx -s reload;
   php /var/www/vhosts/mdotm.com/scripts/systems/datadog-event.php load
fi
#   if [ $Load_AVG -lt $LIMIT2 ]; then
#   /sbin/service nginx start;
#   fi;
#echo "HIGH"
#else
#echo "LOW"
#fi
