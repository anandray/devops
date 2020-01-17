region=$1

# security group
if aws ec2 describe-security-groups --region $region --query SecurityGroups[*].GroupName | grep -i aws-security-group-$region; then
echo -e "\nSecurity Group named aws-security-group-$region already exists..."
else
echo -e "\nCreating Security Group aws-security-group-$region"
aws ec2 create-security-group --group-name aws-security-group-$region --region=$region --description "AWS security group $region" &> /dev/null
fi

# allow ingress traffic
# list group id to use it in the next step
echo -e "\nGet security group id to use it in the next step"
security_group_id=`aws ec2 describe-security-groups --region $region --group-name aws-security-group-$region --query SecurityGroups[*].GroupId --output text`

echo -e "\nSecurity Group ID: $security_group_id"
# add traffic rules to the above security group
if aws ec2 describe-security-groups --region $region --group-name aws-security-group-$region --query SecurityGroups[*].IpPermissions --output text | grep -i tcp | grep 65535; then
echo -e "\nTraffic rules for TCP  exists"
else
echo -e "\nAdding TCP traffic rules to the above security group - $security_group_id/aws-security-group-$region"
aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 0-65535 --cidr 67.180.100.64/32 --region $region &> /dev/null
fi
