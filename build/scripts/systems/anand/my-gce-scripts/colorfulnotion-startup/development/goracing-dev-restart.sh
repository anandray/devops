#!/bin/bash

for goracing in {1..10}
do
if ! ps aux | grep racerver_development | grep -v grep; then
su - goracing << EOF
cd /home/goracing/scripts
./deployDevelopment.sh &
EOF
fi
sleep 5
done
