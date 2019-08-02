#SSH Keys:
#RUN the following on LOG02:
#scp ~/.ssh/* wwwXX:~/.ssh;

# Allow SSH-ing to any server
#yes '' | scp www2042:/etc/ssh/ssh_config /etc/ssh/ && service sshd restart;
sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config
service sshd restart

# Copy /etc/hosts from log04 (10.9.148.198):
scp 10.9.148.198:/etc/hosts /etc/

# Copy /etc/resolv.conf
scp www2042:/etc/resolv.conf /etc/

# Enable histtimeformat
scp www2042:/etc/profile.d/histtimeformat.sh /etc/profile.d/

# Enable IPTABLES
scp -r www2042:/root/scripts /root/
scp -r www2042:/root/treasuredata /root/
sh /root/scripts/iptables.sh

## DISABLE FSCK
tune2fs -c 0 -i 0 /dev/sda1
tune2fs -c 0 -i 0 /dev/sda3

#DISABLE SELINUX:
setenforce 0 && getenforce && setsebool -P httpd_can_network_connect=1;
cp /etc/selinux/config /etc/selinux/config_ORIG;
yes '' | scp www2042:/etc/selinux/config /etc/selinux/
#yes '' | scp www2042:/etc/hosts /etc/
yes '' | scp www2042:/etc/security/limits.conf /etc/security/

#PHP INSTALL:

#### Redhat 6.x ####
rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
#rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-13.ius.el6.noarch.rpm;
rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-14.ius.el6.noarch.rpm
#rpm -Uvh https://mirror.webtatic.com/yum/el6/latest.rpm
########


service syslog stop;
chkconfig syslog off;
service postfix stop;
chkconfig postfix off;
chkconfig --del postfix;
service rsyslog stop;
chkconfig rsyslog off;
chmod 0000 /usr/sbin/postfix;


yum remove -y php*;
yum -y --enablerepo=ius-archive install php54 php54-soap php54-gd php54-ioncube-loader php54-pecl-memcache php54-mcrypt php54-imap php54-devel php54-cli php54-mysql php54-mbstring php54-xml libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql memcached memcached-devel mysql php54-pecl-memcached libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail denyhosts procmail python-argparse *whois;


#Copy GeoIP.dat
yes '' | scp -r www2042:/usr/share/GeoIP /usr/share/;
service httpd restart;

#Install PHP extensions:
cp /etc/php.ini /etc/php.ini_BAK_`date +%m%d%Y`_ORIG

#---- NEW ----

yes '' | pecl install memcached;
yes '' | pecl install geoip; 
yes '' | pecl install -f apc;


yes '' | scp www2042:/etc/php.ini /etc;
yes '' | scp www2042:/etc/php.d/apc.ini /etc/php.d/;
service httpd restart;


#---- NEW ----
#yum -y install memcached memcached-devel php54w-pecl-memcached libmemcached10-devel

yes '' | scp www2042:/etc/sysconfig/memcached /etc/sysconfig;
chkconfig memcached on;
service memcached restart;

# Install Treasuredata
sh /root/treasuredata/treasuredata_agent_install.sh
scp www2042:/etc/td-agent/td-agent.conf /etc/td-agent/
service td-agent restart

#PST date time
ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
yum install ntpdate rdate -y && ntpdate pool.ntp.org && rdate -s time-a.nist.gov;
scp www2042:/etc/cron.hourly/ntpdate.sh /etc/cron.hourly/

#Install EMACS
#yum install -y emacs

#Install syslog-ng:

#service syslog stop;
#chkconfig syslog off;
#service rsyslog stop;
#chkconfig rsyslog off;
#yum -y install syslog-ng syslog-ng-libdbi libdbi-devel telnet
yes '' | scp www2042:/etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/
service syslog-ng restart;
chkconfig syslog-ng on;

#Copy CRONJOBS:

scp www2042:/var/spool/cron/root /var/spool/cron/;
rm -rf /var/log/sites;mkdir /var/log/sites;
rsync -avz www2042:/var/log/sites/ /var/log/sites/;

#Configure services to run on reboot:

service sendmail restart;
chkconfig httpd on;
chkconfig crond on;
chkconfig iptables off;
chkconfig memcached on;
chkconfig sendmail on;
chkconfig syslog-ng on;
chkconfig syslog off;
chkconfig rsyslog off;


