#!/bin/bash

region=`aws ec2 describe-regions | grep RegionName | cut -d "\"" -f4`
region_options=`echo $region`
options=($region_options)
prompt="Select an AWS region:"

PS3="$prompt "
select answer in "${options[@]}"; do
#    zones=`aws ec2 describe-availability-zones --region=$answer --output text | awk '{print$NF}' | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`
    zones=`aws ec2 describe-availability-zones --region=$answer | grep ZoneName | cut -d"\"" -f4 | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`
	echo 'region="'$answer'"' > staginglb-aws-alt-port-$answer.conf
	echo 'awsregion="'$answer'"' >> staginglb-aws-alt-port-$answer.conf
	echo 'zones="'$zones'"' >> staginglb-aws-alt-port-$answer.conf
	echo "Enter node #:"
	read node
	    echo 'node="'$node'"' >> staginglb-aws-alt-port-$answer.conf
	echo "Enter Port:"
	read port
	    echo 'port="'$port'"' >> staginglb-aws-alt-port-$answer.conf
      break 2
done

mv staginglb-aws-alt-port-$answer.conf staginglb-aws-$port-$answer.conf
source staginglb-aws-$port-$answer.conf

# derivative inputs
app="cloudstore"
#port=81
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

# change default region locally
#echo "
#Change default region locally
#"
sed -i '/region/d' ~/.aws/config
echo "region = $region" >> ~/.aws/config

# make sure /root/aws exists
if [ ! -d /root/aws/ ]; then
mkdir -p /root/aws/
fi

# load balancer
echo -e "\nCreate load balancer..."
security_group_id=`aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].GroupId --output text`
subnet_id=`aws ec2 describe-subnets --region=$region | grep -i SubnetId | cut -d"\"" -f4 | awk -vORS=, '{ print $1 }' | sed 's/,/\ /g' | sed 's/$/\n/'`

aws elbv2 create-load-balancer --region $region --name wolk-lb-$region-$port --subnets $subnet_id --security-group $security_group_id

until aws elbv2 describe-load-balancers --region $region --names wolk-lb-$region-$port --query LoadBalancers[*].State --output text | grep active &> /dev/null; do
printf 'Creating Load Balancer'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nLoad Balancer created..."

# target group
echo -e "\nCreate target group..."
vpc_id=`aws ec2 describe-vpcs --region=$region | grep -i VpcId | cut -d"\"" -f4`
aws elbv2 create-target-group --region $region --name wolk-trgt-grp-$region-$port --protocol HTTP --port $port --vpc-id $vpc_id

until aws elbv2 describe-target-groups --region $region --names wolk-trgt-grp-$region-$port --query TargetGroups[*].Protocol --output text | grep HTTP &> /dev/null; do
printf 'Creating Target Group'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nTarget Group created..."

# register target group
echo -e "\nRegister target group..."

# import ssl certificate
# check if certificate already exists
echo -e "\nCheck if a certificate already exists..."
if [[ $(aws acm list-certificates --region $region --query CertificateSummaryList[*].CertificateArn --output text) ]]; then
echo -e "\nSSL certificate exists. No need to import again..."
certificate_arn=`aws acm list-certificates --region $region --query CertificateSummaryList[*].CertificateArn --output text | cut -f1`
else
echo -e "\nImporting SSL Certificate..."
aws acm import-certificate --certificate file:///etc/ssl/certs/wildcard.wolk.com/www.wolk.com.cert --certificate-chain file:///etc/ssl/certs/wildcard.wolk.com/www.wolk.com.pem --private-key file:///etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key --region $region
certificate_arn=`aws acm list-certificates --region $region --query CertificateSummaryList[*].CertificateArn --output text | cut -f1`
fi

instance_id=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | awk -vORS=, '{print"Id="$1}' | sed 's/,/\ /g'`
target_group_arn=`aws elbv2 describe-target-groups --region $region --names wolk-trgt-grp-$region-$port --query TargetGroups[*].TargetGroupArn --output text`
aws elbv2 register-targets --region $region --target-group-arn $target_group_arn --targets $instance_id --region $region

# load balancer arn to create listener in the next step
loadbalancer_arn=`aws elbv2 describe-load-balancers --region $region --names wolk-lb-$region-$port --query LoadBalancers[*].LoadBalancerArn --output text`

# modify target group healthcheck path to /healthcheck
echo -e "\nModify target group healthcheck path to /healthcheck..."
aws elbv2 modify-target-group --region $region --target-group-arn $target_group_arn --health-check-path /healthcheck --health-check-port $port

# create HTTP listener for LB with a default rule that forwards requests to the target group
echo -e "\nCreate HTTP listener for LB with a default rule that forwards requests to the target group..."
aws elbv2 create-listener --region $region --load-balancer-arn $loadbalancer_arn --protocol HTTPS --port $port --certificates CertificateArn=$certificate_arn --default-actions Type=forward,TargetGroupArn=$target_group_arn

until aws elbv2 describe-listeners --region $region --load-balancer-arn $loadbalancer_arn  --query Listeners[*].Protocol | grep HTTP &> /dev/null; do
printf 'Creating Load Balancer Listener'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nHTTP Load Balancer Listener created..."

# create HTTPS listener for LB with a default rule that forwards requests to the target group
echo -e "\nCreate HTTPS listener for LB with a default rule that forwards requests to the target group..."
aws elbv2 create-listener --region $region --load-balancer-arn $loadbalancer_arn --protocol HTTPS --port 443 --certificates CertificateArn=$certificate_arn  --default-actions Type=forward,TargetGroupArn=$target_group_arn

echo -e "\nHTTPS Load Balancer Listener created..."

# attach autoscaling group to load balancer target group
auto_scaling_group_name=`aws autoscaling describe-auto-scaling-groups --region $region --query AutoScalingGroups[*].AutoScalingGroupName --output text`
aws autoscaling --region $region attach-load-balancer-target-groups --auto-scaling-group-name $auto_scaling_group_name --target-group-arns $target_group_arn
