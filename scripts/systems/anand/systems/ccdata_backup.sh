#!/bin/sh
TIMESTAMP=`date +'%b_%d'`
SevendaysAgo=$(date +'%b_%d' --date='7 days ago')
/usr/bin/mysqldump -v -h db03 -udb  -p1wasb0rn2 ccdata > /root/ccdata_backup/ccdata_backup_$TIMESTAMP.sql;
/bin/gzip /root/ccdata_backup/ccdata_backup_$TIMESTAMP.sql;
/usr/bin/rsync -avz /root/ccdata_backup/ www6005:/root/ccdata_backup/;
rm -rf /root/ccdata_backup/ccdata_backup_$SevendaysAgo.sql.gz;
