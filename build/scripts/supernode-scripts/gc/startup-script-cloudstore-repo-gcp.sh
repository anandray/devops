#!/bin/bash

project=`gcloud config list 2> /dev/null | grep project | cut -d"=" -f2 | cut -d " " -f2`

# create /root/scripts dir
sudo mkdir /root/scripts;

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;

sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-sudoers /etc/sudoers;
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-hosts /etc/hosts
#SSH Keys:
sudo gsutil -m cp gs://wolk-scripts/scripts/ssh_keys.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
sudo chmod 0400 /root/.ssh/authorized_keys*
sudo chown root.root /root/.ssh/authorized_keys*
sudo rm -rf /root/.ssh/ssh_keys.tgz

# Allow SSH-ing to any instance/server
sudo cp -rf /etc/ssh/ssh_config /etc/ssh/ssh_config-orig;
sudo cp -rf /etc/ssh/sshd_config /etc/ssh/sshd_config-orig;
sudo gsutil -m cp gs://wolk-scripts/scripts/ssh_config /etc/ssh/
sudo gsutil -m cp gs://wolk-scripts/scripts/sshd_config /etc/ssh/
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
sudo gsutil -m cp gs://wolk-scripts/scripts/histtimeformat.sh /etc/profile.d/

## DISABLE FSCK
sudo tune2fs -c 0 -i 0 /dev/sda1
sudo tune2fs -c 0 -i 0 /dev/sda3

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo gsutil -m cp gs://wolk-scripts/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil -m cp gs://wolk-scripts/scripts/security_limits.conf /etc/security/limits.conf
sudo gsutil -m cp gs://wolk-scripts/scripts/90-nproc.conf /etc/security/limits.d/

sudo yum -y install epel-release gcc make emacs ntpdate telnet screen git python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools mlocate wget sqlite-devel libsecret-devel vim;

sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo gsutil cp gs://wolk-scripts/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;

# syslog-ng
sudo yum -y install syslog-ng syslog-ng-libdbi libdbi-devel
cp /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-orig
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/syslog-ng/syslog-ng.conf /etc/syslog-ng/
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/syslog-ng/dmesg.conf /etc/syslog-ng/conf.d/
#service rsyslog stop
chkconfig rsyslog off
service syslog-ng restart
########

# download cloudstore binary
if [ ! -d /root/go/src/github.com/wolkdb/cloudstore ]; then
        sudo su - << EOF
       	mkdir -p /root/go/src/github.com/wolkdb
	cd /root/go/src/github.com/wolkdb
	git clone -b master --depth 1 --recurse-submodules git@github.com:wolkdb/cloudstore.git
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
sudo gsutil cp gs://wolk-scripts/scripts/profile.d/histtimeformat.sh /etc/profile.d/;
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-bashrc /root/.bashrc;
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-bashrc_aliases /root/.bashrc_aliases;
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/rc.local /etc/rc.d/rc.local
sudo gsutil cp gs://wolk-scripts/scripts/ssh_keys_chk.sh /root/scripts;
sudo gsutil -m cp gs://wolk-scripts/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/syslogtest /root/scripts/;
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-crontab-gcp-new /var/spool/cron/root
sudo service crond stop
sudo chkconfig crond on
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/df.sh /root/scripts/;
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/nrpe-install.sh /root/scripts/;
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/syslog-ng-start.sh /root/scripts/;
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-git-update-gcp-new.sh /root/scripts/;

sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/wolk.service /usr/lib/systemd/system/
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/make-wolk.sh /root/scripts/
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/make-wolk-cron.sh /root/scripts/
#sudo gsutil cp gs://wolk-scripts/scripts/plasma/google.json-$project /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json
sudo systemctl daemon-reload

# prepare for wolk and cloudstore
mkdir /usr/local/wolk /usr/local/cloudstore
chkconfig wolk on
chkconfig cloudstore off

#gsutil -m cp gs://wolk-scripts/scripts/plasma/*start.sh /root/scripts/;

