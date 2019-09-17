#!/bin/bash
if ps aux | grep syslog-ng | grep -v grep | egrep 'supervising|syslog-ng' | awk '{print$11,$12,$13}' && netstat -apn | grep '0.0.0.0:5000' > /dev/null; then
echo syslog-ng is running
sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/syslogchk.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/syslogchk.sh/g' /var/spool/cron/root
else
echo syslog-ng is NOT running
/sbin/service syslog-ng restart

if [ ! -f /etc/syslog-ng/syslog-ng.conf ]; then
gsutil cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/syslog-ng.conf;
service syslog-ng restart;
chkconfig syslog-ng on;
fi
fi
