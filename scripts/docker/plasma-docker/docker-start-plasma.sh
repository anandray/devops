#!/bin/bash

port=$(netstat -apn | grep :30 | grep LISTEN | tail -n1 | cut -d ":" -f4)
#echo "$port"
newport=$(($port + 100))
#echo "$newport"

rpcport=$(netstat -apn | grep :85 | grep LISTEN | tail -n1 | cut -d ":" -f4)
newrpcport=$(($rpcport + 1))
#echo "$newrpcport"

udpport=$(netstat -apn | grep :30 | grep udp | tail -n1 | cut -d ":" -f4)
newudpport=$(($udpport + 100))
#echo "$newudpport"

imgname=$(docker ps -l | grep plasma | awk '{print$2}' | cut -d "/" -f2)
#echo $imgname

imgnum=$(grep -Eo '[[:alpha:]]+|[0-9]+' <<<"$imgname" | tail -n1)
#echo "$imgnum"
newimgnum=$(($imgnum + 1))
#echo "$newimgnum"

#docker build -t wolkinc/plasma$newimgnum . && docker run --name=plasma$newimgnum --rm -it --dns=8.8.8.8 --dns=8.8.4.4 -p 8547:8545 -p 30503:30303  -p 30503:30303/udp -p 30504:30304/udp wolkinc/plasma$newimgnum

if ! docker ps -q &> /dev/null; then
# deploy docker with default port
docker build -t wolkinc/plasma . && docker run --name=plasma --rm -it --dns=8.8.8.8 --dns=8.8.4.4 -p 8545:8545 -p 30303:30303  -p 30303:30303/udp -p 30304:30304/udp wolkinc/plasma

else
docker build -t wolkinc/plasma$newimgnum . && docker run --name=plasma$newimgnum --rm -dit --dns=8.8.8.8 --dns=8.8.4.4 -p $newrpcport:8545 -p $newport:30303  -p $newport:30303/udp -p $newudpport:30304/udp wolkinc/plasma$newimgnum
fi
