#!/bin/sh

if [ -d /var/www/vhosts/mdotm.com ]; then
cd /var/www/vhosts/mdotm.com
git fetch 
LOCAL=$(git rev-parse @{0})
REMOTE=$(git rev-parse @{u})
BASE=$(git merge-base @{0} @{u})

if [ $LOCAL = $REMOTE ]; then 
 echo "no update"
else 
 echo "updating"
 git fetch upstream   
 git merge upstream/master   
 echo "done"
fi

else
 echo "/var/www/vhosts/mdotm.com does NOT exist"
fi
