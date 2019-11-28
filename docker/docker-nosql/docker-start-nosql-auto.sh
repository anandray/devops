#!/bin/bash

"yes" | cp -rf /root/go/src/github.com/wolkdb/plasma/build/bin/nosql /root/go/src/github.com/wolkdb/docker/nosql-docker/nosql/bin/
"yes" | cp -rf /root/go/src/github.com/wolkdb/go-ethereum/build/bin/nosqlnode /root/go/src/github.com/wolkdb/docker/nosql-docker/nosql/bin/

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

imgname=$(docker ps -l | grep nosql | awk '{print$2}' | cut -d "/" -f2)
#echo $imgname

imgnum=$(grep -Eo '[[:alpha:]]+|[0-9]+' <<<"$imgname" | tail -n1)
#echo "$imgnum"
newimgnum=$(($imgnum + 1))
#echo "$newimgnum"

#if docker ps -a -l | grep nosql &> /dev/null; then
##blockchainid=$(docker exec $(docker ps -a -q -l) cat /root/nosql/qdata/dd/genesis.json | grep blockChainId | awk '{print$2}')
#blockchainid=$(docker ps -a -l --no-trunc | grep nosql | awk '{print$5}' | cut -d "\"" -f1)
#newblockchainid=$(($blockchainid + 1))
#fi

# deploy docker first time with default ports
if ! docker ps -l | grep nosql &> /dev/null; then
docker build -t nosql . && docker run --name=nosql -dit --rm -p 5001:5000 -p 34000:34000 -p 31000:31000 nosql /root/nosql/qdata/dd 279

else
blockchainid=$(docker ps -a -l --no-trunc | grep nosql | awk '{print$5}' | cut -d "\"" -f1)
newblockchainid=$(($blockchainid + 1))
docker build -t nosql$newimgnum . && docker run --name=nosql$newimgnum -dit -p $newrpcport:34000 -p $newport:31000  nosql$newimgnum /root/nosql/qdata/dd $newblockchainid
fi
