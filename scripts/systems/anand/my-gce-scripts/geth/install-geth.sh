#!/bin/bash
sudo su - << EOF
find /var/www/vhosts/swarm.wolk.com/src/github.com/ethereum/go-ethereum -name '*.sh' -exec chmod -v +x {} \;
cd /var/www/vhosts/swarm.wolk.com/src/github.com/ethereum/go-ethereum;
make all
EOF
