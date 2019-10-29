#!/bin/bash

for i in {1..18};
do
if ! curl "http://127.0.0.1/ads/systems/bthealthcheck.php" | grep "Hi there:" &> /dev/null; then
        echo "`date +%T` -- Starting crosschannel"
	cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh crosschannel
#    else
#    echo "`date +%T` - crosschannel is running"
fi
sleep 3;
done
