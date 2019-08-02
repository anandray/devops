#!/bin/bash
if [ ! -d /root/go/src/github.com/wolkdb/cloudstore ]; then
        echo "/root/go/src/github.com/wolkdb/cloudstore does NOT exist, proceeding with git clone..."
        cd /var/www/vhosts
        sudo git clone git@github.com:wolkdb/cloudstore.git /root/go/src/github.com/wolkdb/cloudstore;
        cd /root/go/src/github.com/wolkdb/cloudstore/;
        git config core.filemode true;
        git config user.email "sourabh@wolk.com";
        git config user.name "Sourabh Niyogi";
        git fetch origin;
        git merge origin/master;
else
	echo "/root/go/src/github.com/wolkdb/cloudstore exists..."
fi

#for i in {1..11}; # increasing frequency from every 5 seconds to 30 seconds
for i in {1..2};
do
cd /root/go/src/github.com/wolkdb/cloudstore
git fetch
LOCAL=$(git rev-parse @{0})
REMOTE=$(git rev-parse @{u})
BASE=$(git merge-base @{0} @{u})

if [ $LOCAL = $REMOTE ]; then
 echo "`date +'%b %d, %Y, %T'` - Already up-to-date"
else
 echo "`date +'%b %d, %Y, %T'` - updating"
 git fetch origin
 git merge origin/master
 echo "done"
fi
sleep 25;
done
