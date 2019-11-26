#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#BIGTABLE_CHECK=`netstat -apn | grep zn | grep ":::9901" | wc -l`
#BIGTABLE_CHECK=`sh /var/www/vhosts/mdotm.com/httpdocs/ads/systems/znhealthcheck.sh | grep "Escape character is '^]'" | wc -l`
ZN_CHECK=`php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/znhealthcheck.php | grep OK | wc -l`
 
case "${ZN_CHECK}" in
        0)  echo "ZN is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "ZN is running."; exit ${STATE_OK}
        ;;
        *)  echo "ZN is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
