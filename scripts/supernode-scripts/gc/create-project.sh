#!/bin/bash

> /root/.config/gcloud/configurations/config_default

region=`gcloud compute zones list | grep -v NAME | awk '{print$2}' | uniq`
region_options=`echo $region`
options=($region_options)
prompt="Select a GC region:"

PS3="$prompt "
select answer in "${options[@]}"; do
    zones=`gcloud compute zones list | grep -v NAME | grep $answer | awk '{print$1}' | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`

        echo 'region="'$answer'"' > ~/gc/supernode-gc-$answer.conf
#       echo 'region=`echo $answer | sed 's/[0-9]//g'`' >> supernode-gc-$answer.conf
        echo 'zones="'$zones'"' >> ~/gc/supernode-gc-$answer.conf
        echo "Enter node #:"
      break 2
done

source ~/gc/supernode-gc-$answer.conf

# enable anand@wolk.com
gcloud config set account anand@wolk.com

# derivative inputs
#suffix=`date | sha256sum | head -c 4 ; echo`
suffix="wlk"
project="$region-wlk"
wolk_billing_account=`gcloud alpha billing accounts list --filter open=true | grep "wolk billing 042019" | awk '{print$1}'`
# core inputs
echo "project=$project"
echo "region=$region"

# Create Project
if ! gcloud projects list | grep $project; then
gcloud projects create $project
fi

# Set default Project/region
gcloud config set project $project
gcloud config set compute/region $region

# Enable Billing
gcloud alpha billing projects link $project --billing-account $wolk_billing_account

# New Service account created after Billing is enabled
gcloud services enable compute.googleapis.com
serviceAccount=`gcloud projects describe $project | grep projectNumber | cut -d "'" -f2 | awk '{print$1"-compute@developer.gserviceaccount.com"}'`
echo $serviceAccount

# Grant permission to new service account
gcloud projects add-iam-policy-binding $project --member='user:anand@wolk.com' --role='roles/owner'
gcloud projects add-iam-policy-binding $project --member='serviceAccount:422048983785-compute@developer.gserviceaccount.com' --role='roles/owner'
gcloud projects add-iam-policy-binding wolk-1307 --member serviceAccount:$serviceAccount --role roles/editor
