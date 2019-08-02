#!/bin/bash

if ps aux | grep "plasma --bootnode" | grep -v grep &> /dev/null; then
kill -9 $(ps aux | grep "plasma --bootnode" | grep -v grep | awk '{print$2}')
fi

if ! md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | grep 6aa4aa234b0643fc61fa67e5213bda82 &> /dev/null; then
wget -O /root/go/src/github.com/wolkdb/plasma/build/bin/plasma http://www6001.wolk.com/.start/plasma
chmod +x /root/go/src/github.com/wolkdb/plasma/build/bin/plasma
fi

if ! ps aux | grep "plasma --bootnode" | grep -v grep &> /dev/null; then
/root/scripts/plasma-start.sh &
fi
