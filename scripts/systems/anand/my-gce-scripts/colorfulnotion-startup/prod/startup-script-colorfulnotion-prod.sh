#!/bin/bash

# create /root/scripts dir
sudo mkdir /root/scripts;

# install gcloud and activate service account
#sudo tee -a /etc/yum.repos.d/google-cloud-sdk.repo << EOM
#[google-cloud-sdk]
#name=Google Cloud SDK
#baseurl=https://packages.cloud.google.com/yum/repos/cloud-sdk-el7-x86_64
#enabled=1
#gpgcheck=1
#repo_gpgcheck=1
#gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
#       https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
#EOM

#yum -y install google-cloud-sdk wget
#exec -l $SHELL

#wget -O /root/scripts/api-prd-colorfulnotion-2018-bigtable.json http://www6001.wolk.com/.start/api-prd-colorfulnotion-2018-bigtable.json
#gsutil cp gs://startup-scripts-colorfulnotion/scripts/api-prd-colorfulnotion-2018-bigtable.json /root/scripts/
#gcloud auth activate-service-account --key-file /root/scripts/api-prd-colorfulnotion-2018-bigtable.json

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;

#SSH Keys:
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/ssh_keys.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
sudo rm -rf /root/.ssh/ssh_keys.tgz

# Allow SSH-ing to any instance/server
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/ssh_config /etc/ssh/
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/sshd_config /etc/ssh/
sudo service sshd restart

# Enable histtimeformat
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/histtimeformat.sh /etc/profile.d/

## DISABLE FSCK
sudo tune2fs -c 0 -i 0 /dev/sda1

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/security_limits.conf /etc/security/limits.conf
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/90-nproc.conf /etc/security/limits.d/

sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;

sudo yum -y install epel-release gcc make emacs ntpdate telnet screen git python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools mlocate wget sqlite-devel;

# install gcloud cbt
yum makecache &&  yum -y install kubectl google-cloud-sdk google-cloud-sdk-app-engine-grpc google-cloud-sdk-pubsub-emulator google-cloud-sdk-app-engine-go google-cloud-sdk-cloud-build-local google-cloud-sdk-datastore-emulator google-cloud-sdk-app-engine-python google-cloud-sdk-cbt google-cloud-sdk-bigtable-emulator google-cloud-sdk-datalab google-cloud-sdk-app-engine-java

wget -O /root/.cbtrc http://www6001.wolk.com/.start/.cbtrc
#gsutil cp gs://startup-scripts-colorfulnotion/scripts/.cbtrc /root/

# adding log0 and log6 to /etc/hosts
if ! grep log0 /etc/hosts; then
echo '
35.193.168.171    log0' >> /etc/hosts
fi

# syslog-ng
sudo yum -y install syslog-ng syslog-ng-libdbi libdbi-devel
cp /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-orig
wget -O /etc/syslog-ng/syslog-ng.conf http://www6001.wolk.com/.start/syslog-ng.conf
#service rsyslog stop
chkconfig rsyslog off
#service syslog-ng restart
########

# create 'goracing' user/group
sudo /usr/sbin/useradd -d /home/goracing -m goracing
# download goracing binary
if [ ! -d /home/goracing/go_projects/src/goracing.colorfulnotion.com ]; then
        sudo su - goracing << EOF
	wget -O ~/.cbtrc http://www6001.wolk.com/.start/.cbtrc
	gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/id_rsa* ~goracing/.ssh/
	gsutil cp gs://startup-scripts-colorfulnotion/scripts/authorized_keys ~goracing/.ssh/
	chmod 0400 ~goracing/.ssh/id_rsa*; chmod 0400 ~goracing/.ssh/authorized_keys
	chattr +i ~goracing/.ssh/authorized_keys
        if [ ! -d /home/goracing/go_projects/src ]; then
                gsutil cp gs://startup-scripts-colorfulnotion/scripts/goracing.tgz /tmp/
                tar zxvpf /tmp/goracing.tgz -C /tmp
         rsync -avz /tmp/goracing/ /home/goracing/              
        fi
	cd /home/goracing/go_projects/src
	git clone --recurse-submodules -b master git@github.com:colorfulnotion/goracing.colorfulnotion.com.git
	cd /home/goracing/go_projects/src/goracing.colorfulnotion.com
	git config core.filemode true
	git config --global core.filemode true
	mkdir -p /home/goracing/raceFolder/serverLogs
	source /home/goracing/.bashrc
EOF
fi

