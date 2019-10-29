#!/bin/bash

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

# change default region locally
#echo -e "\nChange default region locally"
#sed -i '/region/d' ~/.aws/config
#echo "region = $region" >> ~/.aws/config

# make sure /root/aws exists
if [ ! -d /root/aws/ ]; then
mkdir -p /root/aws/
fi

# key pair
if aws ec2 describe-key-pairs --region $region --query KeyPairs[*].KeyName | grep -i WolkKeyPair-$region; then
echo -e "\nKeypair named WolkKeyPair-$region already exists..."
else
echo -e "\nCreating key pair WolkKeyPair-$region"
aws ec2 create-key-pair --key-name WolkKeyPair-$region --query 'KeyMaterial' --region=$awsregion --output text > /root/aws/WolkKeyPair-$region.pem
chmod 0400 /root/aws/WolkKeyPair-$region.pem
fi

# security group
if aws ec2 describe-security-groups --region $region --query SecurityGroups[*].GroupName | grep -i wolk-sg-$region; then
echo -e "\nSecurity Group named wolk-sg-$region already exists..."
else
echo -e "\nCreating Security Group wolk-sg-$awsregion"
aws ec2 create-security-group --group-name wolk-sg-$awsregion --region=$awsregion --description "wolk security group $region" &> /dev/null
fi

# list group id to use it in the next step
echo -e "\nGet security group id to use it in the next step"
security_group_id=`aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].GroupId --output text`

echo -e "\nSecurity Group ID: $security_group_id"
# add traffic rules to the above security group
if aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].IpPermissions --output text | grep -i tcp | grep 65535; then
echo -e "\nTraffic rules for TCP  exists"
else
echo -e "\nAdding TCP traffic rules to the above security group - $security_group_id/wolk-sg-$region"
aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 0-65535 --cidr 0.0.0.0/0 --region $region &> /dev/null
fi

if aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].IpPermissions --output text | grep -i udp | grep 65535; then
echo -e "\nTraffic rules for UDP  exists"
else
echo -e "\nAdding UDP traffic rules to the above security group - $security_group_id/wolk-sg-$region"
aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol udp --port 0-65535 --cidr 0.0.0.0/0 --region $region &> /dev/null
fi

if aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].IpPermissions --output text | grep -i icmp; then
echo -e "\nTraffic rules for ICMP  exists"
else
echo -e "\nAdding ICMP traffic rules to the above security group - $security_group_id/wolk-sg-$region"
aws ec2 authorize-security-group-ingress --group-id $security_group_id --region $region --ip-permissions IpProtocol=icmp,FromPort=-1,ToPort=-1,IpRanges='[{CidrIp=0.0.0.0/0}]' &> /dev/null
fi


# launch consensus instance with startup script
echo -e "\nLaunch consensus instance with startup script s3://wolk-startup-scripts/scripts/cloudstore/startup-script-cloudstore-repo-aws-$region.sh"

# copy start up script and git-update script to be used to create consensus instance and launch-configuration
echo -e "\nCopy start up script and git-update script to be used to create consensus instance and launch-configuration"
aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/cloudstore-git-update.sh /root/aws/
aws s3 cp s3://wolk-startup-scripts/scripts/plasma/mapping.json /root/aws/
chmod +x /root/aws/cloudstore-git-update.sh

echo -e "\nUse following image to create $fixedinstance"
imageid=`aws ec2 --region $region describe-images --owners 023878629902 --query 'Images[*].ImageId' --output text`

echo -e "\nImageID: $imageid"

if aws ec2 describe-instances --query 'Reservations[*].Instances[*].[State]' --region $region --output text | grep running; then
echo -e "\nRunning instance/s identified.. Check first before recreating..."
else
aws ec2 run-instances --region $region --image-id $imageid --count 1 --instance-type t2.micro --key-name WolkKeyPair-$region --user-data file:///root/aws/cloudstore-git-update.sh --security-group-ids $security_group_id --block-device-mappings file:///root/aws/mapping.json

echo -e "\nVisit https://$region.console.aws.amazon.com/ec2/v2/home?region=$region#Instances:sort=instanceId and wait until the instance creation is completed.."

echo -e "\nWaiting for Cosensus Instance creation to finish..."

