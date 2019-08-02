#!/bin/bash
YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`

# DEBUG = /disk1/log/debug/$YEAR/$MONTH/$DAY/$HOUR$MIN.debug

for i in {1..10};
do stat /disk1/log/debug/$YEAR/$MONTH/$DAY/$HOUR$MIN.debug | grep Modify | grep "`date +'%Y-%m-%d %T'`";
sleep .1
done

#   BID = /disk2/log/bid/$YEAR/$MONTH/$DAY/$HOUR$MIN.bid
for i in {1..10};
do stat /disk2/log/bid/$YEAR/$MONTH/$DAY/$HOUR$MIN.bid | grep Modify | grep "`date +'%Y-%m-%d %T'`";
sleep .1
done
