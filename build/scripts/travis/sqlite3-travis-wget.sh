#!/bin/bash

#sqlite3
sudo wget -O /usr/local/lib/libsqlite3.la http://d5.wolk.com/.start/sqlite3/libsqlite3.la &>/dev/null 2>&1 &
sudo wget -O /usr/local/lib/libsqlite3.so.0.8.6 http://d5.wolk.com/.start/sqlite3/libsqlite3.so.0.8.6 &>/dev/null 2>&1 &
sudo wget -O /usr/local/lib/libsqlite3.a http://d5.wolk.com/.start/sqlite3/libsqlite3.a &>/dev/null 2>&1 &
sudo wget -O /etc/ld.so.conf.d/sqlite3.conf http://d5.wolk.com/.start/sqlite3/sqlite3.conf &>/dev/null 2>&1 &
sudo wget -O /usr/local/bin/sqlite3 http://d5.wolk.com/.start/sqlite3/sqlite3 &>/dev/null 2>&1 &
sudo chmod +x /usr/local/bin/sqlite3 &>/dev/null 2>&1 &
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so &>/dev/null 2>&1 &
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0 &>/dev/null 2>&1 &
sudo ldconfig &>/dev/null 2>&1 &
