#!/bin/bash

disk_size=`df -h | grep disk1 | awk '{print$5}' | cut -d "%" -f1`
if [ $disk_size -ge 75 ];
  then
  echo "Disk size > 75%, deleting logs from 4 days ago..."
  if ls -ld /disk1/log/bid/`date +%Y`/`date +%m`/`date +%d --date='4 days ago'` &> /dev/null;
    then
    rm -rfv /disk1/log/bid/`date +%Y`/`date +%m`/`date +%d --date='4 days ago'`
  else
  echo "/disk1/log/bid/`date +%Y`/`date +%m`/`date +%d --date='4 days ago'` DOES NOT EXIST..."
  fi
else
echo "Disk Use% > 75% but < 85%. 4 days old logs will be deleted when disk exceeds 75%"
fi

if [ $disk_size -ge 85 ];
  then
  echo "Disk size > 85% , deleting logs from 3 days ago..."
  rm -rfv /disk1/log/bid/`date +%Y`/`date +%m`/`date +%d --date='3 days ago'`
else
echo "Disk Use% < 85%. 3 days old logs will be deleted when disk exceeds 85%"
fi
