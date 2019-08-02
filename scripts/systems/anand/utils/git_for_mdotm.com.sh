#!/bin/bash
## Use this script to update(git fetch + merge) mdotm.com
cd /var/www/vhosts
chattr -R -i /var/www/vhosts/mdotm.com
cd /var/www/vhosts/mdotm.com 
git fetch upstream
git merge upstream/master
chown -R mdotm.engineering /var/www/vhosts/mdotm.com
#find /var/www/vhosts/mdotm.com -user root -exec chown -R mdotm.engineering {} \;
#find /var/www/vhosts/mdotm.com -group root -exec chown -R mdotm.engineering {} \;
echo -ne 'Locking /var/www/vhosts/mdotm.com...\n'
sleep 1
echo -ne '######                                  (20%)\r'
sleep 1
echo -ne '##############               	          (40%)\r'
sleep 1
echo -ne '#######################                 (60%)\r'
sleep 1
echo -ne '###############################         (80%)\r'
sleep 1
echo -ne '####################################### (100%)\r'
echo -ne '\n'
chattr -R +i /var/www/vhosts/mdotm.com
