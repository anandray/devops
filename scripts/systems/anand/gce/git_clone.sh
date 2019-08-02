#!/bin/bash

if [ ! -d /var/www/vhosts/mdotm.com ]; then
	echo "/var/www/vhosts/mdotm.com does NOT exist, proceeding with git clone..."
	sudo mkdir -p /var/www/vhosts
	cd /var/www/vhosts
	sudo git clone git@github.com:sourabhniyogi/mdotm.com.git
	cd /var/www/vhosts/mdotm.com
	sudo git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git
	sudo git fetch upstream && git merge upstream/master
	sudo git config core.filemode false
	sudo git config user.email "sourabh@mdotm.com"
	sudo git config user.name "Sourabh Niyogi"
	sudo mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK
	sudo cp -rfv /root/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/
 else
	echo "/var/www/vhosts/mdotm.com already exists, not running git clone..."
#       echo "removing git_clone.sh from crontab..."
#       sed -i '/git_clone.sh/d' /var/spool/cron/root
        echo "commenting out git_clone.sh in crontab..."
        sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/git_clone.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/git_clone.sh/g' /var/spool/cron/root
fi;
