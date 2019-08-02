#!/bin/bash

if ! ps aux | grep 'php /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log6.php' | grep -v grep; then
for i in {1..11};
do php /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log6.php;
sleep 5;
done
fi
