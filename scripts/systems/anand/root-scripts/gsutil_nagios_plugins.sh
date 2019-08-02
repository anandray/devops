#!/bin/bash
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/check_hbase.sh /usr/lib64/nagios/plugins/
chmod +x /usr/lib64/nagios/plugins/*
service nrpe restart
