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
	echo -e "\nEnter node #:"
	read node
	    echo 'node="'$node'"' >> supernode-aws-$answer.conf
      break 2
done

source supernode-aws-$answer.conf

# make sure /root/aws exists
if [ ! -d /root/aws/ ]; then
mkdir -p /root/aws/
fi

security_group_id=`aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].GroupId --output text 2> /dev/null`
key_pair=`aws ec2 describe-key-pairs --region $region --query KeyPairs[*].KeyName --output text 2> /dev/null`

consensus_instance_id=`aws ec2 describe-instances --region $region --filters Name=instance-state-name,Values=running --query Reservations[*].Instances[*].[InstanceId] --output text`
autoscaling_group_name=`aws autoscaling describe-auto-scaling-groups --region $region --query AutoScalingGroups[*].AutoScalingGroupName --output text 2> /dev/null`
launch_configuration_name=`aws autoscaling describe-launch-configurations --region $region --query LaunchConfigurations[*].LaunchConfigurationName --output text 2> /dev/null`

# detach instances from autoscaling group
#autoscaling_instance_id=( $(aws autoscaling describe-auto-scaling-instances --region $region --query AutoScalingInstances[*].InstanceId --output text | cut -f1) )
#for instance_id in "${autoscaling_instance_id[@]}";
#do
#echo -e "\nDetaching instances from Autoscaling Group..."
#aws autoscaling detach-instances --instance-ids ${instance_id} --auto-scaling-group-name wolk-autoscale-$region --region $region --should-decrement-desired-capacity
#done

#autoscaling_instance_id=`aws autoscaling describe-auto-scaling-instances --region $region --query AutoScalingInstances[*].InstanceId --output text | cut -f1`
#autoscaling_size=`aws autoscaling describe-auto-scaling-groups --region $region --query AutoScalingGroups[*].MaxSize --output text`
#if [[ $autoscaling_size -ne 0 ]]; then
#        echo -e "\nAutoscaling Group exists with size > 0 - Resizing to ZER0...\n"
#echo -e "\nDetaching instances from Autoscaling Group..."
#aws autoscaling detach-instances --instance-ids $autoscaling_instance_id --auto-scaling-group-name wolk-autoscale-$region --region $region --should-decrement-desired-capacity
#fi

# resize autoscaling to ZERO
autoscaling_size=`aws autoscaling describe-auto-scaling-groups --region $region --query AutoScalingGroups[*].MaxSize --output text`
if [[ $autoscaling_size -ne 0 ]]; then
        echo -e "\nAutoscaling Group exists with size > 0 - Resizing to ZER0...\n"
        # resize autoscaling group to ZERO
        echo -e "\nResizing autoscaling group to ZERO..."
        aws autoscaling update-auto-scaling-group --auto-scaling-group-name wolk-autoscale-$region --region $region --min-size 0 --max-size 0
else
        echo -e "\nAutoscaling Group size = 0..."
fi

# delete instances
instance_id=( $(aws ec2 describe-instances --region $region --filters Name=instance-state-name,Values=running --query Reservations[*].Instances[*].[InstanceId] --output text) )
for instanceid in "${instance_id[@]}";
do
echo -e "\nDeleting Instance/s..."
aws ec2 terminate-instances --instance-ids ${instanceid} --region $region
done

# wait until instances are deleted
echo -e "\nWait until instances are deleted...\n"

until aws ec2 describe-instances --region $region --filter "Name=instance-state-name,Values=terminated" --output text | grep dynamo 2> /dev/null; do
printf 'Deleting Instance/s'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nInstance Deletion complete..."

# delete dynamoDB
echo -e "\nDeleting dynamoDB\n"
if [[ ! $(aws dynamodb describe-table --table-name wolkdbMaster --region $region --output text 2> /dev/null) ]]; then
echo -e "\nDynamoDB doesn't exist...\n"
else
aws dynamodb delete-table --table-name wolkdbMaster --region $region
fi

# delete listeners
loadbalancer_arn=( $(aws elbv2 describe-load-balancers --region $region --query LoadBalancers[*].LoadBalancerArn --output text) )
for loadbalancer in "${loadbalancer_arn[@]}";
do
listener_arn=( $(aws elbv2 describe-listeners --region $region --load-balancer-arn ${loadbalancer} --query Listeners[*].ListenerArn --output text) )
    for listener in "${listener_arn[@]}";
      do
      echo -e "\nDeleting Listener - ${listener}"
      aws elbv2 delete-listener --region $region --listener-arn ${listener}
    done
done

# delete target group/s
target_group=( $(aws elbv2 describe-target-groups --region $region --query TargetGroups[*].TargetGroupArn --output text) )
for target in "${target_group[@]}";
do
echo -e "\nDeleting Target Group - ${target}"
aws elbv2 delete-target-group --target-group-arn ${target} --region $region;
done

