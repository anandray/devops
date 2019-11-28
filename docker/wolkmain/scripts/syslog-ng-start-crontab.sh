#!/bin/bash
echo "*/1 * * * * /wolk/scripts/syslog-ng-start.sh &>> /var/log/syslog-ng-start.log" >> /var/spool/cron/root
