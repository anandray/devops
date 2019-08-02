#!/bin/bash

# create /root/scripts dir
sudo mkdir /root/scripts;

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo ntpdate -u -b pool.ntp.org;

#sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
#sudo sh /root/scripts/ntpdate.sh;

## Start /etc/sudoers modifications ##
# THE FOLLOWING /etc/sudoers MODIFICATIONS ARE REQUIRED TO ALLOW 'gcloud' and 'gsutil' to run using sudo, WHICH IS OTHERWISE NOT ALLOWED
# Replace /etc/sudoers default 'Defaults secure_path' with following:
#Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin

# Comment out this line in /etc/sudoers:
#Defaults    requiretty

sudo gsutil -m cp gs://startup_scripts_us/scripts/sudoers /etc/
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
sudo gsutil -m cp gs://startup_scripts_us/scripts/hosts /etc/

#SSH Keys:
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
sudo rm -rf /root/.ssh/ssh_keys.tgz

# Allow SSH-ing to any instance/server
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

#PHP INSTALL:

#### Redhat 6.x ####
sudo rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
sudo rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-14.ius.el6.noarch.rpm

if ! rpm -qa | grep epel-release > /dev/null; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/rpms/epel-release-6-8.noarch.rpm /home/anand
sudo rpm -Uvh /home/anand/epel-release-6-8.noarch.rpm
fi

if ! rpm -qa | grep ius-release > /dev/null; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/rpms/ius-release-1.0-14.ius.el6.noarch.rpm /home/anand
sudo rpm -Uvh /home/anand/ius-release-1.0-14.ius.el6.noarch.rpm
fi

sudo su - << EOF
#gsutil gs://startup_scripts_us/scripts/epel.repo /etc/yum.repos.d/epel.repo
/bin/sed -i 's/\#baseurl=http/baseurl=http/g' /etc/yum.repos.d/epel.repo;
/bin/sed -i 's/mirrorlist=http/\#mirrorlist=http/g' /etc/yum.repos.d/epel.repo;
EOF
########

if php -v > /dev/null;
  then
  sudo yum remove -y php*
fi

sudo yum -y --enablerepo=ius-archive install php54 php54-soap php54-gd php54-ioncube-loader php54-pecl-memcache php54-mcrypt php54-imap php54-devel php54-cli php54-mysql php54-mbstring php54-xml libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql php54-pecl-memcached libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail sendmail-cf denyhosts procmail python-argparse *whois openssl openssl-devel php-pear-Net-SMTP php54-pear php-pear-Net-Socket php-pear-Auth-SASL libssh2 libssh2-devel java-1.8.0-openjdk-devel python27-pip lynx;

if ! php -v > /dev/null;
  then
  sudo yum -y --enablerepo=ius-archive install php54 php54-soap php54-gd php54-ioncube-loader php54-pecl-memcache php54-mcrypt php54-imap php54-devel php54-cli php54-mysql php54-mbstring php54-xml libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql php54-pecl-memcached libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail sendmail-cf denyhosts procmail python-argparse *whois openssl openssl-devel php-pear-Net-SMTP php54-pear php-pear-Net-Socket php-pear-Auth-SASL libssh2 libssh2-devel java-1.8.0-openjdk-devel python27-pip lynx
fi

# python2.6 to python2.7
sudo su - << EOF
/usr/bin/pip2.7 install google google_compute_engine
unlink /usr/bin/python2
ln -s /usr/bin/python2.7 /usr/bin/python2
gsutil -m cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
gsutil -m cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
chmod -R 0755 /root/scripts/python_version_change*;
cp -rfv /root/scripts/python_version_change.sh /etc/cron.hourly/
cp -rfv /root/scripts/python_version_change_cron.sh /etc/cron.d/
EOF

sudo service syslog stop;
sudo chkconfig syslog off;
sudo service postfix stop;
sudo chkconfig postfix off;
sudo chkconfig --del postfix;
sudo service rsyslog stop;
sudo chkconfig rsyslog off;

#Copy GeoIP.dat
sudo gsutil -m cp -r gs://startup_scripts_us/scripts/GeoIP /usr/share

# update Geoip databases
#sudo gsutil -m cp gs://startup_scripts_us/scripts/GeoIP.conf /etc/;
#sudo /usr/bin/geoipupdate -v;

