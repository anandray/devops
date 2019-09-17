#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

STORAGE_CHECK=`php -ln "/var/www/vhosts/mdotm.com/include/storage.php" | grep 'No syntax errors detected in' | wc -l`

case "${STORAGE_CHECK}" in
        0)  echo "STORAGE NOT OK. /include/storage.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "STORAGE is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "STORAGE is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
