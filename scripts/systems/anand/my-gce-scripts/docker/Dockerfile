FROM centos

COPY bin/bootnode /usr/bin/bootnode
COPY bin/geth /usr/bin/geth
COPY bin/swarm /usr/bin/swarm

COPY scripts/geth-start.sh /wolk/scripts/geth-start.sh
COPY scripts/geth-start-crontab.sh /wolk/scripts/geth-start-crontab.sh
COPY scripts/swarm-start-firsttime.sh /wolk/scripts/swarm-start-firsttime.sh
COPY scripts/hi.sh /wolk/scripts/hi.sh
COPY scripts/genesis.json.sh /wolk/scripts/genesis.json.sh

RUN unlink /etc/localtime
RUN ln -s /usr/share/zoneinfo/America/Los_Angeles /etc/localtime
RUN yum -y install sudo net-tools mlocate crontabs which vim epel-release iproute telnet
RUN echo "alias vi='vim'" >> ~/.bashrc
ENTRYPOINT "/wolk/scripts/geth-start.sh" ; "/wolk/scripts/hi.sh" ; "/wolk/scripts/geth-start-crontab.sh" ; "crond" ; /bin/bash
