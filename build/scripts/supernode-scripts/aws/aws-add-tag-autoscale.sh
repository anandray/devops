#!/bin/bash

region=`aws ec2 describe-regions | grep RegionName | cut -d "\"" -f4`
region_options=`echo $region`
options=($region_options)
prompt="Select an AWS region:"

PS3="$prompt "
select answer in "${options[@]}"; do
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

# add tags to autoscaling instance
echo "
Adding tags to autoscaling instance..
"

all_instance_ids=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | cut -d"\"" -f2`
echo "$all_instance_ids" > /tmp/all_instance_ids
suffix=`date | sha256sum | head -c 4 ; echo`

echo "source supernode-aws-$answer.conf" > /tmp/aws-add-tag-autoscale.sh
echo "autoscaledinstance=\"wolk-$node-autoscale-$region\"" >> /tmp/aws-add-tag-autoscale.sh
#echo "suffix=\`date | sha256sum | head -c 4 ; echo\`" >> /tmp/aws-add-tag-autoscale.sh

#aws ec2 create-tags --resources $instance_id_1 --tag "Key=Name,Value=$autoscaledinstance-$suffix" --region $region
cat /tmp/all_instance_ids | awk '{print"suffix=\`date | sha256sum | head -c 4 ; echo\`; aws ec2 create-tags --resources",$1,"--tag \"Key=Name,Value=$autoscaledinstance-$suffix\" --region $region"}' >> /tmp/aws-add-tag-autoscale.sh && sh /tmp/aws-add-tag-autoscale.sh

rm -rfv supernode-aws-$answer.conf
