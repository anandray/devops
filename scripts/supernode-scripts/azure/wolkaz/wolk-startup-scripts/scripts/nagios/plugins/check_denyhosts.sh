#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
IP=$(ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}')
DENYHOSTS_CHECK1=`grep $(ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}') /etc/hosts.deny`
DENYHOSTS_CHECK=`grep $(ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}') /etc/hosts.deny | wc -l`
 
case "${DENYHOSTS_CHECK}" in
        1)  echo "DENYHOSTS Blocked `hostname` pvt IP - remove $IP from /etc/hosts.deny and restart denyhosts"; exit ${STATE_CRITICAL}
        ;;
        0)  echo "DENYHOSTS `hostname` pvt IP NOT Blocked - OK"; exit ${STATE_OK}
        ;;
#        *)  echo "DENYHOSTS is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
