#!/bin/bash

if [ ! -d /root/sql/qdata/dd ]; then
sudo mkdir -p /root/sql/qdata/dd
fi

echo "{
  \"blockChainId\": \"0x3b6a2ac8b193b705\"
}" > /root/sql/qdata/dd/genesis.json

if [ ! -d /root/sql/bin ]; then
sudo mkdir -p /root/sql/bin
fi

if [ ! -f /root/sql/bin/sql ]; then
echo "
Downloading the sql binary...
"
sudo wget -O /root/sql/bin/sql www6001.wolk.com/.start/sql
sudo chmod +x /root/sql/bin/sql
fi

# enable sql service
sudo systemctl daemon-reload
sudo chkconfig sql on

MD5=`sudo ssh -q cloudstore.wolk.com md5sum /root/sql/bin/sql | awk '{print$1}'`

if ! ps aux | grep sql | grep -v grep; then
sudo wget -O /root/sql/bin/sql http://www6001.wolk.com/.start/sql &&
sudo chmod +x /root/sql/bin/sql &&
sudo service sql start;
fi

if ! sudo md5sum /root/sql/bin/sql | grep $MD5 &> /dev/null; then
sudo systemctl stop sql.service;
sudo kill -9 $(ps aux | grep sql | grep attach | grep -v grep | awk '{print$2}');
sudo wget -O /root/sql/bin/sql http://www6001.wolk.com/.start/sql &&
sudo chmod +x /root/sql/bin/sql &&
sudo systemctl start sql.service
else
sudo systemctl start sql.service
fi
