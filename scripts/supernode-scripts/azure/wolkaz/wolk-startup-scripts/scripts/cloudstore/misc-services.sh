#!/bin/bash

/root/scripts/nrpe-install.sh &> /var/log/nrpe-install.log
sleep 5
/root/scripts/syslog-ng-start.sh &> /var/log/syslog-ng-start.log
sleep 5
/root/scripts/wolk-start.sh &> /var/log/wolk-start.log
sleep 5
/bin/sed -i 's/\* \* \* \* \* \/root\/scripts\/misc-services.sh/d' /var/spool/cron/root
