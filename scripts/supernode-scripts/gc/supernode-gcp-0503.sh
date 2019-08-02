#!/bin/bash

gcregion=`gcloud compute zones list | grep -v NAME | awk '{print$2}' | uniq`
region_options=`echo $gcregion`
options=($region_options)
prompt="Select a GC region:"

PS3="$prompt "
select answer in "${options[@]}"; do
    zones=`gcloud compute zones list | grep -v NAME | grep $answer | awk '{print$1}' | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`

	echo 'gcregion="'$answer'"' > supernode-gc-$answer.conf
	echo 'region=`echo $answer | sed 's/[0-9]//g'`' >> supernode-gc-$answer.conf
	echo 'zones="'$zones'"' >> supernode-gc-$answer.conf
	echo "Enter node #:"
	read node
	    echo 'node="'$node'"' >> supernode-gc-$answer.conf
      break 2
done

source supernode-gc-$answer.conf

# derivative inputs
suffix=`date | sha256sum | head -c 4 ; echo`
app="cloudstore"
port=80
https_port=443
project="wolk-$region-$suffix"
provider="gc"
fixedinstance="wolk-$node-$provider-$region-datastore"
prefix="$app-$region-$provider"
instancetemplate="$prefix"
instancetemplate_image="$prefix-image"
urlmap="$prefix"
lbname="$prefix-https"
https_lbname="$prefix-https"
globalip="$app-$region-$provider-global-ip"
targetproxy="$prefix-target-proxy-$port"
regionalipname="$app-$region-$provider-regional-ip"
healthcheck="$app-$region-healthcheck"
portname="$app-$port"
https_portname="$app-$https_port"
wolk_billing_account="018164-166386-DAEEED"
image_name="wolk-$gcregion-image"

# core inputs
echo "project=$project"
echo "region=$region"
echo "gcregion=$gcregion"
echo "zones=$zones"
echo "node=$node"
echo "port=$port"
echo "https_port=$https_port"
echo "https_portname=$https_portname"
echo "fixedinstance=$fixedinstance"

read -rsp $'
Press any key to continue...\n
' -n1 key

# Create Project
gcloud projects create $project

# Set default Project
gcloud config set project $project

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

# Create google credential key
"yes" | gcloud iam service-accounts keys create /tmp/google.json-$project --iam-account $serviceAccount

# TODO: streamline SSHKEYs (removed pvt key from ssh_keys.tgz used in startup script)
# TODO: get startup script from github "installer" and put into bucket

## gcloud compute instance-groups managed set-instance-template
gcloud compute --project=$project instance-templates create $instancetemplate_image --machine-type=custom-1-1024 --network=projects/wolk-us-east/global/networks/default --network-tier=PREMIUM --metadata=startup-script-url=gs://wolk-scripts/scripts/cloudstore/cloudstore-git-update-gcp.sh,ssh-keys=root:ssh-rsa\ AAAAB3NzaC1yc2EAAAABIwAAAQEA0\+dWJfxKJKozhrHQ8Zn06CsIXg3nav5tBi5ojZUKjDrOI78P0BPwaIod48fj8er8Z/spyW/pQ5Ys/TAL739TWPMtTQwfsWvsD1B5chOVVWrb5BomcEqWzcg/u6vCUqYFfP8q2p5p46w5U41nO\+S1mO\+NjdWsNn4f2Gqg8xIXZs/BDb02\+AdBZ/DTwj12HAJHoBpUF7OBLVibJwDGX4xj1BQuYtg7\+iCeaW2aW0NDCifU5bPOCZGQ4gAWG7HLGoqEE\+EGvousqXIJ\+K58Ex/G\+21qgaMqLr4QL\+ZBkHGEZ6q72/ziz7Nz9Na3XpanUDYcdfC1ppXlydtxV8yXQgSfNQ==\ root@www6002$'\n'anand:ssh-rsa\ AAAAB3NzaC1yc2EAAAADAQABAAABAQDofwXiKi9VAXlXETZMX16aSu3xZRuwVBB8Jgd/JIT3wG0yfCSymSvSS1Cc8nMm3hwywL5IxwCiIDXxrHsIsBeTuyQWSMZRcNcbEOC4fQxyRNBi1Arqb3t5OPIEZD1Y57C42vV0Hpk08zfysveuww7vJD69inCyGhE8IB4CD6Hn0N2aDQjMp3wtvao6c9aZh9OqtCpyHX0W8EC27RiTOf\+kFy2XxGQ544nJo08g//6bwbZVTS5/Ta4OOjPu92pu40BwnQQyOpPz0FErEGujInFeHhSp3mF3/MFavn5PO8Ne8sZ3CAHg\+NIrQO7i94UYTpnM22E/Xjr3FhN7/qmjHAM/\ anand@MdotMs-MacBook-Pro.local --maintenance-policy=MIGRATE --service-account=222646948547-compute@developer.gserviceaccount.com --scopes=https://www.googleapis.com/auth/cloud-platform --tags=allowall,http-server,https-server --image=$image_name --image-project=$project --boot-disk-size=20GB --boot-disk-type=pd-standard --boot-disk-device-name=$app-$region-$provider

## Create FIXED NODE:
gcloud beta compute --project=$project addresses create $regionalipname --region=$gcregion --network-tier=PREMIUM
regional_ip=`gcloud beta compute --project=$project addresses list | grep $regionalipname | awk '{print$2}'`

## Create FIXED COMPUTE INSTANCE

