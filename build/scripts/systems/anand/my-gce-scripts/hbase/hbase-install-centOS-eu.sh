#!/bin/bash
sudo yum -y install java-1.8.0-openjdk-devel;
sudo rm -rf /usr/local/hbase*;
sudo pkill -9 java;
sudo curl -f -o /usr/local/hbase-1.1.2-bin.tar.gz http://storage.googleapis.com/cloud-bigtable/hbase-dist/hbase-1.1.2/hbase-1.1.2-bin.tar.gz;
sudo tar xvf /usr/local/hbase-1.1.2-bin.tar.gz -C /usr/local/;
sudo mkdir -p /usr/local/hbase-1.1.2/lib/bigtable;
sudo curl http://repo1.maven.org/maven2/com/google/cloud/bigtable/bigtable-hbase-1.1/0.3.0/bigtable-hbase-1.1-0.3.0.jar -f -o /usr/local/hbase-1.1.2/lib/bigtable/bigtable-hbase-1.1-0.3.0.jar;
sudo curl http://repo1.maven.org/maven2/io/netty/netty-tcnative-boringssl-static/1.1.33.Fork13/netty-tcnative-boringssl-static-1.1.33.Fork13-linux-x86_64.jar -f -o /usr/local/hbase-1.1.2/lib/netty-tcnative-boringssl-static-1.1.33.Fork13-linux-x86_64.jar;
sudo curl https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-latest-hadoop2.jar -f -o /usr/local/hbase-1.1.2/lib/gcs-connector-latest-hadoop2.jar;

# copying hbase-site.xml and hbase-env.sh, adding to crontab and starting hbase
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-site-eu.xml /usr/local/hbase-1.1.2/conf/hbase-site.xml;
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-env.sh /usr/local/hbase-1.1.2/conf/;
sed -i '/JAVA_HOME/d' /root/.bashrc;
echo 'export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/' >> /root/.bashrc;
export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/;
mkdir -p /root/scripts/;
gsutil cp gs://startup_scripts_us/scripts/hbase/hbasechk-eu.sh /root/scripts;
echo 1 > /proc/sys/net/ipv4/tcp_tw_reuse;
echo 1 > /proc/sys/net/ipv4/tcp_tw_recycle;
sed -i '/ipv4.tcp_tw/d' /etc/sysctl.conf
sed -i '/Recycle and Reuse TIME_WAIT sockets faster/d' /etc/sysctl.conf
echo '
# Recycle and Reuse TIME_WAIT sockets faster
net.ipv4.tcp_tw_recycle = 1
net.ipv4.tcp_tw_reuse = 1' >> /etc/sysctl.conf;
/sbin/sysctl -p;
sh /root/scripts/hbasechk-eu.sh;
sed -i '/hbasechk-eu/d' /var/spool/cron/root;
echo '*/1 * * * * ssh `hostname` /bin/sh /root/scripts/hbasechk-eu.sh > /var/log/hbase.log 2>&1' >> /var/spool/cron/root;
#echo '*/2 * * * * kill -9 $(ps aux | grep hbasechk | grep -v grep | awk '{print$2}') > /dev/null 2>&1' >> /var/spool/cron/root;
EOF
