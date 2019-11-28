#!/bin/bash

echo "
docker 0
"
docker run --name=nosql -dit -p 34000:34000 -p 31000:31000 wolkinc/nosql /root/nosql/qdata/dd 0xa7bd17935af83b07 34000 31000 32003 cloudflare.wolk.com 80
sleep 4
echo "
docker 1"
/root/scripts/docker-start-nosql-auto.sh 0xb32aaab38d039bfe
sleep 1
echo "
docker 2"
/root/scripts/docker-start-nosql-auto.sh 0x5ae767a8ecc75f9a
sleep 1
echo "
docker 3"
/root/scripts/docker-start-nosql-auto.sh 0x3a4b97a79edc3f29 
sleep 1
echo "
docker 4"
/root/scripts/docker-start-nosql-auto.sh 0x85e7070a5dc6722d  
sleep 1
echo "
docker 5"
/root/scripts/docker-start-nosql-auto.sh 0x4652043ee8860190 
sleep 1
echo "
docker 6"
/root/scripts/docker-start-nosql-auto.sh 0x0526599f15a160a9 
sleep 1
echo "
docker 7"
/root/scripts/docker-start-nosql-auto.sh 0xbf4c0137bd16a155  
sleep 1
echo "
docker 8"
/root/scripts/docker-start-nosql-auto.sh 0x770c6d442d4d00f1 
sleep 1
echo "
docker 9"
/root/scripts/docker-start-nosql-auto.sh 0x824706bce224b198 
sleep 1
echo "
docker 10"
/root/scripts/docker-start-nosql-auto.sh 0x7a9bce98aa73c14c
