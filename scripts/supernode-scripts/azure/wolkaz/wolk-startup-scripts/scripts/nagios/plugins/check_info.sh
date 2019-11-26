#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

INFO_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/ads/systems/inf0.php" | grep 'No syntax errors detected in' | wc -l`

case "${INFO_CHECK}" in
        0)  echo "INFO NOT OK. /httpdocs/ads/systems/inf0.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "INFO is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "INFO is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
