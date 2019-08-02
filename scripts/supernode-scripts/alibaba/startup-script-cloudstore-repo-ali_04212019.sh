#!/bin/bash
#Installing aliyun environment
#Install python and pip
sudo yum install wget curl -y
#wget https://www.python.org/ftp/python/2.7.8/Python-2.7.8.tgz
#tar -zxvf Python-2.7.8.tgz
#cd Python-2.7.8
#./configure
#make
#sudo make install
#curl "https://bootstrap.pypa.io/get-pip.py" -o "pip-install.py"
#sudo python pip-install.py
#sudo pip install -U pip
#sudo yum install python-pip python-devel -y
#sudo pip install aliyuncli
#pip install --upgrade aliyuncli
#sudo pip install aliyun-python-sdk-ecs
#sudo pip install aliyun-python-sdk-rds
#sudo pip install aliyun-python-sdk-slb
#aliyun-python-sdk-ossadmin
#sudo yum install -y zip unzip
wget http://gosspublic.alicdn.com/ossutil/1.4.2/ossutil64
sudo chmod 755 ossutil64
sudo cp ossutil64 /usr/local/bin/ossutil
echo "[Credentials]
language=EN
endpoint=oss-us-east-1.aliyuncs.com
accessKeyID=LTAIObIFiXh5Ks3M
accessKeySecret=qMFluv0YUOxnROsfaMppcmV7CeRxOg
" > /root/.ossutilconfig
# modifying sudoers
sudo sed -i 's/Defaults    secure_path/#Defaults    secure_path/g' /etc/sudoers

sudo sed -i '87 i\Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/root/go/src/github.com/wolkdb/plasma/build/scripts:/root/go/src/github.com/wolkdb/cloudstore/build/scripts:/root/go/src/github.com/wolkdb/plasma/build/bin:/root/go/src/github.com/wolkdb/cloudstore/build/bin:/root/.local/bin' /etc/sudoers

# create /root/scripts dir
sudo mkdir /root/scripts;

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;
sudo cp /etc/sudoers /etc/sudoers_orig
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/sudoers /etc/sudoers
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from ali storage:
sudo cp /etc/hosts /etc/hosts_Orig
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/hosts/hosts /etc/hosts

#SSH Keys:
sudo  ossutil cp  oss://wolk-ali/wolk-startup-scripts/scripts/ssh_keys-cloudstore.tgz /root/.ssh/ssh_keys-cloudstore.tgz
sudo tar zxvpf /root/.ssh/ssh_keys-cloudstore.tgz -C /root/.ssh/
sudo chmod 0400 /root/.ssh/authorized_keys*
sudo chown root.root /root/.ssh/authorized_keys*
sudo rm -rf /root/.ssh/ssh_keys-cloudstore.tgz

# Allow SSH-ing to any instance/server
sudo cp -rf /etc/ssh/ssh_config /etc/ssh/ssh_config-orig;
sudo cp -rf /etc/ssh/sshd_config /etc/ssh/sshd_config-orig;
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/ssh_config /etc/ssh/ssh_config
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/sshd_config /etc/ssh/sshd_config
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/histtimeformat.sh /etc/profile.d/histtimeformat.sh

## DISABLE FSCK
#sudo tune2fs -c 0 -i 0 /dev/sda1
#sudo tune2fs -c 0 -i 0 /dev/sda3

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/security_limits.conf /etc/security/limits.conf
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/90-nproc.conf /etc/security/limits.d/90-nproc.conf

sudo yum -y install epel-release gcc make emacs ntpdate telnet screen git python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools mlocate wget sqlite-devel vim;

sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/rpms/wandisco-git-release-7-2.noarch.rpm /root/scripts/wandisco-git-release-7-2.noarch.rpm  && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;

# adding log0 and log6 to /etc/hosts
if ! grep log0 /etc/hosts; then
echo '
35.193.168.171    log0' >> /etc/hosts
fi

# syslog-ng
sudo yum -y install syslog-ng syslog-ng-libdbi libdbi-devel
sudo cp /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-orig
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf
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
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/cloudstore/cloudstore-bashrc /root/.bashrc
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/ntpdate.sh /root/scripts/ntpdate.sh;
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/ntpdate2.sh /root/scripts/ntpdate2.sh;
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/syslogtest /root/scripts/syslogtest;
sudo ossutil cp -rf oss://wolk-ali/wolk-startup-scripts/scripts/plasma/.google /root/
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/cloudstore-crontab /var/spool/cron/root;
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/cloudstore-git-update.sh /root/scripts/cloudstore-git-update.sh;
sudo systemctl daemon-reload

