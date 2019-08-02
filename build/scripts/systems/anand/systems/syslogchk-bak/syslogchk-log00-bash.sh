#!/bin/bash
YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`

# DEBUG = /disk1/log/debug/$YEAR/$MONTH/$DAY/$HOUR$MIN.debug

if ! stat /disk1/log/debug/$YEAR/$MONTH/$DAY/$HOUR$MIN.debug | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "
  syslog DEBUG FAILED - TRY AGAIN"
if ! stat /disk1/log/debug/$YEAR/$MONTH/$DAY/$HOUR$MIN.debug | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "syslog DEBUG FAILED AGAIN - RESTART SYSLOG-NG"
    if ! netstat -apn | grep '0 0.0.0.0:5000' | grep LISTEN;
    then
    /sbin/service syslog-ng restart
  else
    pkill -9 syslog-ng && /sbin/service syslog-ng restart
    fi
  sleep 5
fi
else
echo "syslog DEBUG is OK
     "
fi

YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`

#   BID = /disk2/log/bid/$YEAR/$MONTH/$DAY/$HOUR$MIN.bid
if ! stat /disk2/log/bid/$YEAR/$MONTH/$DAY/$HOUR$MIN.bid | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "syslog BID FAILED - TRY AGAIN"
if ! stat /disk2/log/bid/$YEAR/$MONTH/$DAY/$HOUR$MIN.bid | grep Modify | grep "`date +'%Y-%m-%d %T'`";
  then
  echo "syslog BID FAILED AGAIN - RESTART SYSLOG-NG"
    if ! netstat -apn | grep '0 0.0.0.0:5000' | grep LISTEN;
    then
    /sbin/service syslog-ng restart
  else
    pkill -9 syslog-ng && /sbin/service syslog-ng restart
    fi
  sleep 5
fi
else
echo "syslog BID is OK
     "
fi
