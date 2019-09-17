#!/bin/bash

if ps aux | grep rsyslogd | grep -v grep; then
service rsyslog stop
chkconfig rsyslog off
wget -O /etc/syslog-ng/syslog-ng.conf http://www6001.wolk.com/.start/syslog-ng.conf
service syslog-ng restart
fi
