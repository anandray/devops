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

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil -m cp gs://startup_scripts_us/scripts/security_limits.conf /etc/security/limits.conf
sudo gsutil -m cp gs://startup_scripts_us/scripts/90-nproc.conf /etc/security/limits.d/

#PHP INSTALL:

#### Redhat 6.x ####
sudo rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
sudo rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-15.ius.el6.noarch.rpm
#sudo gsutil cp gs://startup_scripts_us/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;
#sudo su - << EOF
##gsutil gs://startup_scripts_us/scripts/epel.repo /etc/yum.repos.d/epel.repo
#/bin/sed -i 's/\#baseurl=http/baseurl=http/g' /etc/yum.repos.d/epel.repo;
#/bin/sed -i 's/mirrorlist=http/\#mirrorlist=http/g' /etc/yum.repos.d/epel.repo;
#EOF
########

#sudo yum -y install gcc make emacs ntpdate telnet screen git denyhosts python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx;
sudo yum -y install google-authenticator python-argparse python27-pip;

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

#############


# copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil cp gs://startup_scripts_us/scripts/histtimeformat.sh /etc/profile.d/;
sudo gsutil cp gs://wolk-scripts/gce.sh /etc/profile.d/
sudo gsutil cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
sudo gsutil cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
sudo gsutil cp gs://wolk-scripts/sshd_config /etc/ssh/;
sudo gsutil cp gs://wolk-scripts/sshd /etc/pam.d/;
sudo gsutil cp gs://wolk-scripts/gs-wolk-cdn-acl.sh /root/scripts/;
sudo service sshd restart
sudo chmod -R +x /root/scripts
gsutil cp gs://wolk-scripts/ssh-keys/authorized_keys /home/anand/.ssh
gsutil cp gs://wolk-scripts/enable-google-auth.sh /home/anand/;
sudo chown -R anand.anand /home/anand

# set vim in bashrc
echo "alias vi='/usr/bin/vim'" >> /home/anand/.bashrc
echo "alias vi='/usr/bin/vim'" >> /root/.bashrc

# enable google-autheticator for anand
su - anand << EOF
sh /home/anand/enable-google-auth.sh
EOF

#Copy CRONJOBS:
sudo su - << EOF
gsutil -m cp gs://wolk-scripts/cdn-wolk-cron /var/spool/cron/root;
service crond restart;
EOF

#######

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
