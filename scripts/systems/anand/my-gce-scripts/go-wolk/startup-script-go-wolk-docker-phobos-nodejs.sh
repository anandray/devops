#!/bin/bash

# create /root/scripts dir
sudo mkdir /root/scripts;

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;

# Comment out this line in /etc/sudoers:
#Defaults    requiretty

sudo gsutil cp gs://startup_scripts_us/scripts/swarm/swarm-sudoers /etc/sudoers;
## /End /etc/sudoers modifications ##

#SSH Keys:
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
sudo rm -rf /root/.ssh/ssh_keys.tgz

# Allow SSH-ing to any instance/server
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_host_dsa_key* /etc/ssh/
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_config /etc/ssh/
sudo gsutil -m cp gs://startup_scripts_us/scripts/sshd_config /etc/ssh/
sudo chmod 0400 /etc/ssh/ssh_host_dsa_key*
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
sudo gsutil -m cp gs://startup_scripts_us/scripts/histtimeformat.sh /etc/profile.d/

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo gsutil -m cp gs://startup_scripts_us/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil -m cp gs://startup_scripts_us/scripts/security_limits.conf /etc/security/limits.conf
sudo gsutil -m cp gs://startup_scripts_us/scripts/90-nproc.conf /etc/security/limits.d/

#### Redhat 7.x ####
sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo gsutil cp gs://startup_scripts_us/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;
sudo su - << EOF
#gsutil cp gs://startup_scripts_us/scripts/epel.repo /etc/yum.repos.d/epel.repo
/bin/sed -i 's/\#baseurl=http/baseurl=http/g' /etc/yum.repos.d/epel.repo;
/bin/sed -i 's/mirrorlist=http/\#mirrorlist=http/g' /etc/yum.repos.d/epel.repo;
EOF

#sudo yum -y install gcc make emacs ntpdate telnet screen git denyhosts python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools mlocate;
########

# copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil -m cp gs://startup_scripts_us/scripts/go-wolk/swarm-bashrc-centos7-bootnode-2 /root/.bashrc;
gsutil -m cp gs://startup_scripts_us/scripts/go-wolk/swarm-bashrc-centos7-bootnode-2 /home/anand/.bashrc;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys_chk.sh /root/scripts;
#sudo gsutil cp gs://startup_scripts_us/scripts/go-wolk/docker-image-deployment.sh /root/scripts/;
#sudo gsutil cp gs://startup_scripts_us/scripts/go-wolk/enode.sh /root/scripts/;
#sudo gsutil cp gs://startup_scripts_us/scripts/go-wolk/go-wolk-docker-crontab /root/scripts/; #copying crontab at the end of the script
sudo chmod -R +x /root/scripts
sudo chown anand.anand -R /home/anand;

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
net.ipv4.tcp_tw_reuse = 1

# IPv4 Forwarding
net.ipv4.ip_forward = 1' >> /etc/sysctl.conf;
/sbin/sysctl -p;
EOF

#######

# installing Docker
sudo su - << EOF
yum -y install yum-utils device-mapper-persistent-data lvm2 &&
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo &&
yum -y install docker-ce &&
chkconfig docker on &&
systemctl start docker

# Fix docker firewald issue
nmcli connection modify docker0 connection.zone trusted
systemctl stop NetworkManager.service
firewall-cmd --permanent --zone=trusted --change-interface=docker0
systemctl start NetworkManager.service
nmcli connection modify docker0 connection.zone trusted
systemctl restart docker.service
EOF

# download/pull go-wolk-geth docker images and run:
if yum -q list docker-ce | grep docker-ce &> /dev/null; then

echo "
Docker successfully installed.. Proceed with pulling and deploying Docker image...
"
sudo docker pull wolkinc/go-wolk-phobos-nodejs;
sudo docker run --dns=8.8.8.8 --dns=8.8.4.4 --name=go-wolk-phobos-bootnode-2 --rm -dit -p 8545:8545 -p 30303:30303  -p 30303:30303/udp -p 30304:30304/udp wolkinc/go-wolk-phobos-nodejs

else
echo "Docker is not installed... Installing Docker...
"
sudo su - << EOF
yum -y install yum-utils device-mapper-persistent-data lvm2 &&
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo &&
yum -y install docker-ce &&
chkconfig docker on &&
systemctl restart docker.service

# Fix Docker firewalld issue:
nmcli connection modify docker0 connection.zone trusted
systemctl stop NetworkManager.service
firewall-cmd --permanent --zone=trusted --change-interface=docker0
systemctl start NetworkManager.service
nmcli connection modify docker0 connection.zone trusted
systemctl restart docker.service
EOF
fi

if ! docker images | grep wolkinc | grep -v grep &> /dev/null; then
sudo su - << EOF
docker pull wolkinc/go-wolk-phobos-nodejs;
EOF
else
echo "
Docker image pull successful...
"
fi 

if ! docker ps | grep wolkinc | grep -v grep &> /dev/null; then
sudo docker run --dns=8.8.8.8 --dns=8.8.4.4 --name=go-wolk-phobos-nodejs --rm -dit -p 8545:8545 -p 30303:30303  -p 30303:30303/udp -p 30304:30304/udp wolkinc/go-wolk-phobos-nodejs
else
echo "Docker image run successful...
"
fi

# adding crontab
#sudo gsutil cp gs://startup_scripts_us/scripts/go-wolk/go-wolk-docker-crontab /var/spool/cron/root;
