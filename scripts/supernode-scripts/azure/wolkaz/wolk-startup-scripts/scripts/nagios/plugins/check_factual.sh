#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#FACTUAL_CHECK=`netstat -apn | grep java | grep ':8989' | wc -l`
FACTUAL_CHECK=`lynx --dump "http://146.148.68.46:8989/zz/status" | grep '{"required-memory":0.0,"state":"OK","indices":{}}' | wc -l`
 
case "${FACTUAL_CHECK}" in
        0)  echo "Factual is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Factual is running."; exit ${STATE_OK}
        ;;
        *)  echo "Factual is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
