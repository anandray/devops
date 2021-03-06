#!/bin/bash

# Built with version 1.1: https://cloud.google.com/dataproc/docs/concepts/dataproc-versions

sudo apt-get update --fix-missing
sudo apt-get -y install php5 php5-dev php5-curl php5-geoip php5-mysql ## THIS IS INSTALLING PHP 5.6.27-0+deb8u1 (cli) (built: Oct 15 2016 15:53:28) as of Nov 16, 2016
sudo apt-get -y install emacs git 
sudo apt-get -y install libcurl4-openssl-dev pkg-config libevtlog-dev
sudo apt-get -y install libgtk2.0-dev
sudo apt-get -y install sendmail openssl telnet;
sudo sed -i 's/short_open_tag = Off/short_open_tag = On/g' /etc/php5/cli/php.ini;

# installing php-memcache, php-memcached, memcached:83
sudo apt-get -y install php5-memcache php5-memcached memcached;
sudo gsutil cp gs://startup_scripts_us/scripts/dataproc/memcached.conf /etc/;
sudo /usr/bin/memcached -m 1024 -p 83 -u root -l 127.0.0.1 -d

# copying GeoCity.dat
sudo gsutil cp -r gs://startup_scripts_us/scripts/dataproc/GeoIP /usr/share/

# verify php is installed
sudo dpkg -l | grep php5 || sudo apt-get -y install php5 php5-dev php5-curl php5-geoip php5-mysql
sudo sed -i 's/short_open_tag = Off/short_open_tag = On/g' /etc/php5/cli/php.ini;

# adding alias for ls -l
echo 'alias ll="ls -l"' >> /home/anand/.bashrc;
sudo gsutil cp gs://startup_scripts_us/scripts/dataproc/alias.sh /etc/profile.d/;

# adding Defaults PATH to sudoers
sudo sed -i '/secure_path/d' /etc/sudoers
echo "Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin" >> /etc/sudoers

# mongo php extensions
no '' | sudo pecl install mongo;
sudo pecl install mongodb;
sudo su - << EOF
sed -i '/mongo/d' /etc/php5/cli/php.ini;
echo 'extension=mongo.so' >> /etc/php5/cli/php.ini;
echo ';extension=mongodb.so' >> /etc/php5/cli/php.ini;
EOF

# permit ssh root login
sudo sed -i 's/PermitRootLogin no/PermitRootLogin yes/g' /etc/ssh/sshd_config;
# ssh_config modification
sudo sed -i '51 i\StrictHostKeyChecking no' /etc/ssh/ssh_config
sudo sed -i '52 i\UserKnownHostsFile \/dev\/null' /etc/ssh/ssh_config
sudo /usr/sbin/service ssh restart;

# add .ssh keys
gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz /home/anand/.ssh/; 
sudo gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/; 
tar zxvpf /home/anand/.ssh/ssh_keys.tgz -C /home/anand/.ssh/;
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/;
sudo chown -R root.root /root/.ssh/;


# Adding mdotm.com via git
sudo mkdir -p /var/www/vhosts;
cd /var/www/vhosts;
sudo git clone git@github.com:sourabhniyogi/mdotm.com.git;
cd /var/www/vhosts/mdotm.com;
sudo git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git;
sudo git fetch upstream && git merge upstream/master;
sudo git config core.filemode false;
sudo git config user.email "sourabh@crosschannel.com";
sudo git config user.name "Sourabh Niyogi";

# shortcircuit.php and renaming index.php
#sudo wget -O /var/www/vhosts/mdotm.com/include/shortcircuit.php http://anand.www1001.mdotm.com/gce/shortcircuit.php;
gsutil cp gs://startup_scripts_us/scripts/shortcircuit.php /home/anand/;
sudo cp -rf /home/anand/shortcircuit.php /var/www/vhosts/mdotm.com/include/;
sudo mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK

#USE GIT TO ADD CROSSCHANNEL.COM
#if [ ! -d /var/www/vhosts/crosschannel.com ]; then
#        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
#        sudo su - << EOF
#        cd /var/www/vhosts
#        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
#        cd /var/www/vhosts/crosschannel.com/;
#        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
#        git config core.filemode false;
#        git config user.email "sourabh@crosschannel.com";
#        git config user.name "Sourabh Niyogi";
#        git fetch upstream;
#        git merge upstream/master;
#EOF
#fi

# /etc/hosts with softlayer ha servers public ip
gsutil cp gs://startup_scripts_us/scripts/dataproc/hosts_sl /home/anand;
sudo cat /home/anand/hosts_sl >> /etc/hosts

