#!/bin/bash

#SSH Keys:
#sudo wget -P /root/.ssh/ http://anand.www1001.mdotm.com/gce/ssh_keys/ssh_keys.tgz
#cd /root/.ssh
#sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz .ssh/ && 
sudo cp /home/anand/.ssh/ssh_keys.tgz /root/.ssh/ && 
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/

# Allow SSH-ing to any server
#yes '' | scp 10.128.0.56:/etc/ssh/ssh_config /etc/ssh/ && service sshd restart;
sudo sed -i "/StrictHostKeyChecking/d" /etc/ssh/ssh_config
sudo sed -i "/UserKnownHostsFile/d" /etc/ssh/ssh_config
sudo sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
sudo sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config
sudo service sshd restart

# Copy /etc/hosts from 10.128.0.56(10.128.0.56):
#sudo scp 10.128.0.56:/etc/hosts /etc/
gsutil cp gs://startup_scripts_us/scripts/hosts /home/anand/
sudo cp -rf /home/anand/hosts /etc/

# Copy /etc/resolv.conf
#scp 10.128.0.56:/etc/resolv.conf /etc/

# Enable histtimeformat
#sudo scp 10.128.0.56:/etc/profile.d/histtimeformat.sh /etc/profile.d/
gsutil cp gs://startup_scripts_us/scripts/histtimeformat.sh /home/anand/
sudo cp -rf /home/anand/histtimeformat.sh /etc/profile.d/

## DISABLE FSCK
sudo tune2fs -c 0 -i 0 /dev/sda1

#DISABLE SELINUX:
sudo setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
sudo cp -rf /etc/selinux/config /etc/selinux/config_ORIG;
gsutil cp gs://startup_scripts_us/scripts/config /etc/selinux/
gsutil cp gs://startup_scripts_us/scripts/limits.conf /etc/security/

#sudo scp 10.128.0.56:/etc/selinux/config /etc/selinux/
#sudo scp 10.128.0.56:/etc/security/limits.conf /etc/security/

#APACHE/HTTPD INSTALL
sudo yum -y httpd httpd-devel httpd-tools

#PHP INSTALL:

#### Redhat 6.x ####
sudo rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
sudo rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-14.ius.el6.noarch.rpm
########


sudo service syslog stop;
sudo chkconfig syslog off;
sudo service postfix stop;
sudo chkconfig postfix off;
sudo chkconfig --del postfix;
sudo service rsyslog stop;
sudo chkconfig rsyslog off;
sudo chmod 0000 /usr/sbin/postfix;


sudo yum remove -y php*;
sudo yum --enablerepo=ius-archive -y install php54 php54-soap php54-gd php54-ioncube-loader php54-pecl-memcache php54-mcrypt php54-imap php54-devel php54-cli php54-mysql php54-mbstring php54-xml libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql memcached memcached-devel mysql php54-pecl-memcached libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail denyhosts procmail python-argparse *whois;


#Copy GeoIP.dat
#sudo scp -r 10.128.0.56:/usr/share/GeoIP /usr/share/;
gsutil cp -r gs://startup_scripts_us/scripts/GeoIP /home/anand
sudo cp -rf /home/anand/GeoIP /usr/share/

#Install PHP extensions:
sudo cp /etc/php.ini /etc/php.ini_BAK_`date +%m%d%Y`_ORIG

#---- NEW ----

sudo pecl install memcached << EOF
yes
EOF
sudo pecl install memcache << EOF
yes
EOF
sudo pecl install geoip << EOF
yes
EOF
sudo pecl install -f apc << EOF
yes
EOF
gsutil cp gs://startup_scripts_us/scripts/apc.ini /home/anand
gsutil cp gs://startup_scripts_us/scripts/php.ini /home/anand

sudo cp -rf /home/anand/php.ini /etc/
sudo cp -rf /home/anand/apc.ini /etc/php.d/

#sudo scp 10.128.0.56:/usr/lib64/php/modules/memcache* /usr/lib64/php/modules/
#sudo scp 10.128.0.56:/usr/lib64/php/modules/apc* /usr/lib64/php/modules/
#sudo scp 10.128.0.56:/usr/lib64/php/modules/geoip* /usr/lib64/php/modules/
#sudo scp 10.128.0.56:/etc/php.ini /etc;
#sudo scp 10.128.0.56:/etc/php.d/apc.ini /etc/php.d/;

sudo service httpd restart;


