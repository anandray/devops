#!/bin/bash
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/libsqlite3.la /usr/local/lib/
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/libsqlite3.so.0.8.6 /usr/local/lib/
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/libsqlite3.a /usr/local/lib/
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/sqlite3.conf /etc/ld.so.conf.d/
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0
sudo ldconfig
