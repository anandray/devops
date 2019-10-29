#!/bin/bash

TIMESTAMP=`date +%b_%d`
SevendaysAgo=$(date +'%b_%d' --date='7 days ago')

tables=`mysql --login-path=/usr/local/nagios -hdb03 wolk -e "show tables" | awk '{print$1}' | grep -v Tables_in_wolk`
echo "$tables" > /tmp/wolk_tables
mkdir /disk1/Backups/MySQL_Backups/db03/wolk/wolk_`date +%b_%d`
cat /tmp/wolk_tables | sort | awk '{print"mysqldump -v --opt --force --lock-all-tables --flush-logs --allow-keywords --hex-blob --triggers --routines --events -hdb03 -udb -p1wasb0rn2 wolk",$1,"> /disk1/Backups/MySQL_Backups/db03/wolk/wolk_`date +%b_%d`/wolk_"$1"_`date +%b_%d`.sql"}' > /tmp/wolk_db_backup.sh
sh /tmp/wolk_db_backup.sh &&

#echo "DB03 WOLK database backup completed on `date +'%b %d, %Y, %T'`" | mailx -r "CrossChannel Ad Ops <adops@crosschannel.com>" -s "DB03 WOLK database backup completed on `date +'%b %d, %Y, %T'`" -S smtp="smtp-relay.gmail.com:587" -S smtp-use-starttls -S smtp-auth=login -S smtp-auth-user="adops@crosschannel.com" -S smtp-auth-password='M0r3L0v3!' -S ssl-verify=ignore -S nss-config-dir="/etc/pki/nssdb/" systems@crosschannel.com &&

# copying backups to gs://db03_mysql_backups/db03/wolk/
gsutil -m cp -r /disk1/Backups/MySQL_Backups/db03/wolk/wolk_$TIMESTAMP gs://db03_mysql_backups/db03/wolk/ &&

echo -e "DB03 WOLK database backup completed on `date +'%b %d, %Y, %T'`" | mailx -A gmail -s "DB03 WOLK database backup completed on `date +'%b %d, %Y, %T'`" systems@wolk.com

rm -rfv /disk1/Backups/MySQL_Backups/db03/wolk/wolk_*
gsutil -m rm -r gs://db03_mysql_backups/db03/wolk/wolk_$SevendaysAgo
