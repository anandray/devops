#!/bin/bash

# golang version 1.12.2 install
if go version | grep go1.12.1; then
echo "go1.12.1 is already installed..."
  else
  'yes' | mv -f /usr/local/go /usr/local/`go version | awk '{print$3}'`
  cd /usr/local 
  wget https://dl.google.com/go/go1.12.1.linux-amd64.tar.gz 
  tar zxvpf go1.12.1.linux-amd64.tar.gz
  ln -s /usr/local/go/bin/go /usr/bin/go 
fi

export GOPATH=/root/go
export GOROOT=/usr/local/go

if [ -d /usr/local/wolk ]; then
        sudo /sbin/service wolk stop;
        sudo /usr/bin/pkill -9 wolk;
        sudo rm -rf /usr/local/wolk/*;
fi

# creating wolk0,wolk1-wolk5
if [ -f /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk ]; then
for i in {1..5};
do
'yes' |  cp -rfv /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk$i;
sudo scp c0.wolk.com:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json
#gsutil cp gs://wolk-scripts/scripts/cloudstore/c0-genesis.json /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json;
done
 service wolk restart
  else
    cd /root/go/src/github.com/wolkdb/cloudstore
    make wolk
     service wolk restart
     for i in {1..5};
       do
       'yes' |  cp -rfv /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk$i;
       sudo scp c0.wolk.com:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json
       #gsutil cp gs://wolk-scripts/scripts/cloudstore/c0-genesis.json /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json;
     done
     service wolk restart
fi
