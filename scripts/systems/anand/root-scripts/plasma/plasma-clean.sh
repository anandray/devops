#!/bin/bash
if [ -d /root/data ]; then
kill -9 $(ps aux | grep 'plasma --bootnode' | grep -v grep | awk '{print$2}')
unalias rm
rm -rf /root/data && mkdir /root/data
rm -rf /tmp/plasmachain
#/root/scripts/plasma-start.sh
fi
