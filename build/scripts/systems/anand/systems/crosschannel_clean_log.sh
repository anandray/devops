#!/bin/bash
#unset syslog
#unset crosschannel
#unset syslog1
#unset crosschannel1

#syslog=`ps aux | grep syslog-ng | grep -v grep | grep 'syslog-ng -p /var/run/syslog-ng.pid' | awk '{print$9}'` > /dev/null
#crosschannel=`ps aux | grep crosschannel | grep -v grep | awk '{print$9}'` > /dev/null

echo "
-------------------------LOGROTATION START -------------------------
Before log rotation:
syslog-ng process: `ps aux | grep syslog-ng | grep -v grep | grep 'syslog-ng -p /var/run/syslog-ng.pid'`
crosschannel process: `ps aux | grep crosschannel | grep -vE 'grep|log|tail'`

"
echo "Log Rotation start `date +%T`
======================================================
"
/usr/sbin/logrotate -v -f /etc/logrotate.d/crosschannel
echo "
======================================================
Log Rotation end output - `date +%T`"

sleep 2

echo "
Starting crosschannel..."
cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh crosschannel

#syslog1=`ps aux | grep syslog-ng | grep -v grep | grep 'syslog-ng -p /var/run/syslog-ng.pid' | awk '{print$9}'` > /dev/null
#crosschannel1=`ps aux | grep crosschannel | grep -v grep | awk '{print$9}'` > /dev/null

sleep 10

if ! ps aux | grep crosschannel | grep -vE 'grep|log|tail' > /dev/null;
  then
  cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh crosschannel
fi

echo "
After Log Rotation:
syslog-ng process: `ps aux | grep syslog-ng | grep -v grep | grep 'syslog-ng -p /var/run/syslog-ng.pid'`
crosschannel process: `ps aux | grep crosschannel | grep -vE 'grep|log|tail'`
-------------------------LOGROTATION END -------------------------
"
#unset syslog
#unset crosschannel
#unset syslog1
#unset crosschannel1
