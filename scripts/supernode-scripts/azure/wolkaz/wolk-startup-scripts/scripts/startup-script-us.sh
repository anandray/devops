#!/bin/bash

# adding crontab
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/cron_root /var/spool/cron/root;
service crond restart;
EOF

# running loadAll.php to populate /var/log/sites
sudo /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php;

## populate httpd.conf
#sudo cp -rf /etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf;
#echo '
#LogFormat "%D %{%s}t \"%U\" %>s %O \"%{HTTP_X_FORWARDED_FOR}e\"" smcommon' >> /etc/httpd/conf/httpd.conf;
#
#echo "
#
#NameVirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80
#
#<VirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80>
#  ServerName `hostname`
#  ServerAlias rtb-adx-west.mdotm.com rtb-adx-east.mdotm.com rtb-adx-west.mdotm.co rtb-adx-east.mdotm.co www.mdotm.com ads.mdotm.com secure.mdotm.com
#  DocumentRoot /var/www/vhosts/mdotm.com/httpdocs
#  ErrorLog logs/mdotm.com-error_log
#  CustomLog logs/mdotm.com-access_log smcommon
#</VirtualHost>
#
#SetEnv mach `hostname -s`
#SetEnv sj true
#SetEnv wdc true
#SetEnv wdc2 true
#SetEnv as true
#SetEnv eu true
#SetEnv adx true
#SetEnv rtb true
#
#<Directory />
# Options All
#    AllowOverride All
#</Directory>
#
#ExtendedStatus On
#
#<Location /server-status>
#    SetHandler server-status
#    Order Deny,Allow
#    Deny from all
#    Allow from all
#</Location>
#ServerName `hostname`:80" >> /etc/httpd/conf/httpd.conf;
service httpd restart;

# adding Defaults PATH to sudoers
sudo sed -i '/secure_path/d' /etc/sudoers
echo "Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin" >> /etc/sudoers

# Fetching hbase config file
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-site-us.xml /home/anand;
sudo cp -rf /home/anand/hbase-site-us.xml /usr/local/hbase-1.1.2/conf/hbase-site.xml;

# copying hosts file
gsutil cp gs://startup_scripts_us/scripts/hosts /home/anand/hosts;
sudo cp -rf /home/anand/hosts /etc/;

# Fetching syslog-ng.conf
gsutil cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /home/anand/syslog-ng.conf;
sudo cp -rf /home/anand/syslog-ng.conf /etc/syslog-ng/;
sudo service syslog-ng restart;

# installing openssl libraries required for mongo php extensions
#sudo yum -y install openssl openssl-devel php-pear-Net-SMTP php54-pear php-pear-Net-Socket php-pear-Auth-SASL;

## installing pear mail/Mail.php
#sudo su - << EOF
##yum -y install php-pear-Net-SMTP php54-pear php-pear-Net-Socket php-pear-Auth-SASL # already installed above
#pear channel-update pear.php.net
#pear install mail

## add to include_path in php.ini
#sed -i 's/include_path/;include_path/g' /etc/php.ini && echo 'include_path = ".:/var/www/vhosts/mdotm.com/include/:/usr/share/pear/:/usr/share/pear/Mail/:/usr/share/GeoIP/"' >> /etc/php.ini
#service httpd restart
#EOF

## mongo php extensions
#sudo pecl channel-update pecl.php.net;
#no '' | sudo pecl install mongo;
#sudo pecl install mongodb;
#
## pdf and lua php extensions
#gsutil cp gs://startup_scripts_us/scripts/php_modules/pdf.so /home/anand;
#sudo cp -rf /home/anand/pdf.so /usr/lib64/php/modules/;
#gsutil cp gs://startup_scripts_us/scripts/php_modules/lua.so /home/anand;
#sudo cp -rf /home/anand/lua.so /usr/lib64/php/modules/;
#
## removing extensions from php.ini before adding to avoud duplicacies
#sudo sed -i '/mongo/d' /etc/php.ini;
#sudo sed -i '/ffmpeg/d' /etc/php.ini;
#sudo sed -i '/igbinary/d' /etc/php.ini;
#sudo sed -i '/pdf.so/d' /etc/php.ini;
#sudo sed -i '/lua.so/d' /etc/php.ini;
#
## adding php extensions to php.ini
#sudo su - << EOF
#echo 'extension=mongo.so' >> /etc/php.ini;
#echo 'extension=mongodb.so' >> /etc/php.ini;
#echo '
#extension=igbinary.so
#session.serialize_handler=igbinary
#igbinary.compact_strings=On' >> /etc/php.ini
#echo 'extension=pdf.so' >> /etc/php.ini;
#echo 'extension=lua.so' >> /etc/php.ini;
#EOF
sudo service httpd restart;

sudo touch /root/new-startup-script-used_`date +%m%d%Y-%T`

php /home/anand/mailer.php
