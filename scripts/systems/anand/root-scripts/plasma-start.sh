#!/bin/bash

mkdir -p /root/data

nohup /root/go/src/github.com/wolkdb/plasma/build/bin/plasma \
--bootnodes enode://827192dd0616dc4f4ae9676b7cb2c56f8fdb478afb04a8f1bc74471806379cf27eaa5d6c1262005180cb516bb171060aaac5d287ddbaddde0ca498eb7fab1f3d@35.193.142.191:30303 \
--datadir /root/data \
--verbosity 4 \
--maxpeers  25 \
2>> /root/data/plasma.log &
