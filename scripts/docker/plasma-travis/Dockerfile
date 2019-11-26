FROM centos

COPY swarmdb /usr/local/swarmdb
COPY ssh/* /root/.ssh/
COPY ssh/ssh_config /etc/ssh/

RUN yum -y install crontabs epel-release initscripts iproute mlocate net-tools sudo telnet vim wget which git nmap-ncat gcc-c++ make cpp tcl && rpm -Uvh https://dl.iuscommunity.org/pub/ius/stable/Redhat/7/x86_64/ius-release-1.0-15.ius.el7.noarch.rpm

RUN wget -O /tmp/go1.9.2.linux-amd64.tar.gz https://redirector.gvt1.com/edgedl/go/go1.9.2.linux-amd64.tar.gz
RUN tar zxvpf /tmp/go1.9.2.linux-amd64.tar.gz -C /usr/local
RUN rm -rf /tmp/go1.9.2.linux-amd64.tar.gz

RUN wget -O /tmp/SQLite-3.22.0.tar.gz https://www.sqlite.org/src/tarball/0c55d179/SQLite-0c55d179.tar.gz
RUN tar zxvpf /tmp/SQLite-3.22.0.tar.gz -C /tmp/
RUN mv -f /tmp/SQLite-0c55d179 /tmp/SQLite-3.22.0
RUN mkdir /usr/local/sqlite3
RUN cd /usr/local/sqlite3
RUN /tmp/SQLite-3.22.0/configure && make && make sqlite3.c && make install
RUN rm -rf /tmp/SQLite-3.22.0*

RUN unlink /etc/localtime
RUN ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime

COPY swarmdb/scripts/bashrc /root/.bashrc

ENTRYPOINT "/usr/local/swarmdb/scripts/startup.sh" ; "/usr/sbin/crond" ; "/bin/bash"
