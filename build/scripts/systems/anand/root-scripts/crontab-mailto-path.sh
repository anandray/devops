#!/bin/bash
sed -i "1 i\MAILTO=''" /var/spool/cron/root
sed -i "2 i\PATH=\"/root/sbin:/root/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/go/bin:/root/go/src/github.com/wolkdb/plasma/build/bin:/root/bin\"" /var/spool/cron/root
