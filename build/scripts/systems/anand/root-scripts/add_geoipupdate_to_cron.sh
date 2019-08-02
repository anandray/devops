#!/bin/bash
echo "30 1,13 * * * ssh `hostname` sh /var/www/vhosts/mdotm.com/scripts/systems/geoipupdate.sh > /var/log/geoipupdate.log" >> /var/spool/cron/root
