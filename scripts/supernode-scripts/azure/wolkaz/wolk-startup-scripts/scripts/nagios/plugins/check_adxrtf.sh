#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#LOG_FILE_LAST_UPDATED=`ls -lt /var/log/adxrtf.log | awk '{print$6,$7,$8}'`
LOG_FILE_LAST_UPDATED=`stat /var/log/adxrtf.log | grep Modify | awk '{print$1,$2,$3}'`

###############
#ADXRTF_CHECK=$(ls -lt /var/log/adxrtf.log  | grep "`date +%b\ \ %-d\ %H:`" | wc -l) #UNCOMMENT THIS WHEN DATE IS BETWEEN 1st and 9th
#ADXRTF_CHECK=$(ls -lt /var/log/adxrtf.log  | grep "`date +%b\ %-d\ %H:`" | wc -l) #UNCOMMENT THIS WHEN DATE IS NOT BETWEEN 1st and 9th
###############

#ADXRTF_CHECK=$(ls -lt /var/log/adxrtf.log  | grep -E "`date +%b\ \ %-d\ %H:`|`date +%b\ %-d\ %H:`" | wc -l) #UNCOMMENT THIS WHEN DATE IS NOT BETWEEN 1st and 9th
#TIMESTAMP1=`date +%b\ %-d\ %H:%M`
#TIMESTAMP2=`date +%b\ \ %-d\ %H:%M`
TIMESTAMP=`date +'%Y-%m-%d %H'`
#ADXRTF_CHECK=$(ls -lt /var/log/adxrtf.log  | grep -E "$TIMESTAMP1|$TIMESTAMP2" | wc -l)
ADXRTF_CHECK=$(stat /var/log/adxrtf.log | grep Modify | awk '{print$1,$2,$3}' | cut -d ":" -f1,2 | grep "$TIMESTAMP" | wc -l)

case "${ADXRTF_CHECK}" in
#        0)  echo "ADXRTF NOT OK on www6001 - run \"***\ kill -9 \`ps aux | grep adxrtf.php | grep -v grep | awk '{print\$2}'\` ***\" on www6001"; exit ${STATE_CRITICAL}
        0)  echo "ADXRTF NOT OK on www6001 - RUN \"sh /var/www/vhosts/mdotm.com/scripts/systems/kill_adxrtf.sh\" on www6001"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "ADXRTF is OK - log file \"/var/log/adxrtf.log\" last updated: $LOG_FILE_LAST_UPDATED"; exit ${STATE_OK}
        ;;
#        *)  echo "ADXRTF is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
