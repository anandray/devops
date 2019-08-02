#!/bin/bash

sudo sed -i '/JAVA_HOME/d' /etc/sudoers
sudo sed -i '/JAVA_HOME/d' /home/anand/.bashrc;
sudo sed -i '/JAVA_HOME/d' /root/.bashrc
sudo su - << EOF
echo 'Defaults    env_keep += "JAVA_HOME"' >> /etc/sudoers
echo 'export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/' >> /root/.bashrc
EOF
echo 'export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/' >> /home/anand/.bashrc
echo 'export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/' >> /home/yaron/.bashrc
export JAVA_HOME=/usr/lib/jvm/jre-1.8.0-openjdk.x86_64/
cd /usr/local/hbase-1.1.2
sudo ./bin/hbase rest start&
