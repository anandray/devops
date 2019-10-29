#!/bin/bash

#PST date time

if ! rpm -qa | grep ntpdate > /dev/null; then
sudo yum -y install ntpdate;
fi

if ! rpm -qa | grep rdate > /dev/null; then
sudo yum -y install rdate;
fi

if ! date | grep -E 'PST|PDT';
then
sudo ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
sudo service ntpd stop;
sudo ntpdate -u -b pool.ntp.org;
sudo service ntpd start;

else
sudo service ntpd stop;
sudo ntpdate -u -b pool.ntp.org;
sudo service ntpd start;
fi

## Start /etc/sudoers modifications ##
# THE FOLLOWING /etc/sudoers MODIFICATIONS ARE REQUIRED TO ALLOW 'gcloud' and 'gsutil' to run using sudo, WHICH IS OTHERWISE NOT ALLOWED
# Replace /etc/sudoers default 'Defaults secure_path' with following:
#Defaults secure_path = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/var/www/vhosts/mdotm.com/scripts/utils:/usr/local/share/google/google-cloud-sdk:/usr/local/share/google/google-cloud-sdk/bin

# Comment out this line in /etc/sudoers:
#Defaults    requiretty
if ! grep '#Defaults    requiretty' /etc/sudoers | grep '\#' > /dev/null; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/sudoers /etc/

fi
## /End /etc/sudoers modifications ##

# Copy /etc/hosts from gs:
if ! grep -q db /etc/hosts; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/hosts /etc/;
fi

#SSH Keys:
if [ ! -f /root/.ssh/authorized_keys ]; then
        sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
        sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
        sudo rm -rf /root/.ssh/ssh_keys.tgz
fi

# Allow SSH-ing to any instance/server
if ! grep 'StrictHostKeyChecking no' /etc/ssh/ssh_config; then

        sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_config /etc/ssh/;
        sudo service sshd restart;
fi

if ! grep 'PermitRootLogin yes' /etc/ssh/sshd_config; then

        sudo gsutil -m cp gs://startup_scripts_us/scripts/sshd_config /etc/ssh/
        sudo service sshd restart;
fi
#sed -i "49 i\StrictHostKeyChecking no" /etc/ssh/ssh_config
#sed -i "50 i\UserKnownHostsFile /dev/null" /etc/ssh/ssh_config

# Enable histtimeformat
if ! ls /etc/profile.d/histtimeformat.sh; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/histtimeformat.sh /etc/profile.d/
fi

## DISABLE FSCK
if ! tune2fs -l /dev/sda1 | grep -i 'Maximum mount count:      -1'; then
sudo tune2fs -c 0 -i 0 /dev/sda1;
fi

if ! tune2fs -l /dev/sda1 | grep -i 'Check interval:           0'; then
sudo tune2fs -c 0 -i 0 /dev/sda1;
fi

#DISABLE SELINUX:
if ! grep 0 /selinux/enforce; then
sudo setenforce 0 && setsebool -P httpd_can_network_connect=1;
sudo cp /etc/selinux/config /etc/selinux/config_ORIG;
sudo gsutil -m cp gs://startup_scripts_us/scripts/selinux_config /etc/selinux/config
fi

# limits.conf --> ulimit -a
if ! grep 'root      hard    nofile      500000' /etc/security/limits.conf; then
sudo cp /etc/security/limits.conf /etc/security/limits.conf_ORIG
sudo gsutil -m cp gs://startup_scripts_us/scripts/security_limits.conf /etc/security/limits.conf
fi

if ! grep 'root       soft    nproc     unlimited' /etc/security/limits.d/90-nproc.conf; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/90-nproc.conf /etc/security/limits.d/
fi

#PHP INSTALL:

#### Redhat 6.x ####
if ! rpm -qa | grep epel-release-6-8.noarch; then
sudo rpm -Uvh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
fi

