#!/bin/bash
yum -y remove nagios-plugins nagios-plugins-nrpe nagios-common nrpe
rm -rf /usr/lib64/nagios
yum -y install nagios-plugins nagios-plugins-nrpe nagios-common nrpe
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/
chmod +x /usr/lib64/nagios/plugins/*
/sbin/service nrpe restart
