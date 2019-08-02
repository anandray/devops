#!/bin/bash

enode=`docker exec $(docker ps -q) grep 'enode://' /var/www/vhosts/data/geth.log | head -n1 | awk '{print$NF}' | cut -d "=" -f2 | cut -d "@" -f1`
ip=`gcloud compute instances list | grep $(hostname) | awk '{print"@"$5":30303"}'`

echo "\"$enode$ip\","
