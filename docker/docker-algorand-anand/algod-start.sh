#!/bin/bash
  
rm -rfv /root/wolkdev && \
mkdir -p /root/wolkdev && \
gsutil cp gs://wolk-scripts/scripts/algorand/genesis.json /root/wolkdev/ && \
gsutil cp gs://wolk-scripts/scripts/algorand/algod /root/go/bin/ && \
chmod +x /root/go/bin/algod && \
nohup /root/go/bin/algod -d /root/wolkdev -p 10.138.0.29:323 -l :8080 &> /root/wolkdev/node.log &
