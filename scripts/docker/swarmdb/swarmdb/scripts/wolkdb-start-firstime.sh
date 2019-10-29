#!/bin/bash

echo "
Downloading the wolkdb binary...
"
#wolkdb
rm -rf /usr/local/swarmdb/bin/wolkdb &> /dev/null &
wget -O /usr/local/swarmdb/bin/wolkdb https://github.com/wolkdb/swarmdb/raw/master/src/github.com/ethereum/go-ethereum/swarmdb/s
erver/wolkdb &> /dev/null &

# wolkdb-venus
wget -O /usr/local/swarmdb/bin/wolkdb-venus https://github.com/wolkdb/swarmdb/raw/venus/src/github.com/ethereum/go-ethereum/swar
mdb/server/wolkdb-venus &> /dev/null &

# wolkdb-moon
wget -O /usr/local/swarmdb/bin/wolkdb-moon https://github.com/wolkdb/swarmdb/raw/moon/src/github.com/ethereum/go-ethereum/swarmd
b/server/wolkdb &> /dev/null &

if [ -f /usr/local/swarmdb/bin/wolkdb ]; then
chmod +x /usr/local/swarmdb/bin/wolkdb* &> /dev/null
fi

echo -ne '########                                      (20%)\r'
sleep 2
echo -ne '################                              (40%)\r'
sleep 2
echo -ne '########################                      (60%)\r'
sleep 2
echo -ne '################################              (80%)\r'
sleep 2
echo -ne '##########################################    (100%)\r'
echo -ne '\n'

if ! ps aux | grep wolkdb | grep -vE 'wolkdb-start|grep' &> /dev/null; then
echo "
`date +%T` - wolkdb not running... starting wolkdb...
"

nohup /usr/local/swarmdb/bin/wolkdb-venus &> /usr/local/swarmdb/log/wolkdb.log &
else
echo “`date +%T` - wolkdb is already running...”
fi

# json format swarmdb.conf
#if [ -f /usr/local/swarmdb/etc/swarmdb.conf ]; then
#python -m json.tool /usr/local/swarmdb/etc/swarmdb.conf > /tmp/swarmdb.conf && mv -f /tmp/swarmdb.conf /usr/local/swarmdb/etc/
#fi

# activating crontab
crontab -e &> /dev/null << EOF
:wq
EOF

# install/update nodejs
echo "
Installing nodejs...
"
curl -sL https://rpm.nodesource.com/setup_9.x | bash - &> /dev/null
yum install -y gcc-c++ make nodejs &> /dev/null &

echo -ne '####                                          (10%)\r'
sleep 2
echo -ne '########                                      (20%)\r'
sleep 2
echo -ne '############                                  (30%)\r'
sleep 2
echo -ne '################                              (40%)\r'
sleep 2
echo -ne '####################                          (50%)\r'
sleep 2
echo -ne '########################                      (60%)\r'
sleep 2
echo -ne '############################                  (70%)\r'
sleep 2
echo -ne '################################              (80%)\r'
sleep 2
echo -ne '####################################          (90%)\r'
sleep 2
echo -ne '##########################################    (100%)\r'
echo -ne '\n'

# installing swarmdb.js
if [ ! -d /usr/local/swarmdb/swarmdb.js ]; then
mkdir -p /usr/local/swarmdb/swarmdb.js;
fi

cd /usr/local/swarmdb/swarmdb.js;
echo "
Installing swarmdb.js under /usr/local/swarmdb/swarmdb.js
"

npm install swarmdb.js --save &> /dev/null &
npm install web3@1.0.0-beta.26 &> /dev/null &

echo -ne '####                                          (10%)\r'
sleep 2
echo -ne '########                                      (20%)\r'
sleep 2
echo -ne '############                                  (30%)\r'
sleep 2
echo -ne '################                              (40%)\r'
sleep 2
echo -ne '####################                          (50%)\r'
sleep 2
echo -ne '########################                      (60%)\r'
sleep 2
echo -ne '############################                  (70%)\r'
sleep 2
echo -ne '################################              (80%)\r'
sleep 2
echo -ne '####################################          (90%)\r'
sleep 2
echo -ne '##########################################    (100%)\r'
echo -ne '\n'

# try chmod again
if [ -f /usr/local/swarmdb/bin/wolkdb ]; then
chmod +x /usr/local/swarmdb/bin/wolkdb* &> /dev/null
fi

# making sure wolkdb is started
if ! ps aux | grep wolkdb | grep -vE 'wolkdb-start|grep' &> /dev/null; then
nohup /usr/local/swarmdb/bin/wolkdb &> /usr/local/swarmdb/log/wolkdb.log &
fi

# installing swarmdbstats
if [ ! -d /usr/local/swarmdb/swarmdbstats ]; then
echo "
Installing swarmdbstats
"
/usr/local/swarmdb/scripts/swarmdbstats-init.sh &> /dev/null &

echo -ne '####                                          (10%)\r'
sleep 2
echo -ne '########                                      (20%)\r'
sleep 2
echo -ne '############                                  (30%)\r'
sleep 2
echo -ne '################                              (40%)\r'
sleep 2
echo -ne '####################                          (50%)\r'
sleep 2
echo -ne '########################                      (60%)\r'
sleep 2
echo -ne '############################                  (70%)\r'
sleep 2
echo -ne '################################              (80%)\r'
sleep 2
echo -ne '####################################          (90%)\r'
sleep 2
echo -ne '##########################################    (100%)\r'
echo -ne '\n'
fi

# enter root dir
cd $HOME

# echo swarmdb.conf info
bold=$(tput bold)
normal=$(tput sgr0)

# json format swarmdb.conf
#if [ -f /usr/local/swarmdb/etc/swarmdb.conf ]; then
#python -m json.tool /usr/local/swarmdb/etc/swarmdb.conf > /tmp/swarmdb.conf && mv -f /tmp/swarmdb.conf /usr/local/swarmdb/etc/
#fi

echo "
############################################################
${bold}swarmDB config file: ${normal}/usr/local/swarmdb/etc/swarmdb.conf"
echo "${bold}Your Address: ${normal}\"0x$(grep address /usr/local/swarmdb/etc/swarmdb.conf | tail -n 1 | cut -d "\"" -f4)\""
echo "${bold}Your Private Key: ${normal}\"$(grep privateKey /usr/local/swarmdb/etc/swarmdb.conf | cut -d"\"" -f4)
############################################################"
