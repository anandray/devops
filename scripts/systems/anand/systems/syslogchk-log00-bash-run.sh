#!/bin/bash

if ! ps aux | grep 'sh /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log00-bash.sh' | grep -v grep; then
for i in {1..5};
do sh /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log00-bash-exec.sh;
sleep 10;
done
fi
