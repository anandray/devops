#!/bin/bash
region=${1}
node=${2}
provider=${3}
fixedinstance=wolk-$node-$provider-$region-tablestore
PrimaryZone=$(aliyuncli ecs DescribeZones --RegionId $region --filter Zones.Zone[*].ZoneId --output text | awk '{print$2}')
LBName=wolk-ali-$region-slb
TemplateId=$(aliyuncli ecs DescribeLaunchTemplates --RegionId $region --filter LaunchTemplateSets.LaunchTemplateSet[*].LaunchTemplateId --output text)
certificate_id=$(aliyuncli slb DescribeServerCertificates --RegionId $region --filter ServerCertificates.ServerCertificate[*].ServerCertificateId --output text)
ImageName=`aliyuncli ecs DescribeImages --RegionId $region --filter Images.Image[*].ImageName --output text | cut -f1`
ImageId=`aliyuncli ecs DescribeImages --RegionId $region --filter Images.Image[*].ImageId --output text | cut -f1`
SecurityGroupId=`aliyuncli ecs DescribeSecurityGroups --RegionId $region --filter SecurityGroups.SecurityGroup[*].SecurityGroupId --output text`
KeyPairName=`aliyuncli ecs DescribeKeyPairs --RegionId $region --filter KeyPairs.KeyPair[*].KeyPairName --output text`

# Autoscaling Group
VSwitchId=`aliyuncli ecs DescribeVSwitches --RegionId $region --filter VSwitches.VSwitch[*].VSwitchId --output text`
aliyuncli ess CreateScalingGroup --RegionId $region --ScalingGroupName wolk-$provider-$node-$region-ss --MinSize 1 --MaxSize 1 --VSwitchId $VSwitchId

#Create Storage Instance
suffix=`date | sha256sum | head -c 4 ; echo`
aliyun ecs RunInstances --RegionId $region --LaunchTemplateId $TemplateId --InternetMaxBandwidthOut 1 --InstanceName wolk-ali-$region-vm-$suffix

#Create Load Balancer
aliyuncli slb CreateLoadBalancer --RegionId $region --PrimaryZone $PrimaryZone --LoadBalancerName $LBName --InstanceType Internet --InstanceSpec slb.s1.small

# Add VServerGroup
lb_id=$(aliyuncli slb DescribeLoadBalancers --RegionId $region --filter LoadBalancers.LoadBalancer[*].LoadBalancerId --output text)
echo $lb_id
aliyuncli slb CreateVServerGroup --RegionId $region --VServerGroupName wolk-$provider-$node-$region-vs --LoadBalancerId $lb_id

# Attach VServerGroup to AutoScalingGroup
VServerGroupId=`aliyuncli slb DescribeVServerGroups --RegionId $region --VServerGroupName wolk-$provider-$node-$region-vs --LoadBalancerId $lb_id --filter VServerGroups.VServerGroup[*].VServerGroupId --output text`
ScalingGroupId=`aliyuncli ess DescribeScalingGroups --RegionId $region --filter ScalingGroups.ScalingGroup[*].ScalingGroupId --output text`

# Add Scaling Configuration
aliyuncli ess CreateScalingConfiguration --RegionId $region --ScalingGroupId $ScalingGroupId --ImageName $ImageName --SecurityGroupId $SecurityGroupId --InstanceType ecs.t5-lc1m2.small --SystemDiskCategory cloud_efficiency --SystemDiskSize 20 --InstanceName wolk-$node-$provider-$region --InstanceChargeType Pay-As-You-Go --KeyPairName $KeyPairName --HostName wolk-$node-$provider-$region

#Add instance to Load Balancer Backend Pool
server_id=$(aliyuncli ecs DescribeInstances --RegionId $region --filter Instances.Instance[*].InstanceId --output text)
echo $server_id

#Create HTTPSListener
aliyuncli slb CreateLoadBalancerHTTPSListener --RegionId $region --LoadBalancerId $lb_id --ListenerPort 443 --StickySession off --BackendServerPort 443 --HealthCheck on --HealthCheckURI / --HealthCheckTimeout 3 --HealthCheckInterval 5 --HealthyThreshold 3 --UnhealthyThreshold 3 --Bandwidth -1 --ServerCertificateId $certificate_id

#Start Listener
aliyuncli slb StartLoadBalancerListener --RegionId $region --LoadBalancerId $lb_id --ListenerPort 443

#Add server to Backend
aliyuncli slb AddBackendServers --RegionId $region --LoadBalancerId $lb_id --BackendServers "[{'ServerId':'$server_id'}]"

#Load Balancer IP
LoadBalancer_IP=$(aliyuncli slb DescribeLoadBalancers --RegionId $region --filter LoadBalancers.LoadBalancer[*].Address --output text)
echo $LoadBalancer_IP

#Create Fiixed Instance
aliyun ecs RunInstances --RegionId $region --LaunchTemplateId $TemplateId --InternetMaxBandwidthOut 1 --InstanceName $fixedinstance

# scp wolk.toml
public_ip=$(aliyuncli ecs DescribeInstances --RegionId $region --filter Instances.Instance[*].PublicIpAddress.IpAddress --output text)
