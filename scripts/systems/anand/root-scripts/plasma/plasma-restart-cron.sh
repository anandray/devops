#!/bin/bash

for i in {1..12}
do
if ! ps aux | grep "plasma --bootnode" | grep -v grep &> /dev/null; then
sed -i 's/#nohup/nohup/g' /root/scripts/plasma-start.sh;
/root/scripts/plasma-start.sh &
fi

if ps aux | grep "plasma --bootnode" | grep -v grep &> /dev/null; then
/root/scripts/plasma-cron-disable.sh
fi

sleep 5;
done
