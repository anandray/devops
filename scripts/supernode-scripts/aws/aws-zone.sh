#!/bin/bash

#aws ec2 describe-availability-zones --region=$1 | grep ZoneName | cut -d"\"" -f4
#aws ec2 describe-regions | grep RegionName | cut -d "\"" -f4 | awk '{print"aws ec2 describe-availability-zones --region="$1,"| grep ZoneName | cut -d\"\\\"\" -f4"}' > /tmp/aws-zone.sh && sh /tmp/aws-zone.sh

aws ec2 describe-regions | grep RegionName | cut -d "\"" -f4 | awk '{print"echo ##########","\n","echo region - ",$1,"\n","aws ec2 describe-availability-zones --region="$1,"| grep ZoneName | cut -d\"\\\"\" -f4"}' > /tmp/aws-zone.sh && sh /tmp/aws-zone.sh
