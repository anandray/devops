FROM centos
COPY .ssh /root/.ssh
COPY hosts /etc/hosts
COPY sql /root/sql
COPY sql-start.sh /
COPY sqlite3 /usr/local/sqlite3
COPY lib /usr/local/lib
COPY sqlite3-bin /usr/local/bin/sqlite3
COPY sqlite3.conf /etc/ld.so.conf.d/
RUN yum -y install epel-release 
RUN yum -y install syslog-ng syslog-ng-libdbi libdbi-devel wget git gcc --nogpgcheck
COPY syslog-ng.conf /etc/syslog-ng/

ENTRYPOINT ["/sql-start.sh"]
CMD []
