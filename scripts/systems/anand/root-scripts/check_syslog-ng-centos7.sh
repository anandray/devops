#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
SYSLOG_CHECK=`ps aux | grep syslog-ng|grep -v grep|awk '{print $NF}' | grep -E -e '^(/var/run/syslogd.pid|syslogd)$' | wc -l`
 
case "${SYSLOG_CHECK}" in
        0)  echo "syslog-ng is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "syslog-ng is running."; exit ${STATE_OK}
        ;;
        2)  echo "syslog-ng is not running."; exit ${STATE_CRITICAL}
        ;;
#        *)  echo "More than one syslog-ng process detected / syslog-ng is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac