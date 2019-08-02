#!bin/bash
gsutil cp gs://startup-scripts-colorfulnotion/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm /home/anand/;
yum -y remove nrpe && rpm -Uvh /home/anand/nrpe-2.15-7.el6.x86_64.rpm;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
service nrpe restart;
