#!/bin/sh

# ssh keys
sudo mkdir -p /root/.ssh
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/ssh-keys/* /root/.ssh/
sudo chmod 0400 /root/.ssh/id_rsa*

# copying cloudstore-git-update.sh
sudo gsutil cp gs://startup_scripts_us/scripts/cloudstore/cloudstore-git-update.sh /root/scripts/
sudo chmod +x /root/scripts/cloudstore-git-update.sh

# copying syslog-ng.conf
sudo cp -rf /etc/syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf-bak
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/syslog-ng/syslog-ng.conf /etc/syslog-ng/
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/syslog-ng/dmesg.conf /etc/syslog-ng/conf.d/
sudo service syslog-ng restart

#sqlite3
if ! sqlite3 --version | grep 3.22; then
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/libsqlite3.la /usr/local/lib/
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/libsqlite3.so.0.8.6 /usr/local/lib/
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/libsqlite3.a /usr/local/lib/
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/sqlite3.conf /etc/ld.so.conf.d/
sudo gsutil cp gs://wolk-scripts/scripts/sqlite3/sqlite3 /usr/local/bin/
sudo chmod +x /usr/local/bin/sqlite3
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so
sudo ln -s /usr/local/lib/libsqlite3.so.0.8.6 /usr/local/lib/libsqlite3.so.0
sudo ldconfig
fi

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
 echo "done.. removing cloudstore-git-update.sh from crontab"
 sed -i '/cloudstore-git-update/d' /var/spool/cron/root
fi
fi

sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-bashrc /root/.bashrc
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-bashrc_aliases /root/.bashrc_aliases
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-sudoers /etc/sudoers
if [ -d /usr/lib64/nagios/plugins/plugins ]; then
sudo rm -rf /usr/lib64/nagios/plugins/plugins /usr/lib64/nagios/plugins
fi
sudo gsutil -m cp -r gs://wolk-scripts/scripts/nagios/plugins /usr/lib64/nagios/
sudo gsutil cp gs://wolk-scripts/scripts/nagios/plugins/check_wolk_healthcheck.sh /usr/lib64/nagios/plugins/
sudo gsutil cp gs://wolk-scripts/scripts/nagios/plugins/check_wolk_healthcheck_ssl.sh /usr/lib64/nagios/plugins/
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-crontab-gcp /var/spool/cron/root
sudo gsutil -m cp gs://wolk-scripts/scripts/mysql/* /usr/bin/
sudo chmod +x /usr/bin/mysql* /usr/bin/db03.sh
chmod -R +x /usr/lib64/nagios/plugins
chmod +x /usr/local/bin/sqlite3
sudo gsutil cp gs://wolk-scripts/scripts/nagios/nrpe.cfg /etc/nagios/nrpe.cfg
sudo service nrpe restart

# hosts file
sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/cloudstore-hosts /etc/hosts

if ! ps aux | grep nrpe | grep -v grep; then
sudo /usr/sbin/nrpe -c /etc/nagios/nrpe.cfg -d
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
        sudo /sbin/service wolk stop;
        sudo /usr/bin/pkill -9 wolk;
        sudo rm -rf /usr/local/wolk/*;
fi

# copy ssl certs
sudo gsutil cp gs://wolk-scripts/scripts/certificate/www.wolk.com.crt /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt
sudo gsutil cp gs://wolk-scripts/scripts/certificate/www.wolk.com.key /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key

# create google credential
gcloud services enable compute.googleapis.com
#serviceAccount=`gcloud projects describe $project | grep projectNumber | cut -d "'" -f2 | awk '{print$1"-compute@developer.gserviceaccount.com"}'`
serviceAccount=`gcloud auth list 2> /dev/null | grep developer.gserviceaccount.com | awk '{print$2}'`
echo $serviceAccount

# copy credentials from google storage
hostname=`hostname | cut -d "-" -f4,5`
credentials_name="*$hostname*.json"
echo $credentials_name
sudo gsutil cp  gs://wolk-scripts/scripts/cloudstore/google-credentials/$credentials_name /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json

if [ ! -f /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json ]; then
credentials_file=`gsutil ls gs://wolk-scripts/scripts/cloudstore/google-credentials/$credentials_name | head -n1`
sudo gsutil cp $credentials_file /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json
fi

sudo gsutil cp gs://wolk-scripts/scripts/cloudstore/rc.local /etc/rc.d/rc.local
sudo chmod +x /etc/rc.d/rc.local
sudo chmod +x /etc/rc.local

# populate wolk.toml
region=`hostname | cut -d "-" -f4,5`
project=`grep project_id /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json | awk '{print$2}' | cut -d "\"" -f2`
node=`hostname | cut -d "-" -f2`
gcloud config set project $project
gcloud config set compute/region $region

'yes' | cp -rf /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/wolk.toml-gc-template /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i "s/_ConsensusIdx/$node/g" /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i "s/_NodeType/storage/g" /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i "s/_Region/$region/g" /root/go/src/github.com/wolkdb/cloudstore/wolk.toml
sed -i "s/_GoogleDatastoreProject/$project/g" /root/go/src/github.com/wolkdb/cloudstore/wolk.toml

# git update | make wolk | start wolk | creating wolk1-wolk5
#git --git-dir=/root/go/src/github.com/wolkdb/cloudstore/.git pull 2> /dev/null
sudo su - << EOF
scp c0.wolk.com:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json
gsutil cp gs://wolk-scripts/scripts/cloudstore/make-wolk.sh /root/scripts/
chmod +x /root/scripts/make-wolk.sh
/root/scripts/make-wolk.sh
EOF

if [ -d /usr/local/wolk ]; then
        sudo /sbin/service wolk stop;
        sudo /usr/bin/pkill -9 wolk;
        sudo rm -rf /usr/local/wolk/*;
fi

sh /etc/rc.local