#---- NEW ----
#yum -y install memcached memcached-devel php54w-pecl-memcached libmemcached10-devel

gsutil cp gs://startup_scripts_us/scripts/memcached /home/anand/
sudo cp -rf /home/anand/memcached /etc/sysconfig/

#sudo scp 10.128.0.56:/etc/sysconfig/memcached /etc/sysconfig;
sudo chkconfig memcached on;
sudo service memcached restart;

# Install Treasuredata

#sudo scp -r 10.128.0.56:/root/treasuredata /root/
gsutil cp -r gs://startup_scripts_us/scripts/treasuredata /home/anand
sudo cp -rf /home/anand/treasuredata /root/

sudo sh /root/treasuredata/install-redhat-td-agent2.sh
gsutil cp gs://startup_scripts_us/scripts/treasuredata/td-agent.conf /home/anand
sudo cp -rf /home/anand/td-agent.conf /etc/td-agent/

#sudo scp 10.128.0.56:/etc/td-agent/td-agent.conf /etc/td-agent/
sudo service td-agent restart

#PST date time
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo yum install ntpdate rdate -y && ntpdate pool.ntp.org && rdate -s time-a.nist.gov;
gsutil cp gs://startup_scripts_us/scripts/ntpdate.sh /home/anand
sudo cp -rf /home/anand/ntpdate.sh /etc/cron.hourly/

#sudo scp 10.128.0.56:/etc/cron.hourly/ntpdate.sh /etc/cron.hourly/

#Install EMACS
#yum install -y emacs

#Install syslog-ng:

gsutil cp gs://startup_scripts_us/scripts/syslog-ng.conf /home/anand/
sudo cp -rf /home/anand/syslog-ng.conf /etc/syslog-ng/

#sudo scp 10.128.0.56:/etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/
sudoservice syslog-ng restart;
sudo chkconfig syslog-ng on;

#Copy CRONJOBS:

gsutil cp gs://startup_scripts_us/scripts/cron_root /home/anand/
sudo cp -rf /home/anand/cron_root /var/spool/cron/root

#sudo scp 10.128.0.56:/var/spool/cron/root /var/spool/cron/;
sudo rm -rf /var/log/sites;sudo mkdir /var/log/sites;
sudo rsync -avz 10.128.0.56:/var/log/sites/ /var/log/sites/;

#Configure services to run on reboot:

sudo service sendmail restart;
sudo chkconfig httpd on;
sudo chkconfig crond on;
sudo sudo chkconfig iptables off;
sudo chkconfig memcached on;
sudo chkconfig sendmail on;
sudo chkconfig syslog-ng on;
sudo chkconfig syslog off;
sudo chkconfig rsyslog off;


# Configure httpd.conf
#sudo scp 10.128.0.56:/etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf
gsutil cp gs://startup_scripts_us/scripts/httpd.conf /home/anand
sudo cp -rf /home/anand/httpd.conf /etc/httpd/conf/

#Add LogFormat + vhosts + etc...

sudo su - << EOF
echo '
LogFormat "%D %{%s}t \"%U\" %>s %O \"%{HTTP_X_FORWARDED_FOR}e\"" smcommon' >> /etc/httpd/conf/httpd.conf &&
echo "

NameVirtualHost *:80

<VirtualHost *:80>
  ServerName `hostname`
  ServerAlias *.mdotm.com *.mdotm.co *.crosschannel.com *.crosschannel.co
  DocumentRoot /var/www/vhosts/mdotm.com/httpdocs
  ErrorLog logs/mdotm.com-error_log
  CustomLog logs/mdotm.com-access_log smcommon
</VirtualHost>

SetEnv mach `hostname -s`
SetEnv rtb true
SetEnv adx true
SetEnv sj true
SetEnv wdc true
SetEnv wdc2 true
SetEnv eu true
SetEnv as true

<Directory />
 Options All
    AllowOverride All
</Directory>

ExtendedStatus On

<Location /server-status>
    SetHandler server-status
    Order Deny,Allow
    Deny from all
    Allow from 127.0.0.1 10.84.81.165 75.126.67.187
</Location>

ServerName `hostname`:80" >> /etc/httpd/conf/httpd.conf;
EOF

sudo service httpd restart;

#COPY "/var/www/vhosts/mdotm.com/httpdocs" FROM OTHER SERVERS:

