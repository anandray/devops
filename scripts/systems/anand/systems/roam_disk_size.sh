#!/bin/bash

disk_size=`df -h | grep 'sda1' | awk '{print$5}' | cut -d "%" -f1`
disk_size1=`df -h | grep 'sda1' | awk '{print$5}'`

if [ $disk_size -ge 50 ];
  then
  echo "current disk usage $disk_size1 - `date +%T`"
> /var/log/roam
else
echo "current disk usage: $disk_size1 - `date +%T`"
fi
