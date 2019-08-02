#!/bin/bash

if ! ps aux | grep 'sh /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log00-bash.sh' | grep -v grep; then
for i in {1..11};
do sh /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log00-bash.sh;
sleep 5;
done
fi
