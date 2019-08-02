#!/bin/bash
sudo cp -rf /etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf
echo '
LogFormat "%D %{%s}t \"%U\" %>s %O \"%{HTTP_X_FORWARDED_FOR}e\"" smcommon' >> /etc/httpd/conf/httpd.conf;

echo "

NameVirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80

<VirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80>
  ServerName `hostname`
  ServerAlias rtb-adx-eu.mdotm.com rtb-adx.eu.mdotm.co www.mdotm.com ads.mdotm.com secure.mdotm.com
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

<Directory /var/www/vhosts/mdotm.com/httpdocs/ads>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/api>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/ar>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/assets>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/assets2>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/audience>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/campaign>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/casestudies>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/chatbot>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/chrometheme>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/clients>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/creatives>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/css>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/DataTables-1.9.4>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/gce>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/giving>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/gorilla>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/images>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/img>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/js>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/marketing>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/minigames>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/operations>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/pixel>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

<Directory /var/www/vhosts/mdotm.com/httpdocs/react>
Options Includes FollowSymLinks MultiViews
    AllowOverride All
</Directory>

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