#############

# copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/profile.d/histtimeformat.sh /etc/profile.d/;
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/colorfulnotion-bashrc-repo /root/.bashrc
sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/ssh_keys_chk.sh /root/scripts;
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/goracing-prod-*start.sh /root/scripts
sudo gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/crontab-prod /var/spool/cron/root
sudo chmod +x /root/scripts/*

sudo chmod -R +x /root/scripts
sudo chown anand.anand -R /home/anand;

#Installing golang
if [ ! -d /usr/local/go ]; then
	sudo gsutil cp gs://startup-scripts-colorfulnotion/scripts/go/go1.10.3.linux-amd64.tar.gz /usr/local;
	sudo tar -C /usr/local -xzf /usr/local/go1.10.3.linux-amd64.tar.gz;
fi

#Adding environment variables to /root/.bashrc
if ! sudo grep GOPATH /root/.bashrc; then
sudo su - << EOF
echo '
export PATH="$PATH:/usr/local/go/bin"
export GOPATH=/root
export GOROOT=/home/goracing/go' >> /root/.bashrc
source /root/.bashrc
EOF
fi

# emacs for go-lang
sudo su - << EOF
gsutil -m cp -r gs://startup-scripts-colorfulnotion/scripts/emacs/emacs/site-lisp /usr/share/emacs;
gsutil -m cp -r gs://startup-scripts-colorfulnotion/scripts/emacs/.emacs.d /root/;
EOF

#######

# net.ipv4.tcp_tw_recycle and net.ipv4.tcp_tw_reuse
sudo su - << EOF
echo 1 > /proc/sys/net/ipv4/tcp_tw_reuse;
echo 1 > /proc/sys/net/ipv4/tcp_tw_recycle;
sed -i '/ipv4.tcp_tw/d' /etc/sysctl.conf
sed -i '/Recycle and Reuse TIME_WAIT sockets faster/d' /etc/sysctl.conf
sed -i '/Too many open files/d' /etc/sysctl.conf
sed -i '/fs.file-max/d' /etc/sysctl.conf
sed -i '/fs.nr_open/d' /etc/sysctl.conf
echo '
# Recycle and Reuse TIME_WAIT sockets faster
net.ipv4.tcp_tw_recycle = 1
net.ipv4.tcp_tw_reuse = 1
# Too many open files
fs.file-max = 1000000
fs.nr_open = 2000000' >> /etc/sysctl.conf;
/sbin/sysctl -p;
EOF

################

# applying bashrc
exec -l $SHELL
source /root/.bashrc

# download plasma git repo
if [ ! -d /home/goracing/go_projects/src/goracing.colorfulnotion.com ]; then
        sudo su - goracing << EOF
        gsutil -m cp gs://startup-scripts-colorfulnotion/scripts/id_rsa* ~goracing/.ssh/
	gsutil cp gs://startup-scripts-colorfulnotion/scripts/authorized_keys ~goracing/.ssh/
        chmod 0400 ~goracing/.ssh/id_rsa*; chmod 0400 ~goracing/.ssh/authorized_keys
	chattr +i ~goracing/.ssh/authorized_keys
	if [ ! -d /home/goracing/go_projects/src ]; then
	        gsutil cp gs://startup-scripts-colorfulnotion/scripts/goracing.tgz /tmp/
	        tar zxvpf /tmp/goracing.tgz -C /tmp
       	 rsync -avz /tmp/goracing/ /home/goracing/		
	fi
        cd /home/goracing/go_projects/src
        git clone --recurse-submodules -b master git@github.com:colorfulnotion/goracing.colorfulnotion.com.git
        cd /home/goracing/go_projects/src/goracing.colorfulnotion.com
        git config core.filemode true
        git config --global core.filemode true
	mkdir -p /home/goracing/raceFolder/serverLogs
        source /home/goracing/.bashrc
EOF
fi

# removing postfix
yum -y remove postfix

# start goracing production
#sh /root/scripts/goracing-prod-start.sh

# add ssh key for goracing ssh login
gsutil cp gs://startup-scripts-colorfulnotion/scripts/authorized_keys ~goracing/.ssh/
chmod 0400 ~goracing/.ssh/authorized_keys
chattr +i ~goracing/.ssh/authorized_keys

# cleanup
rm -rf /tmp/goracing*

sleep 15
exec -l $SHELL

# stopping rsyslog to start syslog-ng
service rsyslog stop
chkconfig rsyslog off
service syslog-ng restart

