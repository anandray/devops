FROM centos

COPY swarmdb /usr/local/swarmdb
COPY go /root/go
COPY .google /root/.google
COPY ssh/* /root/.ssh/
COPY ssh/ssh_config /etc/ssh/

#RUN yum -y install crontabs epel-release initscripts iproute mlocate net-tools sudo telnet vim wget which git nmap-ncat && rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
#RUN yum -y install crontabs epel-release initscripts iproute mlocate net-tools sudo telnet vim wget which git nmap-ncat gcc-c++ make cpp tcl && rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm

RUN yum -y install epel-release cronie wget which telnet net-tools && rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
RUN yum -y install syslog-ng syslog-ng-libdbi libdbi-devel

RUN unlink /etc/localtime
RUN ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime

RUN chmod +x /root/go/src/github.com/wolkdb/plasma/build/bin/plasma*
RUN mkdir /root/data
#RUN echo "export PATH=$PATH:/root/go/src/github.com/wolkdb/plasma/build/bin" >> /root/.bashrc
#RUN source /root/.bashrc

#RUN nohup /root/go/src/github.com/wolkdb/plasma/build/bin/plasma --bootnodes enode://827192dd0616dc4f4ae9676b7cb2c56f8fdb478afb04a8f1bc74471806379cf27eaa5d6c1262005180cb516bb171060aaac5d287ddbaddde0ca498eb7fab1f3d@35.193.142.191:30303 --datadir /root/data --verbosity 4 --maxpeers  25 2>> /root/data/plasma.log &

#EXPOSE 8545
#EXPOSE 30303
#EXPOSE 30304

#ENTRYPOINT "/bin/bash"
#ENTRYPOINT "/usr/local/swarmdb/scripts/plasma-start.sh" ; "/bin/bash"
ENTRYPOINT "/usr/local/swarmdb/scripts/plasma-start.sh" ; "/usr/sbin/crond" ; "/bin/bash"
