#!/bin/bash

service rsyslog stop;
wget -O /etc/syslog-ng/syslog-ng.conf http://www6001.wolk.com/.start/syslog-ng.conf;
service syslog-ng restart
