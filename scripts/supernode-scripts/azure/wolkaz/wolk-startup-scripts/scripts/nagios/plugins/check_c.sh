#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

C.PHP_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/ads/c.php" | grep 'No syntax errors detected in' | wc -l`

case "${C.PHP_CHECK}" in
        0)  echo "C.PHP NOT OK. /httpdocs/ads/c.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "C.PHP is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "C.PHP is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
