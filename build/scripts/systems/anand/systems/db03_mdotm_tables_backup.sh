#!/bin/bash

TIMESTAMP=`date +%b_%d`
SevendaysAgo=$(date +'%b_%d' --date='7 days ago')

tables=`mysql --login-path=/usr/local/nagios -hdb03 mdotm -e "show tables" | awk '{print$1}' | grep -v Tables_in_mdotm`
echo "$tables" > /tmp/mdotm_tables
mkdir /disk1/Backups/MySQL_Backups/db03/mdotm/mdotm_`date +%b_%d`
cat /tmp/mdotm_tables | sort | awk '{print"mysqldump -v --opt --force --lock-all-tables --flush-logs --allow-keywords --hex-blob --triggers --routines --events -hdb03 -udb -p1wasb0rn2 mdotm",$1,"> /disk1/Backups/MySQL_Backups/db03/mdotm/mdotm_`date +%b_%d`/mdotm_"$1"_`date +%b_%d`.sql"}' > /tmp/mdotm_db_backup.sh
sh /tmp/mdotm_db_backup.sh &&

#echo "DB03 MDOTM database backup completed on `date +'%b %d, %Y, %T'`" | mailx -r "CrossChannel Ad Ops <adops@crosschannel.com>" -s "DB03 MDOTM database backup completed on `date +'%b %d, %Y, %T'`" -S smtp="smtp-relay.gmail.com:587" -S smtp-use-starttls -S smtp-auth=login -S smtp-auth-user="adops@crosschannel.com" -S smtp-auth-password='M0r3L0v3!' -S ssl-verify=ignore -S nss-config-dir="/etc/pki/nssdb/" systems@crosschannel.com &&

# copying backups to gs://db03_mysql_backups/db03/mdotm/
gsutil -m cp -r /disk1/Backups/MySQL_Backups/db03/mdotm/mdotm_$TIMESTAMP gs://db03_mysql_backups/db03/mdotm/ &&

echo -e "DB03 MDOTM database backup completed on `date +'%b %d, %Y, %T'`" | mailx -A gmail -s "DB03 MDOTM database backup completed on `date +'%b %d, %Y, %T'`" systems@crosschannel.com

rm -rfv /disk1/Backups/MySQL_Backups/db03/mdotm/mdotm_*
gsutil -m rm -r gs://db03_mysql_backups/db03/mdotm/mdotm_$SevendaysAgo