#USE GIT:
#yum -y install git &&
sudo mkdir -p /var/www/vhosts &&
#sudo su -
cd /var/www/vhosts &&
sudo git clone git@github.com:sourabhniyogi/mdotm.com.git &&
cd /var/www/vhosts/mdotm.com/ &&
sudo git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git &&
sudo git fetch upstream &&
sudo git merge upstream/master

ADD shortcircuit.php manually:
gsutil cp gs://startup_scripts_us/scripts/shortcircuit.php /home/anand
sudo cp -rf /home/anand/shortcircuit.php /var/www/vhosts/mdotm.com/include/
#sudo scp 10.128.0.56:/var/www/vhosts/mdotm.com/include/shortcircuit.php /var/www/vhosts/mdotm.com/include/

################
## Install libtool
sudo mkdir -p /root/downloads && cd /root/downloads
sudo yum -y install libtool* git &&
sudo git clone --recursive https://github.com/maxmind/libmaxminddb &&
#sudo su - << EOF
cd /root/downloads/libmaxminddb &&
sudo ./bootstrap &&
sudo ./configure &&
sudo make &&
sudo make check &&
sudo make install &&
sudo ldconfig
#EOF

#Install PHP Extension maxminddb.so:

cd /root &&
sudo curl -sS https://getcomposer.org/installer | sudo php &&
sudo php composer.phar require geoip2/geoip2:~2.0 &&

## This creates a directory named 'vendor'

cd vendor/maxmind-db/reader/ext &&
sudo phpize &&
sudo ./configure &&
sudo make &&
sudo yes | make test &&
sudo make install &&
sudo ldconfig /usr/local/lib/
sudo rsync -avz /usr/local/lib/*maxmind* /usr/lib64/
#######
# Install Kafka
sudo scp 10.128.0.56:/usr/local/lib/librdkafka.so.1 /usr/local/lib/librdkafka.so.1
sudo ln -s /usr/local/lib/librdkafka.so.1 /usr/lib64/librdkafka.so.1
sudo sed -i '/kafka/d' /etc/php.ini && echo 'extension=kafka.so' >> /etc/php.ini
gsutil cp -r gs://startup_scripts_us/scripts/php_modules /home/anand
sudo cp -rf /home/anand/php_modules/* /usr/lib64/php/modules/

#sudo scp 10.128.0.56:/usr/lib64/php/modules/kafka.so /usr/lib64/php/modules/
#sudo scp 10.128.0.56:/usr/lib64/php/modules/msgpack.so /usr/lib64/php/modules/msgpack.so
#sudo scp 10.128.0.56:/usr/lib64/php/modules/citrusleaf.so /usr/lib64/php/modules/
#sudo scp 10.128.0.56:/usr/lib64/php/modules/igbinary.so /usr/lib64/php/modules/
#sudo scp 10.128.0.56:/usr/lib64/php/modules/aerospike.so /usr/lib64/php/modules/
#sudo scp 10.128.0.56:/usr/lib64/php/modules/maxmind* /usr/lib64/php/modules/

#ADD extension=maxminddb.so to /etc/php.ini
#echo extension=maxminddb.so >> /etc/php.ini
sudo sed -i '/maxminddb.so/d' /etc/php.ini &&
sudo sed -i "$ i\extension=maxminddb.so" /etc/php.ini

#Change 'assumeyes=1' --> 'assumeyes=0' in yum.conf
sudo sed -i '/assumeyes/d' /etc/yum.conf
sudo sed -i "$ i\assumeyes=0" /etc/yum.conf

##############
#Install Nagios/cacti client
#yum -y install nagios nagios-plugins nagios-plugins-nrpe nagios-nrpe gd-devel net-snmp;
sudo yum -y install nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
sudo scp 10.128.0.56:/etc/nagios/nrpe.cfg /etc/nagios/;
sudo scp 10.128.0.56:/usr/lib64/nagios/plugins/* /usr/lib64/nagios/plugins/;
sudo chkconfig nrpe on;
sudo service nrpe restart;
sudo chkconfig snmpd on;

################
#Denyhosts
gsutil cp -r gs://startup_scripts_us/scripts/denyhosts/ /home/anand
sudo cp -rf /home/anand/denyhosts/allowed-hosts /var/lib/denyhosts/
sudo cp -rf /home/anand/denyhosts/denyhosts.conf /etc/

#sudo scp 10.128.0.56:/var/lib/denyhosts/allowed-hosts /var/lib/denyhosts/;
#sudo scp 10.128.0.56:/etc/denyhosts.conf /etc;
sudo service denyhosts restart;
sudo chkconfig denyhosts on;
###############
