#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
TIMESTAMP=`date +%s`
TwoMinutesAgo=$(date +'%s' --date='2 minutes ago')
 
#500_CHECK=`ps aux | grep cron|grep -v grep|awk '{print $NF}'|grep -E -e '^(/usr/sbin/cron|crond)$'|wc -l`
#CHECK_500=`tail -n1000 /var/log/httpd/mdotm.com-access_log | grep ' 500 ' | cut -d " " -f4 | uniq -c | wc -l`
#CHECK_500=`grep \`date +%s\` /var/log/httpd/mdotm.com-access_log | grep ' 500 ' | cut -d " " -f4 | uniq -c | wc -l`
CHECK_500=`grep $TwoMinutesAgo /var/log/httpd/mdotm.com-access_log | grep ' 500 ' | cut -d " " -f4 | uniq -c | wc -l`
 
case "${CHECK_500}" in
        0)  echo "There are no 500 ERRORS.."; exit ${STATE_OK}
        ;;
        1)  echo "There are 500 ERRORS.. CHECK \"tail -f /var/log/httpd/mdotm.com-access_log\" for errors"; exit ${STATE_CRITICAL}
        ;;
#        *)  echo "More than one crond process detected / crond is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
