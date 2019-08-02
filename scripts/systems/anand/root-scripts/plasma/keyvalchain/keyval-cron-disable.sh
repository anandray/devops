#!/bin/bash

if grep keyval-restart-cron.sh /var/spool/cron/root | grep ^\* &> /dev/null; then
sed -i 's/\* \* \* \* \* \/root\/scripts\/keyval-restart-cron.sh/\#\* \* \* \* \* \/root\/scripts\/keyval-restart-cron.sh/g' /var/spool/cron/root
fi
