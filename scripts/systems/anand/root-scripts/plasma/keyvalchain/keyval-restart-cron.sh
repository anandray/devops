#!/bin/bash

for i in {1..12}
do
if ! ps aux | grep "keyvalchain --datadir" | grep -v grep &> /dev/null; then
sed -i 's/#nohup/nohup/g' /root/scripts/keyval-start.sh;
/root/scripts/keyval-start.sh &
fi

if ps aux | grep "keyvalchain --datadir" | grep -v grep &> /dev/null; then
/root/scripts/keyval-cron-disable.sh
fi

sleep 5;
done
