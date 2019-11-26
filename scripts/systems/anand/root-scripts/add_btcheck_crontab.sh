#!/bin/bash
sed -i "/btcheck.sh/d" /var/spool/cron/root
sed -i "$ a\* * * * * sh /var/www/vhosts/mdotm.com/scripts/systems/btcheck.sh &>> /var/log/btcheck.log" /var/spool/cron/root
