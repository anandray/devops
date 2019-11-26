FROM centos

COPY bin/bootnode /usr/bin/bootnode
COPY bin/geth /usr/bin/geth
COPY bin/swarm /usr/bin/swarm

COPY scripts/geth-start.sh /wolk/scripts/geth-start.sh
COPY scripts/swarm-start-firsttime.sh /wolk/scripts/swarm-start-firsttime.sh
COPY scripts/hi.sh /wolk/scripts/hi.sh
COPY scripts/genesis.json.sh /wolk/scripts/genesis.json.sh


#CMD ["/wolk/scripts/geth-start.sh"]
#CMD ["/wolk/scripts/swarm-start-firsttime.sh"]
#WORKDIR /wolk/scripts/
#CMD "./hi.sh"
#ENTRYPOINT "/wolk/scripts/hi.sh" ; /bin/bash

#RUN yum -y install sudo

RUN yum -y install net-tools
ENTRYPOINT "/wolk/scripts/geth-start.sh" ; "/wolk/scripts/hi.sh" ; /bin/bash






