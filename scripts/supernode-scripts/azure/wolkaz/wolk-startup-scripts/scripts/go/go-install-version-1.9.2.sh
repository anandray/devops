#!/bin/bash

#Installing golang
if [ ! -d /usr/local/go ]; then
	sudo gsutil cp gs://startup_scripts_us/scripts/go/go1.9.2.linux-amd64.tar.gz /usr/local;
	sudo tar -C /usr/local -xzf /usr/local/go1.9.2.linux-amd64.tar.gz;
fi

#Adding environment variables to /root/.bashrc
if ! sudo grep GOPATH /root/.bashrc; then
sudo su - << EOF
sed -i '/go/d' /root/.bashrc
echo '
export PATH=$PATH:/usr/local/go/bin
#export PATH=$PATH:/usr/local/go/bin:/var/www/vhosts/swarm.wolk.com/src/github.com/ethereum/go-ethereum/build/bin/
#export GOPATH=/var/www/vhosts/swarm.wolk.com
export GOROOT=/usr/local/go' >> /root/.bashrc
source /root/.bashrc
EOF
fi
