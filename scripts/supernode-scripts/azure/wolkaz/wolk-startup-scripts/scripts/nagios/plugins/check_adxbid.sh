#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

ADXBID_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/ads/adxbid.php" | grep 'No syntax errors detected in' | wc -l`

case "${ADXBID_CHECK}" in
        0)  echo "ADXBID NOT OK. /httpdocs/ads/adxbid.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "ADXBID is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "ADXBID is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
