#!/bin/bash
ssh `hostname` /usr/bin/php /var/www/vhosts/mdotm.com/cron/www/loadAll.php
