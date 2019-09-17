#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
AEROSPIKE_CHECK=`ps aux | grep aerospike|grep -v grep|awk '{print $NF}'|grep -E -e '^(/etc/aerospike/aerospike.conf|aerospike)$'|wc -l`
 
case "${AEROSPIKE_CHECK}" in
        0)  echo "Aerospike is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Aerospike is running."; exit ${STATE_OK}
        ;;
        *)  echo "More than one aerospike process detected / aerospike is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
