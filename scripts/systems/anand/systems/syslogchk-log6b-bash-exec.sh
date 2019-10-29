#!/bin/bash
YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`
DATE=`date +%m%d%Y-%T`

syslog=`sh /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log6b-bash.sh | wc -l` > /dev/null
threshold="20"

if [ "$syslog" -lt "$threshold" ]; then
echo "SYSLOG NOT OK"
  if ! netstat -apn | grep '0 0.0.0.0:5000' | grep LISTEN;
  then
  /sbin/service syslog-ng restart
else
  pkill -9 syslog-ng && /sbin/service syslog-ng restart
  fi
else
echo "$DATE - SYSLOG OK"
fi
