#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
#ROAM_CHECK1=`curl -sL -w "%{http_code} %{url_effective}\\n" "http://127.0.0.1/"`
#ROAM_CHECK=`curl -sL -w "%{http_code}\\n" "http://127.0.0.1/" | grep 200 | wc -l`
#ROAM_CHECK1=`curl -sL -w "\\n%{url_effective}\\n" "http://127.0.0.1/health"`
ROAM_CHECK1=`curl -s "http://127.0.0.1/health"`
ROAM_CHECK=`curl -sL "http://127.0.0.1/health" | grep HEALTHCHECK | wc -l`
 
case "${ROAM_CHECK}" in
        0)  echo "Roam is not running - stop httpd - $ROAM_CHECK1 "; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Roam is running - $ROAM_CHECK1 - OK"; exit ${STATE_OK}
        ;;
        *)  echo "Roam is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
