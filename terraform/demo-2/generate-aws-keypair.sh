region=$1
aws ec2 create-key-pair --key-name aws_key_pair.$region --query 'KeyMaterial' --region=$region --output text > aws_key_pair.mykey.$region.pem;
chmod 0400 aws_key_pair.mykey.$region.pem
