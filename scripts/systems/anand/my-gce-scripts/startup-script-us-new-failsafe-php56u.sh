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

#Install php54-5.4.45

if ! rpm -qa | grep php56u; then

#sudo yum remove -y php*;
sudo yum -y install php56u php56u-soap php56u-gd php56u-ioncube-loader php56u-pecl-memcache php56u-mcrypt php56u-imap php56u-devel php56u-cli php56u-mysql php56u-mbstring php56u-xml libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql php56u-pecl-memcached libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail sendmail-cf denyhosts procmail python-argparse *whois openssl openssl-devel php-pear-Net-SMTP php56u-pear php-pear-Net-Socket php-pear-Auth-SASL libssh2 libssh2-devel java-1.8.0-openjdk-devel python27-pip;

# additional php modules
yum -y install php56u-pecl-apc php56u-pecl-ssh2 php56u-pecl-memcache php56u-pecl-memcached php56u-pecl-igbinary php56u-pecl-geoip
fi

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
if ! [ -d /usr/share/GeoIP ]; then
sudo gsutil -m cp -r gs://startup_scripts_us/scripts/GeoIP /usr/share
fi

# changing "max_execution_time = 30" to "max_execution_time = 300" in /etc/php.ini
sudo cp /etc/php.ini /etc/php.ini_BAK_`date +%m%d%Y`_ORIG
if grep "max_execution_time = 300" /etc/php.ini; then
    echo 'Do Nothing'
  else
  sudo sed -i 's/max_execution_time = 30/max_execution_time = 300/g' /etc/php.ini
fi

sudo sed -i 's/short_open_tag = Off/short_open_tag = On/g' /etc/php.ini;
sudo sed -i 's/\;date.timezone \=/date.timezone ="America\/Los_Angeles"/g' /etc/php.ini;

sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/rpms/* /home/anand/;
sudo rpm -Uvh /home/anand/pdflib-lite*.rpm

# installing pear mail/Mail.php required to send mail using smtp-auth
sudo su - << EOF
pear channel-update pear.php.net
pear install mail
EOF

# add to include_path in php.ini
if ! grep '/var/www/vhosts/mdotm.com/include/:/usr/share/pear/:/usr/share/pear/Mail/:/usr/share/GeoIP/' /etc/php.ini
sudo su - << EOF
echo 'error_log = /var/log/httpd/php_errors.log
include_path = ".:/var/www/vhosts/mdotm.com/include/:/usr/share/pear/:/usr/share/pear/Mail/:/usr/share/GeoIP/"' >> /etc/php.ini

chmod 0755 /var/log/httpd

if ps -A -U apache | grep httpd > /dev/null; 
    then
    service httpd graceful
  else
  service httpd restart
fi

EOF


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
gsutil -m cp gs://startup_scripts_us/scripts/cron_root-us /var/spool/cron/root;
service crond restart;
EOF

#Configure services to run on reboot:
service sendmail restart;
chkconfig httpd on;
chkconfig crond on;
chkconfig iptables off;
chkconfig sendmail on;
chkconfig syslog-ng on;
chkconfig syslog off;

#Add LogFormat + vhosts + etc...

sudo gsutil -m cp gs://startup_scripts_us/scripts/httpd1.conf /etc/httpd/conf/;
mkdir -p /root/scripts
sudo gsutil -m cp gs://startup_scripts_us/scripts/httpd_conf-us.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/git_clone.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/shortcircuit.php /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/hbase/hbasechk-us.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/run_newservercheck.php.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslogchk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/nagios/startup-script-us-new-failsafe.sh /root/scripts/;
gsutil -m cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
gsutil -m cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
sudo chmod -R +x /root/scripts
sudo sh /root/scripts/httpd_conf-us.sh

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

#ADD shortcircuit.php manually:
sudo gsutil -m cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/

# loadAll.php
if [ ! -d /var/log/sites ]; then
sudo mkdir -p /var/log/sites;
sudo /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php;
else
echo "/var/log/sites exists"
fi

################
## Install libtool for maxminddb
if ! php -m | grep maxminddb; then
sudo su - << EOF
cd /root;
yum -y install libtool* git && git clone --recursive https://github.com/maxmind/libmaxminddb
cd libmaxminddb &&
./bootstrap &&
./configure &&
make &&
make check &&
make install &&
ldconfig &&

#Install PHP Extension maxminddb.so:

cd /root &&
curl -sS https://getcomposer.org/installer | php &&
php composer.phar require geoip2/geoip2:~2.0 &&

## This creates a directory named 'vendor'

cd vendor/maxmind-db/reader/ext &&
phpize &&
./configure &&
make &&
yes '' | make test &&
make install &&
ldconfig /usr/local/lib/
rsync -avz /usr/local/lib/*maxmind* /usr/lib64/
echo 'extension=maxminddb.so' >> /etc/php.ini
EOF
fi
#######

# installing hbase

if [ ! -d /usr/local/hbase-1.1.2/ ]; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/hbase/hbase-install-centOS-us.sh /home/anand;
sudo sh /home/anand/hbase-install-centOS-us.sh;
else
echo "HBase is already installed..."
fi

# installing GO
if [ ! -d /usr/local/go ]; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/go/go-install.sh /home/anand;
sudo sh /home/anand/go-install.sh;
fi

# running go command
if netstat -apn | egrep '::9900' > /dev/null; then
  echo bt is running
else
sudo su - << EOF
cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php 2> /var/log/git.err  > /var/log/git.log
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
## HADOOP INSTALL ##
#sh /root/scripts/hadoop_install.sh && sh /root/scripts/ha_start_namenode_datanode.sh

# clean up
sudo rm -rf /root/composer*;
sudo rm -rf /root/libmaxminddb;
sudo rm -rf /root/vendor;

# run newservercheck.php
if ! ls -l /root/scripts | grep run_newservercheck.php.sh; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/run_newservercheck.php.sh /root/scripts/;
fi

#if ! grep run_newservercheck.php.sh /var/spool/cron/root; then
#sudo su - << EOF
#echo "*/1 * * * * /bin/sh /root/scripts/run_newservercheck.php.sh > /dev/null 2>&1" >> /var/spool/cron/root
#EOF
#fi

if ! php -v > /dev/null; then
sh /root/scripts/startup*failsafe*.sh;
else
/bin/sed -i 's/\*\/1 \* \* \* \* \/usr\/bin\/flock \-w 0 \/var\/run\/startup-script-us-new-failsafe.lock \/bin\/sh \/root\/scripts\/startup-script-us-new-failsafe.sh/\#\*\/1 \* \* \* \* \/usr\/bin\/flock \-w 0 \/var\/run\/startup-script-us-new-failsafe.lock \/bin\/sh \/root\/scripts\/startup-script-us-new-failsafe.sh/g' /var/spool/cron/root
fi
