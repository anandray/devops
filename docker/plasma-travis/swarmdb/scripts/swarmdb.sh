#!/bin/bash

# Compiling the swarmdb binary
if [ -d /usr/local/go/src/github.com/wolkdb/plasma/build ] && [ ! -f /usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb ]; then
echo "
${bold}`date +%T` - Compiling \"swarmdb\" binary...
"

echo ${normal}

cd /usr/local/go/src/github.com/wolkdb/plasma/
make swarmdb &> /dev/null &
sh /usr/local/swarmdb/scripts/.make-swarmdb-bar.sh

else
echo "
${bold}`date +%T` - \"swarmdb\" binary already exists...
"
fi

echo ${normal}

# Verify if the 'make swarmdb' process is still running.. 
if ps aux | grep -E "go run build|go install" | grep -v grep &> /dev/null; then
echo "
${bold}Looks like SWARMDB is still being complied.. Allow some more time to complete the compilation...
"
sleep 10
fi

echo ${normal}

# Verify if swarmdb was compiled successfully or not. If not, try compiling again...
if [ ! -f /usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb ]; then
echo "
${bold}Looks like SWARMDB is not compiled yet! Try compiling again...
"

echo ${normal}

cd /usr/local/go/src/github.com/wolkdb/plasma/
make swarmdb &> /dev/null &

sh /usr/local/swarmdb/scripts/.make-swarmdb-bar.sh

else
echo "
${bold}swarmdb was successfully compiled...
"
echo ${normal}
fi

#########

if [ -f /usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb ] && ! ps aux | grep "swarmdb --datadir" | grep -v grep &> /dev/null; then
echo "
${bold}#### S T A R T I N G   S W A R M D B ####
"

echo ${normal}

nohup /usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb \
--datadir /usr/local/swarmdb/data \
--port 31313 \
--networkid 7272 \
--mine \
--unlock 0x$acct \
--etherbase 0x$acct \
--verbosity 4 \
--maxpeers 100 \
--rpc \
--rpcaddr $eth \
--rpcport 8545 \
--password /usr/local/swarmdb/data/.wolk \
2>> /usr/local/swarmdb/data/swarmdb.log &

else
echo "
${bold}SWARMDB is already runing...
"
echo ${normal}
fi

if [ -f /usr/local/swarmdb/data/swarmdb.log ]; then
echo "
${bold}Check /usr/local/swarmdb/data/swarmdb.log
"

echo ${normal}
tail -n10 /usr/local/swarmdb/data/swarmdb.log
fi

sleep 2

if [ -d /usr/local/swarmdb/data/keystore ]; then
echo "
${bold}Listing SWARMDB account
"
/usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb --datadir /usr/local/swarmdb/data account list 2> /dev/null | awk "{print$3}" | cut -d "{" -f2 | cut -d "}" -f1

sleep 2

echo "
${bold}nodeInfo:
"
/usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb attach /usr/local/swarmdb/data/swarmdb.ipc --exec admin.nodeInfo
fi

echo ${normal}
