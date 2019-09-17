FROM rook/ceph:v1.0.3
ADD  wolk /usr/local/bin/wolk
ADD  wolk.toml /usr/local/bin/wolk.toml
ADD  www.wolk.com.crt /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt
ADD  www.wolk.com.key /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key
ADD  genesis.json /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json
# CHECK: verbosity=6 --wolklog http,storage,cloud 
RUN  sed 's/# create the keyring file/\/usr\/local\/bin\/wolk --datadir \/usr\/local\/bin --rpc --rpcaddr 0.0.0.0 --rpccorsdomain=* --rpcvhosts=* --rpcport 9902 --port 32300 --httpport 82 --config \/usr\/local\/bin\/wolk.toml > \/usr\/local\/bin\/wolk.log \&/' /usr/local/bin/toolbox.sh > /usr/local/bin/toolbox2.sh
RUN  mv /usr/local/bin/toolbox2.sh /usr/local/bin/toolbox.sh
RUN  chmod 777 /usr/local/bin/toolbox.sh
ENV  PORT 82
