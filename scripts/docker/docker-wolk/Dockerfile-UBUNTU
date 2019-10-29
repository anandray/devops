FROM ubuntu:18.04
ENV GOLANG_VERSION 1.12
ENV DEBIAN_FRONTEND noninteractive

RUN apt update && apt-get install -y git libboost-all-dev wget sqlite3 autoconf sudo tzdata bsdmainutils
#RUN apt update && apt-get install -y git apt-utils libboost-all-dev wget sqlite3 autoconf sudo tzdata bsdmainutils vim cron mlocate curl apt-transport-https ca-certificates

RUN mkdir /root/.ssh
COPY ssh/id_rsa /root/.ssh/id_rsa
COPY ssh/id_rsa.pub /root/.ssh/id_rsa.pub
COPY ssh/authorized_keys /root/.ssh/authorized_keys

RUN echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config
RUN echo "UserKnownHostsFile /dev/null" >> /etc/ssh/ssh_config
COPY .bashrc /root/

WORKDIR /root
RUN wget --quiet https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz && tar -xvf go${GOLANG_VERSION}.linux-amd64.tar.gz && mv go /usr/local
ENV GOROOT /usr/local/go
ENV GOPATH /root/go
ENV GOBIN /root/go/bin
ENV PATH   $GOPATH/bin:$GOROOT/bin:$PATH
RUN mkdir -p $GOPATH/src/github.com/wolkdb
WORKDIR $GOPATH/src/github.com/wolkdb
RUN git clone --depth 1 -b master --recurse-submodules git@github.com:wolkdb/cloudstore.git
WORKDIR $GOPATH/src/github.com/wolkdb/cloudstore
RUN make wolk && for i in {1..5}; do cp -rfv build/bin/wolk build/bin/wolk$i; done
ENTRYPOINT ["/bin/bash"]