if ! rpm -qa | grep ius-release-1.0-15.ius.el6.noarch; then
sudo rpm -Uvh http://dl.iuscommunity.org/pub/ius/stable/Redhat/6/x86_64/ius-release-1.0-15.ius.el6.noarch.rpm
fi

if grep '#baseurl=http' /etc/yum.repos.d/epel.repo; then
sudo su - << EOF
/bin/sed -i 's/\#baseurl=http/baseurl=http/g' /etc/yum.repos.d/epel.repo;
EOF
fi

if ! grep '#mirrorlist=http' /etc/yum.repos.d/epel.repo; then
sudo su - << EOF
/bin/sed -i 's/mirrorlist=http/\#mirrorlist=http/g' /etc/yum.repos.d/epel.repo;
EOF
fi

########

#Install php54-5.4.45

#if ! rpm -qa | grep php54-5.4.45; then

#sudo yum remove -y php*;
#sudo yum -y --enablerepo=ius-archive install php54 php54-soap php54-gd php54-ioncube-loader php54-pecl-memcache php54-mcrypt php54-imap php54-devel php54-cli php54-mysql php54-mbstring php54-xml libxml2 libxml2-devel GeoIP geoip-devel gcc make mysql php54-pecl-memcached libmemcached10-devel emacs ntpdate rdate syslog-ng syslog-ng-libdbi libdbi-devel telnet screen git sendmail sendmail-cf denyhosts procmail python-argparse *whois openssl openssl-devel php-pear-Net-SMTP php54-pear php-pear-Net-Socket php-pear-Auth-SASL libssh2 libssh2-devel java-1.8.0-openjdk-devel python27-pip;

# python2.6 to python2.7
if ! pip2.7 list --format=legacy | grep google-compute-engine; then
sudo su - << EOF
/usr/bin/pip2.7 install google google_compute_engine
unlink /usr/bin/python2
ln -s /usr/bin/python2.7 /usr/bin/python2
cp -rfv /root/scripts/python_version_change.sh /etc/cron.hourly/
cp -rfv /root/scripts/python_version_change_cron.sh /etc/cron.d/
chmod 0755 /etc/cron.hourly/python_version_change.sh
chmod 0755 /etc/cron.d/python_version_change_cron.sh
/usr/bin/pip2.7 install --upgrade pip
EOF

else
/usr/bin/pip2.7 install --upgrade pip
fi

sudo service syslog stop;
sudo chkconfig syslog off;
sudo service postfix stop;
sudo chkconfig postfix off;
sudo chkconfig --del postfix;

#Copy GeoIP.dat
sudo gsutil -m cp -r gs://startup_scripts_us/scripts/GeoIP /usr/share/;

#Install syslog-ng:
#if [ ! -f /etc/syslog-ng/syslog-ng.conf ]; then
if ! grep log6 /etc/syslog-ng/syslog-ng.conf &> /dev/null; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslog/syslog-ng.conf-www6 /etc/syslog-ng/syslog-ng.conf;
service syslog-ng restart;
chkconfig syslog-ng on;

else
echo "syslog-ng is installed"
fi

##############
#Install Nagios/cacti client

