#!/bin/bash
sed -i 's/\/\/echo/echo/g' /var/www/vhosts/mdotm.com/httpdocs/ads/systems/newservercheck.php
# run newservercheck.php
php /var/www/vhosts/mdotm.com/httpdocs/ads/systems/newservercheck.php && 
sed -i '/run_newservercheck.php.sh/d' /var/spool/cron/root
#sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/run_newservercheck.php.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/run_newservercheck.php.sh/g' /var/spool/cron/root
