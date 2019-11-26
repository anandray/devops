#!/bin/bash
#sudo apt-get -y install java-1.8.0-openjdk-devel;
sudo curl -f -o /usr/local/hbase-1.1.2-bin.tar.gz http://storage.googleapis.com/cloud-bigtable/hbase-dist/hbase-1.1.2/hbase-1.1.2-bin.tar.gz;
sudo tar xvf /usr/local/hbase-1.1.2-bin.tar.gz -C /usr/local/;
sudo mkdir -p /usr/local/hbase-1.1.2/lib/bigtable;
sudo curl http://repo1.maven.org/maven2/com/google/cloud/bigtable/bigtable-hbase-1.1/0.3.0/bigtable-hbase-1.1-0.3.0.jar -f -o /usr/local/hbase-1.1.2/lib/bigtable/bigtable-hbase-1.1-0.3.0.jar;
sudo curl http://repo1.maven.org/maven2/io/netty/netty-tcnative-boringssl-static/1.1.33.Fork13/netty-tcnative-boringssl-static-1.1.33.Fork13-linux-x86_64.jar -f -o /usr/local/hbase-1.1.2/lib/netty-tcnative-boringssl-static-1.1.33.Fork13-linux-x86_64.jar;
sudo curl https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-latest-hadoop2.jar -f -o /usr/local/hbase-1.1.2/lib/gcs-connector-latest-hadoop2.jar;

# copying hbase-site.xml and hbase-env.sh, adding to crontab and starting hbase
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-site-us-east.xml /usr/local/hbase-1.1.2/conf/hbase-site.xml;
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-env.sh /usr/local/hbase-1.1.2/conf/;
#echo 'export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/' >> /root/.bashrc;
#export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/;
mkdir /root/scripts/;
gsutil cp gs://startup_scripts_us/scripts/hbase/hbasechk-ha2-go.sh /root/scripts/;
echo 1 > /proc/sys/net/ipv4/tcp_tw_reuse;
echo 1 > /proc/sys/net/ipv4/tcp_tw_recycle;
echo '
# Recycle and Reuse TIME_WAIT sockets faster
net.ipv4.tcp_tw_recycle = 1
net.ipv4.tcp_tw_reuse = 1' >> /etc/sysctl.conf;
/sbin/sysctl -p;
sh /root/scripts/hbasechk-ha2-go.sh;

echo "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
SHELL=/bin/bash
" > /var/spool/cron/crontabs/root;

sed -i '/hbasechk/d' /var/spool/cron/crontabs/root;
echo "* * * * * /bin/sh /root/scripts/hbasechk-ha2-go.sh >> /var/log/hbasechk-ha2-go.log 2>&1" >> /var/spool/cron/crontabs/root;

gsutil cp gs://startup_scripts_us/scripts/dataproc/ha6_crontab_trick.sh /home/anand/;
cp -rf /home/anand/ha6_crontab_trick.sh /root/scripts/;
sh /home/anand/ha6_crontab_trick.sh;
EOF
