#!/bin/bash

echo "
docker 0
"
docker run --name=nosql -dit -p 34000:34000 -p 31000:31000 wolkinc/nosql /root/nosql/qdata/dd `sed "1q;d" blockchainids` 34000 31000 32003 cloudflare.wolk.com 80
sleep 4
echo "
docker 1"
/root/scripts/docker-start-nosql-auto.sh `sed "2q;d" blockchainids`
sleep 1
echo "
docker 2"
/root/scripts/docker-start-nosql-auto.sh `sed "3q;d" blockchainids`
sleep 1
echo "
docker 3"
/root/scripts/docker-start-nosql-auto.sh `sed "4q;d" blockchainids` 
sleep 1
echo "
docker 4"
/root/scripts/docker-start-nosql-auto.sh `sed "5q;d" blockchainids`
sleep 1
echo "
docker 5"
/root/scripts/docker-start-nosql-auto.sh `sed "6q;d" blockchainids` 
sleep 1
echo "
docker 6"
/root/scripts/docker-start-nosql-auto.sh `sed "7q;d" blockchainids` 
sleep 1
echo "
docker 7"
/root/scripts/docker-start-nosql-auto.sh `sed "8q;d" blockchainids`  
sleep 1
echo "
docker 8"
/root/scripts/docker-start-nosql-auto.sh `sed "9q;d" blockchainids` 
sleep 1
echo "
docker 9"
/root/scripts/docker-start-nosql-auto.sh `sed "10q;d" blockchainids` 
sleep 1
echo "
docker 10"
/root/scripts/docker-start-nosql-auto.sh `sed "11q;d" blockchainids`
