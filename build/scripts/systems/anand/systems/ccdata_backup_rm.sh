#!/bin/sh
TIMESTAMP=`date +'%b_%d'`
SevendaysAgo=$(date +'%b_%d' --date='7 days ago')
rm -rf /root/ccdata_backup/ccdata_backup_$SevendaysAgo.sql.gz;
