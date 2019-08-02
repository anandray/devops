#!/bin/bash

if ! ps aux | grep syslogchk-log00.php | grep -v grep; then
for i in {1..11};
do php /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log00.php &>> /var/log/syslog-ng-check.log;
sleep 5;
done
fi
