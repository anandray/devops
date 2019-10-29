#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
OFFICE_IP=50.225.47.189
DENYHOSTS_OFFICE_IP_CHECK=`grep 50.225.47.189 /etc/hosts.deny | wc -l`
 
case "${DENYHOSTS_OFFICE_IP_CHECK}" in
        1)  echo "DENYHOSTS Blocked OFFICE IP - on `hostname` remove $OFFICE_IP from /etc/hosts.deny and restart denyhosts"; exit ${STATE_CRITICAL}
        ;;
        0)  echo "DENYHOSTS `hostname` OFFICE IP $OFFICE_IP NOT Blocked - OK"; exit ${STATE_OK}
        ;;
#        *)  echo "DENYHOSTS is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
