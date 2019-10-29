#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

BQ=`/usr/local/share/google/google-cloud-sdk/bin/bq ls | wc -l`
#BQ_AUTH_CHECK=`/usr/local/share/google/google-cloud-sdk/bin/bq ls bid | grep -c TABLE | grep 28 | wc -l`
GCLOUD_BQ_AUTH_CHECK=`/usr/local/share/google/google-cloud-sdk/bin/bq ls bid | grep -c pico00`

case "${GCLOUD_BQ_AUTH_CHECK}" in
        0)  echo "GCLOUD AUTH NOT OK - run \"bq ls bid\" and follow instructions"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "GCLOUD AUTH is OK - $BQ databases"; exit ${STATE_OK}
        ;;
#        *)  echo "GCLOUD is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
