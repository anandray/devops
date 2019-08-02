#!/bin/bash
## Use this script to update(git fetch + merge) wolkfoundation.org on www6005 + www6006
cd /var/www/vhosts/wolkfoundation.org
git fetch upstream
git merge upstream/master
