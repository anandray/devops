#!/bin/bash
disk1a=$(du -sh /disk1/log/debug/{`date +%Y`,`date +%Y --date='1 year ago'`}/{`date +%m`,`date +%m --date='1 month ago'`}/`date +%d --date='7 days ago'` 2> /dev/null)
disk1b=$(du -sh /disk1/log/debug/{`date +%Y`,`date +%Y --date='1 year ago'`}/{`date +%m`,`date +%m --date='1 month ago'`}/`date +%d --date='6 days ago'` 2> /dev/null)
disk2a=$(du -sh /disk2/log/bid/{`date +%Y`,`date +%Y --date='1 year ago'`}/{`date +%m`,`date +%m --date='1 month ago'`}/`date +%d --date='7 days ago'` 2> /dev/null)
disk2b=$(du -sh /disk2/log/bid/{`date +%Y`,`date +%Y --date='1 year ago'`}/{`date +%m`,`date +%m --date='1 month ago'`}/`date +%d --date='6 days ago'` 2> /dev/null)


# DEBUG log - /disk1/log/debug/$YEAR/$MONTH/$DAY/$HHMM.debug

# 7 days old debug log
if echo $disk1a | grep G &> /dev//null;
  then
  echo -ne "Deleting" `echo $disk1a | awk '{print$2}'`"...\n"
  rm -rfv `echo $disk1a | awk '{print$2}'` &> /var/log/disk_clean_debug.log&
  sleep 2
  echo -ne '######                                  (10%)\r'
  sleep 2
  echo -ne '#########                               (20%)\r'
  sleep 2
  echo -ne '############                            (30%)\r'
  sleep 2
  echo -ne '###############                         (40%)\r'
  sleep 2
  echo -ne '##################                      (50%)\r'
  sleep 2
  echo -ne '#####################                   (60%)\r'
  sleep 2
  echo -ne '#########################               (70%)\r'
  sleep 2
  echo -ne '###############################         (80%)\r'
  sleep 2
  echo -ne '##################################      (90%)\r'
  sleep 2
  echo -ne '####################################### (100%)\r'
  echo -ne '\n'
else
echo -ne "\n7 days old DEBUG log - "
$(/disk1/log/debug/{`date +%Y`,`date +%Y --date='1 year ago'`}/{`date +%m`,`date +%m --date='1 month ago'`}/`date +%d --date='7 days ago'`)
fi

# 6 days old debug log
if echo $disk1b | grep G &> /dev//null;
  then
  echo -ne "Deleting" `echo $disk1b | awk '{print$2}'`"...\n"
  rm -rfv `echo $disk1b | awk '{print$2}'` &> /var/log/disk_clean_debug.log&
  sleep 2
  echo -ne '######                                  (10%)\r'
  sleep 2
  echo -ne '#########                               (20%)\r'
  sleep 2
  echo -ne '############                            (30%)\r'
  sleep 2
  echo -ne '###############                         (40%)\r'
  sleep 2
  echo -ne '##################                      (50%)\r'
  sleep 2
  echo -ne '#####################                   (60%)\r'
  sleep 2
  echo -ne '#########################               (70%)\r'
  sleep 2
  echo -ne '###############################         (80%)\r'
  sleep 2
  echo -ne '##################################      (90%)\r'
  sleep 2
  echo -ne '####################################### (100%)\r'
  echo -ne '\n'
else
echo -ne "\n6 days old DEBUG log - "
$(/disk1/log/debug/{`date +%Y`,`date +%Y --date='1 year ago'`}/{`date +%m`,`date +%m --date='1 month ago'`}/`date +%d --date='6 days ago'`)
fi

# BID log - /disk2/log/bid/$YEAR/$MONTH/$DAY/$HHMM.bid

# 7 days old bid log
if echo $disk2a | grep G &> /dev//null;
  then
  echo -ne "Deleting" `echo $disk2a | awk '{print$2}'`"...\n"
  rm -rfv `echo $disk2a | awk '{print$2}'` &> /var/log/disk_clean_bid.log&
  sleep 2
  echo -ne '######                                  (10%)\r'
  sleep 2
  echo -ne '#########                               (20%)\r'
  sleep 2
  echo -ne '############                            (30%)\r'
  sleep 2 
  echo -ne '###############                         (40%)\r'
  sleep 2
  echo -ne '##################                      (50%)\r'
  sleep 2
  echo -ne '#####################                   (60%)\r'
  sleep 2
  echo -ne '#########################               (70%)\r'
  sleep 2
  echo -ne '###############################         (80%)\r'
  sleep 2
  echo -ne '##################################      (90%)\r'
  sleep 2
  echo -ne '####################################### (100%)\r'
  echo -ne '\n'
else
echo -ne "\n7 days old BID log - "
$(/disk2/log/bid/{`date +%Y`,`date +%Y --date='1 year ago'`}/{`date +%m`,`date +%m --date='1 month ago'`}/`date +%d --date='7 days ago'`)
fi

# 6 days old bid log
if echo $disk2b | grep G &> /dev//null;
  then
  echo -ne "Deleting" `echo $disk2b | awk '{print$2}'`"...\n"
  rm -rfv `echo $disk2b | awk '{print$2}'` &> /var/log/disk_clean_bid.log&
  sleep 2
  echo -ne '######                                  (10%)\r'
  sleep 2
  echo -ne '#########                               (20%)\r'
  sleep 2
  echo -ne '############                            (30%)\r'
  sleep 2
  echo -ne '###############                         (40%)\r'
  sleep 2
  echo -ne '##################                      (50%)\r'
  sleep 2
  echo -ne '#####################                   (60%)\r'
  sleep 2
  echo -ne '#########################               (70%)\r'
  sleep 2
  echo -ne '###############################         (80%)\r'
  sleep 2
  echo -ne '##################################      (90%)\r'
  sleep 2
  echo -ne '####################################### (100%)\r'
  echo -ne '\n'
else
echo -ne "\n6 days old BID log - "
$(/disk2/log/bid/{`date +%Y`,`date +%Y --date='1 year ago'`}/{`date +%m`,`date +%m --date='1 month ago'`}/`date +%d --date='6 days ago'`)
fi
