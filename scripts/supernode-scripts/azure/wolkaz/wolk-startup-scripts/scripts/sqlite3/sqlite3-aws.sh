#!/bin/bash
aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/sqlite3/libsqlite3.la
aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/sqlite3/libsqlite3.so.0.8.6
aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/sqlite3/libsqlite3.a /usr/local/lib/
aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/sqlite3/sqlite3.conf /etc/ld.so.conf.d/
ln -s 
ldconfig
