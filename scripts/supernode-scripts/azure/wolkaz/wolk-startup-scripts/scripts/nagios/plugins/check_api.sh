#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

API_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/ads/api.php" | grep 'No syntax errors detected in' | wc -l`

case "${API_CHECK}" in
        0)  echo "API NOT OK. /httpdocs/ads/api.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "API is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "API is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
