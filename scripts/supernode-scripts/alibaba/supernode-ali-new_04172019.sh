#!/bin/bash
region=us-west-1
node=12
provider=ali
fixedinstance=wolk-$provider-$node-$region-fi
PrimaryZone=$(aliyuncli ecs DescribeVSwitches --RegionId $region --filter VSwitches.VSwitch[*].ZoneId --output text)
LBName=wolk-ali-$region-slb
TemplateId=$(aliyuncli ecs DescribeLaunchTemplates --RegionId $region --filter LaunchTemplateSets.LaunchTemplateSet[*].LaunchTemplateId --output text)
certificate_id=$(aliyuncli slb DescribeServerCertificates --RegionId $region --filter ServerCertificates.ServerCertificate[*].ServerCertificateId --output text)
Vswitch=$(aliyuncli ecs DescribeVSwitches --RegionId $region --filter VSwitches.VSwitch[*].VSwitchId --output text)
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
#Create Fiixed Instance
#aliyun ecs RunInstances --RegionId $region --LaunchTemplateId $TemplateId --InternetMaxBandwidthOut 1 --InstanceName $fixedinstance
