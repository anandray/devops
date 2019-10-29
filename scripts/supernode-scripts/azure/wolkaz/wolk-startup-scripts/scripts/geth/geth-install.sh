#!/bin/bash
if [ ! -d /var/www/vhosts/swarm.wolk.com/src/github.com/ethereum/go-ethereum/build/bin ]; then
echo "compiling geth using make all..."
sudo gsutil cp gs://startup_scripts_us/scripts/geth/install-geth.sh /root/scripts;
sh /root/scripts/install-geth.sh;
else
echo "commenting out..."
sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/geth-install.sh/#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/geth-install.sh/g' /var/spool/cron/root
fi
