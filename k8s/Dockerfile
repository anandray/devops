FROM google/cloud-sdk:latest
#FROM ubuntu:18.04
#RUN unlink /etc/localtime
#RUN ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime
#RUN yum -y install net-tools wget telnet
RUN apt-get update && apt-get install -y librados-dev net-tools wget telnet curl
# binary + startup
RUN  gsutil cp gs://wolk-eu/wolk bin/wolk
ADD  bin/wolk /usr/local/bin/wolk
ADD  conf/startup.sh /usr/local/bin/startup.sh
ADD  conf/curl.sh /usr/local/bin/curl.sh

# toml/genesis/hosts
ADD  conf/wolk.toml /usr/local/bin/wolk.toml
ADD  conf/genesis.json /root/go/src/github.com/wolkdb/cloudstore/genesis.json
ADD  conf/hosts /etc/hosts

# cephx
ADD  ceph/cephx/ceph.conf  /etc/ceph/cephx/ceph.conf
ADD  ceph/cephx/ceph.client.admin.keyring /etc/ceph/cephx/ceph.client.admin.keyring

# ssl
ADD  ssl/www.wolk.com.crt /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt
ADD  ssl/www.wolk.com.key /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key

CMD ["/usr/local/bin/startup.sh"]
