#!/bin/bash

hostname `grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2`
domainname `grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2`
sysctl -w kernel.hostname=`grep HOSTNAME /etc/sysconfig/network | cut -d "=" -f2` && sysctl -p
#service httpd restart

#scp 10.142.2.2:/etc/hosts /etc/
#sed -i '/HOSTNAME/d'  /etc/sysconfig/network
#hostname `grep \`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'\` /etc/hosts | awk '{print$2".mdotm.com"}'`
#hostname `grep "`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'` " /etc/hosts | awk '{print$2".mdotm.com"}'`
#hostname $(grep "`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'` " /etc/hosts | awk '{print$2".mdotm.com"}')
#echo "HOSTNAME=$(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1,"\t"}'` /etc/hosts | awk '{print$2".mdotm.com"}')" >> /etc/sysconfig/network
#hostname $(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1,"\t"}'` /etc/hosts | awk '{print$2".mdotm.com"}')
#sh /root/scripts/www_git_clone.sh
#scp 10.142.2.2:/etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf
#scp 10.142.2.2:/var/www/vhosts/mdotm.com/include/shortcircuit.php /var/www/vhosts/mdotm.com/include/
#mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK
#sh /root/scripts/httpd_conf.sh
##service httpd restart
#cp -rf /var/spool/cron/root_BAK /var/spool/cron/root
#service crond restart
