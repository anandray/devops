#!/bin/bash

AZURE_STORAGE_KEY="CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ=="
if [ -d /root/go/src/github.com/wolkdb/cloudstore ]; then
cd /root/go/src/github.com/wolkdb/cloudstore
git_status=`git status -s | awk '{print$2}'`
git checkout $git_status
git fetch origin && git merge origin/master
LOCAL=$(git rev-parse @{0})
REMOTE=$(git rev-parse @{u})
BASE=$(git merge-base @{0} @{u})
if [ $LOCAL = $REMOTE ]; then
 echo "Already up to date... removing cloudstore-git-update.sh from crontab"
 sed -i '/cloudstore-git-update/d' /var/spool/cron/root
else
 echo "Updating cloudstore repository"
 git_status=`git status -s | awk '{print$2}'`
 git checkout $git_status
 git fetch origin
 git merge origin/master
# /sbin/service wolk restart
 echo "done.. removing cloudstore-git-update.sh from crontab"
 sed -i '/cloudstore-git-update/d' /var/spool/cron/root
fi
fi
