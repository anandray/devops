#!/bin/bash

### VERSION HBASE-1.2.1 ###

if ! rpm -qa | grep java-1.8.0 > /dev/null; then
sudo yum -y install java-1.8.0-openjdk-devel;
else
echo "java-1.8.0-openjdk-devel already installed..."
fi

if [ -d /usr/local/hbase-1.2.1 ]; then
sudo mv -f /usr/local/hbase-1.2.1 /usr/local/hbase-1.2.1_BAK_`date +%m%d-%T`;
fi

if [ -d /usr/local/hbase-profile-c ]; then
sudo mv -f /usr/local/hbase-profile-c /usr/local/hbase-profile-c_BAK_`date +%m%d-%T`
fi

#sudo pkill -9 java;

sudo su - << EOF
if ps aux | grep hbase-profile-c | grep -v grep > /dev/null;
  then
  kill -9 $(ps aux | grep hbase-profile-c | grep -v grep | awk '{print$2}')
fi
EOF

sudo su - << EOF
cd /usr/local;
curl -f -o /usr/local/hbase-1.2.1-bin.tar.gz http://storage.googleapis.com/cloud-bigtable/hbase-dist/hbase-1.2.1/hbase-1.2.1-bin.tar.gz
tar xvf /usr/local/hbase-1.2.1-bin.tar.gz -C /usr/local/
mkdir -p /usr/local/hbase-1.2.1/lib/bigtable;
curl http://repo1.maven.org/maven2/com/google/cloud/bigtable/bigtable-hbase-1.2/0.9.5.1/bigtable-hbase-1.2-0.9.5.1.jar -f -o hbase-1.2.1/lib/bigtable/bigtable-hbase-1.2-0.9.5.1.jar;
curl http://repo1.maven.org/maven2/io/netty/netty-tcnative-boringssl-static/1.1.33.Fork19/netty-tcnative-boringssl-static-1.1.33.Fork19.jar     -f -o hbase-1.2.1/lib/netty-tcnative-boringssl-static-1.1.33.Fork19.jar
mv -f /usr/local/hbase-1.2.1 /usr/local/hbase-profile-c
EOF

# copying hbase-site.xml and hbase-env.sh, adding to crontab and starting hbase
sudo su - << EOF
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-site-profile-c-1.2.1.xml /usr/local/hbase-profile-c/conf/hbase-site.xml;
gsutil cp gs://startup_scripts_us/scripts/hbase/hbase-env-1.2.1.sh /usr/local/hbase-profile-c/conf/hbase-env.sh;
EOF

sudo su - << EOF
if ! echo $JAVA_HOME | grep '/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/' > /dev/null;
  then
  sed -i '/JAVA_HOME/d' /root/.bashrc;
  echo 'export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/' >> /root/.bashrc;
  export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/;
else
echo $JAVA_HOME
fi
EOF

sudo su - << EOF
if ! sysctl -a | grep -E 'net.ipv4.tcp_tw_reuse = 1|net.ipv4.tcp_tw_recycle = 1' > /dev/null;
  then
  echo 1 > /proc/sys/net/ipv4/tcp_tw_reuse;
  echo 1 > /proc/sys/net/ipv4/tcp_tw_recycle;
  sed -i '/ipv4.tcp_tw/d' /etc/sysctl.conf
  sed -i '/Recycle and Reuse TIME_WAIT sockets faster/d' /etc/sysctl.conf
  echo '
  # Recycle and Reuse TIME_WAIT sockets faster
  net.ipv4.tcp_tw_recycle = 1
  net.ipv4.tcp_tw_reuse = 1' >> /etc/sysctl.conf;
  /sbin/sysctl -p;
else
echo 'tcp_tw already set to 1'
sysctl -a | grep tcp_tw
fi
EOF

sudo su - << EOF
if ps aux | grep hbase-profile-c > /dev/null;
  then
  kill -9 `ps aux | grep hbase-profile-c | awk '{print$2}'`
  echo "Starting hbase-profile-c..."
  cd /usr/local/hbase-profile-c && ./bin/hbase rest start -p 8081 --infoport 8086 &
else
echo "Starting hbase-profile-c..."
cd /usr/local/hbase-profile-c && ./bin/hbase rest start -p 8081 --infoport 8086 &
fi
EOF

#sh /root/scripts/hbasechk-us.sh;
#sed -i '/hbasechk-us/d' /var/spool/cron/root;
#echo '*/1 * * * * ssh `hostname` /bin/sh /root/scripts/hbasechk-us.sh > /var/log/hbase.log 2>&1' >> /var/spool/cron/root;
#echo '*/2 * * * * kill -9 $(ps aux | grep hbasechk | grep -v grep | awk '{print$2}') > /dev/null 2>&1' >> /var/spool/cron/root;
