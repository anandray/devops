#!/bin/bash

if [ ! -d /root/plasma/qdata/dd ]; then
sudo mkdir -p /root/plasma/qdata/dd
fi

if [ ! -d /root/plasma/bin ]; then
sudo mkdir -p /root/plasma/bin
fi

if [ ! -f /root/plasma/bin/plasma ]; then
echo "
Downloading the plasma binary...
"
sudo wget -O /root/plasma/bin/plasma www6001.wolk.com/.start/plasma
sudo chmod +x /root/plasma/bin/plasma
fi

# enable plasma service
sudo systemctl daemon-reload
sudo chkconfig plasma on

MD5=`sudo ssh -q cloudstore.wolk.com md5sum /root/plasma/bin/plasma | awk '{print$1}'`

if ! ps aux | grep plasma | grep -v grep; then
sudo wget -O /root/plasma/bin/plasma http://www6001.wolk.com/.start/plasma &&
sudo chmod +x /root/plasma/bin/plasma &&
sudo service plasma start;
fi

if ! sudo md5sum /root/plasma/bin/plasma | grep $MD5 &> /dev/null; then
sudo systemctl stop plasma.service;
sudo kill -9 $(ps aux | grep plasma | grep attach | grep -v grep | awk '{print$2}');
sudo wget -O /root/plasma/bin/plasma http://www6001.wolk.com/.start/plasma &&
sudo chmod +x /root/plasma/bin/plasma &&
sudo systemctl start plasma.service
else
sudo systemctl start plasma.service
fi
