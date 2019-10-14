FROM ubuntu:18.04
#RUN unlink /etc/localtime
RUN ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime
#RUN yum -y install net-tools wget telnet
RUN apt-get update && apt-get install -y librados-dev net-tools wget telnet curl

# binary + startup
ADD  bin/wolk /usr/local/bin/wolk
ADD  conf/startup.sh /usr/local/bin/startup.sh
ADD conf/curl.sh /usr/local/bin/curl.sh

# toml/genesis/hosts
ADD  conf/wolk.toml /usr/local/bin/wolk.toml
ADD  conf/genesis.json /root/go/src/github.com/wolkdb/cloudstore/build/conf/genesis.json
ADD  conf/hosts /etc/hosts

# cephx = alibaba
ADD  ceph/cephx/ceph.conf  /etc/ceph/cephx/ceph.conf
ADD  ceph/cephx/ceph.client.admin.keyring /etc/ceph/cephx/ceph.client.admin.keyring
# cephy = gc stage 0
#ADD  ceph/cephy/ceph.conf  /etc/ceph/cephy/ceph.conf
#ADD  ceph/cephy/ceph.client.admin.keyring /etc/ceph/cephy/ceph.client.admin.keyring
# cephv = azure
ADD  ceph/cephv/ceph.conf  /etc/ceph/cephv/ceph.conf
ADD  ceph/cephv/ceph.client.admin.keyring /etc/ceph/cephv/ceph.client.admin.keyring
# cephw = aws
ADD  ceph/cephw/ceph.conf  /etc/ceph/cephw/ceph.conf
ADD  ceph/cephw/ceph.client.admin.keyring /etc/ceph/cephw/ceph.client.admin.keyring
# cephz = gc stage 1+
ADD  ceph/cephz/ceph.conf  /etc/ceph/cephz/ceph.conf
ADD  ceph/cephz/ceph.client.admin.keyring /etc/ceph/cephz/ceph.client.admin.keyring

# ssl
ADD  ssl/www.wolk.com.crt /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt
ADD  ssl/www.wolk.com.key /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key

CMD ["/usr/local/bin/startup.sh"]
