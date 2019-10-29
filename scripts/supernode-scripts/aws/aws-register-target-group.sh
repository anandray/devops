#!/bin/bash

# usage
#echo ./aws-regions.sh $node_number

region=`aws ec2 describe-regions | grep RegionName | cut -d "\"" -f4`
region_options=`echo $region`
options=($region_options)
prompt="Select an AWS region:"

PS3="$prompt "
select answer in "${options[@]}"; do
#    zones=`aws ec2 describe-availability-zones --region=$answer --output text | awk '{print$NF}' | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`
    zones=`aws ec2 describe-availability-zones --region=$answer | grep ZoneName | cut -d"\"" -f4 | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`
	echo 'region="'$answer'"' > supernode-aws-$answer.conf
	echo 'awsregion="'$answer'"' >> supernode-aws-$answer.conf
	echo 'zones="'$zones'"' >> supernode-aws-$answer.conf
	echo "Enter node #:"
	read node
	    echo 'node="'$node'"' >> supernode-aws-$answer.conf
      break 2
done

source supernode-aws-$answer.conf

# derivative inputs
app="cloudstore"
port=80
project="wolk-$region"
provider="aws"
fixedinstance="wolk-$node-$provider-$region-dynamo"
autoscaledinstance="wolk-$node-autoscale-$region"
prefix="$app-$region-$provider"
instancetemplate="$prefix"
urlmap="$prefix"
instancetemplate="$prefix"
lbname="$prefix-$port"
globalip="$app-$region-$provider-global-ip"
targetproxy="$prefix-target-proxy-$port"
regionalipname="$app-$region-$provider-regional-ip"
healthcheck="$app-$region-healthcheck"
portname="$app-$port"

echo $region
echo $project
echo $provider
echo $fixedinstance
echo $prefix

# register target group
echo "
Register target group
"

instance_id=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | awk -vORS=, '{print"Id="$1}' | sed 's/,/\ /g'`
target_group_arn=`aws elbv2 describe-target-groups --region $region --query TargetGroups[*].TargetGroupArn --output text`
aws elbv2 register-targets --target-group-arn $target_group_arn --targets $instance_id --region $region

rm -rfv /root/aws/supernode-aws-$answer.conf
rm -rfv supernode-aws-$answer.conf