# prepare for wolk and cloudstore
mkdir /usr/local/wolk /usr/local/cloudstore
chkconfig wolk on
chkconfig cloudstore off

sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/wolk-start.sh /root/scripts/wolk-start.sh

sudo chmod -R +x /root/scripts
#sudo chown anand.anand -R /home/anand;

#Installing golang
if [ ! -d /usr/local/go ]; then
        sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/go/go1.10.2.linux-amd64.tar.gz /usr/local/go1.10.2.linux-amd64.tar.gz
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
ossutil cp -rf oss://wolk-ali/wolk-startup-scripts/scripts/emacs/emacs/site-lisp /usr/share/emacs/site-lisp
ossutil cp -rf oss://wolk-ali/wolk-startup-scripts/scripts/emacs/.emacs.d /root/.emacs.d
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
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/cloudstore/cloudstore-bashrc /root/.bashrc;
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/ssh_keys_chk.sh /root/scripts/ssh_keys_chk.sh;
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/.google /root/.google
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/df.sh /root/scripts/df.sh
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/nrpe-install.sh /root/scripts/nrpe-install.sh
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/syslog-ng-start.sh /root/scripts/syslog-ng-start.sh
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/sql.service /usr/lib/systemd/system/sql.service
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/nosql.service /usr/lib/systemd/system/nosql.service
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/plasma.service /usr/lib/systemd/system/plasma.service
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/wolk.service /usr/lib/systemd/system/wolk.service
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/cloudstore.service /usr/lib/systemd/system/cloudstore.service
sudo systemctl daemon-reload

# prepare for wolk and cloudstore
mkdir /usr/local/wolk /usr/local/cloudstore
chkconfig wolk on
chkconfig cloudstore off

ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/nosql-start.sh /root/scripts/nosql-start.sh
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/plasma-start.sh /root/scripts/plasma-start.sh
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/plasma/sql-start.sh /root/scripts/sql-start.sh
sudo systemctl daemon-reload
sudo chmod -R +x /root/scripts
#sudo chown anand.anand -R /home/anand;

#Installing golang
if [ ! -d /usr/local/go ]; then
        sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/go/go1.10.2.linux-amd64.tar.gz /usr/local
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
EOF
#sudo sh /root/scripts/cloustore-git-update.sh

#######
#sqlite3
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/sqlite3/libsqlite3.la /usr/local/lib/libsqlite3.la
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/sqlite3/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0.8.6
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/sqlite3/libsqlite3.a /usr/local/lib/libsqlite3.a
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/sqlite3/sqlite3.conf /etc/ld.so.conf.d/sqlite3.conf
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/sqlite3/sqlite3 /usr/local/bin/sqlite3
sudo chmod +x /usr/local/bin/sqlite3
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0
sudo ldconfig
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/cloudstore/cloudstore-bashrc_aliases /root/.bashrc_aliases
#Nagios-nrpe
sudo yum -y install nagios-plugins nagios-plugins-nrpe nagios-common nrpe nagios-nrpe gd-devel net-snmp &&
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm /root/scripts/nrpe-2.15-7.el6.x86_64.rpm   &&
sudo yum -y remove nrpe && rpm -Uvh /root/scripts/nrpe-2.15-7.el6.x86_64.rpm &&
sudo ossutil cp -rf oss://wolk-ali/wolk-startup-scripts/scripts/nagios/plugins/ /usr/lib64/nagios/plugins/
sudo chmod +x /usr/lib64/nagios/plugins/* &&
sudo chkconfig nrpe on &&
sudo /sbin/service nrpe restart
chmod +x /usr/lib64/nagios/plugins/check_wolk_healthcheck*
chmod +x /usr/local/bin/sqlite3
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/nagios/nrpe.cfg /etc/nagios/nrpe.cfg
service nrpe restart
sudo ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/rc.local /etc/rc.d/rc.local
chmod +x /etc/rc.d/rc.local
if ! ps aux | grep nrpe | grep -v grep; then
/usr/sbin/nrpe -c /etc/nagios/nrpe.cfg -d
fi
/sbin/service crond stop
echo '* * * * * /root/scripts/cloudstore-git-update.sh &> /var/log/cloudstore-git-update.log' >> /var/spool/cron/root

