#!/bin/bash
#if ! netstat -apn | grep -E ':80|crosschannel' | grep LISTEN; then
if ! curl -s "http://127.0.0.1/ads/systems/bthealthcheck.php" | grep "Hi there:" > /dev/null; then
        echo "`date +%T` -- Starting crosschannel"
	cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh crosschannel
    else
    echo "`date +%T` - crosschannel is running"
fi		

#if ! curl -s "http://127.0.0.1:80/ads/systems/bthealthcheck.php" | grep 'Hi there'; then
#        cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh crosschannel 2> /var/log/git-crosschannel.err  > /var/log/git-crosschannel.log;
#        else
#    echo "crosschannel is running"
#fi
