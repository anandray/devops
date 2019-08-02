#!/bin/bash
## Use this script to update(git fetch + merge) api.colorfulnotion.com on www6002 + www6005 + www6006 + api-colorfulnotion-com-3t4c
#cd /var/www/vhosts/colorfulnotion.com
#git fetch upstream && git merge upstream/master
cd /var/www/vhosts/api.colorfulnotion.com
git fetch upstream && git merge upstream/master
sh /var/www/vhosts/api.colorfulnotion.com/scripts/goservice.sh roam
