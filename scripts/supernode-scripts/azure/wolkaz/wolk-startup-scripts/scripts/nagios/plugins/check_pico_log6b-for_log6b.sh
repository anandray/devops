#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

LOG_FILE_LAST_UPDATED=`ls -lt /var/log/pico.log | awk '{print$6,$7,$8}'`

###############
#PICO_CHECK=$(ls -lt /var/log/pico.log  | grep "`date +%b\ \ %-d\ %H:`" | wc -l) #UNCOMMENT THIS WHEN DATE IS BETWEEN 1st and 9th
#PICO_CHECK=$(ls -lt /var/log/pico.log  | grep "`date +%b\ %-d\ %H:`" | wc -l) #UNCOMMENT THIS WHEN DATE IS NOT BETWEEN 1st and 9th
###############

#PICO_CHECK=$(ls -lt /var/log/pico.log  | grep -E "`date +%b\ \ %-d\ %H:`|`date +%b\ %-d\ %H:`" | wc -l) #UNCOMMENT THIS WHEN DATE IS NOT BETWEEN 1st and 9th
TIMESTAMP1=`date +%b\ %-d\ %H:%M`
TIMESTAMP2=`date +%b\ \ %-d\ %H:%M`
PICO_CHECK=$(ls -lt /var/log/pico.log  | grep -E "$TIMESTAMP1|$TIMESTAMP2" | wc -l)

case "${PICO_CHECK}" in
#        0)  echo "PICO NOT OK on log6b - run \"***\ kill -9 \`ps aux | grep pico.php | grep -v grep | awk '{print\$2}'\` ***\" on log6b"; exit ${STATE_CRITICAL}
        0)  echo "PICO NOT OK on log6b - RUN \"sh /var/www/vhosts/mdotm.com/scripts/systems/kill_pico.sh\" on log6b"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "PICO is OK - log file \"/var/log/pico.log\" last updated: $LOG_FILE_LAST_UPDATED"; exit ${STATE_OK}
        ;;
#        *)  echo "BQ is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
