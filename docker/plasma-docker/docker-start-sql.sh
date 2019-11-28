#!/bin/bash

port=$(netstat -apn | grep sql | grep :21 | grep LISTEN | tail -n1 | cut -d ":" -f4)
#echo "$port"
newport=$(($port + 100))
#echo "$newport"

rpcport=$(netstat -apn | grep sql | grep :220 | grep LISTEN | tail -n1 | cut -d ":" -f4)
newrpcport=$(($rpcport + 1))
#echo "$newrpcport"

raftport=$(netstat -apn | grep sql | grep :50 | grep LISTEN | tail -n1 | cut -d ":" -f4)
#echo "$port"
newraftport=$(($port + 100))

imgname=$(docker ps -l | grep sql | awk '{print$2}' | cut -d "/" -f2)
#echo $imgname

imgnum=$(grep -Eo '[[:alpha:]]+|[0-9]+' <<<"$imgname" | tail -n1)
#echo "$imgnum"
newimgnum=$(($imgnum + 1))
#echo "$newimgnum"

if ! docker ps -q &> /dev/null; then
# deploy docker with default port
docker build -t wolkinc/sql . && docker run --name=sql --rm -it --dns=8.8.8.8 --dns=8.8.4.4 -p 22003:22003 -p 50404:50404  -p 21003:21003 -p 21003:21003/udp -p 50404:50404/udp wolkinc/sql

else
docker build -t wolkinc/sql$newimgnum . && docker run --name=sql$newimgnum --rm -dit --dns=8.8.8.8 --dns=8.8.4.4 -p $newrpcport:22003 -p $newraftport:50404 -p $newport:21003  -p $newport:21003/udp -p $newraftport:50404/udp wolkinc/sql$newimgnum
fi
