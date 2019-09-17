#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
DATAPROC_CHECK=`sh /var/www/vhosts/anand/mdotm.com/hadoop/systems/dataproc/check_dataproc.sh | grep "Dataproc Jobs are failing" | wc -l`
 
case "${DATAPROC_CHECK}" in
        0)  echo "Dataproc Jobs are OK."; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Dataproc Jobs are failing"; exit ${STATE_OK}
        ;;
#        *)  echo "BigTable is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