#Install PHP extensions:
sudo cp /etc/php.ini /etc/php.ini_BAK_`date +%m%d%Y`_ORIG
sudo sed -i 's/max_execution_time = 30/max_execution_time = 300/g' /etc/php.ini;
sudo sed -i 's/short_open_tag = Off/short_open_tag = On/g' /etc/php.ini;
sudo sed -i 's/\;date.timezone \=/date.timezone ="America\/Los_Angeles"/g' /etc/php.ini;

#sudo pecl channel-update pecl.php.net;
#yes '' | sudo pecl install memcached; 
#yes '' | sudo pecl install geoip; 
#yes '' | sudo pecl install -f apc;
#yes '' | sudo pecl install ssh2;

# removing extensions from php.ini before adding to avoid duplicacies

sudo sed -i '/geoip/d' /etc/php.ini;
sudo sed -i '/maxminddb.so/d' /etc/php.ini;
sudo sed -i '/memcache.so/d' /etc/php.ini;
sudo sed -i '/memcached.so/d' /etc/php.ini;
sudo sed -i '/ssh2.so/d' /etc/php.ini;
sudo sed -i '/igbinary/d' /etc/php.ini;
sudo sed -i '/pdf.so/d' /etc/php.ini;
sudo sed -i '/lua.so/d' /etc/php.ini;

# php extensions
sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/pdf.so /usr/lib64/php/modules/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/lua.so /usr/lib64/php/modules/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/igbinary.so /usr/lib64/php/modules/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/maxminddb.so /usr/lib64/php/modules/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/geoip.so /usr/lib64/php/modules/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/apc.so /usr/lib64/php/modules/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/ssh2.so /usr/lib64/php/modules/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/apc.ini /etc/php.d/

sudo gsutil -m cp gs://startup_scripts_us/scripts/php_modules/rpms/* /home/anand/;
sudo rpm -Uvh /home/anand/pdflib-lite*.rpm

sudo su - << EOF
echo 'extension=geoip.so' >> /etc/php.ini;
echo 'geoip.custom_directory = /usr/share/GeoIP/' >> /etc/php.ini;
echo 'extension=maxminddb.so' >> /etc/php.ini;
echo 'extension=lua.so' >> /etc/php.ini;
echo 'extension=pdf.so' >> /etc/php.ini;
echo 'extension=ssh2.so' >> /etc/php.ini;
echo 'extension=memcache.so' >> /etc/php.ini;
echo 'extension=memcached.so' >> /etc/php.ini;

echo '
extension=igbinary.so
session.serialize_handler=igbinary
igbinary.compact_strings=On' >> /etc/php.ini

sed -i 's/error_log = php_errors.log/;error_log = php_errors.log/g' /etc/php.ini;
echo 'error_log = /var/log/httpd/php_errors.log' >> /etc/php.ini;
EOF


# installing pear mail/Mail.php required to send mail using smtp-auth
sudo su - << EOF
#yum -y install php-pear-Net-SMTP php54-pear php-pear-Net-Socket php-pear-Auth-SASL # already installed above
pear channel-update pear.php.net
pear install mail

# add to include_path in php.ini
sed -i 's/include_path/;include_path/g' /etc/php.ini && echo 'include_path = ".:/var/www/vhosts/mdotm.com/include/:/usr/share/pear/:/usr/share/pear/Mail/:/usr/share/GeoIP/"' >> /etc/php.ini
chmod 0755 /var/log/httpd
service httpd restart
EOF

################
## Install libtool for maxminddb
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
EOF
#######

# add '/var/www/vhosts/mdotm.com/scripts/utils' to path
sudo su - << EOF
sudo echo "pathmunge /var/www/vhosts/mdotm.com/scripts/utils" > /etc/profile.d/pushcode.sh
EOF

#Install syslog-ng:
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/syslog-ng.conf;
sudo service syslog-ng restart;
sudo chkconfig syslog-ng on;

