#!/bin/bash

# git merge update
cd /var/www/vhosts/mdotm.com
sudo git fetch upstream   
sudo git merge upstream/master

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
sudo service httpd restart;
sudo service memcached restart;
