#!/bin/bash
/bin/sed -i 's/\*\/30 \* \* \* \* sh \/root\/scripts\/ntpdate.sh/\#\*\/30 \* \* \* \* sh \/root\/scripts\/ntpdate.sh/g' /var/spool/cron/root
if ! grep ntpdate /var/spool/cron/root | wc -l | grep 3; then
echo "
23,47 * * * * sh /root/scripts/ntpdate1.sh > /dev/null 2>&1
27,53 * * * * sh /root/scripts/ntpdate2.sh > /dev/null 2>&1" >> /var/spool/cron/root
fi
