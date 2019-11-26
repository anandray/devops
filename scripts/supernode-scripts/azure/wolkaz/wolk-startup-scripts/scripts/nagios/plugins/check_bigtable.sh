#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#BIGTABLE_CHECK=`netstat -apn | grep bt | grep ":::9900" | wc -l`
#BIGTABLE_CHECK=`sh /var/www/vhosts/mdotm.com/httpdocs/ads/systems/bthealthcheck.sh | grep "Escape character is '^]'" | wc -l`
BIGTABLE_CHECK=`php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/bthealthcheck.php | grep OK | wc -l`
 
case "${BIGTABLE_CHECK}" in
        0)  echo "BigTable is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "BigTable is running."; exit ${STATE_OK}
        ;;
        *)  echo "BigTable is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
