#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#Usage: "sh check_syslog-ng-life-log6.sh log6"
 
SYSLOG_LIFE=`ssh -q $1 ps aux | grep syslog-ng | grep -E 'syslog-ng -p /var/run/syslog-ng.pid' | awk '{print$10}' | cut -d":" -f1` > /dev/null
SYSLOG_LIFE1=`ssh -q $1 ps aux | grep syslog-ng | grep -E 'syslog-ng -p /var/run/syslog-ng.pid' | awk '{print$9}'` > /dev/null
SYSLOG_LIFE_CHECK=`ssh -q $1 sh /var/www/vhosts/mdotm.com/scripts/systems/syslog-ng-life-check.sh $1 | grep 'Syslog-ng has not been restarted' | wc -l`
 
case "${SYSLOG_LIFE_CHECK}" in
        0)  echo "Syslog-ng restarted in the last 15 mins, at $SYSLOG_LIFE1 - DO NOTHING - this is for monitoring purpose only"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Syslog-ng has not been restarted in $SYSLOG_LIFE mins."; exit ${STATE_OK}
        ;;
#        *)  echo "More than one syslog-ng process detected / syslog-ng is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac