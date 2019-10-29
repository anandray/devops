#!/bin/bash

region=`aws ec2 describe-regions | grep RegionName | cut -d "\"" -f4`
region_options=`echo $region`
options=($region_options)
prompt="Select an AWS region:"

PS3="$prompt "
select answer in "${options[@]}"; do
#    aws ec2 describe-availability-zones --region=$answer | grep ZoneName | cut -d"\"" -f4
    zones=`aws ec2 describe-availability-zones --region=$answer | grep ZoneName | cut -d"\"" -f4 | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`
        echo 'region="'$answer'"' > update-autoscaling-group-$answer.conf
        echo 'awsregion="'$answer'"' >> update-autoscaling-group-$answer.conf
        echo 'zones="'$zones'"' >> update-autoscaling-group-$answer.conf
        echo "Enter node #:"
        read node
            echo 'node="'$node'"' >> update-autoscaling-group-$answer.conf
      break 2
done

source update-autoscaling-group-$answer.conf

# derivative inputs
app="cloudstore"
port=80
project="wolk-$region"
provider="aws"
fixedinstance="wolk-$node-$provider-$region-dynamo"
autoscaledinstance="wolk-$node-$provider-$region"
prefix="$app-$region-$provider"

# update auto scaling group
aws autoscaling update-auto-scaling-group --auto-scaling-group-name wolk-autoscale-$region --region $region --min-size $1 --max-size $1

until ! aws autoscaling describe-auto-scaling-groups --region $region --query AutoScalingGroups[*].Instances[*].[LifecycleState] | grep -i Pending > /dev/null; do
printf 'Updating Autoscaling Group'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo "
Autoscaling Group updated...
"

instance_id_1=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | tail -n1 | cut -d"\"" -f2 | awk -vORS=, '{ print $1 }' | sed 's/,/\ /g' | sed 's/$/\n/'`
suffix=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | wc -l`

aws ec2 create-tags --resources $instance_id_1 --tag "Key=Name,Value=$autoscaledinstance-$suffix" --region $region

# Re-run register target group to register new instances
instance_id=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | cut -d"\"" -f2 | awk -vORS=, '{ print "Id="$1 }' | sed 's/,/\ /g' | sed 's/$/\n/'`
target_group_arn=`aws elbv2 describe-target-groups --region $region | grep -i TargetGroupArn | cut -d"\"" -f4 | awk -vORS=, '{ print $1 }' | sed 's/,/\ /g' | sed 's/$/\n/'`
aws elbv2 register-targets --target-group-arn $target_group_arn --targets $instance_id
