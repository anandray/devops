#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
GETH_CHECK=`netstat -apn | grep geth | grep ":::30303" | wc -l`
#GETH_CHECK=`ps aux | grep 'geth --bootnodes' | grep -v grep | wc -l`
 
case "${GETH_CHECK}" in
        0)  echo "Geth is not running"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Geth is not running"; exit ${STATE_CRITICAL}
        ;;
        2)  echo "Geth is running"; exit ${STATE_OK}
        ;;
        *)  echo "Geth is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
