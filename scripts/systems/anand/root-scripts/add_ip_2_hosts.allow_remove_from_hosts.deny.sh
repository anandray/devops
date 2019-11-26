#!/bin/bash
> /var/lib/denyhosts/allowed-hosts
echo "127.0.0.1" > /var/lib/denyhosts/allowed-hosts;
echo `ifconfig | grep 'inet addr:10' | awk '{print$2}' | cut -d ":" -f2` >> /var/lib/denyhosts/allowed-hosts;
> /etc/hosts.deny;
#ssh `hostname` /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php
service denyhosts restart;
> /etc/hosts.deny;
service denyhosts restart;
