#!/bin/bash

# To add instances with autoscaling, add this to nginx.conf and comment out the default include file
#    upstream backips  {
#    include /etc/nginx/conf.d/upstream.conf;
#        }

#instance_list=`gcloud compute instances list | grep  $1 | grep RUNNING | awk '{print"\t","server",$4":80 max_fails=0 fail_timeout=5s;"}'`
#instance_list1=`gcloud compute instances list | grep $1 | grep RUNNING | awk '{print"\t","server",$4":80 max_fails=0 fail_timeout=5s;"}' > /tmp/instance_list`
#current_instance_list=`cat /etc/nginx/conf.d/upstream.conf`

for i in {1..10};
do
gcloud compute instances list | grep $1 | grep RUNNING | awk '{print"\t","server",$4":80 max_fails=0 fail_timeout=5s;"}' > /tmp/instance_list;
if diff /tmp/instance_list /etc/nginx/conf.d/upstream.conf; then
  echo "`date +%m-%d-%Y\|%T` - No new instances..."
else
  echo "`date +%m-%d-%Y\|%T` - There are new instances..."
  gcloud compute instances list | grep $1 | grep RUNNING | awk '{print"\t","server",$4":80 max_fails=0 fail_timeout=5s;"}' > /etc/nginx/conf.d/upstream.conf;
  /sbin/service nginx restart;
fi
sleep 5;
done
