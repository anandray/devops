FROM centos

COPY bin/bootnode /usr/bin/bootnode
COPY bin/geth /usr/bin/geth
COPY bin/swarm /usr/bin/swarm

COPY scripts/geth-start.sh /wolk/scripts/geth-start.sh
COPY scripts/geth-start-crontab.sh /wolk/scripts/geth-start-crontab.sh
COPY scripts/syslog-ng-start.sh /wolk/scripts/syslog-ng-start.sh
COPY scripts/syslog-ng-start-crontab.sh /wolk/scripts/syslog-ng-start-crontab.sh
COPY scripts/swarm-start-firsttime.sh /wolk/scripts/swarm-start-firsttime.sh
COPY scripts/hi.sh /wolk/scripts/hi.sh
COPY scripts/genesis.json.sh /wolk/scripts/genesis.json.sh

RUN unlink /etc/localtime
RUN ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime
RUN yum -y install crontabs epel-release initscripts iproute mlocate net-tools sudo telnet vim wget which git

# installing syslog-ng
RUN rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
RUN yum -y install syslog-ng syslog-ng-libdbi libdbi-devel 

# adding syslog config to /etc/syslog-ng/conf.d/swarm-syslog.conf
RUN echo "# start swarm syslogging" >> /etc/syslog-ng/conf.d/swarm-syslog.conf
RUN echo "" >> /etc/syslog-ng/conf.d/swarm-syslog.conf
RUN echo "filter f_log6 { match(\"wolk-debug\" value(\"PROGRAM\")) or match(\"wolk-trace\" value(\"PROGRAM\")) or match(\"wolk-cloud\" value(\"PROGRAM\")) or match(\"wolk-mining\" value(\"PROGRAM\")) or match(\"wolk-mining\" value(\"PROGRAM\")) or match(\"wolk-tcp\" value(\"PROGRAM\")) or match(\"wolk-http\" value(\"PROGRAM\")); };" >> /etc/syslog-ng/conf.d/swarm-syslog.conf
RUN echo "" >> /etc/syslog-ng/conf.d/swarm-syslog.conf
RUN echo "destination log6 {tcp(\"log6\" port (5000));};" >> /etc/syslog-ng/conf.d/swarm-syslog.conf
RUN echo "log { source(s_sys); filter(f_log6); destination(log6); };" >> /etc/syslog-ng/conf.d/swarm-syslog.conf
RUN echo "" >> /etc/syslog-ng/conf.d/swarm-syslog.conf
RUN echo "# /end swarm syslogging" >> /etc/syslog-ng/conf.d/swarm-syslog.conf

#RUN sed -i 's/system/#system/g' /etc/syslog-ng/syslog-ng.conf
RUN	sed -i -E 's/^(\s*)system\(\);/\1unix-stream("\/dev\/log");/' /etc/syslog-ng/syslog-ng.conf
RUN sed -i 's/udp/#udp/g' /etc/syslog-ng/syslog-ng.conf
RUN sed -i '29 i\    tcp(ip(0.0.0.0) port(5000));' /etc/syslog-ng/syslog-ng.conf
RUN /usr/sbin/chkconfig syslog-ng on
#RUN /usr/sbin/syslog-ng -F -p /var/run/syslogd.pid &
RUN echo "alias vi='/usr/bin/vim'" >> ~/.bashrc
RUN echo "" >> ~/.bashrc

RUN echo "if [ -d /var/www/vhosts/data ]; then" >> ~/.bashrc
RUN echo "DATADIR=/var/www/vhosts/data" >> ~/.bashrc
RUN echo "fi" >> ~/.bashrc
RUN echo "" >> ~/.bashrc

RUN echo "if [ -f /var/www/vhosts/data/bootnode.log ]; then" >> ~/.bashrc
RUN echo "enode=`grep "self=enode:" /var/www/vhosts/data/bootnode.log | head -n1 | awk '{print$6}' | cut -d "=" -f2 | cut -d "@" -f1`" >> ~/.bashrc
RUN echo "fi" >> ~/.bashrc
RUN echo "" >> ~/.bashrc

RUN echo "if ifconfig eth0 &> /dev/null; then" >> ~/.bashrc
RUN echo "ip=`ifconfig eth0 | grep inet | awk '{print$2}' | cut -d ":" -f2 | head -n1`" >> ~/.bashrc
RUN echo "fi" >> ~/.bashrc
RUN echo "" >> ~/.bashrc

RUN echo "set -o ignoreeof" >> ~/.bashrc

ENTRYPOINT "/wolk/scripts/geth-start.sh" ; "/wolk/scripts/hi.sh" ; "/wolk/scripts/geth-start-crontab.sh" ; "/wolk/scripts/syslog-ng-start-crontab.sh" ; "crond" ; /bin/bash
