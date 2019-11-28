#!/bin/bash
if ! ps aux | grep '/usr/sbin/syslog-ng -F -p /var/run/syslogd.pid' | grep -v grep &> /dev/null; then
echo "`date +%T` - syslog-ng not runing.. starting syslog-ng..."
/usr/sbin/syslog-ng -F -p /var/run/syslogd.pid &
fi
