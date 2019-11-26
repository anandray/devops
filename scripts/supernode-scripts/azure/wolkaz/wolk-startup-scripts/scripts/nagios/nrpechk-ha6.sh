#!/bin/bash
if ps aux | grep nrpe.cfg | grep -v grep | awk '{print$1,$13}' | egrep 'nagios|nrpe.cfg'> /dev/null; then
echo nrpe is running
else
echo nrpe is NOT running
/etc/init.d/nagios-nrpe-server restart
fi

if [ ! -f /etc/nagios/nrpe.cfg ]; then
apt-get -y install nagios-nrpe-server nagios-nrpe-plugin;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe-ha6.cfg /etc/nagios/nrpe.cfg;
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/check_*.sh /usr/lib/nagios/plugins/;
chmod -R +x /usr/lib/nagios/plugins/*;
/usr/bin/pkill -9 nrpe;
/etc/init.d/nagios-nrpe-server restart;

else

echo "NRPE is installed"
sed -i 's/\* \* \* \* \* \/bin\/sh \/root\/scripts\/nrpechk-ha6.sh/\#\* \* \* \* \* \/bin\/sh \/root\/scripts\/nrpechk-ha6.sh/g' /var/spool/cron/crontabs/root
fi
