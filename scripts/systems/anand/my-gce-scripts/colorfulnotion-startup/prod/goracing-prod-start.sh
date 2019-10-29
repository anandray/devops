#!/bin/bash

if ! ps aux | grep racerver_master | grep -v grep; then
su - goracing << EOF
cd /home/goracing/scripts
./deployStaging.sh &
EOF
fi
