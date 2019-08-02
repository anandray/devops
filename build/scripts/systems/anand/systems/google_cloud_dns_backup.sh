#!/bin/bash

TIMESTAMP=`date +%m%d%Y_%T`
DIR=`date +%m%d%Y`
ThirtyDaysAgo=`date +%m%d%Y --date='30 days ago'`
#echo $ThirtyDaysAgo

mkdir -p /root/dns_record_backup/$DIR;
/usr/local/share/google/google-cloud-sdk/bin/gcloud dns record-sets list --zone=mdotm-com > /root/dns_record_backup/$DIR/mdotm.com_backup_$TIMESTAMP;
/usr/local/share/google/google-cloud-sdk/bin/gcloud dns record-sets list --zone=mdotm-co > /root/dns_record_backup/$DIR/mdotm.co_backup_$TIMESTAMP;
/usr/local/share/google/google-cloud-sdk/bin/gcloud dns record-sets list --zone=crosschannel > /root/dns_record_backup/$DIR/crosschannel_backup_$TIMESTAMP;
rm -rf /root/dns_record_backup/$ThirtyDaysAgo;
#ls -lt /root/dns_record_backup/*_$ThirtyDaysAgo*
