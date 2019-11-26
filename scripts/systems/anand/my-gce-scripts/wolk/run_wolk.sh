#!/bin/bash

if [ ! -d /var/www/vhosts/api.wolk.com ]; then
        echo "/var/www/vhosts/api.wolk.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        cd /var/www/vhosts/;
        git clone git@github.com:wolktoken/api.wolk.com.git /var/www/vhosts/api.wolk.com;
        cd /var/www/vhosts/api.wolk.com;
        git remote add upstream git@github.com:wolktoken/api.wolk.com.git;
        git config user.email sourabh@wolk.com;
        git config user.name "Sourabh Niyogi";
        git config --global core.filemode false;
        git config core.filemode false;
        git fetch upstream;
        git merge upstream/master;
	cd /var/www/vhosts/api.wolk.com/bin && sh goservice.sh wolk &> /var/log/wolk.log
EOF
else
sudo su - << EOF
cd /var/www/vhosts/api.wolk.com/;
git fetch upstream;
git merge upstream/master;
cd /var/www/vhosts/api.wolk.com/bin && sh goservice.sh wolk &> /var/log/wolk.log
EOF
fi
