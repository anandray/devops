#!/bin/bash

sudo yum -y install google-authenticator
/usr/bin/google-authenticator << EOF
y
y
y
y
EOF
sudo mkdir -p /root/scripts
sudo cp gs://wolk-scripts/sshd_config /etc/ssh/
sudo cp gs://wolk-scripts/sshd /etc/pam.d/
sudp cp gs://wolk-scripts/gs-wolk-cdn-acl.sh /root/scripts/
sudo service sshd restart
sudo echo "
MAILTO=''
SHELL=/bin/bash
*/1 * * * * sh /root/scripts/gs-wolk-cdn-acl.sh &> /var/log/gs-wolk-cdn-acl.log" > /var/spool/cron/root
