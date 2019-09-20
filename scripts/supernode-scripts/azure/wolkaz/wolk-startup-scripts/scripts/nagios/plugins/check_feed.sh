#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

FEED_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/ads/feed.php" | grep 'No syntax errors detected in' | wc -l`

case "${FEED_CHECK}" in
        0)  echo "FEED NOT OK. /httpdocs/ads/feed.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "FEED is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "FEED is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac