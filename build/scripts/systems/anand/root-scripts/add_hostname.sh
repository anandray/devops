#!/bin/bash
scp 10.128.0.56:/etc/hosts /etc/
sed -i '/HOSTNAME/d'  /etc/sysconfig/network
#sed -i '/www/d'  /etc/sysconfig/network
#echo "HOSTNAME=$(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'` /etc/hosts | awk '{print$2".mdotm.com"}')" >> /etc/sysconfig/network
#hostname `grep \`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'\` /etc/hosts | awk '{print$2".mdotm.com"}'`
#hostname `grep "`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'` " /etc/hosts | awk '{print$2".mdotm.com"}'`
#hostname $(grep "`ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'` " /etc/hosts | awk '{print$2".mdotm.com"}')
echo "HOSTNAME=$(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1,"\t"}'` /etc/hosts | awk '{print$2".mdotm.com"}')" >> /etc/sysconfig/network
hostname $(grep `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1,"\t"}'` /etc/hosts | awk '{print$2".mdotm.com"}')
sh /root/scripts/www_git_clone.sh
scp 10.128.0.56:/etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf
scp 10.128.0.56:/var/www/vhosts/mdotm.com/include/shortcircuit.php /var/www/vhosts/mdotm.com/include/
mv -fv /var/www/vhosts/mdotm.com/httpdocs/index.php /var/www/vhosts/mdotm.com/httpdocs/index.php_BAK
sh /root/scripts/httpd_conf.sh
#service httpd restart
cp -rf /var/spool/cron/root_BAK /var/spool/cron/root
service crond restart