#Add LogFormat + vhosts + etc...

echo '
LogFormat "%D %{%s}t \"%U\" %>s %O \"%{HTTP_X_FORWARDED_FOR}e\"" smcommon' >> /etc/httpd/conf/httpd.conf &&

echo "

NameVirtualHost `ifconfig bond0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80
NameVirtualHost `ifconfig bond1| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80

<VirtualHost `ifconfig bond0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80>
  ServerName `hostname`
  ServerAlias www.mdotm.com ads-sj.mdotm.co  bidssj.mdotm.com
  DocumentRoot /var/www/vhosts/mdotm.com/httpdocs
  ErrorLog logs/mdotm.com-error_log
  CustomLog logs/mdotm.com-access_log smcommon
</VirtualHost>

<VirtualHost `ifconfig bond1| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80>
  ServerName `hostname`
  ServerAlias www.mdotm.com ads-sj.mdotm.co  bidssj.mdotm.com
  DocumentRoot /var/www/vhosts/mdotm.com/httpdocs
  ErrorLog logs/mdotm.com-error_log
  CustomLog logs/mdotm.com-access_log smcommon
</VirtualHost>

SetEnv mach `hostname -s`
SetEnv sj true
SetEnv rtb true
SetEnv adx true
SetEnv wdc true
SetEnv wdc2 true

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
service httpd restart;



#COPY "/var/www/vhosts/mdotm.com/httpdocs" FROM OTHER SERVERS:

#USE GIT:
#yum -y install git &&
mkdir /var/www/vhosts &&
cd /var/www/vhosts &&
git clone git@github.com:sourabhniyogi/mdotm.com.git &&
cd /var/www/vhosts/mdotm.com/ &&
git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git &&
git fetch upstream &&
git merge upstream/master

ADD shortcircuit.php manually:
scp www2042:/var/www/vhosts/mdotm.com/include/shortcircuit.php /var/www/vhosts/mdotm.com/include/

################
## Install libtool
cd /root
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
#######
# Install Kafka
scp www2042:/usr/local/lib/librdkafka.so.1 /usr/local/lib/librdkafka.so.1
ln -s /usr/local/lib/librdkafka.so.1 /usr/lib64/librdkafka.so.1
sed -i '/kafka/d' /etc/php.ini && echo 'extension=kafka.so' >> /etc/php.ini
scp www2042:/usr/lib64/php/modules/kafka.so /usr/lib64/php/modules/
scp www2042:/usr/lib64/php/modules/msgpack.so /usr/lib64/php/modules/msgpack.so
scp www2042:/usr/lib64/php/modules/citrusleaf.so /usr/lib64/php/modules/
scp www2042:/usr/lib64/php/modules/igbinary.so /usr/lib64/php/modules/
scp www2042:/usr/lib64/php/modules/aerospike.so /usr/lib64/php/modules/

#ADD extension=maxminddb.so to /etc/php.ini
#echo extension=maxminddb.so >> /etc/php.ini
sed -i '/maxminddb.so/d' /etc/php.ini &&
sed -i "$ i\extension=maxminddb.so" /etc/php.ini

#Change 'assumeyes=1' --> 'assumeyes=0' in yum.conf
sed -i '/assumeyes/d' /etc/yum.conf
sed -i "$ i\assumeyes=0" /etc/yum.conf

##############
#Install Nagios/cacti client
#yum -y install nagios nagios-plugins nagios-plugins-nrpe nagios-nrpe gd-devel net-snmp;
yum -y install nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
scp www1016:/etc/nagios/nrpe.cfg /etc/nagios/;
scp www1016:/usr/lib64/nagios/plugins/* /usr/lib64/nagios/plugins/;
chkconfig nrpe on;
service nrpe restart;
chkconfig snmpd on;

################
#Denyhosts
scp www2042:/var/lib/denyhosts/allowed-hosts /var/lib/denyhosts/;
scp www2042:/etc/denyhosts.conf /etc;
service denyhosts restart;
chkconfig denyhosts on;
###############
## HADOOP INSTALL ##
#sh /root/scripts/hadoop_install.sh && sh /root/scripts/ha_start_namenode_datanode.sh
