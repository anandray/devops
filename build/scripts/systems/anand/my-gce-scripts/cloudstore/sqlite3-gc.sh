#!/bin/bash
sudo gsutil -m cp gs://startup_scripts_us/scripts/cloudstore/sqlite3/libsqlite3* /usr/local/lib/
sudo gsutil cp gs://startup_scripts_us/scripts/cloudstore/sqlite3/sqlite3.conf /etc/ld.so.conf.d/
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0
sudo ldconfig
