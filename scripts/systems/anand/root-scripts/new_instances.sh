#!/bin/bash

if [ ! -d /tmp/new_instance ]; then
uptime > /tmp/new_instance;
cat /tmp/new_instance | mail -s"New instance created - $HOSTNAME - `date +%m%d%Y_%T`" engineering@mdotm.com;
 else
	echo "Not a NEW instance...";
	echo "commenting out new_instance.sh in crontab...";
	sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/new_instance.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/new_instance.sh/g' /var/spool/cron/root;
fi
