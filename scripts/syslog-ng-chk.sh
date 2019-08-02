#!/bin/bash
if ! ps -A -U root | grep syslog-ng > /dev/null;
  then
    echo "syslog-ng is NOT running, restarting syslog-ng..."
    service syslog-ng restart;
else
echo syslog-ng is running

#cat /root/scripts/syslog-ngchk.log | mail -s"syslog-ng failed on $HOSTNAME - restart attempted" systems@mdotm.com
#php /var/www/vhosts/mdotm.com/scripts/systems/datadog-event.php syslog
fi

if ! netstat -apn | grep syslog-ng | grep LISTEN | grep ':5000' > /dev/null;
  then
    echo "syslog-ng is NOT running, restarting syslog-ng..."
    service syslog-ng restart;
else
echo syslog-ng is running
fi
