#!/bin/bash

npm=`/usr/bin/npm -v` &> /dev/null
node=`/usr/bin/node -v` &> /dev/null

if [ -f /usr/bin/npm ]; then
echo "npm is installed - version --> "$npm
else
echo "npm NOT installed.. Installing..."
mkdir -p /root/scripts /root/downloads;
gsutil cp gs://startup_scripts_us/scripts/node-v8.1.4-linux-x64.tar.xz /root/downloads;
cd /root/downloads;
tar xvpf node-v8.1.4-linux-x64.tar.xz;
ln -s /root/downloads/node-v8.1.4-linux-x64/lib/node_modules/npm/bin/npm-cli.js  /usr/bin/npm
fi

if [ -f /usr/bin/node ]; then
echo "node is installed - version --> "$node
else
echo "node NOT installed.. Installing..."
cp -rfv /root/downloads/node-v8.1.4-linux-x64/bin/node /usr/bin/
fi
#/usr/bin/npm -v
#/usr/bin/node -v
