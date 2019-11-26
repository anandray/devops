#!/bin/bash
if ! php -v > /dev/null;
  then
if ! ps aux | grep startup | grep -v grep > /dev/null;
  then
  echo "php not installed and startup script is not running"
  sh /root/scripts/startup*failsafe*.sh;
  gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
  chmod +x /usr/lib64/nagios/plugins/*;
  cd /var/www/vhosts/crosschannel.com/;
  git config core.filemode false;
fi
  else
  echo "php installed on `hostname`"
fi
