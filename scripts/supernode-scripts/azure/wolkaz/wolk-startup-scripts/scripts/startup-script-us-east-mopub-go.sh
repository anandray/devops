#!/bin/bash

# create /root/scripts dir
sudo mkdir /root/scripts;

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;

#sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/; #copied below
#sudo sh /root/scripts/ntpdate.sh;

## Start /etc/sudoers modifications ##
# THE FOLLOWING /etc/sudoers MODIFICATIONS ARE REQUIRED TO ALLOW 'gcloud' and 'gsutil' to run using sudo, WHICH IS OTHERWISE NOT ALLOWED
# Replace /etc/sudoers default 'Defaults secure_path' with following:
#Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin

# Comment out this line in /etc/sudoers:
#Defaults    requiretty

sudo gsutil -m cp gs://startup_scripts_us/scripts/sudoers /etc/
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
sudo gsutil -m cp gs://startup_scripts_us/scripts/hosts /etc/

#SSH Keys:
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
sudo rm -rf /root/.ssh/ssh_keys.tgz

# Allow SSH-ing to any instance/server
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_config /etc/ssh/
sudo gsutil -m cp gs://startup_scripts_us/scripts/sshd_config /etc/ssh/
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
sudo gsutil -m cp gs://startup_scripts_us/scripts/histtimeformat.sh /etc/profile.d/

## DISABLE FSCK
sudo tune2fs -c 0 -i 0 /dev/sda1
sudo tune2fs -c 0 -i 0 /dev/sda3

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo gsutil -m cp gs://startup_scripts_us/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil -m cp gs://startup_scripts_us/scripts/security_limits.conf /etc/security/limits.conf
sudo gsutil -m cp gs://startup_scripts_us/scripts/90-nproc.conf /etc/security/limits.d/

#PHP INSTALL:

#### Redhat 6.x ####
sudo rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
sudo rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-15.ius.el6.noarch.rpm

if ! rpm -qa | grep epel-release > /dev/null; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/rpms/epel-release-6-8.noarch.rpm /home/anand
sudo rpm -Uvh /home/anand/epel-release-6-8.noarch.rpm
fi

if ! rpm -qa | grep ius-release > /dev/null; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/rpms/ius-release-1.0-15.ius.el6.noarch.rpm /home/anand
sudo rpm -Uvh /home/anand/ius-release-1.0-15.ius.el6.noarch.rpm
fi

sudo su - << EOF
#gsutil gs://startup_scripts_us/scripts/epel.repo /etc/yum.repos.d/epel.repo
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
gsutil -m cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
gsutil -m cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
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
sudo gsutil -m cp -r gs://startup_scripts_us/scripts/GeoIP /usr/share/
sudo gsutil -m cp gs://startup_scripts_us/scripts/GeoIP/GeoIP.conf /etc/GeoIP.conf

################
## Install libtool for maxminddb
#sudo su - << EOF
#cd /root;
#yum -y install libtool* git && git clone --recursive https://github.com/maxmind/libmaxminddb
#cd libmaxminddb &&
#./bootstrap &&
#./configure &&
#make &&
#make check &&
#make install &&
#ldconfig
#EOF
#######

# add '/var/www/vhosts/mdotm.com/scripts/utils' to path
sudo su - << EOF
sudo echo "pathmunge /var/www/vhosts/mdotm.com/scripts/utils" > /etc/profile.d/pushcode.sh
EOF

#Install syslog-ng:
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/syslog-ng.conf;
sudo service syslog-ng restart;
sudo chkconfig syslog-ng on;

##############
#Install Nagios/cacti client
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
yum -y install nagios-plugins nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm /home/anand/;
yum -y remove nrpe && rpm -Uvh /home/anand/nrpe-2.15-7.el6.x86_64.rpm;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
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

#copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil -m cp gs://startup_scripts_us/scripts/profile.d/* /etc/profile.d/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/cron_env*.bash /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/.bash_profile /root/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/git_clone.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/git_clone_cc.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys_chk.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/shortcircuit.php /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslogchk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/startup-script-us-east-failsafe-mopub-go.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/cron.hourly_crosschannel /etc/cron.hourly/crosschannel;
sudo gsutil -m cp gs://startup_scripts_us/scripts/logrotate.d_crosschannel /etc/logrotate.d/crosschannel;
sudo chmod -R +x /root/scripts

# emacs for go-lang
sudo su - << EOF
gsutil -m cp -r gs://startup_scripts_us/scripts/emacs/emacs/site-lisp /usr/share/emacs;
gsutil -m cp -r gs://startup_scripts_us/scripts/emacs/.emacs.d /root/;
EOF

#Copy CRONJOBS:
sudo su - << EOF
gsutil -m cp gs://startup_scripts_us/scripts/cron_root-us-east-mopub-go /var/spool/cron/root;
service crond restart;
EOF

#ADD shortcircuit.php manually:
sudo gsutil -m cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/

# installing GO
sudo gsutil -m cp gs://startup_scripts_us/scripts/go/go-install.sh /home/anand;
sudo sh /home/anand/go-install.sh;

# running go command
sudo su - << EOF
cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh crosschannel 2> /var/log/git-crosschannel.err  > /var/log/git-crosschannel.log
EOF

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
sudo gsutil -m cp gs://startup_scripts_us/scripts/denyhosts/allowed-hosts /var/lib/denyhosts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/denyhosts/denyhosts.conf /etc;
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
