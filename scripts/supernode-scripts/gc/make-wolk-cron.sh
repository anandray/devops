#!/bin/bash

if [ ! -f /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk ]; then
gsutil cp gs://wolk-scripts/scripts/cloudstore/make-wolk-cron.sh /root/scripts/;
chmod +x /root/scripts/make-wolk-cron.sh;
/root/scripts/make-wolk-cron.sh
service wolk restart;
fi

if [ -f /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk ]; then
sed -i '/make-wolk-cron.sh/d' /var/spool/cron/root;
else
/root/scripts/make-wolk-cron.sh
fi
