#!/bin/bash

sudo apt-get update -q
sudo apt-get install -y tcl tcl-dev
sudo wget -O /tmp/SQLite-3.22.0.tar.gz https://www.sqlite.org/src/tarball/0c55d179/SQLite-0c55d179.tar.gz
sudo tar zxvpf /tmp/SQLite-3.22.0.tar.gz -C /tmp/
sudo mv -f /tmp/SQLite-0c55d179 /tmp/SQLite-3.22.0
sudo mkdir /usr/local/sqlite3
cd /usr/local/sqlite3
sudo /tmp/SQLite-3.22.0/configure
sudo make && sudo make sqlite3.c
sudo make install
echo "/usr/local/lib" > /tmp/sqlite3.conf
sudo cp -rf /tmp/sqlite3.conf /etc/ld.so.conf.d/sqlite3.conf
sudo ldconfig
cd ~/gopath/src/github.com/wolkdb/cloudstore
