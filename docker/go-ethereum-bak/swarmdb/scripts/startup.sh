#!/bin/bash

bold=$(tput bold)
normal=$(tput sgr0)

echo "
${bold}Downloading the go-ethereum repository from github...
"

echo ${normal}

if [ ! -d /var/www/vhosts ]; then
mkdir -p /var/www/vhosts && cd /var/www/vhosts &&
git clone --recurse-submodules git@github.com:ethereum/go-ethereum.git &> /dev/null &

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
${bold}go-ethereum repository already exists...
"
fi

echo ${normal}

# activating crontab
crontab -e &> /dev/null << EOF
:wq
EOF

# Compiling the geth binary
if [ -d /var/www/vhosts/go-ethereum/build ] && [ ! -f /var/www/vhosts/go-ethereum/build/bin/geth ]; then
echo "
${bold}`date +%T` - Compiling \"geth\" binary...
"

echo ${normal}

cd /var/www/vhosts/go-ethereum/
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
if [ ! -f /var/www/vhosts/go-ethereum/build/bin/geth ]; then
echo "
${bold}Looks like GETH is not compiled yet! Try compiling again...
"

echo ${normal}

cd /var/www/vhosts/go-ethereum/
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

echo "
${bold}#################################################
${bold}DATADIR=/usr/local/swarmdb/data
${bold}go-ethereum repository=/var/www/vhosts/go-ethereum
${bold}geth=/var/www/vhosts/go-ethereum/build/bin/geth
${bold}#################################################
"

echo ${normal}

sh /usr/local/swarmdb/scripts/geth-start.sh
