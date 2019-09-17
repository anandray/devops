#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#CRON_CHECK=`ps aux | grep cron|grep -v grep|awk '{print $NF}'|grep -E -e '^(/usr/sbin/cron|DataNode)$'|wc -l`
#DATANODE_CHECK=`netstat -apn | grep :50075 | wc -l`
DATANODE_CHECK=`netstat -apn | grep :50075 | grep '::50075' | wc -l`
 
case "${DATANODE_CHECK}" in
        0)  echo "DataNode is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "DataNode is running."; exit ${STATE_OK}
        ;;
#        *)  echo "More than one DataNode process detected / DataNode is in an unknown state."; exit ${STATE_WARNING}
        *)  echo "DataNode is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
