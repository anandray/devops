#!/bin/bash

## Copying /etc/hosts
scp 10.128.0.56:/etc/hosts /etc/

## Copying important scripts
rsync -avz 10.128.0.56:/root/scripts/ /root/scripts/
scp 10.128.0.56:/var/www/vhosts/mdotm.com/httpdocs/add_hostname.sh /root/scripts/
scp 10.128.0.56:/root/scripts/memcached_igbinary_install.sh /root/scripts
scp 10.128.0.56:/root/scripts/httpd_conf.sh /root/scripts/

## Removing previous HOSTNAME entry from /etc/sysconfig/network
#sed -i '/HOSTNAME/d'  /etc/sysconfig/network


## Adding hostname to /etc/sysconfig/network
#echo "HOSTNAME=$(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1,"\t"}'` /etc/hosts | awk '{print$2".mdotm.com"}')" >> /etc/sysconfig/network

## Applying hostname thru shell with command hostname
hostname `grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2`

## Applying hostname thru shell with command domainname
#domainname `grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2`

## Applying hostname with sysctl command
sysctl -w kernel.hostname=`grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2` && sysctl -p

## Rename /etc/dhcp/dhclient-exit-hooks and /usr/share/google/set-hostname to prevent hostname reset
#mv -fv /etc/dhcp/dhclient-exit-hooks /etc/dhcp/dhclient-exit-hooks_BAK
#mv -fv /usr/share/google/set-hostname /usr/share/google/set-hostname_BAK

## GIT CLONE TO BRING /var/www/vhosts/mdotm.com ##
sh /root/scripts/www_git_clone.sh

## Copying empty httpd.conf ##
scp 10.128.0.56:/etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf

## Setting up httpd.conf
#sh /root/scripts/httpd_conf.sh
echo '
LogFormat "%D %{%s}t \"%U\" %>s %O \"%{HTTP_X_FORWARDED_FOR}e\"" smcommon' >> /etc/httpd/conf/httpd.conf &&

echo "

NameVirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80

<VirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80>
  ServerName `hostname`
  ServerAlias www.mdotm.com rtb-adx-east.mdotm.com rtb-adx-east.mdotm.co ads.mdotm.com secure.mdotm.com
  DocumentRoot /var/www/vhosts/mdotm.com/httpdocs
  ErrorLog logs/mdotm.com-error_log
  CustomLog logs/mdotm.com-access_log smcommon
</VirtualHost>

SetEnv mach `hostname -s`
SetEnv eu true
SetEnv rtb true
SetEnv adx true
SetEnv sj true
SetEnv rtb true
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

## Copying shortcircuit.php
scp 10.128.0.56:/var/www/vhosts/mdotm.com/include/shortcircuit.php /var/www/vhosts/mdotm.com/include/

## Renaming index.php to index.php_BAK so index.html takes precedence and the backends pass the Google cloud healthcheck
mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK

## Copying /root/treasuredata
scp -r 10.128.0.56:/root/treasuredata /root/

## Installing treasuredata
sh /root/treasuredata/install-redhat-td-agent2.sh

## Copying td-agent.conf
scp 10.128.0.56:/etc/td-agent/td-agent.conf /etc/td-agent/

## Start td-agent service
service td-agent restart

## Install igbinary php extension
#sh /root/scripts/memcached_igbinary_install.sh
pecl channel-update pecl.php.net
pecl install igbinary
sed -i '/igbinary/d' /etc/php.ini
echo '
extension=igbinary.so
session.serialize_handler=igbinary
igbinary.compact_strings=On' >> /etc/php.ini
mv /usr/lib64/php/modules/memcached.so /usr/lib64/php/modules/memcached.so_BAK
scp 10.128.0.56:/usr/lib64/php/modules/memcached.so /usr/lib64/php/modules/
service httpd restart

## Copy /etc/sysconfig/memcached
scp 10.128.0.56:/etc/sysconfig/memcached /etc/sysconfig/memcached
service memcached restart

## Adding CRONTAB
cp -rf /var/spool/cron/root_BAK /var/spool/cron/root
service crond restart

