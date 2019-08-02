#!/bin/bash

for goracing in {1..10}
do
if ! ps aux | grep racerver_master | grep -v grep; then
su - goracing << EOF
cd /home/goracing/scripts
./deployProduction.sh &
EOF
fi
sleep 5
done
