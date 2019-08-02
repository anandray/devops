#!/bin/bash
#if ps aux | grep nrpe.cfg | grep -v grep | awk '{print$1,$13}' | egrep 'nagios|nrpe.cfg' | wc -l
if ps aux | grep nrpe.cfg | grep -v grep | awk '{print$1,$13}' | egrep 'nagios|nrpe.cfg'> /dev/null; then
echo nrpe is running
else
echo nrpe is NOT running
/etc/init.d/nrpe stop;
/etc/init.d/nrpe start;
fi
