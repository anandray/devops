#!/bin/bash
logs=$(/root/.config/gcloud/logs/`date +%Y.%m.%d --date='1 days ago'`)
if ls -ld $logs > /dev/null;
  then
  rm -rfv $logs
fi
