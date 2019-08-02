#!/bin/bash
region=ap-southeast-1
node=12
provider=ali
fixedinstance=wolk-$provider-$node-$region-fi
PrimaryZone=$(aliyuncli ecs DescribeZones --RegionId ap-southeast-1 --filter Zones.Zone[*].ZoneId --output text | awk '{print$2}')
LBName=wolk-ali-$region-slb
TemplateId=$(aliyuncli ecs DescribeLaunchTemplates --RegionId $region --filter LaunchTemplateSets.LaunchTemplateSet[*].LaunchTemplateId --output text)
certificate_id=$(aliyuncli slb DescribeServerCertificates --RegionId $region --filter ServerCertificates.ServerCertificate[*].ServerCertificateId --output text)
#Create Storage Instance
aliyun ecs RunInstances --RegionId $region --LaunchTemplateId $TemplateId --InternetMaxBandwidthOut 1 --InstanceName wolk-ali-$region-vm
#Create Load Balancer
aliyuncli slb CreateLoadBalancer --RegionId $region --PrimaryZone $PrimaryZone --LoadBalancerName $LBName --InstanceType Internet --InstanceSpec slb.s1.small
#Create Storage Box
#suffix=`date | sha256sum | head -c 4 ; echo`
#aliyun ecs RunInstances --RegionId $region --LaunchTemplateId $TemplateId --InternetMaxBandwidthOut 1 --InstanceName wolk-$provider-$region-$suffix
#Add instance to Load Balancer Backend Pool
lb_id=$(aliyuncli slb DescribeLoadBalancers --RegionId $region --filter LoadBalancers.LoadBalancer[*].LoadBalancerId --output text)
echo $lb_id
server_id=$(aliyuncli ecs DescribeInstances --RegionId $region --filter Instances.Instance[*].InstanceId --output text)
echo $server_id

#aliyuncli slb AddBackendServers --RegionId $region --LoadBalancerId $lb_id --BackendServers "[{'ServerId':'$server_id'}]"
#Create HTTPSListener
aliyuncli slb CreateLoadBalancerHTTPSListener --RegionId $region --LoadBalancerId $lb_id --ListenerPort 443 --StickySession off --BackendServerPort 443 --HealthCheck on --HealthCheckURI / --HealthCheckTimeout 3 --HealthCheckInterval 5 --HealthyThreshold 3 --UnhealthyThreshold 3 --Bandwidth -1 --ServerCertificateId $certificate_id
#Start Listener
aliyuncli slb StartLoadBalancerListener --RegionId $region --LoadBalancerId $lb_id --ListenerPort 443
#Add server to Backend
aliyuncli slb AddBackendServers --RegionId $region --LoadBalancerId $lb_id --BackendServers "[{'ServerId':'$server_id'}]"
#Load Balancer IP
LoadBalancer_IP=$(aliyuncli slb DescribeLoadBalancers --RegionId ap-southeast-1 --filter LoadBalancers.LoadBalancer[*].Address --output text)
echo $LoadBalancer_IP
#Create Fiixed Instance
aliyun ecs RunInstances --RegionId $region --LaunchTemplateId $TemplateId --InternetMaxBandwidthOut 1 --InstanceName $fixedinstance

