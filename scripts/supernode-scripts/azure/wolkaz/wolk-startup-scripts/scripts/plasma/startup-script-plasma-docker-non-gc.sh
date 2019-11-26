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

gsutil cp gs://startup_scripts_us/scripts/swarm/swarm-sudoers /etc/sudoers;
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
gsutil -m cp gs://startup_scripts_us/scripts/hosts /etc/

#SSH Keys:
if [ ! -f /root/.ssh/authorized_keys ]; then
sudo mkdir -p /root/.ssh
gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
sudo rm -rf /root/.ssh/ssh_keys.tgz
fi

# Allow SSH-ing to any instance/server
gsutil -m cp gs://startup_scripts_us/scripts/ssh_config /etc/ssh/
gsutil -m cp gs://startup_scripts_us/scripts/sshd_config /etc/ssh/
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
gsutil -m cp gs://startup_scripts_us/scripts/histtimeformat.sh /etc/profile.d/

## DISABLE FSCK
sudo tune2fs -c 0 -i 0 /dev/sda1
sudo tune2fs -c 0 -i 0 /dev/sdb1

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
gsutil -m cp gs://startup_scripts_us/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
gsutil -m cp gs://startup_scripts_us/scripts/security_limits.conf /etc/security/limits.conf
gsutil -m cp gs://startup_scripts_us/scripts/90-nproc.conf /etc/security/limits.d/

# Install basic packages + Docker
#sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
#sudo yum -y install epel-release emacs vim ntpdate telnet screen git openssl openssl-devel lynx net-tools mlocate wget yum-utils device-mapper-persistent-data lvm2
sudo yum -y install telnet net-tools yum-utils device-mapper-persistent-data lvm2
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum -y install docker-ce
sudo service docker restart

#############

# copying scripts to /root/scripts
mkdir -p /root/scripts
#gsutil -m cp gs://startup_scripts_us/scripts/profile.d/gce.sh /etc/profile.d/;
gsutil -m cp gs://startup_scripts_us/scripts/profile.d/histtimeformat.sh /etc/profile.d/;
gsutil -m cp gs://startup_scripts_us/scripts/cron_env*.bash /root/scripts;
gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys_chk.sh /root/scripts;
gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
gsutil -m cp gs://startup_scripts_us/scripts/plasma/plasma-start.sh /root/scripts/;
gsutil -m cp -r gs://startup_scripts_us/scripts/plasma/plasma-docker /root/scripts/;
gsutil cp gs://startup_scripts_us/scripts/plasma/plasma-docker-bashrc /root/.bashrc;
sudo chmod -R +x /root/scripts
sudo chown anand.anand -R /home/anand;

#######

# net.ipv4.tcp_tw_recycle and net.ipv4.tcp_tw_reuse
sudo su - << EOF
echo 1 > /proc/sys/net/ipv4/tcp_tw_reuse;
echo 1 > /proc/sys/net/ipv4/tcp_tw_recycle;
echo 1 > /proc/sys/net/ipv4/ip_forward;
sed -i '/ipv4.tcp_tw/d' /etc/sysctl.conf
sed -i '/Recycle and Reuse TIME_WAIT sockets faster/d' /etc/sysctl.conf
echo '
# Recycle and Reuse TIME_WAIT sockets faster
net.ipv4.tcp_tw_recycle = 1
net.ipv4.tcp_tw_reuse = 1

# Enable IPv4 forwarding
net.ipv4.ip_forward = 1' >> /etc/sysctl.conf;
/sbin/sysctl -p;
EOF

#######

# remove rsyslog before starting docker deployment
sudo yum -y remove rsyslog

# Start Docker
if ! docker ps | grep plasma-docker | grep -v grep &> /dev/null; then
cd /root/scripts/plasma-docker
sudo docker build -t wolkinc/plasma-docker . && sudo docker run --name=plasma-docker --rm -dit --dns=8.8.8.8 --dns=8.8.4.4 -p 30303:30303  -p 30303:30303/udp -p 30304:30304/udp -p 5000:5000 wolkinc/plasma-docker
fi

# apply bashrc changes
source /root/.bashrc
exec -l $SHELL
