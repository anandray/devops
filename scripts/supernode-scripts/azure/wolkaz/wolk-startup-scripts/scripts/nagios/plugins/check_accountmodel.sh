#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

ACCOUNTMODEL_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/system/application/models/accountmodel.php" | grep 'No syntax errors detected in' | wc -l`

case "${ACCOUNTMODEL_CHECK}" in
        0)  echo "ACCOUNTMODEL NOT OK. /httpdocs/system/application/models/accountmodel.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "ACCOUNTMODEL is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "ACCOUNTMODEL is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
