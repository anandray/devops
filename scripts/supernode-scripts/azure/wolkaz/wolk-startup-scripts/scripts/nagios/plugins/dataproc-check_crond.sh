#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#CRON_CHECK=`ps aux | grep cron|grep -v grep|awk '{print $NF}'|grep -E -e '^(/usr/sbin/cron|crond)$'|wc -l`
CRON_CHECK=`ps aux | grep cron|grep -v grep|awk '{print $11}'|grep -E -e '^(/usr/sbin/cron)$'|wc -l`
 
case "${CRON_CHECK}" in
        0)  echo "Crond is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Crond is running."; exit ${STATE_OK}
        ;;
        *)  echo "More than one crond process detected / crond is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
