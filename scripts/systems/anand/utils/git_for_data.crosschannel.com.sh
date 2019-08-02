#!/bin/bash
## Use this script to update(git fetch + merge) data.crosschannel.com on www6005 + www6006
cd /var/www/vhosts
cd /var/www/vhosts/data.crosschannel.com
git fetch upstream
git merge upstream/master
