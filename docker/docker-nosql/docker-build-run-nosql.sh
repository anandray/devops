#!/bin/bash

#datadir = /root/nosql/qdata/dd
#blockChainId = 199
cd /var/www/vhosts/docker/docker/docker-nosql
docker build -t nosql .
docker run -dit -p 22000:22000 -p 21000:21000 nosql /root/nosql/qdata/dd 199
#docker run -dit -p 50400:50400 -p 22000:22000 -p 21000:21000 nosql /root/nosql/qdata/dd 199
