#!/bin/bash
if ps -A -U apache | grep httpd > /dev/null; then
echo httpd is running
else
echo httpd is NOT running
pkill -9 httpd;
#kill -9 `ps aux | grep httpdse | awk '{print$2}'`
#killall httpd
/etc/init.d/httpd stop;
/bin/sh /root/scripts/ipcs_remove.sh;
/etc/init.d/httpd start;
#php /var/www/vhosts/mdotm.com/httpdocs/ads/logreset.php;
ps aux | grep httpd > /root/scripts/apachechk.log
cat /root/scripts/apachechk.log | mail -s"apache failed on $HOSTNAME - restart attempted" engineering@mdotm.com
fi
