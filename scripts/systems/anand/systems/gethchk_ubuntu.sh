#!/bin/bash

#geth --rpc --rpcaddr 10.128.0.230 &>> /var/log/geth_rpc.log
psaux=`ps aux | grep 'geth --rpc --rpcaddr 10.128.0.230' | grep -v grep` > /dev/null 2>&1
if ! ps aux | grep 'geth --rpc --rpcaddr 10.128.0.230' | grep -v grep > /dev/null 2>&1;
  then
  echo "`date +%m-%d-%T` - restarting geth rpc"
  geth --rpc --rpcaddr 10.128.0.230 &>> /var/log/geth_rpc.log
else
echo "

---- Start of geth check ----
`date +%m-%d-%T`:
---------------------------------------------------------------------------------------------------
$psaux
---------------------------------------------------------------------------------------------------

Check [1] geth rpc is running"
fi

netstat_port=`netstat -apn | grep -E 'geth|udp|tcp' | grep ':::30303'` > /dev/null 2>&1
netstat_rpc_port=`netstat -apn | grep -E ':30303|:8545' | grep LISTEN` > /dev/null 2>&1
#if ! netstat -apn | grep -E 'geth|udp|tcp' | grep ':::30303' > /dev/null;
#if ! netstat -apn | grep '10.128.0.230:8545' | grep LISTEN > /dev/null 2>&1;
if ! netstat -apn | grep -E ':30303|:8545' | grep LISTEN > /dev/null 2>&1;
  then
  echo "`date +%m-%d-%T` - restarting geth rpc"
  geth --rpc --rpcaddr 10.128.0.230 &>> /var/log/geth_rpc.log
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
