#!/bin/bash
YEAR=`date +%Y`
MONTH=`date +%m`
DAY=`date +%d`
HOUR=`date +%H`
MIN=`date +%M`

# LOG = /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.log

for i in {1..10};
do stat /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.log | grep Modify | grep "`date +'%Y-%m-%d %T'`";
sleep .1
done

for i in {1..10};
# TRACK = /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.track
do stat /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.track | grep Modify | grep "`date +'%Y-%m-%d %T'`";
sleep .1
done

for i in {1..10};
# CONVERSION = /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.conversion
do stat /disk1/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.conversion | grep Modify | grep "`date +'%Y-%m-%d %T'`";
sleep .1
done
