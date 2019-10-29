#!/bin/bash

#PST date time

if ! rpm -qa | grep ntpdate > /dev/null; then
sudo yum -y install ntpdate;
fi

if ! rpm -qa | grep rdate > /dev/null; then
sudo yum -y install rdate;
fi

if ! date | grep -E 'PST|PDT';
then
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo service ntpd stop;
sudo ntpdate -u -b pool.ntp.org;
sudo service ntpd start;

else
sudo service ntpd stop;
sudo ntpdate -u -b pool.ntp.org;
sudo service ntpd start;
fi

## Start /etc/sudoers modifications ##
# THE FOLLOWING /etc/sudoers MODIFICATIONS ARE REQUIRED TO ALLOW 'gcloud' and 'gsutil' to run using sudo, WHICH IS OTHERWISE NOT ALLOWED
# Replace /etc/sudoers default 'Defaults secure_path' with following:
#Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin

# Comment out this line in /etc/sudoers:
#Defaults    requiretty
if ! grep '#Defaults    requiretty' /etc/sudoers | grep '\#' > /dev/null; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/sudoers /etc/

fi
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
if ! grep -q db /etc/hosts; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/hosts /etc/;
fi

#SSH Keys:
if [ ! -f /root/.ssh/authorized_keys ]; then
        sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
        sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
        sudo rm -rf /root/.ssh/ssh_keys.tgz
fi

# Allow SSH-ing to any instance/server
if ! grep 'StrictHostKeyChecking no' /etc/ssh/ssh_config; then

        sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_config /etc/ssh/;
        sudo service sshd restart;
fi

if ! grep 'PermitRootLogin yes' /etc/ssh/sshd_config; then

        sudo gsutil -m cp gs://startup_scripts_us/scripts/sshd_config /etc/ssh/
        sudo service sshd restart;
fi
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
if ! ls /etc/profile.d/histtimeformat.sh; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/histtimeformat.sh /etc/profile.d/
fi

## DISABLE FSCK
if ! tune2fs -l /dev/sda1 | grep -i 'Maximum mount count:      -1'; then
sudo tune2fs -c 0 -i 0 /dev/sda1;
fi

if ! tune2fs -l /dev/sda1 | grep -i 'Check interval:           0'; then
sudo tune2fs -c 0 -i 0 /dev/sda1;
fi

#DISABLE SELINUX:
if ! grep 0 /selinux/enforce; then
sudo setenforce 0 && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo gsutil -m cp gs://startup_scripts_us/scripts/selinux_config /etc/selinux/config
fi

# limits.conf --> ulimit -a
if ! grep 'root      hard    nofile      500000' /etc/security/limits.conf; then
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil -m cp gs://startup_scripts_us/scripts/security_limits.conf /etc/security/limits.conf
fi

if ! grep 'root       soft    nproc     unlimited' /etc/security/limits.d/90-nproc.conf; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/90-nproc.conf /etc/security/limits.d/
fi

#PHP INSTALL:

#### Redhat 6.x ####
if ! rpm -qa | grep epel-release-6-8.noarch; then
sudo rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
fi

if ! rpm -qa | grep ius-release-1.0-15.ius.el6.noarch; then
sudo rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-15.ius.el6.noarch.rpm
fi

if grep '#baseurl=http' /etc/yum.repos.d/epel.repo; then
sudo su - << EOF
/bin/sed -i 's/\#baseurl=http/baseurl=http/g' /etc/yum.repos.d/epel.repo;
EOF
fi

if ! grep '#mirrorlist=http' /etc/yum.repos.d/epel.repo; then
sudo su - << EOF
/bin/sed -i 's/mirrorlist=http/\#mirrorlist=http/g' /etc/yum.repos.d/epel.repo;
EOF
fi

########

# check whether all necessary RPMs are installed or not
sudo su - << EOF
if ! rpm -qa | grep -E -c 'libxml2|libxml2-devel|GeoIP|geoip-devel|gcc|make|mysql|libmemcached10-devel|emacs|ntpdate|rdate|syslog-ng|syslog-ng-libdbi|libdbi-devel|telnet|screen|git|sendmail|sendmail-cf|denyhosts|procmail|python-argparse|*whois|openssl|openssl-devel|libssh2|libssh2-devel|java-1.8.0-openjdk-devel|python27-pip|lynx|libmaxminddb' | grep 46;
  then
  yum -y install libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail sendmail-cf denyhosts procmail python-argparse *whois openssl openssl-devel libssh2 libssh2-devel java-1.8.0-openjdk-devel python27-pip lynx libmaxminddb libmaxminddb-devel;
  else
  echo "All RPMs are installed"
fi
EOF

# python2.6 to python2.7
if ! pip2.7 list --format=legacy | grep google-compute-engine; then
sudo su - << EOF
/usr/bin/pip2.7 install google google_compute_engine
unlink /usr/bin/python2
ln -s /usr/bin/python2.7 /usr/bin/python2
cp -rfv /root/scripts/python_version_change.sh /etc/cron.hourly/
cp -rfv /root/scripts/python_version_change_cron.sh /etc/cron.d/
chmod 0755 /etc/cron.hourly/python_version_change.sh
chmod 0755 /etc/cron.d/python_version_change_cron.sh
/usr/bin/pip2.7 install --upgrade pip
EOF

