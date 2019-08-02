#!/bin/bash

for goracing in {1..10}
do
if ! ps aux | grep racerver_staging | grep -v grep; then
su - goracing << EOF
cd /home/goracing/scripts
./deployStaging.sh &
EOF
fi
sleep 5
done
