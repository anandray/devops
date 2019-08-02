#!/bin/bash

sudo mkdir /root/.aws
echo "[default]
output = json
region = ap-southeast-1" > /root/.aws/config

echo "[default]
aws_access_key_id = AKIAIMAXWBRFN5PPRCZA
aws_secret_access_key = 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" >  /root/.aws/credentials

# install aws cli
curl -O https://bootstrap.pypa.io/get-pip.py
python get-pip.py --user
/root/.local/bin/pip install awscli --upgrade --user
/root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore-bashrc /root/.bashrc
/root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore-bashrc_aliases /root/.bashrc_aliases
source /root/.bashrc

# modifying sudoers
sudo sed -i 's/Defaults    secure_path/#Defaults    secure_path/g' /etc/sudoers

sudo sed -i '87 i\Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/root/go/src/github.com/wolkdb/plasma/build/scripts:/root/go/src/github.com/wolkdb/cloudstore/build/scripts:/root/go/src/github.com/wolkdb/plasma/build/bin:/root/go/src/github.com/wolkdb/cloudstore/build/bin:/root/.local/bin' /etc/sudoers

# create /root/scripts dir
sudo mkdir /root/scripts;

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;

sudo aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/cloudstore-sudoers-aws /etc/sudoers;
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma-hosts-aws /etc/hosts

#SSH Keys:
sudo aws s3 cp s3://wolk-startup-scripts/scripts/ssh_keys-cloudstore.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys-cloudstore.tgz -C /root/.ssh/
sudo chmod 0400 /root/.ssh/authorized_keys*
sudo chown root.root /root/.ssh/authorized_keys*
sudo rm -rf /root/.ssh/ssh_keys-cloudstore.tgz

# Allow SSH-ing to any instance/server
sudo cp -rf /etc/ssh/ssh_config /etc/ssh/ssh_config-orig;
sudo cp -rf /etc/ssh/sshd_config /etc/ssh/sshd_config-orig;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/ssh_config /etc/ssh/
sudo aws s3 cp s3://wolk-startup-scripts/scripts/sshd_config /etc/ssh/
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
sudo aws s3 cp s3://wolk-startup-scripts/scripts/histtimeformat.sh /etc/profile.d/

## DISABLE FSCK
sudo tune2fs -c 0 -i 0 /dev/sda1
sudo tune2fs -c 0 -i 0 /dev/sda3

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo aws s3 cp s3://wolk-startup-scripts/scripts/security_limits.conf /etc/security/limits.conf
sudo aws s3 cp s3://wolk-startup-scripts/scripts/90-nproc.conf /etc/security/limits.d/

sudo yum -y install epel-release gcc make emacs ntpdate telnet screen git python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools mlocate wget sqlite-devel vim;

sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo aws s3 cp s3://wolk-startup-scripts/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;

# adding log0 and log6 to /etc/hosts
if ! grep log0 /etc/hosts; then
echo '
35.193.168.171    log0' >> /etc/hosts
fi

# syslog-ng
sudo wget -O /etc/yum.repos.d/czanik-syslog-ng319-epel-7.repo https://copr.fedorainfracloud.org/coprs/czanik/syslog-ng319/repo/epel-7/czanik-syslog-ng319-epel-7.repo
sudo yum -y install syslog-ng syslog-ng-libdbi libdbi-devel librabbitmq
systemctl enable syslog-ng
cp /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-orig
aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf
#service rsyslog stop
chkconfig rsyslog off
#service syslog-ng restart
########

# download cloudstore repository
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
sudo aws s3 cp s3://wolk-startup-scripts/scripts/profile.d/histtimeformat.sh /etc/profile.d/;
#sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore-bashrc /root/.bashrc
sudo aws s3 cp s3://wolk-startup-scripts/scripts/ssh_keys_chk.sh /root/scripts;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/ntpdate.sh /root/scripts/;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/ntpdate2.sh /root/scripts/;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/syslogtest /root/scripts/;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/.google /root/ --recursive;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore-crontab-aws /var/spool/cron/root;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/df.sh /root/scripts/;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore-git-update.sh /root/scripts/;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/nrpe-install.sh /root/scripts/;
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/syslog-ng-start.sh /root/scripts/;

sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/sql.service /usr/lib/systemd/system/
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/nosql.service /usr/lib/systemd/system/
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/plasma.service /usr/lib/systemd/system/
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/wolk.service /usr/lib/systemd/system/
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore.service /usr/lib/systemd/system/
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore.toml /root/go/src/github.com/wolkdb/cloudstore/cloudstore.toml
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/wolk-start.sh /root/scripts/
sudo aws s3 cp s3://wolk-startup-scripts/scripts/plasma/.aws /root/ --recursive;
sudo systemctl daemon-reload

# prepare for wolk and cloudstore
mkdir /usr/local/wolk /usr/local/cloudstore
chkconfig wolk on
chkconfig cloudstore on

aws s3 cp s3://wolk-startup-scripts/scripts/plasma/nosql-start.sh /root/scripts/;
aws s3 cp s3://wolk-startup-scripts/scripts/plasma/plasma-start.sh /root/scripts/;
aws s3 cp s3://wolk-startup-scripts/scripts/plasma/sql-start.sh /root/scripts/;
aws s3 cp s3://wolk-startup-scripts/scripts/plasma/syslog-ng-start.sh /root/scripts/;
aws s3 cp s3://wolk-startup-scripts/scripts/plasma/wolk-start.sh /root/scripts/;

sudo chmod -R +x /root/scripts
sudo chown anand.anand -R /home/anand;

#Installing golang
if [ ! -d /usr/local/go ]; then
	sudo aws s3 cp s3://wolk-startup-scripts/scripts/go/go1.10.2.linux-amd64.tar.gz /usr/local;
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
aws s3 cp s3://wolk-startup-scripts/scripts/emacs/emacs/site-lisp /usr/share/emacs --recursive;
aws s3 cp s3://wolk-startup-scripts/scripts/emacs/.emacs.d /root/ --recursive;
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
fi

## compile wolk/cloudstore
##make wolk #not necessary - fetched from git repo
#sudo su - << EOF
#aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore.toml /root/go/src/github.com/wolkdb/cloudstore/cloudstore.toml;
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
sudo aws s3 cp s3://wolk-startup-scripts/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm . &&
sudo yum -y remove nrpe && rpm -Uvh nrpe-2.15-7.el6.x86_64.rpm &&
sudo aws s3 cp s3://wolk-startup-scripts/scripts/nagios/nrpe.cfg /etc/nagios/ &&
sudo aws s3 cp s3://wolk-startup-scripts/scripts/nagios/plugins/ /usr/lib64/nagios/plugins/ --recursive &&
sudo chmod +x /usr/lib64/nagios/plugins/* &&
sudo chkconfig nrpe on &&
sudo /sbin/service nrpe restart

# toml + datastore credentials
region=`cat ~/.aws/config | grep region | head -n1 | awk '{print$NF}'`
AmazonCredentials="/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/credentials"

sudo cp /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/wolk.toml-aws-template /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i 's/"_AmazonRegion"/"'$region'"/g' /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i 's|"_AmazonCredentials"|"'$AmazonCredentials'"|g' /root/go/src/github.com/wolkdb/cloudstore/wolk.toml # using '|' instead of '/' because of error: "sed: -e expression #1, char 35: unknown option to `s'"

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
