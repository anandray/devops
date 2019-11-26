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
sudo /usr/bin/gsutil cp gs://startup_scripts_us/scripts/dataproc/memcached.conf /etc/;
sudo /usr/bin/memcached -m 1024 -p 83 -u root -l 127.0.0.1 -d

# verify php is installed
sudo dpkg -l | grep php5 || sudo apt-get -y install php5 php5-dev php5-curl php5-geoip php5-mysql
sudo sed -i 's/short_open_tag = Off/short_open_tag = On/g' /etc/php5/cli/php.ini;

# adding Defaults PATH to sudoers
sudo sed -i '/secure_path/d' /etc/sudoers
echo "Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin" >> /etc/sudoers

# permit ssh root login
sudo sed -i 's/PermitRootLogin no/PermitRootLogin yes/g' /etc/ssh/sshd_config;
# ssh_config modification
sudo sed -i '51 i\StrictHostKeyChecking no' /etc/ssh/ssh_config
sudo sed -i '52 i\UserKnownHostsFile \/dev\/null' /etc/ssh/ssh_config
sudo /usr/sbin/service ssh restart;

# add .ssh keys
/usr/bin/gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz /home/anand/.ssh/; 
sudo /usr/bin/gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/; 
tar zxvpf /home/anand/.ssh/ssh_keys.tgz -C /home/anand/.ssh/;
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/;
sudo chown -R root.root /root/.ssh/;


# GIT CLONE IF NEEDED

# installing GO
sudo su - << EOF
/usr/bin/gsutil cp gs://startup_scripts_us/scripts/go/go-install-version-1.9.2.sh /home/anand;
sh /home/anand/go-install-version-1.9.2.sh;
ln -s /usr/local/go/bin/go /usr/bin
EOF

# change timezone to PDT
sudo mv /etc/localtime /etc/localtime_BAK;
sudo ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;

# shutdown instance if php is not installed
sudo dpkg -l | grep php5 || sudo shutdown -h now

# adding PATH to bashrc
sudo su - << EOF
/usr/bin/gsutil cp gs://startup_scripts_us/scripts/dataproc/dataproc_bashrc /root/.bashrc;
/usr/bin/gsutil cp gs://startup_scripts_us/scripts/dataproc/alias.sh /etc/profile.d/;
EOF

# emacs for go-lang
sudo su - << EOF
/usr/bin/gsutil cp -r gs://startup_scripts_us/scripts/emacs/emacs/site-lisp /usr/share/emacs;
/usr/bin/gsutil cp -r gs://startup_scripts_us/scripts/emacs/.emacs.d /root/;
EOF

# START copying fair-scheduler.xml and setting up yarn and restarting
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/dataproc/fair-scheduler.xml /etc/hadoop/conf/;
gsutil cp gs://startup_scripts_us/scripts/dataproc/hadoop-yarn-resourcemanager-chk.sh /home/anand;

bdconfig set_property \
    --configuration_file /etc/hadoop/conf/yarn-site.xml \
    --name yarn.resourcemanager.scheduler.class \
    --value org.apache.hadoop.yarn.server.resourcemanager.scheduler.fair.FairScheduler

bdconfig set_property \
    --configuration_file /etc/hadoop/conf/yarn-site.xml \
    --name yarn.scheduler.fair.allocation.file \
    --value /etc/hadoop/conf/fair-scheduler.xml
EOF

sudo su - << EOF
ROLE=$(curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/dataproc-role)
if echo $ROLE | grep Master > /dev/null; then
sudo /bin/systemctl restart hadoop-yarn-resourcemanager.service > /root/hadoop-yarn-resourcemanager-restart.log 2>&1
sh /home/anand/hadoop-yarn-resourcemanager-chk.sh;
fi
EOF
