#!/bin/bash
region=$1
node=$2
provider=ali
fixedinstance=wolk-$provider-$node-$region-tablestore

# Add bashrc and ConsensusIdx + NodeType to wolk.toml
consensus_IP=`aliyuncli ecs DescribeInstances --RegionId $region --filter Instances.Instance[*].PublicIpAddress.IpAddress --output text | tail -n1`
storage_IP=`aliyuncli ecs DescribeInstances --RegionId $region --filter Instances.Instance[*].PublicIpAddress.IpAddress --output text | head -n1`

# Pushing bashrc to Consensus node
ossutil cp -f oss://wolk-ali/wolk-startup-scripts/scripts/cloudstore/cloudstore-bashrc .
scp -q cloudstore-bashrc $consensus_IP:/root/.bashrc

# Pushing bashrc to Storage nodes
for storageIP in "${storage_IP[@]}";
do
scp -q cloudstore-bashrc ${storageIP}:/root/.bashrc
done

# Pushing wolk.toml
echo -e "\nAdding wolk.toml to Consensus node --> $consensus_IP"
scp -q $consensus_IP:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/wolk.toml-ali-template wolk.toml-ali
sed -i "s/_ConsensusIdx/$node/g" wolk.toml-ali
sed -i "s/_NodeType/consensus/g" wolk.toml-ali
sed -i "s/_AlibabaRegion/$region/g" wolk.toml-ali
scp -q wolk.toml-ali $consensus_IP:/root/go/src/github.com/wolkdb/cloudstore/wolk.toml

# make wolk and copy wolk1/2/3/4/5 on Storage Nodes
ssh $consensus_IP git --git-dir=/root/go/src/github.com/wolkdb/cloudstore/.git pull 2> /dev/null
ssh -q $consensus_IP make wolk -C /root/go/src/github.com/wolkdb/cloudstore
for i in {1..5};
do ssh -q $consensus_IP cp -rfv /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk$i;
done

# Pushing wolk.toml to Storage nodes
for storageIP in "${storage_IP[@]}";
do
echo -e "\nAdding wolk.toml to Storage nodes --> ${storageIP}"
sed -i "s|consensus|storage|g" wolk.toml-ali
scp -q wolk.toml-ali ${storageIP}:/root/go/src/github.com/wolkdb/cloudstore/wolk.toml
done

# make wolk and copy wolk1/2/3/4/5 on Storage Nodes
for storageIP in "${storage_IP[@]}";
do
ssh -q ${storageIP} git --git-dir=/root/go/src/github.com/wolkdb/cloudstore/.git pull 2> /dev/null
ssh -q ${storageIP} make wolk -C /root/go/src/github.com/wolkdb/cloudstore
  for i in {1..5};
    do ssh -q ${storageIP} cp -rfv /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk$i;
  done
done