##############
#Install Nagios/cacti client
sudo su - << EOF
gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
yum -y install nagios-plugins nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
chmod +x /usr/lib64/nagios/plugins/*
chkconfig nrpe on;
service nrpe restart;
chkconfig snmpd on;
EOF

#############

#Configure services to run on reboot:
service sendmail restart;
chkconfig httpd on;
chkconfig crond on;
chkconfig iptables off;
chkconfig sendmail on;
chkconfig syslog-ng on;
chkconfig syslog off;
chkconfig rsyslog off;

#USE GIT:
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

#Add LogFormat + vhosts + etc...
sudo gsutil -m cp gs://startup_scripts_us/scripts/httpd1.conf /etc/httpd/conf/;

# copying scripts to /root/scripts
mkdir -p /root/scripts
sudo gsutil -m cp gs://startup_scripts_us/scripts/profile.d/* /etc/profile.d/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/cron_env*.bash /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/.bash_profile /root/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/httpd_conf-eu.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/git_clone.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/git_clone_cc.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys_chk.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/shortcircuit.php /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/hbase/hbasechk-eu.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/run_newservercheck.php.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslogchk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/startup-script-eu-new-failsafe.sh /root/scripts/;
gsutil -m cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
gsutil -m cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
sudo chmod -R +x /root/scripts
sudo sh /root/scripts/httpd_conf-eu.sh

# emacs for go-lang
sudo su - << EOF
gsutil -m cp -r gs://startup_scripts_us/scripts/emacs/emacs/site-lisp /usr/share/emacs;
gsutil -m cp -r gs://startup_scripts_us/scripts/emacs/.emacs.d /root/;
EOF

#Copy CRONJOBS:
sudo su - << EOF
gsutil -m cp gs://startup_scripts_us/scripts/cron_root-eu /var/spool/cron/root;
service crond restart;
EOF

#ADD shortcircuit.php manually:
sudo gsutil -m cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/

# loadAll.php
sudo mkdir -p /var/log/sites;
ssh 'hostname' /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php > /var/log/loadall.log 2>&1;

# installing hbase
sudo gsutil -m cp gs://startup_scripts_us/scripts/hbase/hbase-install-centOS-eu.sh /home/anand;
sudo sh /home/anand/hbase-install-centOS-eu.sh;

# installing GO
sudo gsutil -m cp gs://startup_scripts_us/scripts/go/go-install.sh /home/anand;
sudo sh /home/anand/go-install.sh;

# running go command
#cd /var/www/vhosts/crosschannel.com/bidder/bin && git fetch upstream && git merge upstream/master && php goservice.php bt 2> /var/log/git.err  > /var/log/git.log
sudo su - << EOF
cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php bt 2> /var/log/git-bt.err  > /var/log/git-bt.log
cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php zn 2> /var/log/git-zn.err  > /var/log/git-zn.log
EOF

#######

#Change 'assumeyes=1' --> 'assumeyes=0' in yum.conf
sudo su - << EOF
/bin/sed -i '/assumeyes/d' /etc/yum.conf
/bin/sed -i "$ i\assumeyes=0" /etc/yum.conf
EO

################

# python2.6 to python2.7
sudo su - << EOF
/usr/bin/pip2.7 install google google_compute_engine
unlink /usr/bin/python2
ln -s /usr/bin/python2.7 /usr/bin/python2
cp -rfv /root/scripts/python_version_change.sh /etc/cron.hourly/
cp -rfv /root/scripts/python_version_change_cron.sh /etc/cron.d/
chmod 0755 /etc/cron.hourly/python_version_change.sh
chmod 0755 /etc/cron.d/python_version_change_cron.sh
EOF

################
#Denyhosts
sudo gsutil -m cp gs://startup_scripts_us/scripts/denyhosts/allowed-hosts /var/lib/denyhosts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/denyhosts/denyhosts.conf /etc;
sudo su - << EOF
echo `ifconfig | grep 'inet addr:10' | awk '{print$2}' | cut -d ":" -f2` >> /var/lib/denyhosts/allowed-hosts
EOF
service denyhosts restart;
chkconfig denyhosts on;
###############

## HADOOP INSTALL ##
#sh /root/scripts/hadoop_install.sh && sh /root/scripts/ha_start_namenode_datanode.sh

# clean up
sudo rm -rf /root/composer*;
sudo rm -rf /root/libmaxminddb;
sudo rm -rf /root/vendor;

# run newservercheck.php
sudo gsutil -m cp gs://startup_scripts_us/scripts/run_newservercheck.php.sh /root/scripts/;
#sudo su - << EOF
#echo "*/1 * * * * /bin/sh /root/scripts/run_newservercheck.php.sh > /dev/null 2>&1" >> /var/spool/cron/root
#EOF

# update gcloud
#sudo gcloud -q components update
