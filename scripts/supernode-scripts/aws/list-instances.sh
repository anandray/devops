regions=( $(aws ec2 describe-regions --query Regions[*].RegionName --output text) )
for Regions in "${regions[@]}"
do
echo ${Regions}
aws ec2 describe-instances --query 'Reservations[*].Instances[*].[State]' --region ${Regions}
done
