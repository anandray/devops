#!/bin/bash
cd /var/www/vhosts/mdotm.com && git fetch upstream && git merge upstream/master
