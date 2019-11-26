#!/bin/bash
sed -i 's/KeepAlive On/KeepAlive Off/g' /etc/httpd/conf/httpd.conf
#sed -i 's/KeepAlive 650/KeepAlive On/g' /etc/httpd/conf/httpd.conf
sed -i 's/MaxKeepAliveRequests 10000/MaxKeepAliveRequests 100/g' /etc/httpd/conf/httpd.conf
#sed -i 's/KeepAliveTimeout 15/KeepAliveTimeout 650/g' /etc/httpd/conf/httpd.conf
sed -i 's/KeepAliveTimeout 650/KeepAliveTimeout 15/g' /etc/httpd/conf/httpd.conf
service httpd restart
