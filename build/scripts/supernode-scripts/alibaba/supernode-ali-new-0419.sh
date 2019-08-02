#!/bin/bash
region=$1
node=$2
provider=ali
fixedinstance=wolk-$provider-$node-$region-tablestore
PrimaryZone=$(aliyuncli ecs DescribeVSwitches --RegionId $region --filter VSwitches.VSwitch[*].ZoneId --output text)
LBName=wolk-ali-$region-slb
TemplateId=$(aliyuncli ecs DescribeLaunchTemplates --RegionId $region --filter LaunchTemplateSets.LaunchTemplateSet[*].LaunchTemplateId --output text)
certificate_id=$(aliyuncli slb DescribeServerCertificates --RegionId $region --filter ServerCertificates.ServerCertificate[*].ServerCertificateId --output text)
Vswitch=$(aliyuncli ecs DescribeVSwitches --RegionId $region --filter VSwitches.VSwitch[*].VSwitchId --output text)
#Create Fiixed Instance
aliyuncli ecs RunInstances --RegionId $region --LaunchTemplateId $TemplateId --InternetMaxBandwidthOut 1 --InstanceName $fixedinstance
#Create Load Balancer
aliyuncli slb CreateLoadBalancer --RegionId $region --PrimaryZone $PrimaryZone --LoadBalancerName $LBName --InstanceType Internet --InstanceSpec slb.s1.small
lb_id=$(aliyuncli slb DescribeLoadBalancers --RegionId $region --filter LoadBalancers.LoadBalancer[*].LoadBalancerId --output text)
echo $lb_id

#Create HTTPSListener
aliyuncli slb CreateLoadBalancerTCPListener --RegionId $region --LoadBalancerId $lb_id --ListenerPort 443 --StickySession off --BackendServerPort 443 --HealthCheck on --HealthCheckURI / --HealthCheckTimeout 3 --HealthCheckInterval 5 --HealthyThreshold 3 --UnhealthyThreshold 3 --Bandwidth -1 --ServerCertificateId $certificate_id
#Start Listener
aliyuncli slb StartLoadBalancerListener --RegionId $region --LoadBalancerId $lb_id --ListenerPort 443
#Create AutoScaling Group
aliyuncli ess CreateScalingGroup --RegionId $region --ScalingGroupName wolk-ali-$node-$region-ss --MinSize 1 --MaxSize 1 --VSwitchId $Vswitch --LaunchTemplateId $TemplateId --LoadBalanacerId $lb_id
#Enable Scaling Group
sleep 90
ScalingGroupId=$(aliyuncli ess DescribeScalingGroups --RegionId $region --filter ScalingGroups.ScalingGroup[*].ScalingGroupId --output text)
echo $ScalingGroupId
aliyuncli ess EnableScalingGroup --ScalingGroupId $ScalingGroupId --RegionId $region
sleep 60
#AttachLoadBalancer 
aliyun ess AttachLoadBalancers  --ScalingGroupId $ScalingGroupId --LoadBalancer.1 $lb_id

#Modify Sacling Group
aliyuncli ess ModifyScalingGroup --ReionId $region --ScalingGroupId $ScalingGroupId
#resize Scaling Group to 1 
#aliyuncli ess ModifyScalingGroup --RegionId $region --ScalingGroupId $ScalingGroupId MaxSize 1 MinSize 1

#Load Balancer IP
LoadBalancer_IP=$(aliyuncli slb DescribeLoadBalancers --RegionId $region --filter LoadBalancers.LoadBalancer[*].Address --output text)
echo $LoadBalancer_IP

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
