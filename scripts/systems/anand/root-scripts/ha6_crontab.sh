echo "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
SHELL=/bin/bash
" > /var/spool/cron/crontabs/root;
echo "* * * * * /bin/sh /root/scripts/hbasechk-ha6-go.sh >> /var/log/hbasechk-ha6-go.log 2>&1" >> /var/spool/cron/crontabs/root;
