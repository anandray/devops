#!/bin/bash

for i in {1..12};
do

cd /go/src/github.com/algorand/go-algorand
LOCAL=$(git rev-parse @{0})
REMOTE=$(git rev-parse @{u})
BASE=$(git merge-base @{0} @{u})
git fetch

if [ $LOCAL = $REMOTE ]; then
 echo "
 `date +%m/%d/%Y-%T` - Already up to date... 
 "
else
 echo "
 `date +%m/%d/%Y-%T` - Updating go-algorand repository
 "
 git pull
 scp d0.wolk.com:/root/wolkdev/Primary/genesis.json /root/wolkdev/ && /root/go/bin/algod -d /root/wolkdev -p 10.138.0.29:323 -l :8080 &
 echo "
 `date +%m/%d/%Y-%T` - Updated repository... 
 "
fi
sleep 5
done
