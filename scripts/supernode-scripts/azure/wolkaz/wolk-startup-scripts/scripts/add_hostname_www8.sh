#!/bin/bash

## Copying /etc/hosts
scp 10.132.0.5:/etc/hosts /etc/

## Copying important scripts
rsync -avz 10.132.0.5:/root/scripts/ /root/scripts/
scp 10.132.0.5:/root/scripts/memcached_igbinary_install.sh /root/scripts

## Removing previous HOSTNAME entry from /etc/sysconfig/network
sed -i '/HOSTNAME/d'  /etc/sysconfig/network
#sed -i '/www/d'  /etc/sysconfig/network


## Adding hostname to /etc/sysconfig/network
echo "HOSTNAME=$(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1,"\t"}'` /etc/hosts | awk '{print$2".mdotm.com"}')" >> /etc/sysconfig/network
## echo "HOSTNAME=$(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'` /etc/hosts | awk '{print$2".mdotm.com"}')" >> /etc/sysconfig/network

## Applying hostname thru shell with command hostname
hostname `grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2`
## hostname $(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1,"\t"}'` /etc/hosts | awk '{print$2".mdotm.com"}') ## not using this
## hostname `grep \`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'\` /etc/hosts | awk '{print$2".mdotm.com"}'`
## hostname `grep "`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'` " /etc/hosts | awk '{print$2".mdotm.com"}'`
## hostname $(grep "`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'` " /etc/hosts | awk '{print$2".mdotm.com"}')

## Applying hostname thru shell with command domainname
domainname `grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2`

## Applying hostname with sysctl command
sysctl -w kernel.hostname=`grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2` && sysctl -p

## Rename /etc/dhcp/dhclient-exit-hooks and /usr/share/google/set-hostname to prevent hostname reset
mv -fv /etc/dhcp/dhclient-exit-hooks /etc/dhcp/dhclient-exit-hooks_BAK
mv -fv /usr/share/google/set-hostname /usr/share/google/set-hostname_BAK

## GIT CLONE TO BRING /var/www/vhosts/mdotm.com ##
sh /root/scripts/www_git_clone.sh

## Copying empty httpd.conf ##
scp 10.132.0.5:/etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf

## Setting up httpd.conf
sh /root/scripts/httpd_conf.sh

## Copying shortcircuit.php
scp 10.132.0.5:/var/www/vhosts/mdotm.com/include/shortcircuit.php /var/www/vhosts/mdotm.com/include/

## Renaming index.php to index.php_BAK so index.html takes precedence and the backends pass the Google cloud healthcheck
mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK

## Copying /root/treasuredata
scp -r 10.132.0.5:/root/treasuredata /root/

## Installing treasuredata
sh /root/treasuredata/install-redhat-td-agent2.sh

## Copying td-agent.conf
scp 10.132.0.5:/etc/td-agent/td-agent.conf /etc/td-agent/

## Start td-agent service
service td-agent restart

## Install igbinary php extension
sh /root/scripts/memcached_igbinary_install.sh

## Copy /etc/sysconfig/memcached
scp 10.132.0.5:/etc/sysconfig/memcached /etc/sysconfig/memcached
service memcached restart

## Adding CRONTAB
cp -rf /var/spool/cron/root_BAK /var/spool/cron/root
service crond restart

