#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

NETWORKMODEL_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/system/application/models/networkmodel.php" | grep 'No syntax errors detected in' | wc -l`

case "${NETWORKMODEL_CHECK}" in
        0)  echo "NETWORKMODEL NOT OK. httpdocs/system/application/models/networkmodel.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "NETWORKMODEL is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "NETWORKMODEL is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
