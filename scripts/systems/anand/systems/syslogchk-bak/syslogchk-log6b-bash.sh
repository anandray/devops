#!/bin/bash
YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`

# LOG = /disk1/log/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.log

if ! stat /disk1/log/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.log | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "
  syslog LOG FAILED - TRY AGAIN"
if ! stat /disk1/log/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.log | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "syslog LOG FAILED AGAIN - RESTART SYSLOG-NG"
    if ! netstat -apn | grep '0 0.0.0.0:5000' | grep LISTEN;
    then
    /sbin/service syslog-ng restart
  else
    pkill -9 syslog-ng && /sbin/service syslog-ng restart
    fi
  sleep 5
fi
else
echo "syslog LOG is OK
     "
fi

YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`

# TRACK = /disk1/log/track/$YEAR/$MONTH/$DAY/$HOUR$MIN.track
if ! stat /disk1/log/track/$YEAR/$MONTH/$DAY/$HOUR$MIN.track | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "syslog TRACK FAILED - TRY AGAIN"
if ! stat /disk1/log/track/$YEAR/$MONTH/$DAY/$HOUR$MIN.track | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "syslog TRACK FAILED AGAIN - RESTART SYSLOG-NG"
    if ! netstat -apn | grep '0 0.0.0.0:5000' | grep LISTEN;
    then
    /sbin/service syslog-ng restart
  else
    pkill -9 syslog-ng && /sbin/service syslog-ng restart
    fi
  sleep 5
fi
else
echo "syslog TRACK is OK
     "
fi

YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`

# CONVERSION = /disk1/log/conversion/$YEAR/$MONTH/$DAY/$HOUR$MIN.conversion
if ! stat /disk1/log/conversion/$YEAR/$MONTH/$DAY/$HOUR$MIN.conversion | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "syslog CONVERSION FAILED - TRY AGAIN"
if ! stat /disk1/log/conversion/$YEAR/$MONTH/$DAY/$HOUR$MIN.conversion | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "syslog CONVERSION FAILED AGAIN - RESTART SYSLOG-NG"
    if ! netstat -apn | grep '0 0.0.0.0:5000' | grep LISTEN;
    then
    /sbin/service syslog-ng restart
  else
    pkill -9 syslog-ng && /sbin/service syslog-ng restart
    fi
  sleep 5
fi
else
echo "syslog CONVERSION is OK
---------------------------------------------------
---------------------------------------------------"
fi
