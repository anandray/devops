FROM centos

COPY swarmdb /usr/local/swarmdb
COPY go /root/go
COPY sql /root/sql
COPY .google /root/.google
COPY .aws /root/.aws
COPY ssh/* /root/.ssh/
COPY ssh/ssh_config /etc/ssh/
COPY swarmdb/scripts/bashrc /root/.bashrc

#RUN yum -y install crontabs epel-release mlocate net-tools sudo telnet vim wget which openssh-clients wget &> /dev/null

#RUN yum -y install crontabs epel-release initscripts iproute mlocate net-tools sudo telnet vim wget which git nmap-ncat && rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
#RUN yum -y install crontabs epel-release initscripts iproute mlocate net-tools sudo telnet vim wget which git nmap-ncat gcc-c++ make cpp tcl && rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm

#RUN yum -y install epel-release cronie wget which telnet net-tools && rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm
#RUN yum -y install syslog-ng syslog-ng-libdbi libdbi-devel

RUN unlink /etc/localtime
RUN ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime

RUN chmod +x /root/go/src/github.com/wolkdb/plasma/build/bin/*
RUN chmod +x /root/sql/bin/*

#ENTRYPOINT "/bin/bash"
ENTRYPOINT ["/usr/local/swarmdb/scripts/anand-start.sh","bash"]
CMD ["param1"]
