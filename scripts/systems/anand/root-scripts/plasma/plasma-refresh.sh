#!/bin/bash

if ps aux | grep "plasma --bootnode" | grep -v grep &> /dev/null; then
kill -9 $(ps aux | grep "plasma --bootnode" | grep -v grep | awk '{print$2}');
sed -i 's/nohup/#nohup/g' /root/scripts/plasma-start.sh;
fi

MD5=`ssh -q 35.193.142.191 md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | awk '{print$1}'`

if ! md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | grep $MD5 &> /dev/null; then
scp -C -p 35.193.142.191:/root/go/src/github.com/wolkdb/plasma/build/bin/plasma /root/go/src/github.com/wolkdb/plasma/build/bin/ &> /dev/null
chmod +x /root/go/src/github.com/wolkdb/plasma/build/bin/plasma
sed -i 's/#nohup/nohup/g' /root/scripts/plasma-start.sh;
fi
