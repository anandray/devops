#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#CCDEX_CHECK=`netstat -apn | grep ccdex | grep ":::80" | wc -l`
CCDEX_CHECK=`curl -s "http://127.0.0.1/data/healthcheck" | grep OK | wc -l`

case "${CCDEX_CHECK}" in
        0)  echo "CCdex is not running - run \"sh /var/www/vhosts/mdotm.com/scripts/systems/ccdexchk.sh\" to start CCDEX"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "CCdex is running."; exit ${STATE_OK}
        ;;
        *)  echo "CCdex is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
