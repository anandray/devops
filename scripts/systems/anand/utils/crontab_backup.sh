#!/bin/bash

chattr -R -i /var/www/vhosts/mdotm.com/
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/www6001/cron; /usr/bin/rsync -avz www6001:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/www6001/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/admin/cron; /usr/bin/rsync -avz admin:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/admin/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/www6003/cron; /usr/bin/rsync -avz www6003:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/www6003/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/www6005/cron; /usr/bin/rsync -avz www6005:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/www6005/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/www6006/cron; /usr/bin/rsync -avz www6006:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/www6006/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/www6007/cron; /usr/bin/rsync -avz www6007:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/www6007/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/www6008/cron; /usr/bin/rsync -avz www6008:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/www6008/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/www6009/cron; /usr/bin/rsync -avz www6009:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/www6009/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/www6010/cron; /usr/bin/rsync -avz www6010:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/www6010/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/log00/cron; /usr/bin/rsync -avz log00:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/log00/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/log6/cron; /usr/bin/rsync -avz log6:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/log6/cron/ 
mkdir -p /var/www/vhosts/mdotm.com/cron/crontabs/log6b/cron; /usr/bin/rsync -avz log6b:/var/spool/cron/* /var/www/vhosts/mdotm.com/cron/crontabs/log6b/cron/ 

chattr -R -i /var/www/vhosts/mdotm.com
cd /var/www/vhosts/mdotm.com/cron/crontabs/
/usr/bin/git fetch upstream && /usr/bin/git merge upstream/master

chattr -R -i /var/www/vhosts/mdotm.com
chown -R mdotm /var/www/vhosts/mdotm.com/cron/
chmod -R 0755 /var/www/vhosts/mdotm.com/cron/

chattr -R -i /var/www/vhosts/mdotm.com
/usr/bin/git add /var/www/vhosts/mdotm.com/cron/crontabs/*
/usr/bin/git commit -m "adding crontabs backup `date +%m-%d-%Y_%T`"
/usr/bin/git push origin master

chattr -R -i /var/www/vhosts/mdotm.com
chown -R admin.engineering /var/www/vhosts/mdotm.com/cron/
chmod -R 0755 /var/www/vhosts/mdotm.com/cron/

#/usr/bin/git add `ls -lt | grep ^d | cut -d " " -f11` crontab_backup.sh 
#/usr/bin/git fetch upstream && /usr/bin/git merge upstream/master
#/usr/bin/git add /var/www/vhosts/mdotm.com/cron/crontabs/*
#/usr/bin/git commit -m "adding crontabs to /usr/bin/git"
#/usr/bin/git push origin master
