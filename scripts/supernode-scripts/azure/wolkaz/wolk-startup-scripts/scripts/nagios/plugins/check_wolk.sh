#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#WOLK_CHECK=`netstat -apn | grep wolk | grep ":::80" | wc -l`
WOLK_CHECK=`curl -s "http://127.0.0.1/data/healthcheck" | grep OK | wc -l`

case "${WOLK_CHECK}" in
        0)  echo "WOLK is not running - run \"sh /var/www/vhosts/mdotm.com/scripts/systems/wolkchk.sh\" to start WOLK"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "WOLK is running."; exit ${STATE_OK}
        ;;
        *)  echo "WOLK is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
