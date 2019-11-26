#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

EMAIL_CHECK=`php /var/www/vhosts/mdotm.com/scripts/systems/email_test.php | grep "Message successfully sent!!" | wc -l`

case "${EMAIL_CHECK}" in
        0)  echo "PUSHCODE EMAIL NOT OK"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "PUSHCODE EMAIL is OK"; exit ${STATE_OK}
        ;;
#        *)  echo "ADXMONITOR is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
