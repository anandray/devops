#!/bin/bash
  
for i in {1..5};
do

if ! ps aux | grep "/root/go/bin/algod -d" | grep -v grep &> /dev/null; then
 echo "
 `date +%m/%d/%Y-%T` - algod not running... Starting...
 "
rm -rfv /root/wolkdev && \
mkdir -p /root/wolkdev && \
gsutil cp gs://wolk-scripts/scripts/algorand/genesis.json /root/wolkdev/ && \
gsutil cp gs://wolk-scripts/scripts/algorand/algod /root/go/bin/ && \
chmod +x /root/go/bin/algod && \
nohup /root/go/bin/algod -d /root/wolkdev -p 10.138.0.29:323 -l :8080 &> /root/wolkdev/node.log &

else
 echo "
 `date +%m/%d/%Y-%T` - ALGOD is already running...
 "
fi
sleep 11
done
