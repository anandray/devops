#!/bin/bash

	if [ ! -d /var/www/vhosts/api.colorfulnotion.com ]; then
        echo "/var/www/vhosts/api.colorfulnotion.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        cd /var/www/vhosts/;
        git clone git@github.com:sourabhniyogi/api.colorfulnotion.com.git /var/www/vhosts/api.colorfulnotion.com;
        cd /var/www/vhosts/api.colorfulnotion.com;
        git remote add upstream git@github.com:sourabhniyogi/api.colorfulnotion.com.git;
        git config user.email sourabh@crosschannel.com;
        git config user.name "Sourabh Niyogi";
        git config --global core.filemode false;
        git config core.filemode false;
        git fetch upstream;
        git merge upstream/master;
	cd /var/www/vhosts/api.colorfulnotion.com/scripts && sh goservice.sh roam &> /var/log/roam.log
EOF
fi

if ! curl -s  "http://127.0.0.1/health" | grep HEALTHCHECK &> /dev/null; then
sudo su - << EOF
cd /var/www/vhosts/api.colorfulnotion.com/;
git fetch upstream;
git merge upstream/master;
cd /var/www/vhosts/api.colorfulnotion.com/scripts && sh goservice.sh roam &> /var/log/roam.log
EOF
fi

if ! ps aux | grep '/var/www/vhosts/api.colorfulnotion.com/roam/roam' | grep -v grep &> /dev/null; then
echo "
`date +%m-%d-%T`:
-------------------------------------
roam is NOT running.. Restarting....
-------------------------------------
"
sudo su - << EOF
cd /var/www/vhosts/api.colorfulnotion.com/;
git fetch upstream;
git merge upstream/master;
cd /var/www/vhosts/api.colorfulnotion.com/scripts && sh goservice.sh roam &> /var/log/roam.log
EOF
fi
