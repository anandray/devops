#!/bin/bash

if ! docker images | grep wolkinc | grep -v grep &> /dev/null; then
sudo su - << EOF
docker pull wolkinc/go-wolk-geth
EOF
else
echo "
`date +%m-%d-%T` - Docker image pull successful...
"
fi

if ! docker ps | grep wolkinc | grep -v grep &> /dev/null; then
docker run --dns=8.8.8.8 --dns=8.8.4.4 --name=go-wolk-geth --rm -dit -p 30303:30303  -p 30303:30303/udp -p 30304:30304/udp wolkinc/go-wolk-geth
else
echo "`date +%m-%d-%T` - Docker image run successful...
"
#sed -i 's/\* \* \* \* \*/\#\* \* \* \* \*/g' /var/spool/cron/root
fi
