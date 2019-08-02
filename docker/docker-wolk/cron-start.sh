#!/bin/bash
  
/usr/sbin/cron -n
crontab -e << EOF
:wq
EOF
