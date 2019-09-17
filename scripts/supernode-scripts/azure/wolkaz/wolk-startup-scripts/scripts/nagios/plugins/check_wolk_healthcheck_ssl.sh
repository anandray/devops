#!/bin/bash

STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

HEALTHCHECK=`curl -k -s https://localhost/healthcheck | grep OK | wc -l`
HEALTHCHECK1=`curl -k -s https://localhost/healthcheck`

case "${HEALTHCHECK}" in
        0)  echo "HEALTHCHECK Failed - $HEALTHCHECK1"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "HEALTHY - $HEALTHCHECK1"; exit ${STATE_OK}
        ;;
        *)  echo "Unknown state."; exit ${STATE_UNKNOWN}
       ;;
esac
