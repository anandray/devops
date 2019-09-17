#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#HBASE_CHECK=`curl -s http://\`hostname\`/ads/systems/hbasehealthcheck.php | grep OK | wc -l`
HBASE_CHECK=`php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/hbasehealthcheck.php | grep OK | wc -l`
 
case "${HBASE_CHECK}" in
        0)  echo "HBase is not running."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "HBase is running."; exit ${STATE_OK}
        ;;
        *)  echo "HBase is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
