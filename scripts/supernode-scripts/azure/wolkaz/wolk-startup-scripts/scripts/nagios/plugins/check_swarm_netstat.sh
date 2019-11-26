#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
SWARM_CHECK=`netstat -apn | grep swarm | grep -E ':::8500|:::30399' | wc -l`
 
case "${SWARM_CHECK}" in
        0)  echo "Swarm is not running"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Swarm is not running"; exit ${STATE_CRITICAL}
        ;;
        2)  echo "Swarm is not running"; exit ${STATE_CRITICAL}
        ;;
        3)  echo "Swarm is running"; exit ${STATE_OK}
        ;;
        *)  echo "Swarm is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
