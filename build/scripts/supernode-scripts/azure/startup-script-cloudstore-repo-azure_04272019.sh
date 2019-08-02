#!/bin/bash
#Install azcopy
yum install -y libunwind icu
yum install -y wget
wget -O azcopy.tar.gz https://aka.ms/downloadazcopylinux64
tar -xf azcopy.tar.gz
sudo ./install.sh

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
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/sudoers --destination /etc/sudoers --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from ali storage:
sudo cp /etc/hosts /etc/hosts_Orig
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/hosts/hosts --destination /etc/hosts --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet

#SSH Keys:
sudo  azcopy --source  https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/ssh_keys-cloudstore.tgz --destination /root/.ssh/ssh_keys-cloudstore.tgz --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo tar zxvpf /root/.ssh/ssh_keys-cloudstore.tgz -C /root/.ssh/
sudo chmod 0400 /root/.ssh/authorized_keys*
sudo chown root.root /root/.ssh/authorized_keys*
sudo rm -rf /root/.ssh/ssh_keys-cloudstore.tgz

# Allow SSH-ing to any instance/server
sudo cp -rf /etc/ssh/ssh_config /etc/ssh/ssh_config-orig;
sudo cp -rf /etc/ssh/sshd_config /etc/ssh/sshd_config-orig;
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/ssh_config --destination /etc/ssh/ssh_config --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/sshd_config --destination /etc/ssh/sshd_config --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/histtimeformat.sh --destination /etc/profile.d/histtimeformat.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet 

## DISABLE FSCK
#sudo tune2fs -c 0 -i 0 /dev/sda1
#sudo tune2fs -c 0 -i 0 /dev/sda3

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/selinux_config --destination /etc/selinux/config --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/security_limits.conf --destination /etc/security/limits.conf --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/90-nproc.conf --destination /etc/security/limits.d/90-nproc.conf --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet 

sudo yum -y install epel-release gcc make emacs ntpdate telnet screen git python-argparse *whois openssl openssl-devel java-1.8.0-openjdk-devel python27-pip lynx net-tools mlocate wget sqlite-devel vim;

sudo rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/rpms/wandisco-git-release-7-2.noarch.rpm --destination /root/scripts/wandisco-git-release-7-2.noarch.rpm  && rpm -Uvh /root/scripts/wandisco-git-release-7-2.noarch.rpm;

# adding log0 and log6 to /etc/hosts
if ! grep log0 /etc/hosts; then
echo '
35.224.9.111    log0' >> /etc/hosts
fi

# syslog-ng
sudo yum -y install syslog-ng syslog-ng-libdbi libdbi-devel
sudo cp /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-orig
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/syslog-ng.conf --destination /etc/syslog-ng/syslog-ng.conf --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
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
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/cloudstore/cloudstore-bashrc --destination /root/.bashrc --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/ntpdate.sh --destination /root/scripts/ntpdate.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet;
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/ntpdate2.sh --destination /root/scripts/ntpdate2.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet;
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/syslogtest --destination /root/scripts/syslogtest --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet;
sudo azcopy --source  https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/.google --destination /root/ --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --recursive
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/cloudstore-crontab --destination /var/spool/cron/root --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet;
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/cloudstore-git-update.sh --destination /root/scripts/cloudstore-git-update.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet;
sudo systemctl daemon-reload

# prepare for wolk and cloudstore
mkdir /usr/local/wolk /usr/local/cloudstore
chkconfig wolk on
chkconfig cloudstore off

sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/wolk-start.sh --destination /root/scripts/wolk-start.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet

sudo chmod -R +x /root/scripts
#sudo chown anand.anand -R /home/anand;

#Installing golang
if [ ! -d /usr/local/go ]; then
	sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/go/go1.10.2.linux-amd64.tar.gz --destination /usr/local/go1.10.2.linux-amd64.tar.gz --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
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
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/emacs/emacs/site-lisp --destination /usr/share/emacs/site-lisp --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --recursive
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/emacs/.emacs.d --destination /root/.emacs.d --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --recursive
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
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/cloudstore/cloudstore-bashrc --destination /root/.bashrc --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet;
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/ssh_keys_chk.sh --destination /root/scripts/ssh_keys_chk.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet;
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/.google --destination /root/.google --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --recursive
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/df.sh --destination /root/scripts/df.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/nrpe-install.sh --destination /root/scripts/nrpe-install.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/syslog-ng-start.sh --destination /root/scripts/syslog-ng-start.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/sql.service --destination /usr/lib/systemd/system/sql.service --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/nosql.service --destination /usr/lib/systemd/system/nosql.service --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/plasma.service --destination /usr/lib/systemd/system/plasma.service --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/wolk.service --destination /usr/lib/systemd/system/wolk.service --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/cloudstore.service --destination /usr/lib/systemd/system/cloudstore.service --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo systemctl daemon-reload

# prepare for wolk and cloudstore
mkdir /usr/local/wolk /usr/local/cloudstore
chkconfig wolk on
chkconfig cloudstore off

azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/nosql-start.sh --destination /root/scripts/nosql-start.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/plasma-start.sh --destination /root/scripts/plasma-start.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/sql-start.sh --destination /root/scripts/sql-start.sh --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo systemctl daemon-reload
sudo chmod -R +x /root/scripts
#sudo chown anand.anand -R /home/anand;


#######
#sqlite3
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/sqlite3/libsqlite3.la --destination /usr/local/lib/libsqlite3.la --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/sqlite3/libsqlite3.so.0.8.6 --destination /usr/local/lib/libsqlite3.so.0.8.6 --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/sqlite3/libsqlite3.a --destination /usr/local/lib/libsqlite3.a --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/sqlite3/sqlite3.conf --destination /etc/ld.so.conf.d/sqlite3.conf --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/sqlite3/sqlite3 --destination /usr/local/bin/sqlite3 --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo chmod +x /usr/local/bin/sqlite3
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0
sudo ldconfig
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/cloudstore/cloudstore-bashrc_aliases --destination /root/.bashrc_aliases --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
#Nagios-nrpe
sudo yum -y install nagios-plugins nagios-plugins-nrpe nagios-common nrpe nagios-nrpe gd-devel net-snmp &&
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm --destination /root/scripts/nrpe-2.15-7.el6.x86_64.rpm --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet  &&
sudo yum -y remove nrpe && rpm -Uvh /root/scripts/nrpe-2.15-7.el6.x86_64.rpm &&
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/nagios/plugins/ --destination /usr/lib64/nagios/plugins/ --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --recursive --quiet
sudo chmod +x /usr/lib64/nagios/plugins/* &&
sudo chkconfig nrpe on &&
sudo /sbin/service nrpe restart
chmod +x /usr/lib64/nagios/plugins/check_wolk_healthcheck*
chmod +x /usr/local/bin/sqlite3
azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/nagios/nrpe.cfg --destination /etc/nagios/nrpe.cfg --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
service nrpe restart
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/rc.local --destination /etc/rc.d/rc.local --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
chmod +x /etc/rc.d/rc.local
#copy certificate
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/certificate/www.wolk.com.crt --destination /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
sudo azcopy --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/certificate/www.wolk.com.key --destination /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key --source-key CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ== --quiet
if ! ps aux | grep nrpe | grep -v grep; then
/usr/sbin/nrpe -c /etc/nagios/nrpe.cfg -d
fi
/sbin/service crond stop
echo '* * * * * /root/scripts/cloudstore-git-update.sh &> /var/log/cloudstore-git-update.log' >> /var/spool/cron/root
history -c