# syslog-ng
sudo /etc/init.d/rsyslog stop;
sudo pkill -9 syslog-ng;
sudo apt-get -y remove rsyslog syslog-ng;
sudo apt-get -y autoremove;
sudo rm -rf /usr/local/sbin/syslog*;
#sudo apt-get -y install pkg-config libgtk2.0-dev libevtlog-dev;
wget -O /home/anand/syslog-ng_3.2.5.tar.gz https://my.balabit.com/downloads/syslog-ng/open-source-edition/3.2.5/source/syslog-ng_3.2.5.tar.gz;
tar zxvpf /home/anand/syslog-ng_3.2.5.tar.gz -C /home/anand/;
cd /home/anand/syslog-ng-3.2.5;
./configure && make;
sudo make install;
sudo cp -rf /home/anand/syslog-ng-3.2.5/syslog-ng/syslog-ng /usr/local/sbin/;
sudo mkdir -p /etc/syslog-ng/;
sudo mkdir -p /usr/local/var;
sudo gsutil cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/syslog-ng.conf;
sudo /usr/local/sbin/syslog-ng -f /etc/syslog-ng/syslog-ng.conf;
sudo ps aux | grep syslog-ng > /tmp/syslog-ng.log;
#cat /tmp/syslog-ng.log | sudo mail -s"Dataproc Cluster created - $HOSTNAME `date +%m%d%Y_%T`" engineering@crosschannel.com
#cat /tmp/syslog-ng.log | sudo mail -s"Dataproc Cluster created - $HOSTNAME `date +%m%d%Y_%T`" engineering@mdotm.com

# change timezone to PDT
sudo mv /etc/localtime /etc/localtime_BAK;
sudo ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;

# shutdown instance if php is not installed
sudo dpkg -l | grep php5 || sudo shutdown -h now

# installing hbase
sudo gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-install-haa-go.sh /home/anand;
sudo sh /home/anand/hbase-install-haa-go.sh;

# install bt/zn
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, copying from google storage..."
        sudo mkdir -p /var/www/vhosts/crosschannel.com/bidder
        sudo gsutil -m cp -r gs://startup_scripts_us/scripts/go/bin /var/www/vhosts/crosschannel.com/bidder/
        sudo chmod -R 0755 /var/www/vhosts/crosschannel.com/bidder/bin
fi

if netstat -apn | egrep '::9900' > /dev/null; then
  echo bt is running
else
  echo bt is NOT running
  cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php bt 2> /var/log/git-bt.err  > /var/log/git-bt.log
fi

if netstat -apn | egrep '::9901' > /dev/null; then
  echo zn is running
else
  echo zn is NOT running
  cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php zn 2> /var/log/git-zn.err  > /var/log/git-zn.log
fi

# emacs for go-lang
sudo su - << EOF
gsutil cp -r gs://startup_scripts_us/scripts/emacs/emacs/site-lisp /usr/share/emacs;
gsutil cp -r gs://startup_scripts_us/scripts/emacs/.emacs.d /root/;
EOF

# install nrpe
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe_install-ha6.sh /home/anand;
sh /home/anand/nrpe_install-ha6.sh;

# START copying fair-scheduler.xml and setting up yarn and restarting
sudo gsutil cp gs://startup_scripts_us/scripts/dataproc/fair-scheduler.xml /home/anand;
sudo gsutil cp gs://startup_scripts_us/scripts/dataproc/hadoop-yarn-resourcemanager-chk.sh /home/anand;
sudo cp -rf /home/anand/fair-scheduler.xml /etc/hadoop/conf/

sudo bdconfig set_property \
    --configuration_file /etc/hadoop/conf/yarn-site.xml \
    --name yarn.resourcemanager.scheduler.class \
    --value org.apache.hadoop.yarn.server.resourcemanager.scheduler.fair.FairScheduler

sudo bdconfig set_property \
    --configuration_file /etc/hadoop/conf/yarn-site.xml \
    --name yarn.scheduler.fair.allocation.file \
    --value /etc/hadoop/conf/fair-scheduler.xml

ROLE=$(curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/dataproc-role)
if echo $ROLE | grep Master > /dev/null; then
sudo /bin/systemctl restart hadoop-yarn-resourcemanager.service > /root/hadoop-yarn-resourcemanager-restart.log 2>&1
sh /home/anand/hadoop-yarn-resourcemanager-chk.sh;
fi

#Making sure to hadoop-yar-resourcemanager is restarted after copying the .xmls
#sudo gsutil cp gs://startup_scripts_us/scripts/dataproc/hadoop-yarn-resourcemanager-chk.sh /root/scripts;
#sudo su - << EOF
#echo '*/1 * * * * /bin/ssh `hostname` sh /root/scripts/hadoop-yarn-resourcemanager-chk.sh > /var/log/hadoop-yarn-resourcemanager-chk.log 2>&1' >> /var/spool/cron/crontabs/root
#EOF

#ROLE=$(curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/dataproc-role)
#if [[ "${ROLE}" == 'Master' ]]; then
#  /bin/systemctl restart hadoop-yarn-resourcemanager.service
#fi

# /END copying fair-scheduler.xml and setting up yarn and restarting #

## Sending status to stackdriver
#sudo /usr/bin/logger -p local3.info -t CROSSCHANNEL "NEW INSTANCE CREATED|$HOSTNAME|`date +%m%d-%T`"
