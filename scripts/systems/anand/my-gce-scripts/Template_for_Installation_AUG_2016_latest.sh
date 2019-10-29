#!/bin/bash

#SSH Keys:
sudo gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
sudo rm -rf /root/.ssh/ssh_keys.tgz

# Allow SSH-ing to any instance/server
sudo gsutil cp gs://startup_scripts_us/scripts/ssh_config /etc/ssh/
sudo gsutil cp gs://startup_scripts_us/scripts/sshd_config /etc/ssh/
sudo service sshd restart
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Replace /etc/sudoers default 'Defaults secure_path' with following:
#Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin

# Comment out this line in /etc/sudoers:
#Defaults    requiretty

sudo gsutil cp gs://startup_scripts_us/scripts/sudoers /etc/

# Copy /etc/hosts from gs:
sudo gsutil cp gs://startup_scripts_us/scripts/hosts /etc/

# Enable histtimeformat
sudo gsutil cp gs://startup_scripts_us/scripts/histtimeformat.sh /etc/profile.d/

## DISABLE FSCK
sudo tune2fs -c 0 -i 0 /dev/sda1
sudo tune2fs -c 0 -i 0 /dev/sda3

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo gsutil cp gs://startup_scripts_us/scripts/selinux_config /etc/selinux/config

# limits.conf --> ulimit -a
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil cp gs://startup_scripts_us/scripts/security_limits.conf /etc/security/limits.conf

#PHP INSTALL:

#### Redhat 6.x ####
sudo rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
sudo rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-14.ius.el6.noarch.rpm
########

sudo yum remove -y php*;
sudo yum -y --enablerepo=ius-archive install php54 php54-soap php54-gd php54-ioncube-loader php54-pecl-memcache php54-mcrypt php54-imap php54-devel php54-cli php54-mysql php54-mbstring php54-xml libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql mysql php54-pecl-memcached libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail sendmail-cf denyhosts procmail python-argparse *whois openssl openssl-devel php-pear-Net-SMTP php54-pear php-pear-Net-Socket php-pear-Auth-SASL libssh2 libssh2-devel java-1.8.0-openjdk-devel;

sudo service syslog stop;
sudo chkconfig syslog off;
sudo service postfix stop;
sudo chkconfig postfix off;
sudo chkconfig --del postfix;
sudo service rsyslog stop;
sudo chkconfig rsyslog off;

#Copy GeoIP.dat
sudo gsutil cp -r gs://startup_scripts_us/scripts/GeoIP /usr/share

#Install PHP extensions:
sudo cp /etc/php.ini /etc/php.ini_BAK_`date +%m%d%Y`_ORIG
sudo sed -i 's/short_open_tag = Off/short_open_tag = On/g' /etc/php.ini;

sudo pecl channel-update pecl.php.net;
yes '' | sudo pecl install memcached;
yes '' | sudo pecl install geoip; 
yes '' | sudo pecl install -f apc;

sudo rpm -Uvh ftp://ftp.pbone.net/mirror/rpmfusion.org/nonfree/el/updates/6/x86_64/pdflib-lite-7.0.5-1.el6.1.x86_64.rpm;
sudo rpm -Uvh ftp://ftp.pbone.net/mirror/rpmfusion.org/nonfree/el/updates/6/x86_64/pdflib-lite-devel-7.0.5-1.el6.1.x86_64.rpm;
yes '' | sudo pecl install ssh2;

# removing extensions from php.ini before adding to avoid duplicacies
sudo sed -i '/mongo/d' /etc/php.ini;
sudo sed -i '/ffmpeg/d' /etc/php.ini;
sudo sed -i '/igbinary/d' /etc/php.ini;
sudo sed -i '/pdf.so/d' /etc/php.ini;
sudo sed -i '/lua.so/d' /etc/php.ini;

# php extensions
sudo gsutil cp gs://startup_scripts_us/scripts/php_modules/pdf.so /usr/lib64/php/modules/;
sudo gsutil cp gs://startup_scripts_us/scripts/php_modules/lua.so /usr/lib64/php/modules/;
sudo gsutil cp gs://startup_scripts_us/scripts/php_modules/mongo*.so /usr/lib64/php/modules/;
sudo gsutil cp gs://startup_scripts_us/scripts/php_modules/igbinary.so /usr/lib64/php/modules/;
sudo gsutil cp gs://startup_scripts_us/scripts/php_modules/aerospike.so /usr/lib64/php/modules/;
sudo gsutil cp gs://startup_scripts_us/scripts/php_modules/maxminddb.so /usr/lib64/php/modules/;
sudo gsutil cp gs://startup_scripts_us/scripts/php_modules/apc.ini /etc/php.d/

