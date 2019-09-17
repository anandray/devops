#!/bin/bash

/root/scripts/nrpe-install.sh &> /var/log/nrpe-install.log
sleep 5
/root/scripts/syslog-ng-start.sh &> /var/log/syslog-ng-start.log

