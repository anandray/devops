#!/bin/bash
cd /var/www/vhosts/mdotm.com
git fetch upstream && git merge upstream/master

cd /var/www/vhosts/crosschannel.com
git fetch upstream && git merge upstream/master

sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/git_merge.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/git_merge.sh/g' /var/spool/cron/root
