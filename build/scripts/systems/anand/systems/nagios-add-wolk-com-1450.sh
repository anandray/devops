#!/bin/bash

for i in {1..4};
do
# remove previous server.cfg first
echo "`date +%m-%d-%Y\|%T` - Removing old servers..."
rm -rfv /usr/local/nagios/etc/objects/servers/www-wolk-com-1450*.cfg &> /var/log/nagios-add-wolk-com-1450.log

echo "`date +%m-%d-%Y\|%T` - Adding new servers..."

/usr/local/share/google/google-cloud-sdk/bin/gcloud compute instances list --project wolk-1307 | grep www-wolk-com-1450 | awk '{print"echo \"\n\#\""$1"\" host definition\n","define host\{\n","\t","host_name","\t","\t","\t",$1,"\n","\t","alias","\t","\t","\t","\t","WWW-WOLK-COM\n","\t","address","\t","\t","\t",$4,"\n","\t","contact_groups","\t","\t","oncall-admins,oncall-admins2\n","\t","check_command","\t","\t","\t","check-host-alive\n","\t","max_check_attempts","\t","\t","10\n","\t","notification_interval","\t","\t","120\n","\t","notification_period","\t","\t","24x7\n","\t","notification_options","\t","\t","d,u,r\n","\}\"","> /usr/local/nagios/etc/objects/servers/"$1".cfg"}' > /tmp/www-wolk-com-1450.sh && sh /tmp/www-wolk-com-1450.sh &>> /var/log/nagios-add-wolk-com-1450.log
sleep 15;
done
