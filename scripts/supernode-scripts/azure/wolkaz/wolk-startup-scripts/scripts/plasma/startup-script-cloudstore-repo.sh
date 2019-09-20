#!/bin/bash

project=`gcloud config list 2> /dev/null | grep project | cut -d"=" -f2 | cut -d " " -f2`

# create /root/scripts dir
sudo mkdir /root/scripts;

## install gcloud and activate service account

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

#mkdir /root/.google; wget -O /root/.google/crosschannel-1307-520dd999e93e.json http://www6001.wolk.com/.start/crosschannel-1307-520dd999e93e.json
#gcloud auth activate-service-account --key-file /root/.google/crosschannel-1307-520dd999e93e.json

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;

sudo gsutil cp gs://startup_scripts_us/scripts/swarm/swarm-sudoers /etc/sudoers;
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
sudo gsutil -m cp gs://startup_scripts_us/scripts/plasma-hosts /etc/hosts

#SSH Keys:
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys-cloudstore.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys-cloudstore.tgz -C /root/.ssh/
sudo chmod 0400 /root/.ssh/authorized_keys*
sudo chown root.root /root/.ssh/authorized_keys*
sudo rm -rf /root/.ssh/ssh_keys-cloudstore.tgz

# Allow SSH-ing to any instance/server
sudo cp -rf /etc/ssh/ssh_config /etc/ssh/ssh_config-orig;
sudo cp -rf /etc/ssh/sshd_config /etc/ssh/sshd_config-orig;
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

sudo yum -y install epel-release gcc make emacs ntpdate telnet screen git python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools mlocate wget sqlite-devel vim;

sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo gsutil cp gs://startup_scripts_us/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;

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

# download cloudstore binary
if [ ! -d /root/go/src/github.com/wolkdb/cloudstore ]; then
        sudo su - << EOF
       	mkdir -p /root/go/src/github.com/wolkdb
	cd /root/go/src/github.com/wolkdb
	git clone --recurse-submodules git@github.com:wolkdb/cloudstore.git
	cd /root/go/src/github.com/wolkdb/cloudstore
	git config --global user.name "anand ray"
	git config --global user.email "anand@wolk.com"
        git config user.name "anand ray"
        git config user.email "anand@wolk.com"
	git config core.filemode true
	git config --global core.filemode true
EOF
fi

#############

# copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil cp gs://startup_scripts_us/scripts/profile.d/histtimeformat.sh /etc/profile.d/;
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/plasma-bashrc-repo /root/.bashrc
sudo gsutil cp gs://startup_scripts_us/scripts/ssh_keys_chk.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/syslogtest /root/scripts/;
sudo gsutil -m cp -r gs://startup_scripts_us/scripts/plasma/.google /root/;
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/cloudstore-crontab /var/spool/cron/root;
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/df.sh /root/scripts/;
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/nrpe-install.sh /root/scripts/;
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/syslog-ng-start.sh /root/scripts/;

sudo gsutil cp gs://startup_scripts_us/scripts/plasma/sql.service /usr/lib/systemd/system/
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/nosql.service /usr/lib/systemd/system/
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/plasma.service /usr/lib/systemd/system/
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/wolk.service /usr/lib/systemd/system/
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/cloudstore.service /usr/lib/systemd/system/
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/cloudstore.toml /root/go/src/github.com/wolkdb/cloudstore/cloudstore.toml
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/wolk-start.sh /root/scripts/
sudo gsutil cp gs://startup_scripts_us/scripts/plasma/google.json-$project /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json
sudo systemctl daemon-reload

# prepare for wolk and cloudstore
mkdir /usr/local/wolk /usr/local/cloudstore
chkconfig wolk on
chkconfig cloudstore on

gsutil -m cp gs://startup_scripts_us/scripts/plasma/*start.sh /root/scripts/;

sudo chmod -R +x /root/scripts
sudo chown anand.anand -R /home/anand;

#Installing golang
if [ ! -d /usr/local/go ]; then
	sudo gsutil cp gs://startup_scripts_us/scripts/go/go1.10.2.linux-amd64.tar.gz /usr/local;
	sudo tar -C /usr/local -xzf /usr/local/go1.10.2.linux-amd64.tar.gz;
fi

#Adding environment variables to /root/.bashrc
if ! sudo grep GOPATH /root/.bashrc; then
sudo su - << EOF
echo '
export PATH="$PATH:/usr/local/go/bin"
export GOPATH=/root/go
export GOROOT=/usr/local/go' >> /root/.bashrc
EOF
fi

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

#######

#Change 'assumeyes=1' --> 'assumeyes=0' in yum.conf
sudo su - << EOF
/bin/sed -i '/assumeyes/d' /etc/yum.conf
/bin/sed -i "$ i\assumeyes=0" /etc/yum.conf
EOF

################

# exporting GOOGLE_APPLICATION_CREDENTIALS
export GOOGLE_APPLICATION_CREDENTIALS="/root/.google/crosschannel-1307-520dd999e93e.json"

