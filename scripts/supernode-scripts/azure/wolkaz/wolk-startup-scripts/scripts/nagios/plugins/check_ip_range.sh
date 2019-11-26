#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

IP_RANGE=`sh /root/scripts/ip_range.sh | awk '{print$2}' | sort | uniq -c | awk '{print$2}' | wc -l | grep 7 | wc -l`

case "${IP_RANGE}" in
        0)  echo "NEW IP Range \"sh /root/scripts/ip_range.sh | awk '{print$2}' | sort | uniq -c | awk '{print$2}' | wc -l\" to see NEW range."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "NO NEW IP Range."; exit ${STATE_OK}
        ;;
#        *)  echo "BigTable is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
