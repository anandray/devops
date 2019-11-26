#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

LOGGER_CHECK=`php -ln "/var/www/vhosts/mdotm.com/include/logger.php" | grep 'No syntax errors detected in' | wc -l`

case "${LOGGER_CHECK}" in
        0)  echo "LOGGER NOT OK. /httpdocs/ads/logger.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "LOGGER is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "LOGGER is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
