#!/bin/bash
sudo su - << EOF
apt-get -y install nagios-nrpe-server nagios-nrpe-plugin;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe-ha6.cfg /etc/nagios/nrpe.cfg;
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/check_*.sh /usr/lib/nagios/plugins/;

# copy check_syslog-ng-ha6.sh as check_syslog-ng.sh
gsutil cp gs://startup_scripts_us/scripts/nagios/plugins/check_syslog-ng-ha6.sh /usr/lib/nagios/plugins/check_syslog-ng.sh

chmod -R +x /usr/lib/nagios/plugins/*
/etc/init.d/nagios-nrpe-server restart;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpechk-ha6.sh /root/scripts/;
sed -i '/nrpechk/d' /var/spool/cron/crontabs/root;
touch /var/log/nrpechk.log;
echo '* * * * * /bin/sh /root/scripts/nrpechk-ha6.sh > /var/log/nrpechk.log 2>&1' >> /var/spool/cron/crontabs/root;
EOF
