#!/bin/bash

# populate httpd.conf
sudo cp -rf /etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf;
echo '
LogFormat "%D %{%s}t \"%U\" %>s %O \"%{HTTP_X_FORWARDED_FOR}e\"" smcommon' >> /etc/httpd/conf/httpd.conf;

echo "

NameVirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80

<VirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80>
  ServerName `hostname`
  ServerAlias rtb-adx-west.mdotm.com rtb-adx-east.mdotm.com rtb-adx-east.mdotm.co www.mdotm.com ads.mdotm.com secure.mdotm.com
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
#sudo wget -N -O /usr/local/hbase-1.1.2/conf/hbase-site.xml http://anand.www1001.mdotm.com/gce/hbase/hbase-site-us.xml;
sudo wget -O /usr/local/hbase-1.1.2/conf/hbase-site.xml http://anand.www1001.mdotm.com/gce/hbase/hbase-site-us.xml;

# Fetching syslog-ng.conf
gsutil cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-log /home/anand/syslog-ng.conf;
sudo cp -rf /home/anand/syslog-ng.conf /etc/syslog-ng/;
sudo service syslog-ng restart;

# mongo php extensions
no '' | sudo pecl install mongo;
sudo pecl install mongodb;
sudo su - << EOF
echo 'extension=mongo.so' >> /etc/php.ini;
echo 'extension=mongodb.so' >> /etc/php.ini;
EOF
sudo service httpd restart;

# Installing google-fluentd
gsutil cp gs://startup_scripts_us/scripts/install-logging-agent.sh /home/anand;
sudo bash  /home/anand/install-logging-agent.sh;
