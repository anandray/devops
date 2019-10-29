#!/bin/bash
if lsattr /var/www/vhosts/mdotm.com | grep "\-\-\-\-i\-\-\-\-\-\-\-\-e\-" &> /dev/null;
  then
  chattr -R -i /var/www/vhosts/mdotm.com
fi

if ps aux | grep '/var/www/vhosts/mdotm.com/cron/db/adxrtb/report.php' | grep -v grep &> /dev/null;
  then
  kill -9 $(ps aux | grep '/var/www/vhosts/mdotm.com/cron/db/adxrtb/report.php' | grep -v grep | awk '{print$2}')
  echo "DONE.. NOW WAIT FOR THE CRONJOB TO RUN THE PROCESS.."
fi
