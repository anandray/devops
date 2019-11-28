#!/bin/bash

# Download the new 'wolkdb' and 'wolk' binaries

# wolkdb binary --> https://github.com/wolktoken/swarm.wolk.com/raw/master/src/github.com/ethereum/go-ethereum/swarmdb/server/wolkdb
#wolk cli binary --> https://github.com/wolktoken/swarm.wolk.com/raw/master/src/github.com/ethereum/go-ethereum/swarmdb/client/wolk/wolk

#wolkdb
rm -rfv /usr/local/swarmdb/bin/wolkdb && wget -O /usr/local/swarmdb/bin/wolkdb https://github.com/wolktoken/swarm.wolk.com/raw/master/src/github.com/ethereum/go-ethereum/swarmdb/server/wolkdb && chmod +x /usr/local/swarmdb/bin/wolkdb && sed -i 's/rm -rfv \/usr\/local\/swarmdb\/bin\/wolkdb/#rm -rfv \/usr\/local\/swarmdb\/bin\/wolkdb/g' /usr/local/swarmdb/scripts/wolkdb-start.sh

#wolk cli
rm -rfv /usr/local/swarmdb/bin/wolk && wget -O /usr/local/swarmdb/bin/wolk https://github.com/wolktoken/swarm.wolk.com/raw/master/src/github.com/ethereum/go-ethereum/swarmdb/client/wolk/wolk && chmod +x /usr/local/swarmdb/bin/wolk && sed -i 's/rm -rfv \/usr\/local\/swarmdb\/bin\/wolk/#rm -rfv \/usr\/local\/swarmdb\/bin\/wolk/g' /usr/local/swarmdb/scripts/wolkdb-start.sh

# json format swarmdb.conf
if [ -f /usr/local/swarmdb/etc/swarmdb.conf ]; then
echo "json formatting swarmdb.conf..."
python -m json.tool /usr/local/swarmdb/etc/swarmdb.conf > /tmp/swarmdb.conf && mv -f /tmp/swarmdb.conf /usr/local/swarmdb/etc/
fi

for i in {1..12}
do
if ! ps aux | grep wolkdb | grep -vE 'wolkdb-start|grep' &> /dev/null; then
echo “`date +%T` - wolkdb not running... starting wolkdb...”

nohup /usr/local/swarmdb/bin/wolkdb &> /usr/local/swarmdb/log/wolkdb.log &
/usr/bin/python -m json.tool /usr/local/swarmdb/etc/swarmdb.conf > /tmp/swarmdb.conf && mv -f /tmp/swarmdb.conf /usr/local/swarmdb/etc/ && sed -i 's/\/usr\/bin\/python -m json.tool/#\/usr\/bin\/python -m json.tool/g' /usr/local/swarmdb/scripts/wolkdb-start.sh

else
echo “`date +%T` - wolkdb is already running...”
fi
sleep 5;
done
