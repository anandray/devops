#!/bin/bash
echo "
MAILTO=''
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/usr/lib64/google-cloud-sdk/bin:/root/bin:/root/sql/bin

*/1 * * * * /sql-start-cron.sh &>> /var/log/sql-start-cron.log" >> /var/spool/cron/root
