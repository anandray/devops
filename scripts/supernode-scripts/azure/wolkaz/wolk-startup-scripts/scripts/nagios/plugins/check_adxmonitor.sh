#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#LOG_FILE_LAST_UPDATED=`stat /var/log/report.log | grep Modify | awk '{print$1,$2,$3}'`
LOG_FILE_LAST_UPDATED=`stat /var/log/report.log | grep Modify | awk '{print$1,$2,$3}' | cut -d ":" -f1,2,3,4 | cut -d "." -f1`

TIMESTAMP=`date +'%Y-%m-%d %H'`
TIMESTAMP1=`date +'%Y-%m-%d %H' --date='1 hour ago'`
#ADXMONITOR_CHECK=$(stat /var/log/report.log | grep Modify | awk '{print$1,$2,$3}' | cut -d ":" -f1,2 | grep "$TIMESTAMP" | wc -l)
ADXMONITOR_CHECK=$(stat /var/log/report.log | grep Modify | awk '{print$1,$2,$3}' | cut -d ":" -f2,3 | grep -E "$TIMESTAMP:18|$TIMESTAMP:19|$TIMESTAMP:20|$TIMESTAMP:21|$TIMESTAMP:22|$TIMESTAMP:41|$TIMESTAMP:42|$TIMESTAMP:43|$TIMESTAMP:44|$TIMESTAMP:45|$TIMESTAMP:46|$TIMESTAMP1:18|$TIMESTAMP1:19|$TIMESTAMP1:20|$TIMESTAMP1:21|$TIMESTAMP1:22|$TIMESTAMP1:41|$TIMESTAMP1:42|$TIMESTAMP1:43|$TIMESTAMP1:44|$TIMESTAMP1:45|$TIMESTAMP1:46" | wc -l)

case "${ADXMONITOR_CHECK}" in
        0)  echo "ADXMONITOR NOT OK on www6002 - RUN \"sh /var/www/vhosts/mdotm.com/scripts/systems/kill_adxmonitor.sh\" on www6002"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "ADXMONITOR is OK - log file \"/var/log/report.log\" last updated: $LOG_FILE_LAST_UPDATED"; exit ${STATE_OK}
        ;;
#        *)  echo "ADXMONITOR is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
