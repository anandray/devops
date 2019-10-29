#!/bin/bash

if [ ! -d /var/www/vhosts/mdotm.com ]; then
	echo "/var/www/vhosts/mdotm.com does NOT exist, proceeding with git clone..."
	sudo mkdir -p /var/www/vhosts
	sudo su - << EOF
	cd /var/www/vhosts
	sudo git clone git@github.com:sourabhniyogi/mdotm.com.git /var/www/vhosts/mdotm.com;
	cd /var/www/vhosts/mdotm.com/;
	git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git;
	git config core.filemode false;
	git config user.email "sourabh@crosschannel.com";
	git config user.name "Sourabh Niyogi";
	git fetch upstream;
	git merge upstream/master;
	gsutil cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/
	mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK

        /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php > /var/log/loadAll.log 2>&1;
        gsutil cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/syslog-ng.conf;
        service syslog-ng restart;
        gcloud components update --quiet;
        gcloud components install alpha --quiet;
        gcloud components install beta --quiet;
EOF
#        sudo php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/newservercheck.php;
 else
	echo "/var/www/vhosts/mdotm.com already exists, not running git clone..."
#	Commenting out git_clone.sh in crontab...
	sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/git_clone.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/git_clone.sh/g' /var/spool/cron/root
fi;
