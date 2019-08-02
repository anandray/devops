#!/bin/bash
if ! php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/bthealthcheck.php | grep OK; then
  cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php bt;
else
  echo bt is running
fi
