#!/bin/bash

for i in {1..18};
do
if ! php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/bthealthcheck.php | grep OK; then
cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php bt;
else
  echo bt is running
fi
sleep 3;
done
