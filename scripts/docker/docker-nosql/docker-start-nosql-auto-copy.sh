#!/bin/bash

unset newimgnum imgnum imgname
unset port newport rpcport newrpcport newsyslogport

port=$(netstat -apn | grep docker-proxy | grep :31 | grep LISTEN | sort -k4 | tail -n1 | cut -d ":" -f4)
#echo "$port"
newport=$(($port + 1))
#echo "$newport"

rpcport=$(netstat -apn | grep docker-proxy | grep :34 | grep LISTEN | sort -k4 | tail -n1 | cut -d ":" -f4)
newrpcport=$(($rpcport + 1))
#echo "$newrpcport"

#raftport=$(netstat -apn | grep docker-proxy | grep :40 | grep LISTEN | tail -n1 | cut -d ":" -f4)
##echo "$port"
#newraftport=$(($raftport + 1))

imgname=$(docker ps -l | grep nosql | awk '{print$NF}')
#imgname=$(docker ps -l | grep nosql | awk '{print$2}' | cut -d "/" -f2)
#echo $imgname

imgnum=$(grep -Eo '[[:alpha:]]+|[0-9]+' <<<"$imgname" | tail -n1)
#echo "$imgnum"
newimgnum=$(($imgnum + 1))
#echo "$newimgnum"

syslogport=$(netstat -apn | grep docker-proxy | grep :51 | grep LISTEN | sort -k4 | tail -n1 | cut -d ":" -f4)
newsyslogport=$(($syslogport + 1))

# deploy docker first time with default ports
if ! docker ps -l | grep nosql &> /dev/null; then
docker run --name=nosql -dit -p 34000:34000 -p 31000:31000 wolkinc/nosql /root/nosql/qdata/dd 0xef741ae5e8f66bb8 34000 31000 32003 cloudflare.wolk.com 80
sleep 1
else
docker run --name=nosql$newimgnum -dit -p $newrpcport:34000 -p $newport:31000  wolkinc/nosql /root/nosql/qdata/dd $1 34000 31000 32003 cloudflare.wolk.com 80
#sleep 5
fi

unset newimgnum imgnum imgname; unset port newport rpcport newrpcport newsyslogport
