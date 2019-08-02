#!/bin/bash
echo "
MAILTO=''
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/usr/lib64/google-cloud-sdk/bin:/root/bin

*/1 * * * * /wolk/scripts/container-start.sh &>> /var/log/container-start.log" >> /var/spool/cron/root
