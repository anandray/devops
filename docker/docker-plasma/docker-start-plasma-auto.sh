#!/bin/bash

"yes" | cp -rf /root/go/src/github.com/wolkdb/plasma/build/bin/plasma /root/go/src/github.com/wolkdb/docker/plasma-docker/plasma/bin/
"yes" | cp -rf /root/go/src/github.com/wolkdb/go-ethereum/build/bin/plasmanode /root/go/src/github.com/wolkdb/docker/plasma-docker/plasma/bin/

port=$(netstat -apn | grep docker-proxy | grep :31 | grep LISTEN | sort -k4 | tail -n1 | cut -d ":" -f4)
#echo "$port"
newport=$(($port + 1))
#echo "$newport"

rpcport=$(netstat -apn | grep docker-proxy | grep :32 | grep LISTEN | sort -k4 | tail -n1 | cut -d ":" -f4)
newrpcport=$(($rpcport + 1))
#echo "$newrpcport"

#raftport=$(netstat -apn | grep docker-proxy | grep :40 | grep LISTEN | tail -n1 | cut -d ":" -f4)
##echo "$port"
#newraftport=$(($raftport + 1))

imgname=$(docker ps -l | grep plasma | awk '{print$2}' | cut -d "/" -f2)
#echo $imgname

imgnum=$(grep -Eo '[[:alpha:]]+|[0-9]+' <<<"$imgname" | tail -n1)
#echo "$imgnum"
newimgnum=$(($imgnum + 1))
#echo "$newimgnum"

#if docker ps -a -l | grep plasma &> /dev/null; then
##blockchainid=$(docker exec $(docker ps -a -q -l) cat /root/plasma/qdata/dd/genesis.json | grep blockChainId | awk '{print$2}')
#blockchainid=$(docker ps -a -l --no-trunc | grep plasma | awk '{print$5}' | cut -d "\"" -f1)
#newblockchainid=$(($blockchainid + 1))
#fi

# deploy docker first time with default ports
if ! docker ps -l | grep plasma &> /dev/null; then
docker build -t plasma . && docker run --name=plasma -dit --rm -p 32003:32003 -p 31003:31003 plasma /root/plasma/qdata/dd

else
blockchainid=$(docker ps -a -l --no-trunc | grep plasma | awk '{print$5}' | cut -d "\"" -f1)
newblockchainid=$(($blockchainid + 1))
docker build -t plasma$newimgnum . && docker run --name=plasma$newimgnum -dit -p $newrpcport:32003 -p $newport:31003  plasma$newimgnum /root/plasma/qdata/dd $newblockchainid
fi