until aws ec2 describe-instances --query 'Reservations[*].Instances[*].[State]' --region $region --output text | grep running > /dev/null; do
printf 'Creating Consensus Instance'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nInstance creation complete..."
fi

# obtain instance-id of the instance created above(consensus instance) to use it to create HOSTNAME TAG
echo -e "\nObtain Consensus Instance ID created above to use it to create HOSTNAME TAG"
consensus_instance_id=`aws ec2 describe-instances --region $region --filters Name=instance-state-name,Values=running --query Reservations[*].Instances[*].[InstanceId] --output text`
echo -e "\nConsensus InstanceID: $consensus_instance_id"

# add hostname tag
if aws ec2 describe-instances --region $region --filters Name=instance-state-name,Values=running --query Reservations[*].Instances[*].Tags --output text | grep $fixedinstance; then
echo -e "\nHostname Tag already exists..."
else
echo -e "\nAdd hostname tag $fixedinstance"
aws ec2 create-tags --resources $consensus_instance_id --tag "Key=Name,Value=$fixedinstance" --region $region
fi

# checking whether new consensus instance can be accessed via ssh yet
public_IP=`aws ec2 describe-instances --region $region --query Reservations[*].Instances[*].PublicIpAddress --output text`

echo -e "\nChecking if new Consensus Instance can be accessed via ssh yet..."

while ! ssh -q $public_IP ls -lt /root/go/src/github.com/wolkdb/cloudstore/wolk.toml &> /dev/null; do
printf 'Checking if new Consensus Instance can be accessed via ssh yet...'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nConsensus Instance is now accessible via SSH..."

# Add ConsensusIdx = $node
AmazonCredentials="/root/.aws/credentials"
ssh -q $public_IP mkdir -p /root/.aws
scp -q /root/aws/aws-credentials $public_IP:/root/.aws/credentials
echo -e "\nAdding ConsensusIdx = $node to wolk.toml"
scp -q $public_IP:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/wolk.toml-aws-template /tmp/wolk.toml
sed -i "s|\_ConsensusIdx|$node|" /tmp/wolk.toml
sed -i "s|\_AmazonRegion|$region|" /tmp/wolk.toml
sed -i "s|\_AmazonCredentials|$AmazonCredentials|" /tmp/wolk.toml

sed -i '/SSL/d' /tmp/wolk.toml
echo -e "\nSSLCertFile = \"/etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt\"
SSLKeyFile = \"/etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key\"" >> /tmp/wolk.toml
scp -q /tmp/wolk.toml $public_IP:/root/go/src/github.com/wolkdb/cloudstore/wolk.toml

echo -e "\nAdding the ssl certificates..."
mkdir -p wildcard.wolk.com
aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/wildcard.wolk.com/www.wolk.com.crt /root/aws/wildcard.wolk.com/
aws s3 cp s3://wolk-startup-scripts/scripts/cloudstore/wildcard.wolk.com/www.wolk.com.key /root/aws/wildcard.wolk.com/
scp -r wildcard.wolk.com $public_IP:/etc/ssl/certs/ 2> /dev/null
ssh -q $public_IP service wolk restart

# launch configuration
echo -e "\nCreate launch configuration"
aws autoscaling create-launch-configuration --region $region --image-id $imageid --instance-type t2.micro --key-name WolkKeyPair-$region --user-data file:///root/aws/cloudstore-git-update.sh --security-group $security_group_id --block-device-mappings file:///root/aws/mapping.json --launch-configuration-name wolk-launch-config-$region

until aws autoscaling describe-launch-configurations --region $region --query LaunchConfigurations[*].LaunchConfigurationName --output text | grep wolk-launch-config-$region > /dev/null; do
printf 'Creating Launch Configuration'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nLaunch Configuration is ready..."

# auto scaling group
echo -e "\nCreate auto scaling group"
availability_zones=`aws ec2 describe-availability-zones --region=$region | grep ZoneName | cut -d"\"" -f4 | awk -vORS=, '{ print $1 }' | sed 's/,/\ /g' | sed 's/$/\n/'`

aws autoscaling create-auto-scaling-group --region $region --auto-scaling-group-name wolk-autoscale-$region --availability-zones $availability_zones --launch-configuration-name wolk-launch-config-$region --max-size 1 --min-size 1

