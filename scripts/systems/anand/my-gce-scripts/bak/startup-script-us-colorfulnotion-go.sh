#!/bin/bash

# create /root/scripts dir
sudo mkdir /root/scripts;

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;

#sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/ntpdate*.sh /root/scripts/; #copied below
#sudo sh /root/scripts/ntpdate.sh;

## Start /etc/sudoers modifications ##
# THE FOLLOWING /etc/sudoers MODIFICATIONS ARE REQUIRED TO ALLOW 'gcloud' and 'gsutil' to run using sudo, WHICH IS OTHERWISE NOT ALLOWED
# Replace /etc/sudoers default 'Defaults secure_path' with following:
#Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin

# Comment out this line in /etc/sudoers:
#Defaults    requiretty

sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/sudoers /etc/
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/hosts /root/scripts/
sudo cat /root/scripts/hosts >> /etc/hosts

#SSH Keys:
sudo mkdir -p /root/.ssh
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/ssh_keys.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
sudo rm -rf /root/.ssh/ssh_keys.tgz

# Allow SSH-ing to any instance/server
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/ssh_config /etc/ssh/
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/sshd_config /etc/ssh/
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/histtimeformat.sh /etc/profile.d/

## DISABLE FSCK
sudo tune2fs -c 0 -i 0 /dev/sda1
sudo tune2fs -c 0 -i 0 /dev/sda3

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/security_limits.conf /etc/security/limits.conf
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/90-nproc.conf /etc/security/limits.d/

#PHP INSTALL:

#### Redhat 6.x ####
#sudo rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
#sudo rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-14.ius.el6.noarch.rpm
sudo rpm -Uvh https://storage.googleapis.com/startup-scripts-colorfulnotion/scripts/rpm_packages/epel_ius_rpms/epel-release-6-8.noarch.rpm
sudo rpm -Uvh https://storage.googleapis.com/startup-scripts-colorfulnotion/scripts/rpm_packages/epel_ius_rpms/ius-release-1.0-14.ius.el6.noarch.rpm
sudo su - << EOF
#gsutil gs://startup-scripts-colorfulnotion/scripts/epel.repo /etc/yum.repos.d/epel.repo
/bin/sed -i 's/\#baseurl=http/baseurl=http/g' /etc/yum.repos.d/epel.repo;
/bin/sed -i 's/mirrorlist=http/\#mirrorlist=http/g' /etc/yum.repos.d/epel.repo;
EOF
########

sudo yum -y install libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail sendmail-cf denyhosts procmail python-argparse *whois openssl openssl-devel libssh2 libssh2-devel java-1.8.0-openjdk-devel python27-pip lynx;

# python2.6 to python2.7
sudo su - << EOF
/usr/bin/pip2.7 install google google_compute_engine
unlink /usr/bin/python2
ln -s /usr/bin/python2.7 /usr/bin/python2
gsutil cp gs://startup-scripts-colorfulnotion/scripts/python_version_change.sh /root/scripts/;
gsutil cp gs://startup-scripts-colorfulnotion/scripts/python_version_change_cron.sh /root/scripts/;
chmod -R 0755 /root/scripts/python_version_change*;
cp -rfv /root/scripts/python_version_change.sh /etc/cron.hourly/
cp -rfv /root/scripts/python_version_change_cron.sh /etc/cron.d/
EOF

sudo service syslog stop;
sudo chkconfig syslog off;
sudo service postfix stop;
sudo chkconfig postfix off;
sudo chkconfig --del postfix;
sudo service rsyslog stop;
sudo chkconfig rsyslog off;

#Copy GeoIP.dat
sudo gsutil -m cp -r gs://startup-scripts-colorfulnotion/scripts/GeoIP /usr/share/
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/GeoIP/GeoIP.conf /etc/GeoIP.conf

#######

# add '/var/www/vhosts/mdotm.com/scripts/utils' to path
sudo su - << EOF
sudo echo "pathmunge /var/www/vhosts/mdotm.com/scripts/utils" > /etc/profile.d/pushcode.sh
EOF

#Install syslog-ng:
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/syslog/syslog-ng.conf-colorfulnotion /etc/syslog-ng/syslog-ng.conf;
sudo service syslog-ng restart;
sudo chkconfig syslog-ng on;

