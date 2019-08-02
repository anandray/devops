#!/bin/bash

> /root/.config/gcloud/configurations/config_default

#node=`mysql --defaults-extra-file=~/.mysql wolk -e "select node, projectID, region from project3;" | grep -v node | awk '{print$1}'`
#project=`mysql --defaults-extra-file=~/.mysql wolk -e "select node, projectID, region from project3;" | grep -v node | awk '{print$2}'`
#region=`mysql --defaults-extra-file=~/.mysql wolk -e "select node, projectID, region from project3;" | grep -v node | awk '{print$3}'`

project=$1
region=$2
node=$3
zones=`gcloud compute zones list | grep -v NAME | grep $region | awk '{print$1}' | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`
echo 'project="'$project'"' > /root/gc/supernode-gc-$region.conf
echo 'region="'$region'"' >> /root/gc/supernode-gc-$region.conf
echo 'zones="'$zones'"' >> /root/gc/supernode-gc-$region.conf
echo 'node="'$node'"' >> /root/gc/supernode-gc-$region.conf

source /root/gc/supernode-gc-$region.conf

# enable anand@wolk.com
gcloud config set account anand@wolk.com

# derivative inputs
#suffix=`date | sha256sum | head -c 4 ; echo`
app="cloudstore"
port=80
https_port=443
provider="gc"
fixedinstance="wolk-$node-$provider-$region-datastore"
instance_group="$fixedinstance"
prefix="$app-$region-$provider"
instancetemplate="$prefix"
instancetemplate_image="$prefix-image"
urlmap="$prefix"
lbname="$prefix-https"
https_lbname="$prefix-https"
globalip="$app-$region-$provider-global-ip"
targetproxy="$prefix-target-proxy-$https_port"
regionalipname="$app-$region-$provider-regional-ip"
healthcheck="$app-$region-healthcheck"
portname="$app-$port"
https_portname="$app-$https_port"
wolk_billing_account=`gcloud alpha billing accounts list --filter open=true | grep "wolk billing 042019" | awk '{print$1}'`
image_name="wolk-$region-image"
primaryzone=$(gcloud compute zones list | grep $region | awk '{print$1}' | head -n1)
# core inputs
echo "project=$project"
echo "region=$region"
echo "zones=$zones"
echo "node=$node"
echo "port=$port"
echo "https_port=$https_port"
echo "https_portname=$https_portname"
echo "fixedinstance=$fixedinstance"
#fixedinstance list
#gcloud compute instances list --filter="name=( '$fixedinstance' )" --project=$project

# Set default Project/region
gcloud config set project $project
gcloud config set compute/region $region

serviceAccount=`gcloud projects describe $project | grep projectNumber | cut -d "'" -f2 | awk '{print$1"-compute@developer.gserviceaccount.com"}'`
keyID=`gcloud -q iam service-accounts keys list --iam-account $serviceAccount | grep -v KEY_ID | head -1 | awk '{print$1}'`
gcloud -q iam service-accounts keys delete $keyID --iam-account $serviceAccount


# Delete consensus instance
zone=`gcloud compute instances --verbosity none  list --project=$project --filter="name=( '$fixedinstance' )" | grep $fixedinstance | awk '{print$2}'`
gcloud -q compute --project=$project instances delete $fixedinstance --zone $zone

# Delete Static IP for Consensus Instance:
regional_ip=`gcloud beta compute --project=$project addresses list | grep $regionalipname | awk '{print$2}'`
gcloud -q beta compute --project=$project addresses delete $regionalipname --region $region

#### DELETE LOAD BALANCER ####
# 1. Delete GLOBAL forwarding-rules
gcloud -q compute --project=$project forwarding-rules delete $https_lbname --global

# 2. Delete TARGET PROXY "cloudstore-sa-gc"/"$app-$region-$provider" USING THE ABOVE URL MAP
gcloud -q compute --project=$project target-https-proxies delete $targetproxy

# 3. Delete URL MAP
gcloud -q compute url-maps delete $urlmap --project $project

# 1a. gcloud compute backend-services delete
gcloud -q compute --project $project backend-services delete $https_lbname --global

# 4. Delete Static IP
gcloud -q beta compute --project=$project addresses delete $globalip --global

# 6 Delete Firewall
gcloud -q compute firewall-rules delete allow-all --project=$project

# Delete ssl certificate
ssl_certificates=`gcloud compute ssl-certificates list --project=$project | tail -n1 | cut -d " " -f1`
gcloud -q compute ssl-certificates delete wolk-ssl-cert --project=$project

## gcloud compute http-healsth-checks delete
gcloud -q compute --project "$project" http-health-checks delete "$healthcheck"
gcloud -q compute --project "$project" https-health-checks delete "$healthcheck"

#Delete Managed Instance Group
gcloud -q beta compute --project=$project instance-groups managed delete $instance_group --region $region

# delete template
gcloud -q compute --project=$project instance-templates delete $region-template

# Delete custom image
if gcloud compute images list --project=$project | grep wolk; then
gcloud -q compute images delete wolk-$region-image --project=$project
fi
