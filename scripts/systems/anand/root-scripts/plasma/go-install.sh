#!/bin/bash

if ! go version &> /dev/null; then
echo "
GO is not installed.. Installing GoLang..
"
cd /usr/local && \
wget https://redirector.gvt1.com/edgedl/go/go1.9.2.linux-amd64.tar.gz && \
tar zxvpf go1.9.2.linux-amd64.tar.gz && \
ln -s /usr/local/go/bin/go /usr/bin/go && \
exec -l $SHELL && \
cd - && \
which go && \
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc && \
sourcebashrc && \
exec -l $SHELL && \
cd /root/go/src/github.com/wolkdb/plasma && \
go get cloud.google.com/go/bigtable && \
go get github.com/ethereum/go-ethereum/common && \
go get github.com/ethereum/go-ethereum/crypto && \
go version
else
echo "
[`go version`] is already installed...
"
fi
