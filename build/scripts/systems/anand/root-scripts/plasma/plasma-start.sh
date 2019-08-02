#!/bin/bash

mkdir -p /root/data

nohup /root/go/src/github.com/wolkdb/plasma/build/bin/plasma \
--bootnodes enode://f6f6eb1b0dd01c86713d2055f1bd579e51b2afee5c8a13286a6662428dc7e57363c9d570f5f428ca1d8f3017da307ea5687a13c9021291483bd17a4ec8bc1829@35.193.142.191:30303 \
--datadir /root/data \
--verbosity 4 \
--maxpeers  25 \
2>> /root/data/plasma.log &
