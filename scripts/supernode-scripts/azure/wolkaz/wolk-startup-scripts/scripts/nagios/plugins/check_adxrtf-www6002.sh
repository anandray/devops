#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

LOG_FILE_LAST_UPDATED=`stat /var/log/adxrtf.log | grep Modify | awk '{print$1,$2,$3}'`

TIMESTAMP=`date +'%Y-%m-%d %H'`
ADXRTF_CHECK=$(ssh -q www6001 stat /var/log/adxrtf.log | grep Modify | awk '{print$1,$2,$3}' | cut -d ":" -f1,2 | grep "$TIMESTAMP" | wc -l)

case "${ADXRTF_CHECK}" in
        0)  echo "ADXRTF NOT OK on www6001 - RUN \"sh /var/www/vhosts/mdotm.com/scripts/systems/kill_adxrtf.sh\" on www6001"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "ADXRTF is OK - log file \"/var/log/adxrtf.log\" last updated: $LOG_FILE_LAST_UPDATED"; exit ${STATE_OK}
        ;;
#        *)  echo "ADXRTF is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
