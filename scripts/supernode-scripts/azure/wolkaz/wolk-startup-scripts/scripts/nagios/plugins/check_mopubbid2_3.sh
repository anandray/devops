#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

MOPUBBID2_3_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/ads/mopubbid2_3.php" | grep 'No syntax errors detected in' | wc -l`

case "${MOPUBBID2_3_CHECK}" in
        0)  echo "MOPUBBID2_3 NOT OK. /httpdocs/ads/mopubbid2_3.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "MOPUBBID2_3 is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "MOPUBBID2_3 is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
