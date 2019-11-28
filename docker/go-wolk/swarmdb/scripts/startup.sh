#!/bin/bash

bold=$(tput bold)
normal=$(tput sgr0)

echo "
${bold}Downloading the go-wolk repository from github...
"

echo ${normal}

if [ ! -d /var/www/vhosts ]; then
mkdir -p /var/www/vhosts && cd /var/www/vhosts &&
git clone --recurse-submodules -b phobos git@github.com:wolkdb/go-wolk.git &> /dev/null &

echo -ne '####                                          (10%)\r'
sleep 10
echo -ne '########                                      (20%)\r'
sleep 10
echo -ne '############                                  (30%)\r'
sleep 10
echo -ne '################                              (40%)\r'
sleep 10
echo -ne '####################                          (50%)\r'
sleep 10
echo -ne '########################                      (60%)\r'
sleep 10
echo -ne '############################                  (70%)\r'
sleep 10
echo -ne '################################              (80%)\r'
sleep 10
echo -ne '####################################          (90%)\r'
sleep 10
echo -ne '##########################################    (100%)\r'
echo -ne '\n'

else
echo "
${bold}go-wolk repository already exists...
"
fi

echo ${normal}

# Commenting out chunkBroadcastLoop section
#if [ -f /var/www/vhosts/go-wolk/eth/handler.go ] &> /dev/null; then
#sed -i '859 i\\/\*' /var/www/vhosts/go-wolk/eth/handler.go
#sed -i '873 i\\*\/' /var/www/vhosts/go-wolk/eth/handler.go
#fi

# activating crontab
crontab -e &> /dev/null << EOF
:wq
EOF

# Compiling the geth binary
if [ -d /var/www/vhosts/go-wolk/build ] && [ ! -f /var/www/vhosts/go-wolk/build/bin/geth ]; then
echo "
${bold}`date +%T` - Compiling \"geth\" binary...
"

echo ${normal}

cd /var/www/vhosts/go-wolk/
make geth &> /dev/null &

echo -ne '####                                          (10%)\r'
sleep 12
echo -ne '########                                      (20%)\r'
sleep 12
echo -ne '############                                  (30%)\r'
sleep 12
echo -ne '################                              (40%)\r'
sleep 12
echo -ne '####################                          (50%)\r'
sleep 12
echo -ne '########################                      (60%)\r'
sleep 12
echo -ne '############################                  (70%)\r'
sleep 12
echo -ne '################################              (80%)\r'
sleep 12
echo -ne '####################################          (90%)\r'
sleep 12
echo -ne '##########################################    (100%)\r'
sleep 12
echo -ne '\n'

else
echo "
${bold}`date +%T` - \"geth\" binary already exists...
"
fi

echo ${normal}

# Verify if the 'make geth' process is still running.. 
if ps aux | grep -E "go run build|go install" | grep -v grep &> /dev/null; then
echo "
${bold}Looks like GETH is still being complied.. Allow some more time to complete the compilation...
"
sleep 10
fi

echo ${normal}

# Verify if geth was compiled successfully or not. If not, try compiling again...
if [ ! -f /var/www/vhosts/go-wolk/build/bin/geth ]; then
echo "
${bold}Looks like GETH is not compiled yet! Try compiling again...
"

echo ${normal}

cd /var/www/vhosts/go-wolk/
make geth &> /dev/null &

echo -ne '####                                          (10%)\r'
sleep 12
echo -ne '########                                      (20%)\r'
sleep 12
echo -ne '############                                  (30%)\r'
sleep 12
echo -ne '################                              (40%)\r'
sleep 12
echo -ne '####################                          (50%)\r'
sleep 12
echo -ne '########################                      (60%)\r'
sleep 12
echo -ne '############################                  (70%)\r'
sleep 12
echo -ne '################################              (80%)\r'
sleep 12
echo -ne '####################################          (90%)\r'
sleep 12
echo -ne '##########################################    (100%)\r'
sleep 12
echo -ne '\n'

else
echo "
${bold}geth was successfully compiled...
"
echo ${normal}
fi

if [ -d /var/www/vhosts/data ]; then
export DATADIR=/var/www/vhosts/data
else
mkdir -p /var/www/vhosts/data
export DATADIR=/var/www/vhosts/data
fi

if [ -d /usr/local/swarmdb/apptxn ]; then
export APPTXNDIR=/usr/local/swarmdb/apptxn
fi

# enter APPTXNDIR
cd $APPTXNDIR

echo "
${bold}#################################################
${bold}DATADIR=/usr/local/swarmdb/data
${bold}go-wolk repository=/var/www/vhosts/go-wolk
${bold}geth=/var/www/vhosts/go-wolk/build/bin/geth
${bold}apptxn=/usr/local/swarmdb/apptxn
${bold}#################################################
"

echo ${normal}

sh /usr/local/swarmdb/scripts/geth-start.sh
