#!/bin/bash
region=$1
node=$2
lb_id=$(aliyuncli slb DescribeLoadBalancers --RegionId $region --filter LoadBalancers.LoadBalancer[*].LoadBalancerId --output text)
#Delete Load Balancer
aliyuncli slb DeleteLoadBalancer --RegionId $region --LoadBalancerId $lb_id
#Modify Scaleset to 0
ScalingGroupId=$(aliyuncli ess DescribeScalingGroups --RegionId $region --filter ScalingGroups.ScalingGroup[*].ScalingGroupId --output text)
aliyuncli ess ModifyScalingGroup --RegionId $region --ScalingGroupId $ScalingGroupId MaxSize 0 MinSize 0
sleep 60
#Delete ScalingGroup
aliyuncli ess DeleteScalingGroup --RegionId $region --ScalingGroupId $ScalingGroupId
sleep 60
#Stop Fixed Instance 
InstanceId=$(aliyuncli ecs DescribeInstances --RegionId $region --filter Instances.Instance[*].InstanceId --output text)
aliyuncli ecs StopInstance --RegionId $region --InstanceId $InstanceId
sleep 60
#delete Fixed Instance 
aliyuncli ecs DeleteInstance --RegionId $region --InstanceId $InstanceId