else
/usr/bin/pip2.7 install --upgrade pip
fi

sudo service syslog stop;
sudo chkconfig syslog off;
sudo service postfix stop;
sudo chkconfig postfix off;
sudo chkconfig --del postfix;

#Copy GeoIP.dat
sudo gsutil -m cp -r gs://startup_scripts_us/scripts/GeoIP /usr/share

# add '/var/www/vhosts/mdotm.com/scripts/utils' to path
sudo su - << EOF
sudo echo "pathmunge /var/www/vhosts/mdotm.com/scripts/utils" > /etc/profile.d/pushcode.sh
EOF

#Install syslog-ng:
#if [ ! -f /etc/syslog-ng/syslog-ng.conf ]; then
if ! grep log6 /etc/syslog-ng/syslog-ng.conf &> /dev/null; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/syslog-ng.conf;
service syslog-ng restart;
chkconfig syslog-ng on;

else
echo "syslog-ng is installed"
fi

##############
#Install Nagios/cacti client

if [ -d /usr/lib64/nagios/plugins ]; then
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
yum -y install nagios-plugins nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm /home/anand/;
yum -y remove nrpe && rpm -Uvh /home/anand/nrpe-2.15-7.el6.x86_64.rpm;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
chmod +x /usr/lib64/nagios/plugins/*
chkconfig nrpe on;
service nrpe restart;
chkconfig snmpd on;
EOF

else
echo "nagios is installed"
fi

#check nrpe allowed_hosts
if ! grep 'allowed_hosts=10.128.1.15,104.197.43.125,50.225.47.189' /etc/nagios/nrpe.cfg; then
gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
/sbin/service nrpe restart;
fi

#############

#Copy CRONJOBS:
sudo su - << EOF
gsutil -m cp gs://startup_scripts_us/scripts/cron_root-us-ccdex-go /var/spool/cron/root;
service crond restart;
EOF

#Configure services to run on reboot:

if [ -f /etc/init.d/httpd ]; then
chkconfig httpd off;
else
echo "httpd is NOT installed, so running \"chkconfig httpd off\" not necessary"
fi

service sendmail restart;
chkconfig crond on;
chkconfig iptables off;
chkconfig sendmail on;
chkconfig syslog-ng on;
chkconfig syslog off;

# Copyinfg scripts to /root/scripts

mkdir -p /root/scripts
sudo gsutil -m cp gs://startup_scripts_us/scripts/git_clone.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/git_clone_cc.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/shortcircuit.php /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslogchk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/startup-script-us-new-failsafe-ccdex-go.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/cron.hourly_crosschannel /etc/cron.hourly/crosschannel;
sudo gsutil -m cp gs://startup_scripts_us/scripts/logrotate.d_crosschannel /etc/logrotate.d/crosschannel;
sudo chmod -R +x /root/scripts

#USE GIT:
if [ ! -d /var/www/vhosts/mdotm.com ]; then
sudo mkdir -p /var/www/vhosts;
sudo su - << EOF
cd /var/www/vhosts
sudo git clone git@github.com:sourabhniyogi/mdotm.com.git /var/www/vhosts/mdotm.com;
cd /var/www/vhosts/mdotm.com/;
sudo git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git;
sudo git config core.filemode false;
sudo git config user.email "sourabh@crosschannel.com";
sudo git config user.name "Sourabh Niyogi";
sudo git fetch upstream;
sudo git merge upstream/master;
mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK
EOF

else
echo "/var/www/vhosts/mdotm.com exists"
fi

#USE GIT TO ADD CROSSCHANNEL.COM
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        cd /var/www/vhosts
        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        git config core.filemode false;
        git config user.email "sourabh@crosschannel.com";
        git config user.name "Sourabh Niyogi";
        git fetch upstream;
        git merge upstream/master;
EOF
fi

# GIT CLONE API.WOLK.COM
if [ ! -d /var/www/vhosts/api.wolk.com ]; then
        echo "/var/www/vhosts/api.wolk.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        cd /var/www/vhosts/;
        git clone git@github.com:wolkdb/api.wolk.com.git /var/www/vhosts/api.wolk.com;
        cd /var/www/vhosts/api.wolk.com;
        git remote add upstream git@github.com:wolkdb/api.wolk.com.git;
        git config user.email sourabh@crosschannel.com;
        git config user.name "Sourabh Niyogi";
        git config --global core.filemode false;
        git config core.filemode false;
        git fetch upstream;
        git merge upstream/master;
EOF
fi

# GIT CLONE WWW.WOLK.COM
if [ ! -d /var/www/vhosts/www.wolk.com ]; then
        echo "/var/www/vhosts/www.wolk.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        cd /var/www/vhosts/;
        git clone git@github.com:wolkdb/www.wolk.com.git /var/www/vhosts/www.wolk.com;
        cd /var/www/vhosts/www.wolk.com;
        git remote add upstream git@github.com:wolkdb/www.wolk.com.git;
        git config user.email sourabh@crosschannel.com;
        git config user.name "Sourabh Niyogi";
        git config --global core.filemode false;
        git config core.filemode false;
        git fetch upstream;
        git merge upstream/master;
EOF
fi

#ADD shortcircuit.php manually:
sudo gsutil -m cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/

################
## Install libtool for maxminddb
#if ! ls -l /usr/local/lib/libmaxminddb.so*; then
#sudo su - << EOF
#cd /root;
#yum -y install libtool* git && git clone --recursive https://github.com/maxmind/libmaxminddb
#cd libmaxminddb &&
#./bootstrap &&
#./configure &&
#make &&
#make check &&
#make install &&
#ldconfig
#EOF
#fi
#######

# installing GO
if [ ! -d /usr/local/go ]; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/go/go-crosschannel-install.sh /home/anand;
sudo sh /home/anand/go-crosschannel-install.sh;
fi

#######

# net.ipv4.tcp_tw_recycle and net.ipv4.tcp_tw_reuse
if ! grep 1 /proc/sys/net/ipv4/tcp_tw_reuse > /dev/null; then
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
fi

######

# running go command

if netstat -apn | grep crosschannel > /dev/null; then
  echo "crosschannel is running... killing"
  kill -9 $(ps aux | grep '/var/www/vhosts/crosschannel.com/bidder/bin/crosschannel' | grep -v grep | awk '{print$2}')
fi

if ! ps aux | grep ccdex | grep -v grep; then
sudo su - << EOF
cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh ccdex &>> /var/log/ccdex.log
EOF
fi

if ! netstat -apn | grep ccdex > /dev/null; then
  sudo su - << EOF
  cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh ccdex &>> /var/log/ccdex.log
EOF
fi

#######

#Change 'assumeyes=1' --> 'assumeyes=0' in yum.conf
sed -i '/assumeyes/d' /etc/yum.conf
sed -i "$ i\assumeyes=0" /etc/yum.conf

################

#Denyhosts
if [ ! -d /var/lib/denyhosts ]; then
sudo cp gs://startup_scripts_us/scripts/denyhosts/allowed-hosts /var/lib/denyhosts/;
sudo cp gs://startup_scripts_us/scripts/denyhosts/denyhosts.conf /etc;
sudo su - << EOF
echo `ifconfig | grep 'inet addr:10' | awk '{print$2}' | cut -d ":" -f2` >> /var/lib/denyhosts/allowed-hosts
EOF
service denyhosts restart;
chkconfig denyhosts on;
else
echo "Denyhosts is already installed"
fi

###############

# Try yum install rpms again, in case they failed earlier:
sudo su - << EOF
if ! rpm -qa | grep -E -c 'libxml2|libxml2-devel|GeoIP|geoip-devel|gcc|make|mysql|libmemcached10-devel|emacs|ntpdate|rdate|syslog-ng|syslog-ng-libdbi|libdbi-devel|telnet|screen|git|sendmail|sendmail-cf|denyhosts|procmail|python-argparse|*whois|openssl|openssl-devel|libssh2|libssh2-devel|java-1.8.0-openjdk-devel|python27-pip|lynx|libmaxminddb' | grep 46;
  then
  sudo yum -y install libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail sendmail-cf denyhosts procmail python-argparse *whois openssl openssl-devel libssh2 libssh2-devel java-1.8.0-openjdk-devel python27-pip lynx libmaxminddb libmaxminddb-devel;
  else
  echo "All RPMs are installed"
fi
EOF


# Try git cloning again, in case they failed earlier:
#USE GIT:
if [ ! -d /var/www/vhosts/mdotm.com ]; then
sudo mkdir -p /var/www/vhosts;
sudo su - << EOF
cd /var/www/vhosts
sudo git clone git@github.com:sourabhniyogi/mdotm.com.git /var/www/vhosts/mdotm.com;
cd /var/www/vhosts/mdotm.com/;
sudo git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git;
sudo git config core.filemode false;
sudo git config user.email "sourabh@crosschannel.com";
sudo git config user.name "Sourabh Niyogi";
sudo git fetch upstream;
sudo git merge upstream/master;
mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK
EOF

else
echo "/var/www/vhosts/mdotm.com exists"
fi

#USE GIT TO ADD CROSSCHANNEL.COM
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        cd /var/www/vhosts
        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        git config core.filemode false;
        git config user.email "sourabh@crosschannel.com";
        git config user.name "Sourabh Niyogi";
        git fetch upstream;
        git merge upstream/master;
EOF
fi

sed -i 's/\*\/1 \* \* \* \* \/usr\/bin\/flock -w 0 \/var\/run\/startup-script-us-new-failsafe-ccdex-go.lock/#\*\/1 \* \* \* \* \/usr\/bin\/flock -w 0 \/var\/run\/startup-script-us-new-failsafe-ccdex-go.lock/g' /var/spool/cron/root
