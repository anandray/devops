#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#INSTALLED=`/usr/bin/pecl remote-info geoip | grep Installed`
INSTALLED=`php -i | grep 'geoip extension version' | awk '{print$5}'`
LATEST=`/usr/bin/pecl remote-info geoip | grep Latest | awk '{print$2}'`

GEOIP_VERSION_CHECK1=`/usr/bin/pecl remote-info geoip | grep Latest | grep 1.1.1 > /tmp/geoip_version`
GEOIP_VERSION_CHECK=`cat /tmp/geoip_version | grep Latest | grep 1.1.1 | wc -l`

case "${GEOIP_VERSION_CHECK}" in
        0)  echo "GEOIP_VERSION $LATEST > $INSTALLED"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "GEOIP_VERSION $INSTALLED = $LATEST"; exit ${STATE_OK}
        ;;
#        *)  echo "GEOIP_VERSION is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
