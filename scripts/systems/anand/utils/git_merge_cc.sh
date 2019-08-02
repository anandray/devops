#!/bin/bash
cd /var/www/vhosts/crosschannel.com/ && git fetch upstream && git merge upstream/master
cd -
