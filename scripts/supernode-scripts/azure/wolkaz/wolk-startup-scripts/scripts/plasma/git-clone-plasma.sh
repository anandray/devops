#!/bin/bash

if [ ! -d /root/go/src/github.com/wolkdb/plasma ]; then
        mkdir -p /root/go/src/github.com/wolkdb
	cd /root/go/src/github.com/wolkdb
        git clone --recurse-submodules git@github.com:wolkdb/plasma.git
        cd /root/go/src/github.com/wolkdb/plasma
        git config --global user.name "Sourabh Niyogi"
        git config --global user.email "sourabh@wolk.com"
        git config core.filemode true
        git config --global core.filemode true
        echo "export PATH=$PATH:/root/go/src/github.com/wolkdb/plasma/build/bin" >> /root/.bashrc
        source /root/.bashrc
fi
