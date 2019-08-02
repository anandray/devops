#!/bin/bash
if [ ! -d /var/www/vhosts/www.wolk.com ]; then
        echo "/var/www/vhosts/www.wolk.com does NOT exist, proceeding with git clone..."
        cd /var/www/vhosts
        sudo git clone git@github.com:wolktoken/www.wolk.com.git /var/www/vhosts/www.wolk.com;
        cd /var/www/vhosts/www.wolk.com/;
        git remote add upstream git@github.com:wolktoken/www.wolk.com.git;
        git config core.filemode false;
        git config user.email "sourabh@wolk.com";
        git config user.name "Sourabh Niyogi";
        git fetch upstream;
        git merge upstream/master;
else
	echo "/var/www/vhosts/www.wolk.com exists..."
fi

for i in {1..12};
do
#if ! cd /var/www/vhosts/www.wolk.com/ && git status | grep 'nothing to commit (working directory clean)' &> /dev/null; then
	cd /var/www/vhosts/www.wolk.com/ &&
	git fetch upstream &&
        git merge upstream/master;
#else
#	echo "`date +%m-%d-%Y\|%T` - www.wolk.com is already updated..."
#fi
sleep 5;
done
