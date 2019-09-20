#!/bin/sh

#sqlite3
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/libsqlite3.la /usr/local/lib/
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/libsqlite3.so.0.8.6 /usr/local/lib/
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/libsqlite3.a /usr/local/lib/
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/sqlite3.conf /etc/ld.so.conf.d/
sudo /root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/sqlite3/sqlite3 /usr/local/bin/
sudo chmod +x /usr/local/bin/sqlite3
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0
sudo ldconfig

if [ -d /root/go/src/github.com/wolkdb/cloudstore ]; then
cd /root/go/src/github.com/wolkdb/cloudstore
git fetch
LOCAL=$(git rev-parse @{0})
REMOTE=$(git rev-parse @{u})
BASE=$(git merge-base @{0} @{u})

if [ $LOCAL = $REMOTE ]; then
 echo "Already up to date... removing cloudstore-git-update.sh from crontab"
 sed -i '/cloudstore-git-update/d' /var/spool/cron/root
else
 echo "Updating cloudstore repository"
 git fetch origin
 git merge origin/master
 /sbin/service wolk restart
 echo "done.. removing cloudstore-git-update.sh from crontab"
 sed -i '/cloudstore-git-update/d' /var/spool/cron/root
fi
fi

/root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore-bashrc /root/.bashrc
/root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore-bashrc_aliases /root/.bashrc_aliases
/root/.local/bin/aws s3 -m cp s3://wolk-startup-scripts/scripts/nagios/plugins/check_wolk_healthcheck.sh /usr/lib64/nagios/plugins/
/root/.local/bin/aws s3 -m cp s3://wolk-startup-scripts/scripts/nagios/plugins/check_wolk_healthcheck_ssl.sh /usr/lib64/nagios/plugins/
/root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/cloudstore-sudoers-aws /etc/sudoers
/root/.local/bin/aws s3 cp s3://wolk-startup-scripts/scripts/plasma/cloudstore-crontab-aws /var/spool/cron/root
chmod +x /usr/lib64/nagios/plugins/check_wolk_healthcheck*
chmod +x /usr/local/bin/sqlite3
/root/.local/bin/aws s3 cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/nrpe.cfg
service nrpe restart

if ! ps aux | grep nrpe | grep -v grep; then
/usr/sbin/nrpe -c /etc/nagios/nrpe.cfg -d
fi

#Adding environment variables to /root/.bashrc
if ! sudo grep GOPATH /root/.bashrc; then
sudo su - << EOF
echo '
export PATH="$PATH:/usr/local/go/bin"
export GOPATH=/root/go
export GOROOT=/usr/local/go' >> /root/.bashrc
EOF
fi

if [ -d /usr/local/wolk ]; then
        /sbin/service wolk stop;
        /usr/bin/pkill -9 wolk;
        rm -rf /usr/local/wolk/*;
        /sbin/service wolk restart;
fi

exec -l $SHELL
source ~/.bashrc