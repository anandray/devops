#!/bin/bash

"yes" | cp -rf /root/go/src/github.com/wolkdb/plasma/build/bin/sql /root/go/src/github.com/wolkdb/docker/sql-docker/sql/bin/
"yes" | cp -rf /root/go/src/github.com/wolkdb/go-ethereum/build/bin/sqlnode /root/go/src/github.com/wolkdb/docker/sql-docker/sql/bin/

port=$(netstat -apn | grep docker-proxy | grep :21 | grep LISTEN | sort -k4 | tail -n1 | cut -d ":" -f4)
#echo "$port"
newport=$(($port + 1))
#echo "$newport"

rpcport=$(netstat -apn | grep docker-proxy | grep :24 | grep LISTEN | sort -k4 | tail -n1 | cut -d ":" -f4)
newrpcport=$(($rpcport + 1))
#echo "$newrpcport"

#raftport=$(netstat -apn | grep docker-proxy | grep :40 | grep LISTEN | tail -n1 | cut -d ":" -f4)
##echo "$port"
#newraftport=$(($raftport + 1))

imgname=$(docker ps -l | grep sql | awk '{print$2}' | cut -d "/" -f2)
#echo $imgname

imgnum=$(grep -Eo '[[:alpha:]]+|[0-9]+' <<<"$imgname" | tail -n1)
#echo "$imgnum"
newimgnum=$(($imgnum + 1))
#echo "$newimgnum"

#if docker ps -a -l | grep sql &> /dev/null; then
##blockchainid=$(docker exec $(docker ps -a -q -l) cat /root/sql/qdata/dd/genesis.json | grep blockChainId | awk '{print$2}')
#blockchainid=$(docker ps -a -l --no-trunc | grep sql | awk '{print$5}' | cut -d "\"" -f1)
#newblockchainid=$(($blockchainid + 1))
#fi

# deploy docker first time with default ports
if ! docker ps -l | grep sql &> /dev/null; then
docker build -t sql . && docker run --name=sql -dit --rm -p 24000:24000 -p 21000:21000 sql /root/sql/qdata/dd 0x3b6a2ac8b193b705

else
blockchainid=$(docker ps -a -l --no-trunc | grep sql | awk '{print$5}' | cut -d "\"" -f1)
newblockchainid=$(($blockchainid + 1))
docker build -t sql$newimgnum . && docker run --name=sql$newimgnum -dit -p $newrpcport:24000 -p $newport:21000  sql$newimgnum /root/sql/qdata/dd $newblockchainid
fi
