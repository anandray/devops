FROM centos
COPY nosql /root/nosql
COPY nosql-start.sh /
COPY nosql-start.sh /
COPY sqlite3 /usr/local/sqlite3
COPY lib /usr/local/lib
COPY sqlite3-bin /usr/local/bin/sqlite3
COPY sqlite3.conf /etc/ld.so.conf.d/
RUN yum -y install epel-release 
RUN yum -y install syslog-ng syslog-ng-libdbi libdbi-devel --nogpgcheck
COPY syslog-ng.conf /etc/syslog-ng/

ENTRYPOINT ["/nosql-start.sh"]
CMD []
