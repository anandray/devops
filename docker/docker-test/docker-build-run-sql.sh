#!/bin/bash

#datadir = /root/sql/qdata/dd
#blockChainId = 199
cd /var/www/vhosts/docker/docker/docker-sql
docker build -t sql .
docker run -dit -p 22000:22000 -p 21000:21000 sql /root/sql/qdata/dd 199
#docker run -dit -p 50400:50400 -p 22000:22000 -p 21000:21000 sql /root/sql/qdata/dd 199
