#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#SYNTAX_CHECK=`php -ln "/var/www/vhosts/mdotm.com/include/storage.php" | grep 'No syntax errors detected in' | wc -l`
#SYNTAX_CHECK=`/bin/find /var/www/vhosts/mdotm.com \( -name "inf0.php" -or -name "storage.php" -or -name "logger.php" -or -name "c.php" -or -name "api.php" -or -name "mdotmfeed-pico.php"-or -name "feed.php" -or -name "adxbid.php" -or -name "mopubbid2_3.php" -or -name "networkmodel.php" -or -name "accountmodel.php" -or -name "campaignmodel.php" -or -name "affiliatemodel.php" \) -print0 | xargs -0 -n 1 php -l | grep -v 'api/1.0/api.php' | grep -c 'No syntax errors detected in' | grep 14 | wc -l`
 
SYNTAX_CHECK=`/bin/find /var/www/vhosts/mdotm.com \( -name "storage.php" -or -name "logger.php" -or -name "c.php" -or -name "api.php" -or -name "mdotmfeed-pico.php"-or -name "feed.php" -or -name "adxbid.php" -or -name "mopubbid2_3.php" -or -name "networkmodel.php" -or -name "accountmodel.php" -or -name "campaignmodel.php" -or -name "affiliatemodel.php" \) -print0 | xargs -0 -n 1 php -l | grep -v 'api/1.0/api.php' | grep -c 'No syntax errors detected in' | grep 12 | wc -l`
 
case "${SYNTAX_CHECK}" in
        0)  echo "SYNTAX NOT OK."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "SYNTAX is OK."; exit ${STATE_OK}
        ;;
#        *)  echo "SYNTAX is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