# delete Loadbalancer
loadbalancer_arn=( $(aws elbv2 describe-load-balancers --region $region --query LoadBalancers[*].LoadBalancerArn --output text) )
for loadbalancer in "${loadbalancer_arn[@]}";
do
echo -e "\nDeleting Load Balancer - ${loadbalancer}"
aws elbv2 delete-load-balancer --load-balancer-arn ${loadbalancer} --region $region
done

# delete autoscaling group
if [[ $(aws autoscaling describe-auto-scaling-groups --region $region --output text 2> /dev/null) ]]; then
	echo -e "\nAutoscaling Group exists - Deleting...\n"
	aws autoscaling delete-auto-scaling-group --auto-scaling-group-name $autoscaling_group_name --region $region 2> /dev/null
else
	echo -e "\nAutoscaling Group wolk-autoscale-$region doesn't exist..."
fi

# making sure all instances are deleted
if [[ $(aws ec2 describe-instances --region $region --filter "Name=instance-state-name,Values=running" --output text 2> /dev/null) ]]; then
	echo -e "\nInstances exist...\n"
	aws ec2 terminate-instances --instance-ids $consensus_instance_id --region $region 2> /dev/null
else
	echo -e "\nNo instances are in Running state..."
fi

# delete Autoscaling Launch Configuration
if [[ $(aws autoscaling describe-launch-configurations --region $region --output text 2> /dev/null) ]]; then
	echo -e "\nAutoscaling Launch Configuration exists - Deleting...\n"
	aws autoscaling delete-launch-configuration --launch-configuration-name $launch_configuration_name --region $region
else
        echo -e "\nAutoscaling Launch Configuration wolk-launch-config--$region doesn't exist..."
fi

# security groups
#security_group=( $(aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].GroupId --output text) )
#for securitygroup_id in "${security_group[@]}";
#do
#echo -e "\nDeleting Security Group - ${securitygroup_id}"
#aws ec2 delete-security-group --group-id ${securitygroup_id} --region $region
#done

# key pairs
if [[ $(aws ec2 describe-key-pairs --region $region --output text 2> /dev/null) ]]; then
echo -e "\nKey Pair exists - Deleting...\n"
aws ec2 delete-key-pair --key-name $key_pair --region $region
else
	echo -e "\nKey Pair doesn't exist...\n"
fi

# making sure the target group/s are deleted
target_group=( $(aws elbv2 describe-target-groups --region $region --query TargetGroups[*].TargetGroupArn --output text) )
for target in "${target_group[@]}";
do
echo -e "\nTrying to delete Target Group/s Again..."
aws elbv2 delete-target-group --target-group-arn ${target} --region $region;
done

# attempting to delete security group if it still exists
#if [[ $(aws ec2 describe-security-groups --group-names wolk-sg-$region --region $region --output text 2> /dev/null) ]]; then
#echo -e "\nTrying to delete Security Group again...\n"
#aws ec2 delete-security-group --group-id $security_group_id --region $region
#else
#        echo -e "\nSecurity Group doesn't exist...\n"
#fi

# attempting to delete autoscaling group if it still exists
if [[ $(aws autoscaling describe-auto-scaling-groups --region $region --output text 2> /dev/null) ]]; then
        echo -e "\nTying to delete Autoscaling Group again...\n"
        # resize autoscaling group to ZERO
        aws autoscaling delete-auto-scaling-group --auto-scaling-group-name $autoscaling_group_name --region $region 2> /dev/null
else
        echo -e "\nAutoscaling Group wolk-autoscale-$region doesn't exist..."
fi

# attempting to delete Autoscaling Launch Configuration if it still exists
if [[ $(aws autoscaling describe-launch-configurations --region $region --output text 2> /dev/null) ]]; then
        echo -e "\nTrying to delete Autoscaling Launch Configuration again...\n"
        aws autoscaling delete-launch-configuration --launch-configuration-name $launch_configuration_name --region $region 2> /dev/null
else
        echo -e "\nAutoscaling Launch Configuration wolk-launch-config--$region doesn't exist..."
fi

certificate_arn=( $(aws acm list-certificates --region $region --query CertificateSummaryList[*].CertificateArn --output text) )
for certificate in "${certificate_arn[@]}";
do
echo -e "\nDeleting SSL Certificate/s..."
aws acm delete-certificate --certificate-arn ${certificate} --region $region
done

if [[ $(aws ec2 describe-security-groups --group-names wolk-sg-$region --region $region --output text 2> /dev/null) ]]; then
security_group=( $(aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].GroupId --output text) )
for securitygroup_id in "${security_group[@]}";
do
echo -e "\nTrying to delete Security Group ${securitygroup_id} again...\n"
aws ec2 delete-security-group --group-id ${securitygroup_id} --region $region 2> /dev/null
if [[ ! $(aws ec2 describe-security-groups --group-names wolk-sg-$region --region $region --output text 2> /dev/null) ]]; then
	echo -e "\nSecurity Group deletion successful...\n"
fi
done
else
	echo -e "\nSecurity Group doesn't exist...\n"
fi