# download cloudstore git repo
if [ ! -d /root/go/src/github.com/wolkdb/cloudstore ]; then
        mkdir -p /root/go/src/github.com/wolkdb
	cd /root/go/src/github.com/wolkdb
        git clone --recurse-submodules git@github.com:wolkdb/cloudstore.git
        cd /root/go/src/github.com/wolkdb/cloudstore
        git config --global user.name "Sourabh Niyogi"
        git config --global user.email "sourabh@wolk.com"
        git config user.name "Sourabh Niyogi"
        git config user.email "sourabh@wolk.com"
        git config core.filemode true
        git config --global core.filemode true
        echo "export PATH=$PATH:/root/go/src/github.com/wolkdb/cloudstore/build/bin" >> /root/.bashrc
fi

## compile wolk/cloudstore
##make wolk #not necessary - fetched from git repo
#sudo su - << EOF
#gsutil cp gs://startup_scripts_us/scripts/plasma/cloudstore.toml /root/go/src/github.com/wolkdb/cloudstore/cloudstore.toml;
#if [ ! -f /root/go/src/github.com/wolkdb/cloudstore/build/bin/cloudstore ]; then
#cd /root/go/src/github.com/wolkdb/cloudstore
#export PATH="$PATH:/usr/local/go/bin"
#export GOPATH=/root/go
#export GOROOT=/usr/local/go
#make cloudstore
#fi
#EOF

#nrpe
sudo yum -y install nagios-plugins nagios-plugins-nrpe nagios-common nrpe nagios-nrpe gd-devel net-snmp &&
sudo gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm . &&
sudo yum -y remove nrpe && rpm -Uvh nrpe-2.15-7.el6.x86_64.rpm &&
sudo gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/ &&
sudo gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/ &&
sudo chmod +x /usr/lib64/nagios/plugins/* &&
sudo chkconfig nrpe on &&
sudo /sbin/service nrpe restart

# toml + datastore credentials
# Enable API [cloudresourcemanager.googleapis.com] on project
"yes" | gcloud projects describe $project | grep projectNumber | cut -d "'" -f2 | awk '{print$1"-compute@developer.gserviceaccount.com"}'
project=`gcloud config list 2> /dev/null | grep project | cut -d"=" -f2 | cut -d " " -f2`
serviceAccount=`gcloud projects describe $project | grep projectNumber | cut -d "'" -f2 | awk '{print$1"-compute@developer.gserviceaccount.com"}'`
region=`gcloud compute instance-groups list | grep -v NAME | awk '{print$2}' | cut -d "1" -f1`
ConsensusIdx=`gcloud compute instance-groups list | grep wolk | cut -d "-" -f2`
GoogleDatastoreCredentials="/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json"
Provider1=`echo $GoogleDatastoreCredentials | cut -d "/" -f11 | cut -d "." -f1`
Provider2=`gcloud compute instance-groups list | grep wolk | cut -d "-" -f6 | awk '{print$1}'`
Provider=`echo "$Provider1"_"$Provider2"`

sudo cp /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/wolk.toml-template /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i 's/"_GoogleDatastoreProject"/"'$project'"/g' /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i 's/"_Region"/"'$region'"/g' /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i 's/_ConsensusIdx/'$ConsensusIdx'/g' /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i 's/"_Provider"/"'$Provider'"/g' /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i 's|"_GoogleDatastoreCredentials"|"'$GoogleDatastoreCredentials'"|g' /root/go/src/github.com/wolkdb/cloudstore/wolk.toml # using '|' instead of '/' because of error: "sed: -e expression #1, char 35: unknown option to `s'"

if [ ! -f /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json ]; then
"yes" | gcloud iam service-accounts keys create /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json --iam-account $serviceAccount
sudo gsutil cp /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json gs://startup_scripts_us/scripts/plasma/google.json-$project
    elif [ ! -s /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json ]; then
    "yes" | gcloud iam service-accounts keys create /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json --iam-account $serviceAccount
    sudo gsutil cp /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json gs://startup_scripts_us/scripts/plasma/google.json-$project
fi

# starting wolk
if ! ps aux | grep wolk | grep -v grep; then
echo "
WOLK is NOT running. Starting WOLK...
"
sudo systemctl start wolk.service
fi

# stopping rsyslog to start syslog-ng
if ps aux | grep rsyslogd | grep -v grep; then
sudo service rsyslog stop
sudo chkconfig rsyslog off
sudo /sbin/service syslog-ng restart
fi

# clean yum db
sudo yum clean all

#stop/disable yum-cron
sudo systemctl stop yum-cron.service
sudo chkconfig yum-cron off

#if [ ! -f /root/go/src/github.com/wolkdb/cloudstore/build/bin/cloudstore ]; then
#cd /root/go/src/github.com/wolkdb/cloudstore
#export PATH="$PATH:/usr/local/go/bin"
#export GOPATH=/root/go
#export GOROOT=/usr/local/go
#make cloudstore
#fi

if ! ps aux | grep wolk | grep -v grep; then
echo "
WOLK is NOT running. Starting WOLK...
"
/root/scripts/wolk-start.sh
fi

exec -l $SHELL
source ~/.bashrc