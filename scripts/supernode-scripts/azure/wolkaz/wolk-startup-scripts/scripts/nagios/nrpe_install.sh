#!/bin/bash
sudo su - << EOF
yum -y install nagios-plugins nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
chmod +x /usr/lib64/nagios/plugins/*
chkconfig nrpe on;
service nrpe restart;
chkconfig snmpd on;
EOF
