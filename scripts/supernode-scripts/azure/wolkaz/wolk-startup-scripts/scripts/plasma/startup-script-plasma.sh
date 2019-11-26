#!/bin/bash

# create /root/scripts dir
sudo mkdir /root/scripts;

# install gcloud and activate service account

sudo tee -a /etc/yum.repos.d/google-cloud-sdk.repo << EOM
[google-cloud-sdk]
name=Google Cloud SDK
baseurl=https://packages.cloud.google.com/yum/repos/cloud-sdk-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
       https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOM

yum -y install google-cloud-sdk wget
exec -l $SHELL

wget -O /root/scripts/crosschannel-1307-520dd999e93e.json http://www6001.wolk.com/.start/crosschannel-1307-520dd999e93e.json
gcloud auth activate-service-account --key-file /root/scripts/crosschannel-1307-520dd999e93e.json

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;

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

sudo yum -y install epel-release gcc make emacs ntpdate telnet screen git python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools mlocate wget sqlite-devel;

sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo gsutil cp gs://startup_scripts_us/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;

# adding log0 and log6 to /etc/hosts
if ! grep log0 /etc/hosts; then
echo '
104.154.36.231    log6
35.193.168.171    log0' >> /etc/hosts
fi

# syslog-ng
sudo yum -y install syslog-ng syslog-ng-libdbi libdbi-devel
cp /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-orig
wget -O /etc/syslog-ng/syslog-ng.conf http://www6001.wolk.com/.start/syslog-ng.conf
service rsyslog stop
chkconfig rsyslog off
service syslog-ng restart
########

# download plasma binary
if [ ! -d /root/go/src/github.com/wolkdb/plasma/build/bin ]; then
        sudo su - << EOF
       	mkdir -p /root/go/src/github.com/wolkdb/plasma/build/bin
	gsutil cp gs://startup_scripts_us/scripts/plasma/bin/plasma /root/go/src/github.com/wolkdb/plasma/build/bin
	gsutil cp gs://startup_scripts_us/scripts/plasma/bin/plasma-gw /root/go/src/github.com/wolkdb/plasma/build/bin
	chmod +x /root/go/src/github.com/wolkdb/plasma/build/bin/plasma*
	echo "export PATH=$PATH:/root/go/src/github.com/wolkdb/plasma/build/bin" >> /root/.bashrc
	source /root/.bashrc
EOF
fi

#############

# copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil -m cp gs://startup_scripts_us/scripts/profile.d/histtimeformat.sh /etc/profile.d/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/cron_env*.bash /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/plasma/plasma-bashrc /root/.bashrc
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys_chk.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/plasma/plasma-start.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/plasma/syslogtest /root/scripts/;
sudo gsutil -m cp -r gs://startup_scripts_us/scripts/plasma/.google /root/;
sudo chmod -R +x /root/scripts
sudo chown anand.anand -R /home/anand;

# emacs for go-lang
sudo su - << EOF
gsutil -m cp -r gs://startup_scripts_us/scripts/emacs/emacs/site-lisp /usr/share/emacs;
gsutil -m cp -r gs://startup_scripts_us/scripts/emacs/.emacs.d /root/;
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

# applying bashrc
exec -l $SHELL
source /root/.bashrc

# exporting GOOGLE_APPLICATION_CREDENTIALS
export GOOGLE_APPLICATION_CREDENTIALS="/root/.google/crosschannel-1307-520dd999e93e.json"

# Start plasma
if ! ps aux | grep plasma | grep -v grep; then
echo "Starting plasma..."
/root/scripts/plasma-start.sh
fi

ps aux  | grep -E 'plasma|syslog' | grep -v grep
/root/scripts/syslogtest
