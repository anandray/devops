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

sudo gsutil cp gs://startup_scripts_us/scripts/swarm/swarm-sudoers /etc/sudoers;
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

# yum installations

sudo yum -y install epel-release gcc make emacs ntpdate telnet screen git denyhosts python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools sudo telnet vim wget initscripts iproute mlocate;

sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo gsutil cp gs://startup_scripts_us/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;

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

# python3

yum -y install python36u python36u-devel python36u-libs python36u-pip python36u-setuptools python36u-tools;
ln -s /usr/bin/python3.6 /usr/bin/python3;
ln -s /usr/bin/pip3.6 /usr/bin/pip3;
pip3 install --upgrade setuptools;

yum -y groupinstall "Development tools";
yum -y install openssl-devel;
pip3 install google google_compute_engine;

#############

# installing syslog-ng
rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
RUN yum -y install syslog-ng syslog-ng-libdbi libdbi-devel

#############

# GIT CLONE SWARM.WOLK.COM
if [ ! -d /var/www/vhosts/swarm.wolk.com ]; then
        echo "/var/www/vhosts/swarm.wolk.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
	mkdir -p /var/www/vhosts;
        cd /var/www/vhosts/;
	git clone -v git@github.com:wolkdb/swarm.wolk.com.git
	cd swarm.wolk.com
	git remote -v add upstream git@github.com:wolkdb/swarm.wolk.com.git
	git config user.email sourabh@wolk.com
	git config user.name "Sourabh Niyogi"
	git config --global core.filemode false
	git config core.filemode false
        git fetch upstream;
        git merge upstream/master;
EOF
fi


#############

# installing GO
#sudo gsutil -m cp gs://startup_scripts_us/scripts/go/go-install-swarm.sh /home/anand;
sudo gsutil -m cp gs://startup_scripts_us/scripts/go/go-install-swarm-1.9.2.sh /home/anand;
sudo sh /home/anand/go-install-swarm-1.9.2.sh;

#############

# copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil -m cp gs://startup_scripts_us/scripts/profile.d/* /etc/profile.d/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/cron_env*.bash /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/geth/geth-bashrc-centos7 /root/.bashrc;
gsutil -m cp gs://startup_scripts_us/scripts/geth/geth-bashrc-centos7 /home/anand/.bashrc;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys_chk.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/swarm/geth-install.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/geth/geth-ropsten-start-nomine.sh /root/scripts/;
sudo chmod -R +x /root/scripts
sudo chown anand.anand -R /home/anand;

# emacs for go-lang
sudo su - << EOF
gsutil -m cp -r gs://startup_scripts_us/scripts/emacs/emacs/site-lisp /usr/share/emacs;
gsutil -m cp -r gs://startup_scripts_us/scripts/emacs/.emacs.d /root/;
EOF

#Copy CRONJOBS:
sudo su - << EOF
gsutil -m cp gs://startup_scripts_us/scripts/geth/cron-ropsten-geth.sh /var/spool/cron/root;
service crond restart;
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
rm -rf /usr/bin/gsutil;source /root/.bashrc
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

# start geth using ropsten enode #
sudo sh /root/scripts/geth-ropsten-start-nomine.sh
#sleep 10

# start swarm using geth address
#sudo sh /root/scripts/swarm-start.sh
