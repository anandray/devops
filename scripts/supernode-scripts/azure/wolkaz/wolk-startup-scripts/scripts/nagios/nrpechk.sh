#!/bin/bash

if ! ps aux | grep nrpe.cfg | grep -v grep | awk '{print$1,$13}' | egrep 'nagios|nrpe.cfg'> /dev/null; then
#echo nrpe is running
#else
echo nrpe is NOT running
/sbin/service nrpe restart
fi

if [ ! -f /etc/nagios/nrpe.cfg ]; then
yum -y install nagios-plugins nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm /home/anand/;
yum -y remove nrpe && sudo rpm -Uvh /home/anand/nrpe-2.15-7.el6.x86_64.rpm;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
chmod +x /usr/lib64/nagios/plugins/*
chkconfig nrpe on;
service nrpe restart;
chkconfig snmpd on;

else

echo "NRPE is installed"

if ! grep 'allowed_hosts=10.128.1.15,104.197.43.125,50.225.47.189' /etc/nagios/nrpe.cfg; then
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
/sbin/service nrpe restart;

if ! ls -l /usr/lib64/nagios/plugins/ | wc -l | grep 180 > /dev/null;
  then
  /usr/local/share/google/google-cloud-sdk/bin/gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
  chmod +x /usr/lib64/nagios/plugins/*
fi

else

echo "Correct allowed_hosts in /etc/nagios/nrpe.cfg"

sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/nrpechk.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/nrpechk.sh/g' /var/spool/cron/root
fi
fi