##############
#Install Nagios/cacti client
sudo su - << EOF
gsutil cp gs://startup-scripts-colorfulnotion/scripts/nagios/nrpechk.sh /root/scripts/;
yum -y install nagios-plugins nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
gsutil cp gs://startup-scripts-colorfulnotion/scripts/nagios/nrpe.cfg /etc/nagios/;
gsutil -m cp -r gs://startup-scripts-colorfulnotion/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
chmod +x /usr/lib64/nagios/plugins/*
chkconfig nrpe on;
service nrpe restart;
chkconfig snmpd on;
EOF

#############

#Configure services to run on reboot:
service sendmail restart;
chkconfig httpd off;
chkconfig crond on;
chkconfig iptables off;
chkconfig sendmail on;
chkconfig syslog-ng on;
chkconfig syslog off;
chkconfig rsyslog off;

#USE GIT:
sudo mkdir -p /var/www/vhosts;
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
mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK
EOF

#USE GIT TO ADD COLORFULNOTION.COM
if [ ! -d /var/www/vhosts/api.colorfulnotion.com ]; then
        echo "/var/www/vhosts/api.colorfulnotion.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        cd /var/www/vhosts
        sudo git clone git@github.com:sourabhniyogi/api.colorfulnotion.com.git /var/www/vhosts/api.colorfulnotion.com;
        cd /var/www/vhosts/api.colorfulnotion.com/;
        git remote add upstream git@github.com:sourabhniyogi/api.colorfulnotion.com.git;
        git config core.filemode false;
        git config user.email "sourabh@crosschannel.com";
        git config user.name "Sourabh Niyogi";
        git fetch upstream;
        git merge upstream/master;
EOF
fi

#USE GIT TO ADD CROSSCHANNEL.COM
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
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
EOF
fi

# copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil -m cp -r gs://startup-scripts-colorfulnotion/scripts/roam_scripts/* /root/scripts/;
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/profile.d/* /etc/profile.d/;
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/.bash_profile /root/;
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/startup-script-us-colorfulnotion-go-failsafe.sh /root/scripts/;
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/cron.hourly_crosschannel /etc/cron.hourly/crosschannel;
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/logrotate.d_crosschannel /etc/logrotate.d/crosschannel;
sudo chmod -R +x /root/scripts

# emacs for go-lang
sudo su - << EOF
gsutil -m cp -r gs://startup-scripts-colorfulnotion/scripts/emacs/emacs/site-lisp /usr/share/emacs;
gsutil -m cp -r gs://startup-scripts-colorfulnotion/scripts/emacs/.emacs.d /root/;
EOF

#Copy CRONJOBS:
sudo su - << EOF
gsutil cp gs://startup-scripts-colorfulnotion/scripts/cron_root-us-colorfulnotion-go /var/spool/cron/root;
service crond restart;
EOF

#ADD shortcircuit.php manually:
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/

# installing hbase
if [ ! -d /usr/local/hbase-1.1.2/ ]; then
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/hbase/hbase-install-us-colorfulnotion.sh /root/scripts;
sudo sh /root/scripts/hbase-install-us-colorfulnotion.sh;
fi

# installing GO
#sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/go/go-install.sh /home/anand;
#sudo sh /home/anand/go-install.sh;

#######

# net.ipv4.tcp_tw_recycle and net.ipv4.tcp_tw_reuse
sudo su - << EOF
echo 1 > /proc/sys/net/ipv4/tcp_tw_reuse;
echo 1 > /proc/sys/net/ipv4/tcp_tw_recycle;
sed -i '/ipv4.tcp_tw/d' /etc/sysctl.conf
sed -i '/Recycle and Reuse TIME_WAIT sockets faster/d' /etc/sysctl.conf
echo '
# Recycle and Reuse TIME_WAIT sockets faster
net.ipv4.tcp_tw_recycle = 1
net.ipv4.tcp_tw_reuse = 1' >> /etc/sysctl.conf;
/sbin/sysctl -p;
EOF

#######

# Starting ROAM
sudo su - << EOF
cd /var/www/vhosts/api.colorfulnotion.com/scripts && sh goservice.sh roam &> /var/log/roam.log
EOF

#######

#Change 'assumeyes=1' --> 'assumeyes=0' in yum.conf
sudo su - << EOF
/bin/sed -i '/assumeyes/d' /etc/yum.conf
/bin/sed -i "$ i\assumeyes=0" /etc/yum.conf
EOF

################

# python2.6 to python2.7
sudo su - << EOF
/usr/bin/pip2.7 install google google_compute_engine
unlink /usr/bin/python2
ln -s /usr/bin/python2.7 /usr/bin/python2
cp -rfv /root/scripts/python_version_change.sh /etc/cron.hourly/
cp -rfv /root/scripts/python_version_change_cron.sh /etc/cron.d/
chmod 0755 /etc/cron.hourly/python_version_change.sh
chmod 0755 /etc/cron.d/python_version_change_cron.sh
EOF

################
#Denyhosts
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/denyhosts/allowed-hosts /var/lib/denyhosts/;
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/denyhosts/denyhosts.conf /etc;
sudo su - << EOF
echo `ifconfig | grep 'inet addr:10' | awk '{print$2}' | cut -d ":" -f2` >> /var/lib/denyhosts/allowed-hosts
EOF
service denyhosts restart;
chkconfig denyhosts on;
###############

# clean up
#sudo rm -rf /root/composer*;
#sudo rm -rf /root/libmaxminddb;
#sudo rm -rf /root/vendor;

# update gcloud
#sudo gcloud -q components update
