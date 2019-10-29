#!/bin/bash

if ! ps aux | grep etherbase > /dev/null;
  then
  export LD_LIBRARY_PATH=/opt/glibc-2.14/lib
#  echo $LD_LIBRARY_PATH
  geth --etherbase '213cb7006456d1fcca47f0c8e72ffc20fdcd811f' --mine  &>> /var/log/geth.log
else
echo "geth mine is running"
fi
