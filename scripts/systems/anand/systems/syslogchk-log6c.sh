#!/bin/bash

if ! ps aux | grep 'php /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log6c.php' | grep -v grep; then
for i in {1..11};
do php /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log6c.php;
sleep 5;
done
fi
