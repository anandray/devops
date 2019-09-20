#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#SENDMAIL_CHECK=`ps aux | grep cron|grep -v grep|awk '{print $NF}'|grep -E -e '^(/usr/sbin/cron|crond)$'|wc -l`
SENDMAIL_CHECK=`ps aux | grep sendmail|grep -v grep| grep -v 'accepting connections'|awk '{print $NF}'|grep -E -e '^(/var/spool/clientmqueue|sendmail)$'|wc -l`
 
case "${SENDMAIL_CHECK}" in
        0)  echo "Sendmail is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Sendmail is running."; exit ${STATE_OK}
        ;;
#        *)  echo "More than one crond process detected / crond is in an unknown state."; exit ${STATE_WARNING}
#        *)  echo "More than one crond process detected, BUT FUNCTIONAL."; exit ${STATE_OK}
        *)  echo "Sendmail is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac