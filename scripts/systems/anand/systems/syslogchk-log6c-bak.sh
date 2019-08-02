#!/bin/bash
YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`

# LOG = /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.log

if ! stat /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.log | grep Modify | grep "`date +'%Y-%m-%d %T'`";
	then
	/sbin/service syslog-ng restart
else
echo "syslog LOG is OK"
fi

# TRACK = /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.track
if ! stat /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.track | grep Modify | grep "`date +'%Y-%m-%d %T'`";
        then
       /sbin/service syslog-ng restart
else
echo "syslog TRACK is OK"
fi

# CONVERSION = /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.conversion
if ! stat /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.conversion | grep Modify | grep "`date +'%Y-%m-%d %T'`";
        then
       /sbin/service syslog-ng restart
else
echo "syslog CONVERSION is OK"
fi
