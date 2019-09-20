#!/bin/bash

if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        sudo mkdir -p /var/www/vhosts
        sudo su - << EOF
        cd /var/www/vhosts
        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        git config core.filemode false;
        git config user.email "sourabh@crosschannel.com";
        git config user.name "Sourabh Niyogi";
        git fetch upstream;
        git merge upstream/master;
	cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh ccdex &> /var/log/ccdex.log
EOF
else
sudo su - << EOF
pkill -9 crosschannel
cd /var/www/vhosts/crosschannel.com/;
git fetch upstream;
git merge upstream/master;
cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh ccdex &> /var/log/ccdex.log
EOF
fi