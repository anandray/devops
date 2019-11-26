#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

MDOTMFEEDPICO_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/ads/mdotmfeed-pico.php" | grep 'No syntax errors detected in' | wc -l`

case "${MDOTMFEEDPICO_CHECK}" in
        0)  echo "MDOTMFEEDPICO NOT OK. /httpdocs/ads/mdotmfeed-pico.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "MDOTMFEEDPICO is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "MDOTMFEEDPICO is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
