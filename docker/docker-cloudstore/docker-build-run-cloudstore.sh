#!/bin/bash

cd /var/www/vhosts/docker/docker/docker-cloudstore
docker build -t cloudstore .
docker run -it -p 8546:8546 cloudstore /root/cloudstore/dd 8546
