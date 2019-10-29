#!/bin/bash
if ! php -v > /dev/null;
  then
if ! ps aux | grep startup | grep -v grep > /dev/null;
  then
  echo "php not installed and startup script is not running"
  sh /root/scripts/startup*failsafe*.sh;
  rsync -avz www6002:/usr/lib64/nagios/plugins/ /usr/lib64/nagios/plugins/;
  cd /var/www/vhosts/crosschannel.com/;
  git config core.filemode false;
fi
  else
  echo "php installed on `hostname`"
fi
