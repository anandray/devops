#!/bin/bash
sudo apt-get update --fix-missing
sudo apt-get -y install php5 php5-dev php5-curl php5-geoip php5-mysql 
sudo apt-get -y install emacs git 
sudo apt-get -y install pkg-config libgtk2.0-dev libevtlog-dev 
sudo apt-get -y install sendmail openssl telnet;
sudo sed -i 's/short_open_tag = Off/short_open_tag = On/g' /etc/php5/cli/php.ini;

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
echo 'extension=mongo.so' >> /etc/php5/cli/php.ini;
echo 'extension=mongodb.so' >> /etc/php5/cli/php.ini;
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
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-install.sh /home/anand;
sh /home/anand/hbase-install.sh;

# copying fair-scheduler.xml and setting up yarn and restarting - start
gsutil cp gs://startup_scripts_us/scripts/dataproc/fair-scheduler.xml /etc/hadoop/conf/;

bdconfig set_property \
    --configuration_file /etc/hadoop/conf/yarn-site.xml \
    --name yarn.resourcemanager.scheduler.class \
    --value org.apache.hadoop.yarn.server.resourcemanager.scheduler.fair.FairScheduler

bdconfig set_property \
    --configuration_file /etc/hadoop/conf/yarn-site.xml \
    --name yarn.scheduler.fair.allocation.file \
    --value /etc/hadoop/conf/fair-scheduler.xml

ROLE=$(curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/dataproc-role)
if [[ "${ROLE}" == 'Master' ]]; then
  systemctl restart hadoop-yarn-resourcemanager.service
fi
# copying fair-scheduler.xml and setting up yarn and restarting - end

# Installing google-fluentd
sudo apt-get -y install google-fluentd google-fluentd-catch-all-config;
sudo rm -rf /etc/google-fluentd/config.d/*;
sudo gsutil cp gs://startup_scripts_us/scripts/google-fluentd-syslog.conf /etc/google-fluentd/config.d/syslog.conf;
sudo service google-fluentd restart;

# Sending status to stackdriver
sudo /usr/bin/logger -p local3.info -t CROSSCHANNEL "NEW INSTANCE CREATED|$HOSTNAME|`date +%m%d-%T`"
