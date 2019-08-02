#!/bin/bash

if grep plasma-restart-cron.sh /var/spool/cron/root | grep ^\* &> /dev/null; then
sed -i 's/\* \* \* \* \* \/root\/scripts\/plasma-restart-cron.sh/\#\* \* \* \* \* \/root\/scripts\/plasma-restart-cron.sh/g' /var/spool/cron/root
fi
