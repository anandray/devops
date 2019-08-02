#!/bin/bash
cd /var/www/vhosts/api.wolk.com/ && git fetch upstream && git merge upstream/master
cd -
