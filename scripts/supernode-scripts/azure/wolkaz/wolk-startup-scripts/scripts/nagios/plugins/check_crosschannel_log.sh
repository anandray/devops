#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#LOG_FILE_SIZE=`du -sh /var/log/crosschannel.log | awk '{print$1}' | cut -d "M" -f1` #MB
LOG_FILE_SIZE=`du -s /var/log/crosschannel.log | awk '{print$1}'` #KB

LOG_FILE_CHECK=$(sh /var/www/vhosts/mdotm.com/scripts/systems/crosschannel_log_file_size.sh | grep 'is larger than 102400000' | wc -l)

case "${LOG_FILE_CHECK}" in
        1)  echo "CROSSCHANNEL_LOG greater than 10 GB - run \"sh /etc/cron.hourly/crosschannel\""; exit ${STATE_CRITICAL}
        ;;
        0)  echo "CROSSCHANNEL_LOG is OK - log file $LOG_FILE_SIZE"; exit ${STATE_OK}
        ;;
#        *)  echo "CROSSCHANNEL_LOG is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