sudo gsutil gs://startup_scripts_us/scripts/php_modules/rpms/*;
sudo rpm -Uvh pdflib-lite*.rpm
sudo su - << EOF
echo 'extension=geoip.so' >> /etc/php.ini;
echo 'extension=aerospike.so' >> /etc/php.ini;
echo 'extension=maxminddb.so' >> /etc/php.ini;
echo 'extension=mongo.so' >> /etc/php.ini;
echo 'extension=mongodb.so' >> /etc/php.ini;
echo 'extension=lua.so' >> /etc/php.ini;
echo 'extension=pdf.so' >> /etc/php.ini;
echo 'extension=ssh2.so' >> /etc/php.ini;

echo '
extension=igbinary.so
session.serialize_handler=igbinary
igbinary.compact_strings=On' >> /etc/php.ini

echo ';extension=ffmpeg.so' >> /etc/php.ini;
echo ';extension=kafka.so' >> /etc/php.ini;
echo ';extension=citrusleaf.so' >> /etc/php.ini;

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
service httpd restart
EOF


#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum -y install ntpdate rdate;
sudo gsutil cp gs://startup_scripts_us/scripts/ntpdate.sh /etc/cron.hourly/;
sudo sh /etc/cron.hourly/ntpdate.sh;

#Install syslog-ng:

sudo gsutil cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/;
service syslog-ng restart;
chkconfig syslog-ng on;

#Copy CRONJOBS:
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/cron_root /var/spool/cron/root;
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
chkconfig rsyslog off;

#Add LogFormat + vhosts + etc...

sudo gsutil cp gs://startup_scripts_us/scripts/httpd1.conf /etc/httpd/conf/;
mkdir -p /root/scripts
sudo gsutil cp gs://startup_scripts_us/scripts/httpd_conf.sh /root/scripts/;
sudo gsutil cp gs://startup_scripts_us/scripts/git_clone.sh /root/scripts;
sudo gsutil cp gs://startup_scripts_us/scripts/shortcircuit.php /root/scripts;
sudo gsutil cp gs://startup_scripts_us/scripts/hbase/hbasechk.sh /root/scripts;
sudo sh /root/scripts/httpd_conf.sh

#USE GIT:
sudo mkdir -p /var/www/vhosts;
sudo cd /var/www/vhosts;
sudo git clone git@github.com:sourabhniyogi/mdotm.com.git;
sudo cd /var/www/vhosts/mdotm.com/;
sudo git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git;
sudo git config core.filemode false;
sudo git config user.email "sourabh@crosschannel.com";
sudo git config user.name "Sourabh Niyogi";
sudo git fetch upstream;
sudo git merge upstream/master;
sudo mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK;

#ADD shortcircuit.php manually:
sudo gsutil cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/

# loadAll.php
sudo mkdir -p /var/log/sites;
sudo /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php;

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

# installing hbase

sudo gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-install-centOS.sh /home/anand;
sudo sh /home/anand/hbase-install-centOS.sh;

#######
# Install Kafka
#scp www2042:/usr/local/lib/librdkafka.so.1 /usr/local/lib/librdkafka.so.1
#ln -s /usr/local/lib/librdkafka.so.1 /usr/lib64/librdkafka.so.1

#Change 'assumeyes=1' --> 'assumeyes=0' in yum.conf
sed -i '/assumeyes/d' /etc/yum.conf
sed -i "$ i\assumeyes=0" /etc/yum.conf

##############
#Install Nagios/cacti client
#yum -y install nagios nagios-plugins nagios-plugins-nrpe nagios-nrpe gd-devel net-snmp;
sudo yum -y install nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
sudo gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
sudo cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
chkconfig nrpe on;
service nrpe restart;
chkconfig snmpd on;

################
#Denyhosts
sudo cp gs://startup_scripts_us/scripts/denyhosts/allowed-hosts /var/lib/denyhosts/;
sudo cp gs://startup_scripts_us/scripts/denyhosts/denyhosts.conf /etc;
service denyhosts restart;
chkconfig denyhosts on;
###############
## HADOOP INSTALL ##
#sh /root/scripts/hadoop_install.sh && sh /root/scripts/ha_start_namenode_datanode.sh

# clean up
sudo rm -rf /root/composer*;
sudo rm -rf /root/libmaxminddb;
sudo rm -rf /root/vendor;
