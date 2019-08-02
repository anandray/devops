#!/bin/bash

for i in {1..12};
do
if ! curl -s  "http://127.0.0.1/health" | grep HEALTHCHECK; then
   echo "`date +'%m%d%Y %T'` - roam NOT running- restarting..."
   cd /var/www/vhosts/api.colorfulnotion.com/scripts && sh goservice.sh roam &
#   /var/www/vhosts/api.colorfulnotion.com/roam/roam &
else
  echo "`date +'%m%d%Y %T'` - roam is running"
fi
sleep 5;
done
