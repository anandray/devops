#!/bin/bash
yum -y install nagios-plugins nagios-plugins-nrpe nagios-common nrpe nagios-nrpe gd-devel net-snmp &&
/usr/local/share/google/google-cloud-sdk/bin/gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm . &&
yum -y remove nrpe && rpm -Uvh nrpe-2.15-7.el6.x86_64.rpm &&
/usr/local/share/google/google-cloud-sdk/bin/gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/ &&
/usr/local/share/google/google-cloud-sdk/bin/gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/ &&
chmod +x /usr/lib64/nagios/plugins/* &&
chkconfig nrpe on &&
/sbin/service nrpe restart
