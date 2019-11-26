#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#CRON_CHECK=`ps aux | grep cron|grep -v grep|awk '{print $NF}'|grep -E -e '^(/usr/sbin/cron|crond)$'|wc -l`
#TD_CHECK=`netstat -apn | egrep ':8888|:24220|:24224|:24230|td-agent.sock' |awk '{print $NF}' | wc -l`
TD_CHECK=`ps aux | grep ruby | grep -v grep | awk '{print $NF}' | egrep 'td-agent|td-agent.log|td-agent.pid' | wc -l`
 
case "${TD_CHECK}" in
        0)  echo "TreasureData is not running."; exit ${STATE_CRITICAL}
        ;;
        2)  echo "TreasureData is running."; exit ${STATE_OK}
        ;;
        *)  echo "More than one TreasureData process detected / TreasureData is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
