#!/bin/bash

syslog_life=`ps aux | grep syslog-ng | grep -v grep | grep -E 'syslog-ng -p /var/run/syslog-ng.pid' | awk '{print$10}' | cut -d":" -f1` > /dev/null
#pid=`ps aux | grep 'syslog-ng -p /var/run/syslog-ng.pid'| grep -v grep | awk '{print$2}'` &> /dev/null
#syslog_life_hr=`ps -p "$pid" -o etime= | cut -d":" -f1` &> /dev/null
#syslog_life_min=`ps -p "$pid" -o etime= | cut -d":" -f2` &> /dev/null
#syslog_life=$((syslog_life_hr * 60 + syslog_life_min))

#echo "$syslog_life"

if [ $syslog_life -lt 15 ];
  then
  echo "Syslog-ng was restarted in the last 15 mins"
else
echo "Syslog-ng has not been restarted in $syslog_life mins"
fi
