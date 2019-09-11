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
autoscaledinstance="wolk-$node-$provider-$region"
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
echo "
Change default region locally
"
sed -i '/region/d' ~/.aws/config
echo "region = $region" >> ~/.aws/config

# key pair
if aws ec2 describe-key-pairs --region $region --query KeyPairs[*].KeyName | grep -i WolkKeyPair-$region; then
echo "Keypair named WolkKeyPair-$region already exists...
"
else
echo "
Creating key pair WolkKeyPair-$region
"
aws ec2 create-key-pair --key-name WolkKeyPair-$region --query 'KeyMaterial' --region=$awsregion --output text > /root/aws/WolkKeyPair-$region.pem
fi

# security group
if aws ec2 describe-security-groups --region $region --query SecurityGroups[*].GroupName | grep -i wolk-sg-$region; then
echo "Security Group named wolk-sg-$region already exists...
"
else
echo "
Creating Security Group wolk-sg-$awsregion
"
aws ec2 create-security-group --group-name wolk-sg-$awsregion --region=$awsregion --description "wolk security group $region" &> /dev/null
fi

# list group id to use it in the next step
echo "
Get security group id to use it in the next step
"
security_group_id=`aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].GroupId --output text`

echo "Security Group ID: $security_group_id"

# add traffic rules to the above security group
if aws ec2 describe-security-groups --region $region --group-name wolk-sg-$region --query SecurityGroups[*].IpPermissions --output text; then
echo "
Traffic rules already exists
"
else
echo "
Add traffic rules to the above security group
"
aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 0-65535 --cidr 0.0.0.0/0 --region $region &> /dev/null
aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol udp --port 0-65535 --cidr 0.0.0.0/0 --region $region &> /dev/null
fi
