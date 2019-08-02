#!/bin/bash

wget -O /tmp/SQLite-3.22.0.tar.gz https://www.sqlite.org/src/tarball/0c55d179/SQLite-0c55d179.tar.gz &&
tar zxvpf /tmp/SQLite-3.22.0.tar.gz -C /tmp/ &&
mv -f /tmp/SQLite-0c55d179 /tmp/SQLite-3.22.0 &&
mkdir /usr/local/sqlite3 &&
cd /usr/local/sqlite3 &&
/tmp/SQLite-3.22.0/configure &&
make && make sqlite3.c  
make install
echo "/usr/local/lib" > /etc/ld.so.conf.d/sqlite3.conf && ldconfig -v
