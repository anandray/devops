#!/bin/bash

if ! ps aux | grep 'sh /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log2-bash.sh' | grep -v grep; then
for i in {1..5};
do sh /var/www/vhosts/mdotm.com/scripts/systems/syslogchk-log2-bash-exec.sh;
sleep 10;
done
fi
