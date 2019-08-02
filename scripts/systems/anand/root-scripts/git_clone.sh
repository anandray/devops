#!/bin/bash

#SSH Keys:
if [ ! -f /root/.ssh/authorized_keys ]; then
        sudo gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
        sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
        sudo rm -rf /root/.ssh/ssh_keys.tgz
else
#        echo "ssh keys are present"

if [ ! -d /var/www/vhosts/mdotm.com ]; then
        echo "/var/www/vhosts/mdotm.com does NOT exist, proceeding with git clone..."
        sudo mkdir -p /var/www/vhosts
        sudo su - << EOF
        cd /var/www/vhosts
        sudo git clone git@github.com:sourabhniyogi/mdotm.com.git /var/www/vhosts/mdotm.com;
        cd /var/www/vhosts/mdotm.com/;
        sudo git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git;
        sudo git config core.filemode false;
        sudo git config user.email "sourabh@crosschannel.com";
        sudo git config user.name "Sourabh Niyogi";
        sudo git fetch upstream;
        sudo git merge upstream/master;
        gsutil cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/
        mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK
EOF

        sudo /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php;
        sudo /bin/sh /root/scripts/httpd_conf.sh;
        sudo gsutil cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/syslog-ng.conf;
        sudo service syslog-ng restart;
        gcloud components update --quiet;
        gcloud components install alpha --quiet;
        gcloud components install beta --quiet;
#        sudo php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/newservercheck.php;
 else
        echo "/var/www/vhosts/mdotm.com already exists, not running git clone..."
#       echo "removing git_clone.sh from crontab..."
#       sed -i '/git_clone.sh/d' /var/spool/cron/root
        echo "commenting out git_clone.sh in crontab..."
        sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/git_clone.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/git_clone.sh/g' /var/spool/cron/root
fi;
#       sudo /usr/bin/php /root/scripts/mailer.php;
fi;
