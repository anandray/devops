#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#CRON_CHECK=`ps aux | grep cron|grep -v grep|awk '{print $NF}'|grep -E -e '^(/usr/sbin/cron|TaskTracker)$'|wc -l`
#TASKTRACKER_CHECK=`jps | grep Task | grep -v grep | awk '{print $NF}' | grep -E -e '^(TaskTracker)$'|wc -l`
#TASKTRACKER_CHECK=`netstat -apn | grep :50060 | wc -l`
TASKTRACKER_CHECK=`netstat -apn | grep :50060 | grep '::50060' | wc -l`
 
case "${TASKTRACKER_CHECK}" in
        0)  echo "TaskTracker is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "TaskTracker is running."; exit ${STATE_OK}
        ;;
#        *)  echo "More than one TaskTracker process detected / TaskTracker is in an unknown state."; exit ${STATE_WARNING}
        *)  echo "TaskTracker is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
