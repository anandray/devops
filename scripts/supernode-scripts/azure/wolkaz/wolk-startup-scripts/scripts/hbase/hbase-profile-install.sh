#!/bin/bash
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-site-us-profile-all.xml /usr/local/hbase-profile/conf/hbase-site.xml;
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-profile-chk.sh /root/scripts/;
sh /root/scripts/hbase-profile-chk.sh;
echo '*/1 * * * * ssh `hostname` /bin/sh /root/scripts/hbase-profile-chk.sh > /var/log/hbase.log 2>&1' >> /var/spool/cron/crontabs/root;
EOF
