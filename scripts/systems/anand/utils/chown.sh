#!/bin/bash
##chattr -R -i /var/www/vhosts/mdotm.com 2>&1 &
##find /var/www/vhosts/mdotm.com -user root -exec chown -R mdotm.engineering {} \; 2>&1 &
##find /var/www/vhosts/mdotm.com -user root -exec chown -R mdotm.engineering {} \; 2>&1 &
##chattr -R +i /var/www/vhosts/mdotm.com 2>&1 &
#find /var/www/vhosts/adops -user root -exec chown -R adops.engineering {} \; 2>&1 &
#find /var/www/vhosts/alina -user root -exec chown -R alina.engineering {} \; 2>&1 &
#find /var/www/vhosts/anand -user root -exec chown -R anand.engineering {} \; 2>&1 &
#find /var/www/vhosts/bruce -user root -exec chown -R bruce.engineering {} \; 2>&1 &
#find /var/www/vhosts/huisi -user root -exec chown -R huisi.engineering {} \; 2>&1 &
#find /var/www/vhosts/mayumi -user root -exec chown -R mayumi.engineering {} \; 2>&1 &
#find /var/www/vhosts/rodney -user root -exec chown -R rodney.engineering {} \; 2>&1 &
#find /var/www/vhosts/sourabh -user root -exec chown -R sourabh.engineering {} \; 2>&1 &
#find /var/www/vhosts/yaron -user root -exec chown -R yaron.engineering {} \; 2>&1 &
chown -R mdotm.engineering /var/www/vhosts/mdotm.com 2>&1 &
#chown -R mdotm.engineering /var/www/html/phpMyAdmin 2>&1 &
chown -R adops.engineering /var/www/vhosts/adops 2>&1 &
chown -R alina.engineering /var/www/vhosts/alina 2>&1 &
chown -R anand.engineering /var/www/vhosts/anand 2>&1 &
chown -R bruce.engineering /var/www/vhosts/bruce 2>&1 &
#chown -R huisi.engineering /var/www/vhosts/huisi 2>&1 &
chown -R mayumi.engineering /var/www/vhosts/mayumi 2>&1 &
chown -R rodney.engineering /var/www/vhosts/rodney 2>&1 &
chown -R sourabh.engineering /var/www/vhosts/sourabh 2>&1 &
#chown -R amelia.engineering /var/www/vhosts/amelia 2>&1 &
chown -R upload.upload /home/upload 2>&1 &
chown -R yaron.engineering /var/www/vhosts/yaron 2>&1 &
#chown -R minigames.engineering /var/www/vhosts/minigames 2>&1 &
#chown -R nitin.engineering /var/www/vhosts/nitin 2>&1 &
##chown -R apache.apache /var/log/uploadjob/ 2>&1 &
##chmod -R 0777 /var/log/uploadjob/ 2>&1 &
#chown -R sourabh.engineering /var/www/vhosts/crosschannel.com 2>&1 &
chown -R michael.engineering /var/www/vhosts/michael 2>&1 &
chmod -R +x /var/www/vhosts/mdotm.com/scripts/utils 2>&1 &
