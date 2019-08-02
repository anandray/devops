#!/bin/bash
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        cd /var/www/vhosts
        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        git config core.filemode false;
        git config user.email "sourabh@crosschannel.com";
        git config user.name "Sourabh Niyogi";
        git fetch upstream;
        git merge upstream/master;
else
	echo "/var/www/vhosts/crosschannel.com exists..."
fi

for i in {1..10};
do
cd /var/www/vhosts/crosschannel.com
git fetch
LOCAL=$(git rev-parse @{0})
REMOTE=$(git rev-parse @{u})
BASE=$(git merge-base @{0} @{u})

if [ $LOCAL = $REMOTE ]; then
 echo "`date +'%b %d, %Y, %T'` - Already up-to-date"
else
 echo "`date +'%b %d, %Y, %T'` - updating"
 git fetch upstream
 git merge upstream/master
 echo "done"
fi
sleep 5;
done
