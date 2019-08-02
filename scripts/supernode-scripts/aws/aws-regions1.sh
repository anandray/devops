#!/bin/bash

region=`aws ec2 describe-regions | grep RegionName | cut -d "\"" -f4`
region_options=`echo $region`
options=($region_options)
prompt="Pick an option:"

PS3="$prompt "
select answer in "${options[@]}"; do
#    aws ec2 describe-availability-zones --region=$answer | grep ZoneName | cut -d"\"" -f4
    sed -i 's/region="_region"/region="'$answer'"/g' supernode-aws.conf
    sed -i 's/awsregion="_awsregion"/awsregion="'$answer'"/g' supernode-aws.conf
    zones=`aws ec2 describe-availability-zones --region=$answer | grep ZoneName | cut -d"\"" -f4 | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`
    sed -i 's/zones="_zones"/zones="'$zones'"/g' supernode-aws.conf
      break 2
done
