regions=( $(aws ec2 describe-regions --query Regions[*].RegionName --output text) )
for Regions in "${regions[@]}"
do
echo ${Regions}
aws ec2 describe-key-pairs --query 'KeyPairs[*]' --region ${Regions}
done