until aws autoscaling describe-auto-scaling-groups --region $region --query AutoScalingGroups[*].Instances[*].[LifecycleState] | grep -i InService > /dev/null; do
printf 'Creating Autoscaling Group'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nAutoscaling Group created..."

# add tags to autoscaling instance
echo -e "\nAdding tags to autoscaling instance.."

#instance_id_1=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | cut -d"\"" -f2 | awk -vORS=, '{ print $1 }' | sed 's/,/\ /g' | sed 's/$/\n/'`
instance_id_1=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' --output text | awk -vORS=, '{ print $1 }' | sed 's/,/\ /g' | sed 's/$/\n/'`

#suffix=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | wc -l`
#suffix=`echo $((5555 + RANDOM % 1000))` #numeric only
suffix=`date | sha256sum | head -c 4 ; echo`

aws ec2 create-tags --resources $instance_id_1 --tag "Key=Name,Value=$autoscaledinstance-$suffix" --region $region

# load balancer
echo -e "\nCreate load balancer"
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
echo -e "\nCreate target group"
vpc_id=`aws ec2 describe-vpcs --region $region --query Vpcs[*].VpcId --output text`
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
echo -e "\nRegister target group"

instance_id=`aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E "\[|\]" | awk -vORS=, '{print"Id="$1}' | sed 's/,/\ /g'`
target_group_arn=`aws elbv2 describe-target-groups --region $region --names wolk-trgt-grp-$region-$port --query TargetGroups[*].TargetGroupArn --output text`
aws elbv2 register-targets --target-group-arn $target_group_arn --targets $instance_id --region $region

# modify target group healthcheck path to /healthcheck
echo -e "\nModify target group healthcheck path to /healthcheck"
aws elbv2 modify-target-group --region $region --target-group-arn $target_group_arn --health-check-path /healthcheck --health-check-port $port

# load balancer arn to create listener in the next step
loadbalancer_arn=`aws elbv2 describe-load-balancers --region $region --names wolk-lb-$region-$port --query LoadBalancers[*].LoadBalancerArn --output text`

# create HTTP listener for LB with a default rule that forwards requests to the target group
echo -e "\nCreate HTTP listener for LB with a default rule that forwards requests to the target group"
aws elbv2 create-listener --region $region --load-balancer-arn $loadbalancer_arn --protocol HTTP --port $port  --default-actions Type=forward,TargetGroupArn=$target_group_arn

until aws elbv2 describe-listeners --region $region --load-balancer-arn $loadbalancer_arn  --query Listeners[*].Protocol | grep HTTP > /dev/null; do
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

# create HTTPS listener for LB with a default rule that forwards requests to the target group
echo -e "\nCreate HTTPS listener for LB with a default rule that forwards requests to the target group"
aws elbv2 create-listener --region $region --load-balancer-arn $loadbalancer_arn --protocol HTTPS --port 443 --certificates CertificateArn=$certificate_arn  --default-actions Type=forward,TargetGroupArn=$target_group_arn

echo -e "\nHTTPS Load Balancer Listener created..."

# attach autoscaling group to load balancer target group
auto_scaling_group_name=`aws autoscaling describe-auto-scaling-groups --region $region --query AutoScalingGroups[*].AutoScalingGroupName --output text`
aws autoscaling attach-load-balancer-target-groups --region $region --auto-scaling-group-name $auto_scaling_group_name --target-group-arns $target_group_arn

# create dynamoDB
echo -e "\nCreate dynamoDB"
if aws dynamodb describe-table --table-name wolkdbMaster --region $region &> /dev/null; then
echo -e "\nDynamoDB already exists...
"
else
aws dynamodb create-table --table-name wolkdbMaster --attribute-definitions AttributeName=chunkID,AttributeType=B --key-schema AttributeName=chunkID,KeyType=HASH --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 --region $region

until aws dynamodb describe-table --table-name wolkdbMaster --region $region | grep -i ACTIVE > /dev/null; do
printf 'Creating DynamoDB'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nDynamoDB created..."
fi

#echo "
#AWS cluster successfully created in $region.. Deleting supernode-aws-$answer.conf...
#"
#rm -rfv supernode-aws-$answer.conf
