#!/bin/sh

# ssh keys
if ! grep log04 /root/.ssh/id_rsa.pub; then
sudo mkdir -p /root/.ssh
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/ssh-keys/* /root/.ssh/
sudo chmod 0400 /root/.ssh/id_rsa*
fi

if [ -d /root/go/src/github.com/wolkdb/cloudstore ]; then
cd /root/go/src/github.com/wolkdb/cloudstore
git fetch
LOCAL=$(git rev-parse @{0})
REMOTE=$(git rev-parse @{u})
BASE=$(git merge-base @{0} @{u})

if [ $LOCAL = $REMOTE ]; then
 echo "Already up to date... removing cloudstore-git-update.sh from crontab"
 sed -i '/cloudstore-git-update/d' /var/spool/cron/root
else
 echo "Updating cloudstore repository"
 git fetch origin
 git merge origin/master
 echo "done.. removing cloudstore-git-update.sh from crontab"
 sed -i '/cloudstore-git-update/d' /var/spool/cron/root
fi
fi

if [ -d /usr/local/wolk ]; then
        sudo /sbin/service wolk stop;
        sudo /usr/bin/pkill -9 wolk;
        sudo rm -rf /usr/local/wolk/*;
fi

# populate wolk.toml
sudo su - << EOF
sed -i "s/consensus/storage/g" /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
EOF

if [ -d /usr/local/wolk ]; then
        sudo /sbin/service wolk stop;
        sudo /usr/bin/pkill -9 wolk;
        sudo rm -rf /usr/local/wolk/*;
fi

# make wolk | start wolk | creating wolk1-wolk5
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/rc.local /etc/rc.d/rc.local
sudo chmod +x /etc/rc.d/rc.local
sudo chmod +x /etc/rc.local
sudo scp c0.wolk.com:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/make-wolk.sh /root/scripts/
sudo chmod +x /root/scripts/make-wolk.sh
sudo sh /root/scripts/make-wolk.sh
