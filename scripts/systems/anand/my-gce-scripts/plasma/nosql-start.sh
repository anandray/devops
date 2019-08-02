#!/bin/bash

if [ ! -d /root/nosql/qdata/dd ]; then
sudo mkdir -p /root/nosql/qdata/dd
fi

echo "{
   \"blockChainId\": 276
}" > /root/nosql/qdata/dd/genesis.json

if [ ! -d /root/nosql/bin ]; then
sudo mkdir -p /root/nosql/bin
fi

if [ ! -f /root/nosql/bin/nosql ]; then
echo "
Downloading the nosql binary...
"
sudo wget -O /root/nosql/bin/nosql www6001.wolk.com/.start/nosql
sudo chmod +x /root/nosql/bin/nosql
fi

MD5=`sudo ssh -q cloudstore.wolk.com md5sum /root/nosql/bin/nosql | awk '{print$1}'`

if ! ps aux | grep "nosql --datadir" | grep -v grep; then
sudo wget -O /root/nosql/bin/nosql http://www6001.wolk.com/.start/nosql &&
sudo chmod +x /root/nosql/bin/nosql &&
service nosql start;
fi

# enable nosql service
sudo systemctl daemon-reload
sudo chkconfig nosql on

if ! sudo md5sum /root/nosql/bin/nosql | grep $MD5 &> /dev/null; then
sudo systemctl stop nosql.service;
sudo kill -9 $(ps aux | grep nosql | grep attach | grep -v grep | awk '{print$2}');
sudo wget -O /root/nosql/bin/nosql http://www6001.wolk.com/.start/nosql && 
sudo chmod +x /root/nosql/bin/nosql && 
sudo systemctl start nosql.service;
else
sudo systemctl start nosql.service;
fi
