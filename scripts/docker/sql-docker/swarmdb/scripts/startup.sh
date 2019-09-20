#!/bin/bash

bold=$(tput bold)
normal=$(tput sgr0)

source /root/.bashrc

echo "
${bold}Downloading the plasma repository from github...
"

echo ${normal}

if [ ! -d /usr/local/go/src/github.com/wolkdb/plasma ]; then
mkdir -p /usr/local/go/src/github.com/wolkdb && cd /usr/local/go/src/github.com/wolkdb &&
git clone --recurse-submodules git@github.com:wolkdb/plasma.git &> /dev/null &

sh /usr/local/swarmdb/scripts/.git-clone-bar.sh

else
echo "
${bold}plasma repository already exists...
"
fi

echo ${normal}

echo "
${bold}Downloading and Installing GO-DEP:
"

echo ${normal}

if [ ! -f /usr/local/go/bin/dep ]; then
mkdir -p /usr/local/go/bin
export GOPATH="/usr/local/go"
curl -s https://raw.githubusercontent.com/golang/dep/master/install.sh | sh &> /dev/null &

sh /usr/local/swarmdb/scripts/.dep-install-bar.sh
fi

# run dep status
echo "
${bold}Running \"dep status\":
"
if [ -f /usr/local/go/bin/dep ]; then
cd /usr/local/go/src/github.com/wolkdb/plasma
dep status &
sh /usr/local/swarmdb/scripts/.dep-status-bar.sh
fi

# activating crontab
crontab -e &> /dev/null << EOF
:wq
EOF

# Compiling the plasma binary
if [ -d /usr/local/go/src/github.com/wolkdb/plasma/build ] && [ ! -f /usr/local/go/src/github.com/wolkdb/plasma/build/bin/plasma ]; then
echo "
${bold}`date +%T` - Compiling \"plasma\" binary...
"

echo ${normal}

cd /usr/local/go/src/github.com/wolkdb/plasma/
make plasma &> /dev/null &

sh /usr/local/swarmdb/scripts/.make-plasma-bar.sh

else
echo "
${bold}`date +%T` - \"plasma\" binary already exists...
"
fi

echo ${normal}

# Verify if the 'make plasma' process is still running.. 
if ps aux | grep -E "go run build|go install" | grep -v grep &> /dev/null; then
echo "
${bold}Looks like PLASMA is still being complied.. Allow some more time to complete the compilation...
"
sleep 10
fi

echo ${normal}

# Verify if plasma was compiled successfully or not. If not, try compiling again...
if [ ! -f /usr/local/go/src/github.com/wolkdb/plasma/build/bin/plasma ]; then
echo "
${bold}Looks like PLASMA is not compiled yet! Try compiling again...
"

echo ${normal}

cd /usr/local/go/src/github.com/wolkdb/plasma/
make plasma &> /dev/null &

sh /usr/local/swarmdb/scripts/.make-plasma-bar.sh

else
echo "
${bold}plasma was successfully compiled...
"
echo ${normal}
fi

# Compiling the swarmdb binary
if [ -d /usr/local/go/src/github.com/wolkdb/plasma/build ] && [ ! -f /usr/local/go/src/github.com/wolkdb/plasma/build/bin/swarmdb ]; then
/usr/local/swarmdb/scripts/swarmdb.sh
fi

if [ -d /usr/local/go/src/github.com/wolkdb/data ]; then
export DATADIR=/usr/local/go/src/github.com/wolkdb/data
else
mkdir -p /usr/local/go/src/github.com/wolkdb/data
export DATADIR=/usr/local/go/src/github.com/wolkdb/data
fi

# enter APPTXNDIR
cd $APPTXNDIR

echo "
${bold}#################################################
${bold}DATADIR=/usr/local/swarmdb/data
${bold}plasma repository=/usr/local/go/src/github.com/wolkdb/plasma
${bold}plasma=/usr/local/go/src/github.com/wolkdb/plasma/build/bin/plasma
${bold}#################################################
"

echo ${normal}

sh /usr/local/swarmdb/scripts/plasma-start.sh