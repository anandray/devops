#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

a=`ps aux | grep -c httpd`
b=400

MAXCLIENTS_CHECK=`sh /var/www/vhosts/mdotm.com/httpdocs/ads/systems/maxclients_check.sh | grep LESS | wc -l`

case "${MAXCLIENTS_CHECK}" in
        0)  echo "httpd connections $a > $b"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "MAXCLIENTS is OK - $a httpd connections"; exit ${STATE_OK}
        ;;
#        *)  echo "MAXCLIENTS is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