if [ -d /usr/lib64/nagios/plugins ]; then
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
yum -y install nagios-plugins nagios-plugins-nrpe nrpe nagios-nrpe gd-devel net-snmp;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe-2.15-7.el6.x86_64.rpm /home/anand/;
yum -y remove nrpe && rpm -Uvh /home/anand/nrpe-2.15-7.el6.x86_64.rpm;
gsutil cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
gsutil -m cp -r gs://startup_scripts_us/scripts/nagios/plugins/* /usr/lib64/nagios/plugins/;
chmod +x /usr/lib64/nagios/plugins/*
chkconfig nrpe on;
service nrpe restart;
chkconfig snmpd on;
EOF

else
echo "nagios is installed"
fi

#check nrpe allowed_hosts
if ! grep 'allowed_hosts=10.128.1.15,104.197.43.125,50.225.47.189' /etc/nagios/nrpe.cfg; then
gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpe.cfg /etc/nagios/;
/sbin/service nrpe restart;
fi

#############

#Copy CRONJOBS:
sudo su - << EOF
gsutil -m cp gs://startup_scripts_us/scripts/cron_root-us-east-go /var/spool/cron/root;
service crond restart;
EOF

#Configure services to run on reboot:
service sendmail restart;
chkconfig httpd off;
chkconfig crond on;
chkconfig iptables off;
chkconfig sendmail on;
chkconfig syslog-ng on;
chkconfig syslog off;

#USE GIT TO ADD CROSSCHANNEL.COM
if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        sudo su - << EOF
        sudo mkdir -p /var/www/vhosts;
        cd /var/www/vhosts
        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        git config core.filemode false;
        git config user.email "sourabh@crosschannel.com";
        git config user.name "Sourabh Niyogi";
        git fetch upstream;
        git merge upstream/master;
EOF
fi

# Copying scripts to /root/scripts

mkdir -p /root/scripts
sudo gsutil -m cp gs://startup_scripts_us/scripts/git_clone_cc.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ssh_keys_chk.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/go/go-crosschannel-install.sh /root/scripts;
sudo gsutil -m cp gs://startup_scripts_us/scripts/ntpdate*.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/nagios/nrpechk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/syslogchk.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/startup-script-us-east-failsafe-go.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/python_version_change.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/python_version_change_cron.sh /root/scripts/;
sudo gsutil -m cp gs://startup_scripts_us/scripts/cron.hourly_crosschannel /etc/cron.hourly/crosschannel;
sudo gsutil -m cp gs://startup_scripts_us/scripts/logrotate.d_crosschannel /etc/logrotate.d/crosschannel;
sudo chmod -R +x /root/scripts


# installing GO
if [ ! -d /usr/local/go ]; then
sudo gsutil -m cp gs://startup_scripts_us/scripts/go/go-crosschannel-install.sh /home/anand;
sudo sh /home/anand/go-crosschannel-install.sh;
fi

# running go/crosschannel command
if netstat -apn | egrep '::80' > /dev/null; then
  echo crosschannel is running
else
sudo su - << EOF
cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh crosschannel 2> /var/log/git-crosschannel.err  > /var/log/git-crosschannel.log
EOF
fi

#######

#Change 'assumeyes=1' --> 'assumeyes=0' in yum.conf
sed -i '/assumeyes/d' /etc/yum.conf
sed -i "$ i\assumeyes=0" /etc/yum.conf

################

#Denyhosts
if [ ! -d /var/lib/denyhosts ]; then
sudo cp gs://startup_scripts_us/scripts/denyhosts/allowed-hosts /var/lib/denyhosts/;
sudo cp gs://startup_scripts_us/scripts/denyhosts/denyhosts.conf /etc;
sudo su - << EOF
echo `ifconfig | grep 'inet addr:10' | awk '{print$2}' | cut -d ":" -f2` >> /var/lib/denyhosts/allowed-hosts
EOF
service denyhosts restart;
chkconfig denyhosts on;
else
echo "Denyhosts is already installed"
fi

###############

# Commenting out the startup-script-us-east-failsafe-go.sh line in crontab after the first run
/bin/sed -i 's/\*\/1 \* \* \* \* \/usr\/bin\/flock \-w 0 \/var\/run\/startup-script-us-east-failsafe-go.lock \/bin\/sh \/root\/scripts\/startup-script-us-east-failsafe-go.sh/\#\*\/1 \* \* \* \* \/usr\/bin\/flock \-w 0 \/var\/run\/startup-script-us-east-failsafe-go.lock \/bin\/sh \/root\/scripts\/startup-script-us-east-failsafe-mopub-go.sh/g' /var/spool/cron/root
