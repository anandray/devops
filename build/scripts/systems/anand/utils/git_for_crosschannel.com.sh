#!/bin/bash
## Use this script to update(git fetch + merge) crosschannel.com
cd /var/www/vhosts
chattr -R -i /var/www/vhosts/crosschannel.com
cd /var/www/vhosts/crosschannel.com
git fetch upstream
git merge upstream/master
#chown -R crosschannel.engineering /var/www/vhosts/crosschannel.com
#find /var/www/vhosts/crosschannel.com -user root -exec chown -R crosschannel.engineering {} \;
#find /var/www/vhosts/crosschannel.com -group root -exec chown -R crosschannel.engineering {} \;
chattr -R +i /var/www/vhosts/crosschannel.com
