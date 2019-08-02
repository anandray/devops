#!/bin/bash
TIMESTAMP=`date +%Y.%m.%d`
SevenDaysAgo=$(date +%Y.%m.%d  --date='7 days ago')
#/root/.config/gcloud/logs/`date +%Y.%m.%d`
rm -rfv /root/.config/gcloud/logs/$SevenDaysAgo
