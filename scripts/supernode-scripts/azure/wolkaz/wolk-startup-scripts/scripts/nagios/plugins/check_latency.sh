#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

latency=$(grep `date +%s --date='2 second ago'` /var/log/httpd/mdotm.com-access_log | grep -E 'check.php|chartboostbid2_3|mopubbid|adxbid|spotxbid' | grep -v '204 192' | cut -d " " -f1 | head -n1) > /dev/null
LATENCY_CHECK=`sh /var/www/vhosts/mdotm.com/scripts/systems/latency.sh  | head -n20 | grep -c '<' | grep 20 | wc -l`

case "${LATENCY_CHECK}" in
        0)  echo "LATENCY NOT OK - $latency"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "LATENCY is OK - $latency"; exit ${STATE_OK}
        ;;
#        *)  echo "BQ is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
