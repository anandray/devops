#!/bin/bash

#psaux=`ps aux | grep etherbase | grep -v grep` > /dev/null
#geth --rpc --rpcaddr 50.225.47.153 &>> /var/log/geth_rpc.log
psaux=`ps aux | grep 'geth --rpc --rpcaddr 50.225.47.153' | grep -v grep` > /dev/null
if ! ps aux | grep 'geth --rpc --rpcaddr 50.225.47.153' | grep -v grep > /dev/null;
  then
  echo "`date +%m-%d-%T` - restarting geth rpc"
  export LD_LIBRARY_PATH=/opt/glibc-2.14/lib
#  echo $LD_LIBRARY_PATH
#  geth --etherbase '7a4ab9b9dba6528b7c957386991fc98ca1d6d3ef' --mine --minerthreads=32  &>> /var/log/geth.log
  geth --rpc --rpcaddr 50.225.47.153 &>> /var/log/geth_rpc.log
else
echo "

---- Start of geth check ----
`date +%m-%d-%T`:
---------------------------------------------------------------------------------------------------
$psaux
---------------------------------------------------------------------------------------------------

Check [1] geth rpc is running"
fi

netstat_port=`netstat -apn | grep -E 'geth|udp|tcp' | grep ':::30303'` > /dev/null
netstat_rpc_port=`netstat -apn | grep -E ':30303|:8545' | grep LISTEN` &> /dev/null
#if ! netstat -apn | grep -E 'geth|udp|tcp' | grep ':::30303' > /dev/null;
if ! netstat -apn | grep '50.225.47.153:8545' | grep LISTEN &> /dev/null;
  then
  echo "`date +%m-%d-%T` - restarting geth rpc"
  export LD_LIBRARY_PATH=/opt/glibc-2.14/lib
#  geth --etherbase '7a4ab9b9dba6528b7c957386991fc98ca1d6d3ef' --mine  &>> /var/log/geth.log
  geth --rpc --rpcaddr 50.225.47.153 &>> /var/log/geth_rpc.log
else
echo "
`date +%m-%d-%T`:
--------------------------------------------------------------------------------------------------
$netstat_rpc_port
$netstat_port
--------------------------------------------------------------------------------------------------

Check [2] geth rpc is running
---- /End of geth check ----
"
fi
