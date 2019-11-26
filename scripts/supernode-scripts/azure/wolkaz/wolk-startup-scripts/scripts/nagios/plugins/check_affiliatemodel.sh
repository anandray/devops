#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

AFFILIATEMODEL_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/system/application/models/affiliatemodel.php" | grep 'No syntax errors detected in' | wc -l`

case "${AFFILIATEMODEL_CHECK}" in
        0)  echo "AFFILIATEMODEL NOT OK. /httpdocs/system/application/models/affiliatemodel.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "AFFILIATEMODEL is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "AFFILIATEMODEL is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
