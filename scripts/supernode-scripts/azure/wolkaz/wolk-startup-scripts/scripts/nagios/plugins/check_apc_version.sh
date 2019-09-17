#!/bin/bash
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3

#INSTALLED=`/usr/bin/pecl remote-info apc | grep Installed`
INSTALLED=`php -i | grep -a1 'APC Support' | grep Version | awk '{print$3}'`
LATEST=`/usr/bin/pecl remote-info apc | grep Latest | awk '{print$2}'`

#if (( $(echo "$result1 > $result2" | bc -l) ));

#if [ $LATEST > $INSTALLED | bc -l]; then
#if [ "$(echo $LATEST '>' $INSTALLED | bc -l)" -eq 1 ]; then
#if [ (( $(echo "$LATEST > $INSTALLED" | bc -l) )) -eq 1 ]; then
#	echo "APC_VERSION is $LATEST"; exit ${STATE_OK}
#if [ $LATEST -gt $INSTALLED ]; then
#else
#	echo "APC_VERSION $LATEST > $INSTALLED"; exit ${STATE_CRITICAL}
#fi

#if [ "$count" = "0" ]; then # no matches, exit with no error
#    $ECHO "Log check ok - 0 pattern matches found\n"
#    exitstatus=$STATE_OK
#else # Print total matche count and the last entry we found
#    $ECHO "($count) $lastentry"
#    exitstatus=$STATE_CRITICAL
#fi

#exit $exitstatus

APC_VERSION_CHECK1=`/usr/bin/pecl remote-info apc | grep Latest | grep 3.1.13 > /tmp/apc_version`
APC_VERSION_CHECK=`cat /tmp/apc_version | grep Latest | grep 3.1.13 | wc -l`

case "${APC_VERSION_CHECK}" in
        0)  echo "APC_VERSION $LATEST > $INSTALLED"; exit ${STATE_CRITICAL}
        ;;
        1)  echo "APC_VERSION $INSTALLED = $LATEST"; exit ${STATE_OK}
        ;;
#        *)  echo "APC_VERSION is in an unknown state."; exit ${STATE_WARNING}
#        ;;
esac
