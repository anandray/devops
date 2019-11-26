#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

CAMPAIGNMODEL_CHECK=`php -ln "/var/www/vhosts/mdotm.com/httpdocs/system/application/models/campaignmodel.php" | grep 'No syntax errors detected in' | wc -l`

case "${CAMPAIGNMODEL_CHECK}" in
        0)  echo "CAMPAIGNMODEL NOT OK. /httpdocs/system/application/models/campaignmodel.php"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "CAMPAIGNMODEL is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "CAMPAIGNMODEL is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
