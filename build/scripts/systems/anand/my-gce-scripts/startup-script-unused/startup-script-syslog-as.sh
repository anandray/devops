#!/bin/bash

# running loadAll.php to populate /var/log/sites
sudo /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php;

# populate httpd.conf
sudo cp -rf /etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf;
echo '
LogFormat "%D %{%s}t \"%U\" %>s %O \"%{HTTP_X_FORWARDED_FOR}e\"" smcommon' >> /etc/httpd/conf/httpd.conf;

echo "

NameVirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80

<VirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80>
  ServerName `hostname`
  ServerAlias rtb-adx-as.mdotm.com rtb-adx-as.mdotm.co www.mdotm.com ads.mdotm.com secure.mdotm.com
  DocumentRoot /var/www/vhosts/mdotm.com/httpdocs
  ErrorLog logs/mdotm.com-error_log
  CustomLog logs/mdotm.com-access_log smcommon
</VirtualHost>

SetEnv mach `hostname -s`
SetEnv sj true
SetEnv wdc true
SetEnv wdc2 true
SetEnv as true
SetEnv eu true
SetEnv adx true
SetEnv rtb true

<Directory />
 Options All
    AllowOverride All
</Directory>

ExtendedStatus On

<Location /server-status>
    SetHandler server-status
    Order Deny,Allow
#    Deny from all
    Allow from all
</Location>
ServerName `hostname`:80" >> /etc/httpd/conf/httpd.conf;
service httpd restart;

# Fetching hbase config file
#sudo wget -N -O /usr/local/hbase-1.1.2/conf/hbase-site.xml http://anand.www1001.mdotm.com/gce/hbase/hbase-site-as.xml;
#sudo wget -O /usr/local/hbase-1.1.2/conf/hbase-site.xml http://anand.www1001.mdotm.com/gce/hbase/hbase-site-as.xml;
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-site-as.xml /home/anand;
sudo cp -rf /home/anand/hbase-site-as.xml /usr/local/hbase-1.1.2/conf/hbase-site.xml;

# Fetching syslog-ng.conf
gsutil cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-wwwa /home/anand/syslog-ng.conf;
sudo cp -rf /home/anand/syslog-ng.conf /etc/syslog-ng/;
sudo service syslog-ng restart;

# installing openssl libraries required for mongo php extensions
sudo yum -y install openssl openssl-devel;

# mongo php extensions
sudo pecl channel-update pecl.php.net;
no '' | sudo pecl install mongo;
sudo pecl install mongodb;

# pdf and lua php extensions
gsutil cp gs://startup_scripts_us/scripts/php_modules/pdf.so /home/anand;
sudo cp -rf /home/anand/pdf.so /usr/lib64/php/modules/;
gsutil cp gs://startup_scripts_us/scripts/php_modules/lua.so /home/anand;
sudo cp -rf /home/anand/lua.so /usr/lib64/php/modules/;

# removing extensions from php.ini before adding to avoud duplicacies
sudo sed -i '/mongo/d' /etc/php.ini;
sudo sed -i '/ffmpeg/d' /etc/php.ini;
sudo sed -i '/igbinary/d' /etc/php.ini;
sudo sed -i '/pdf.so/d' /etc/php.ini;
sudo sed -i '/lua.so/d' /etc/php.ini;

# adding php extensions to php.ini
sudo su - << EOF
echo 'extension=mongo.so' >> /etc/php.ini;
echo 'extension=mongodb.so' >> /etc/php.ini;
echo '
extension=igbinary.so
session.serialize_handler=igbinary
igbinary.compact_strings=On' >> /etc/php.ini
echo 'extension=pdf.so' >> /etc/php.ini;
echo 'extension=lua.so' >> /etc/php.ini;
EOF
sudo service httpd restart;

# OR download the mongo.sh script and execute
# wget -O /home/anand/mongo.sh http://anand.www1001.mdotm.com/gce/mongo.sh;
# sudo /bin/sh /home/anand/mongo.sh;

# Installing google-fluentd
#gsutil cp gs://startup_scripts_us/scripts/install-logging-agent.sh /home/anand;
#sudo bash  /home/anand/install-logging-agent.sh;

# stop/remove google-fluentd
sudo /sbin/service google-fluentd stop
sudo yum -y remove google-fluentd google-fluentd-catch-all-config
