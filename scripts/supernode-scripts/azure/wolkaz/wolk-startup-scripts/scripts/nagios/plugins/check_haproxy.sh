#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#HAPROXY_CHECK=`ps aux | grep haproxy|grep -v grep|awk '{print $NF}'|wc -l`
#HAPROXY_CHECK=`ps aux | grep haproxy|grep -v grep| awk '{print$NF}'|grep -E -e haproxy| wc -l`
HAPROXY_CHECK=`ps aux | grep docker|grep -v grep| awk '{print$NF}'|grep -E -e haproxy| wc -l`
 
case "${HAPROXY_CHECK}" in
        0)  echo "Haproxy is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Haproxy is running."; exit ${STATE_OK}
        ;;
        *)  echo "Haproxy is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