sudo chmod -R +x /root/scripts

#Installing golang
if [ ! -d /usr/local/go ]; then
	sudo gsutil cp gs://wolk-scripts/scripts/go/go1.12.1.linux-amd64.tar.gz /usr/local;
	sudo tar -C /usr/local -xzf /usr/local/go1.12.1.linux-amd64.tar.gz;
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
gsutil -m cp -r gs://wolk-scripts/scripts/emacs/emacs/site-lisp /usr/share/emacs;
gsutil -m cp -r gs://wolk-scripts/scripts/emacs/.emacs.d /root/;
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
if grep "assumeyes=1" /etc/yum.conf; then
sudo su - << EOF
/bin/sed -i '/assumeyes/d' /etc/yum.conf
/bin/sed -i "$ i\assumeyes=0" /etc/yum.conf
EOF
fi

################

# download cloudstore git repo
if [ ! -d /root/go/src/github.com/wolkdb/cloudstore ]; then
        mkdir -p /root/go/src/github.com/wolkdb
	cd /root/go/src/github.com/wolkdb
        git clone -b master --depth 1 --recurse-submodules git@github.com:wolkdb/cloudstore.git
        cd /root/go/src/github.com/wolkdb/cloudstore
        git config --global user.name "Sourabh Niyogi"
        git config --global user.email "sourabh@wolk.com"
        git config user.name "Sourabh Niyogi"
        git config user.email "sourabh@wolk.com"
        git config core.filemode true
        git config --global core.filemode true
fi

#nrpe + snmpd
sudo yum -y install nagios-plugins nagios-plugins-nrpe nagios-common nrpe nagios-nrpe gd-devel net-snmp &&
sudo gsutil cp gs://wolk-scripts/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm . &&
sudo yum -y remove nrpe && rpm -Uvh nrpe-2.15-7.el6.x86_64.rpm &&
sudo gsutil cp gs://wolk-scripts/scripts/nagios/nrpe.cfg /etc/nagios/ &&
sudo gsutil cp gs://wolk-scripts/scripts/cacti/snmpd.conf /etc/snmp/ &&
sudo gsutil -m cp -r gs://wolk-scripts/scripts/nagios/plugins /usr/lib64/nagios/
sudo chmod +x /usr/lib64/nagios/plugins/* &&
sudo chkconfig nrpe on &&
sudo chkconfig snmpd on &&
sudo /sbin/service nrpe restart
sudo /sbin/service snmpd restart

#Uploading Certificate to /etc/ssl/certs/wildcard.wolk.com
sudo gsutil cp gs://wolk-scripts/scripts/certificate/www.wolk.com.crt /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt
sudo gsutil cp gs://wolk-scripts/scripts/certificate/www.wolk.com.key /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key

#Copy rc.local
sudo gsutil cp gs://wolk-scripts/scripts/rc.local /etc/rc.d/rc.local
chmod +x /etc/rc.d/rc.local

#sqlite3
if ! sqlite3 --version | grep 3.22; then
sudo rm -rf /usr/local/bin/sqlite3 /usr/bin/sqlite3
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/libsqlite3.la /usr/local/lib/
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/libsqlite3.so.0.8.6 /usr/local/lib/
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/libsqlite3.a /usr/local/lib/
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/sqlite3.conf /etc/ld.so.conf.d/
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/sqlite3 /usr/local/bin/
sudo chmod +x /usr/local/bin/sqlite3
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0
sudo ldconfig
fi
chmod +x /usr/local/bin/sqlite3
sudo gsutil -m cp gs://wolk-scripts/scripts/mysql/* /usr/bin/
sudo chmod +x /usr/bin/mysql* /usr/bin/db03.sh

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

# make wolk
#sudo chmod +x /root/scripts/make-wolk.sh
#sudo sh /root/scripts/make-wolk.sh

# remove cloudstore-git-update.sh from cron
sudo sed -i '/cloudstore-git-update/d' /var/spool/cron/root
