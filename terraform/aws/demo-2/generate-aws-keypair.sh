region=$1
  
# key pair
if aws ec2 describe-key-pairs --region $region --query KeyPairs[*].KeyName | grep -i "aws_key_pair.$region"; then
echo -e "\nKeypair named aws_key_pair.$region already exists..."
else
echo -e "\nCreating key pair aws_key_pair.$region"
aws ec2 create-key-pair --key-name aws_key_pair.$region --query 'KeyMaterial' --region=$region --output text > aws_key_pair.$region.pem;
chmod 0400 aws_key_pair.$region.pem
fi