gcloud compute --project=$project instances create $fixedinstance --zone=$primaryzone --machine-type=custom-1-1024 --address $regional_ip --subnet=default --network-tier=PREMIUM --metadata=startup-script-url=gs://wolk-scripts/scripts/cloudstore/cloudstore-git-update-gcp.sh,ssh-keys=root:ssh-rsa\ AAAAB3NzaC1yc2EAAAABIwAAAQEA0\+dWJfxKJKozhrHQ8Zn06CsIXg3nav5tBi5ojZUKjDrOI78P0BPwaIod48fj8er8Z/spyW/pQ5Ys/TAL739TWPMtTQwfsWvsD1B5chOVVWrb5BomcEqWzcg/u6vCUqYFfP8q2p5p46w5U41nO\+S1mO\+NjdWsNn4f2Gqg8xIXZs/BDb02\+AdBZ/DTwj12HAJHoBpUF7OBLVibJwDGX4xj1BQuYtg7\+iCeaW2aW0NDCifU5bPOCZGQ4gAWG7HLGoqEE\+EGvousqXIJ\+K58Ex/G\+21qgaMqLr4QL\+ZBkHGEZ6q72/ziz7Nz9Na3XpanUDYcdfC1ppXlydtxV8yXQgSfNQ==\ root@www6002$'\n'anand:ssh-rsa\ AAAAB3NzaC1yc2EAAAADAQABAAABAQDofwXiKi9VAXlXETZMX16aSu3xZRuwVBB8Jgd/JIT3wG0yfCSymSvSS1Cc8nMm3hwywL5IxwCiIDXxrHsIsBeTuyQWSMZRcNcbEOC4fQxyRNBi1Arqb3t5OPIEZD1Y57C42vV0Hpk08zfysveuww7vJD69inCyGhE8IB4CD6Hn0N2aDQjMp3wtvao6c9aZh9OqtCpyHX0W8EC27RiTOf\+kFy2XxGQ544nJo08g//6bwbZVTS5/Ta4OOjPu92pu40BwnQQyOpPz0FErEGujInFeHhSp3mF3/MFavn5PO8Ne8sZ3CAHg\+NIrQO7i94UYTpnM22E/Xjr3FhN7/qmjHAM/\ anand@MdotMs-MacBook-Pro.local --maintenance-policy=MIGRATE --service-account=$serviceAccount --scopes=https://www.googleapis.com/auth/cloud-platform --tags=allowall,http-server,https-server --image=$image_name --image-project=$project --boot-disk-size=20GB --boot-disk-type=pd-standard --boot-disk-device-name=$fixedinstance

## gcloud compute instance-groups managed create
gcloud -q beta compute --project=$project instance-groups managed create $fixedinstance --base-instance-name=$fixedinstance --template=$instancetemplate_image --size=1 --zones=$zones --initial-delay=300

## gcloud compute instance-groups managed set-autoscaling
#gcloud compute --project "$project" instance-groups managed set-autoscaling "$fixedinstance" --region "$gcregion" --cool-down-period "60" --max-num-replicas "1" --min-num-replicas "1" --target-cpu-utilization "0.6"

# resize instance (with Autoscaling OFF)
gcloud beta compute instance-groups managed resize $fixedinstance --size=3 --region=$gcregion --project $project

## gcloud compute http-health-checks create
gcloud compute --project "$project" http-health-checks create "$healthcheck" --port "$port" --request-path "/healthcheck" --check-interval "5" --timeout "5" --unhealthy-threshold "2" --healthy-threshold "2"
gcloud compute --project "$project" https-health-checks create "$healthcheck" --port "$https_port" --request-path "/healthcheck" --check-interval "5" --timeout "5" --unhealthy-threshold "2" --healthy-threshold "2"

## Add/Associate named-ports with compute instance group (required by backend)
gcloud compute instance-groups set-named-ports $fixedinstance --named-ports "" --region $gcregion
gcloud compute instance-groups set-named-ports --region=$gcregion --named-ports=$portname:$port $fixedinstance
gcloud compute instance-groups set-named-ports --region=$gcregion --named-ports=$https_portname:$https_port $fixedinstance

#### LOAD BALANCER ####
# 1a. gcloud compute backend-services create
gcloud compute --project $project backend-services create $https_lbname --global --http-health-checks $healthcheck --load-balancing-scheme=EXTERNAL --port-name=$https_portname --protocol=HTTPS
# 1b. gcloud compute backend-services add-backend
gcloud compute backend-services add-backend $https_lbname --instance-group=$fixedinstance --instance-group-region=$gcregion --balancing-mode=UTILIZATION --global --max-utilization=0.8 --max-rate-per-instance=1000

# 2. Add URL MAP
gcloud compute url-maps create $urlmap --default-service $https_lbname --description "Backend Service for LB"

# 3. CREATE TARGET PROXY "cloudstore-sa-gc"/"$app-$region-$provider" USING THE ABOVE URL MAP
gcloud compute --project=$project target-http-proxies create $targetproxy --url-map=$urlmap

# 4. Create Static IP
gcloud beta compute --project=$project addresses create $globalip --global --network-tier=PREMIUM

# 5. Create GLOBAL forwarding-rules
gcloud compute --project=$project forwarding-rules create $https_lbname --global --address=$globalip --ip-protocol=TCP --ports=$https_port --target-http-proxy=$targetproxy

# 7 Firewall
gcloud compute firewall-rules create allow-all --allow tcp,udp --source-ranges=0.0.0.0/0 --target-tags=allowall

# TODO: get datastore credentials setup with right region, get table setup
# TODO: print out the LB IP + consensus IP so that the user can submit a registerNode transaction
# TODO: run wb test and keep trying every minute until it passes
# /root/go/src/github.com/wolkdb/cloudstore/wolk/tests/wb/wb -n=1000 -c=100 -server $lbip -run=share
