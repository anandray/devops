#!/bin/bash
# golang version 1.12.2 install
if go version | grep go1.10; then
'yes' | mv -f /usr/local/go /usr/local/`go version | awk '{print$3}'`
cd /usr/local
wget https://dl.google.com/go/go1.12.1.linux-amd64.tar.gz
tar zxvpf go1.12.1.linux-amd64.tar.gz
ln -s /usr/local/go/bin/go /usr/bin/go
  else
  cd /usr/local 
  wget https://dl.google.com/go/go1.12.1.linux-amd64.tar.gz 
  tar zxvpf go1.12.1.linux-amd64.tar.gz
  ln -s /usr/local/go/bin/go /usr/bin/go
fi
