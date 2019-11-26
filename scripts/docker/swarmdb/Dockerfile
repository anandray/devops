FROM centos

COPY swarmdb /usr/local/swarmdb
RUN /usr/local/swarmdb/bin/wget -O /usr/local/swarmdb/bin/wolkdb https://github.com/wolktoken/swarm.wolk.com/raw/master/src/github.com/ethereum/go-ethereum/swarmdb/server/wolkdb
RUN /usr/local/swarmdb/bin/wget -O /usr/local/swarmdb/bin/wolk https://github.com/wolktoken/swarm.wolk.com/raw/master/src/github.com/ethereum/go-ethereum/swarmdb/client/wolk/wolk

RUN chmod +x /usr/local/swarmdb/bin/wolkdb
RUN chmod +x /usr/local/swarmdb/bin/wolk

RUN yum -y install crontabs epel-release initscripts iproute mlocate net-tools sudo telnet vim wget which git && rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm

RUN unlink /etc/localtime
RUN ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime

RUN echo "export PATH=$PATH:/usr/local/swarmdb/bin" >> ~/.bashrc
RUN echo "alias vi='/usr/bin/vim'" >> ~/.bashrc
RUN echo "" >> ~/.bashrc

COPY swarmdb/scripts/wolkdb-start-crontab /var/spool/cron/root

RUN echo "set -o ignoreeof" >> ~/.bashrc

ENTRYPOINT "/usr/local/swarmdb/scripts/wolkdb-start-firstime.sh" ; "/usr/sbin/crond" ; "/bin/bash"
