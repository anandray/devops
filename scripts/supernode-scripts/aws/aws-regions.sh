#!/bin/bash

region=`aws ec2 describe-regions | grep RegionName | cut -d "\"" -f4`
region_options=`echo $region`
options=($region_options)
prompt="Pick an option:"

PS3="$prompt "
select answer in "${options[@]}"; do
#    aws ec2 describe-availability-zones --region=$answer | grep ZoneName | cut -d"\"" -f4
    zones=`aws ec2 describe-availability-zones --region=$answer | grep ZoneName | cut -d"\"" -f4 | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`
	echo 'region="'$answer'"' > supernode-aws-$answer.conf
	echo 'awsregion="'$answer'"' >> supernode-aws-$answer.conf
	echo 'zones="'$zones'"' >> supernode-aws-$answer.conf
      break 2
done
