#!/bin/bash

> /root/.config/gcloud/configurations/config_default

region=`gcloud compute zones list | grep -v NAME | awk '{print$2}' | uniq`
region_options=`echo $region`
options=($region_options)
prompt="Select a GC region:"

PS3="$prompt "
select answer in "${options[@]}"; do
    zones=`gcloud compute zones list | grep -v NAME | grep $answer | awk '{print$1}' | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/'`

        echo 'region="'$answer'"' > /root/gc/supernode-gc-$answer.conf
#       echo 'region=`echo $answer | sed 's/[0-9]//g'`' >> /root/gc/supernode-gc-$answer.conf
        echo 'zones="'$zones'"' >> /root/gc/supernode-gc-$answer.conf
        echo "Enter node #:"
        read node
            echo 'node="'$node'"' >> /root/gc/supernode-gc-$answer.conf
      break 2
done

source /root/gc/supernode-gc-$answer.conf

# enable anand@wolk.com
gcloud config set account anand@wolk.com

# Enable API
#'yes' | gcloud services list --available &> /dev/null
#'yes' | gcloud services enable cloudbilling.googleapis.com
#'yes' | gcloud services enable cloudresourcemanager.googleapis.com

# derivative inputs
#suffix=`date | sha256sum | head -c 4 ; echo`
suffix="wlk"
app="cloudstore"
port=80
https_port=443
project="$region-$suffix"
#project=wolk-asia-south1-5bda
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

#read -rsp $'
#Press any key to continue...\n
#' -n1 key

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

# Create google credential key
"yes" | gcloud iam service-accounts keys create /tmp/google.json-$project --iam-account $serviceAccount
gsutil cp /tmp/google.json-$project gs://wolk-scripts/scripts/cloudstore/google-credentials/google.json-$project

# Create Static IP for Consensus Instance:
gcloud beta compute --project=$project addresses create $regionalipname --region=$region --network-tier=PREMIUM
regional_ip=`gcloud beta compute --project=$project addresses list | grep $regionalipname | awk '{print$2}'`

# creating fixed instance
gcloud compute --project=$project instances create $fixedinstance --zone=$primaryzone --machine-type=custom-2-4096 --address $regional_ip --subnet=default --network-tier=PREMIUM --metadata=startup-script-url=gs://wolk-scripts/scripts/cloudstore/startup-script-cloudstore-repo-gcp.sh,ssh-keys=root:ssh-rsa\ AAAAB3NzaC1yc2EAAAABIwAAAQEA0\+dWJfxKJKozhrHQ8Zn06CsIXg3nav5tBi5ojZUKjDrOI78P0BPwaIod48fj8er8Z/spyW/pQ5Ys/TAL739TWPMtTQwfsWvsD1B5chOVVWrb5BomcEqWzcg/u6vCUqYFfP8q2p5p46w5U41nO\+S1mO\+NjdWsNn4f2Gqg8xIXZs/BDb02\+AdBZ/DTwj12HAJHoBpUF7OBLVibJwDGX4xj1BQuYtg7\+iCeaW2aW0NDCifU5bPOCZGQ4gAWG7HLGoqEE\+EGvousqXIJ\+K58Ex/G\+21qgaMqLr4QL\+ZBkHGEZ6q72/ziz7Nz9Na3XpanUDYcdfC1ppXlydtxV8yXQgSfNQ==\ root$'\n'Tapas:ssh-rsa\ AAAAB3NzaC1yc2EAAAABJQAAAQEAoOsmmxsNw4dvzOKUks3F1wDhEL43L\+7hCGhkOhX1gI04pFiaktMXoyEklN9ue8ESDWtTTL2D9mwBvlysexTG7FhjAVXfOnWByjz5mXoEzOIrM4bcrKuhTHCUujTXFDnkVJaKo5czFsobs/F7TNBMsIq788x7GfI1ntwc27kCFvAjsVSkwoF/VTb4/O7SyyDVbyf6F6jvOjRZKO/6JVzqJyMN5S740xAfp42I/cqUdM8t\+MUpQh/1EgzBlUiG4WFEDrcbmL/TWuRTa91BcHc57ym7ddTsljzS0I5iU5y1t2GCE3dhrRj\+KysxMgc3GmH3saEAymYt8PAPDNSdHkqgiQ==\ Tapas$'\n'anand:ssh-rsa\ AAAAB3NzaC1yc2EAAAADAQABAAABAQDofwXiKi9VAXlXETZMX16aSu3xZRuwVBB8Jgd/JIT3wG0yfCSymSvSS1Cc8nMm3hwywL5IxwCiIDXxrHsIsBeTuyQWSMZRcNcbEOC4fQxyRNBi1Arqb3t5OPIEZD1Y57C42vV0Hpk08zfysveuww7vJD69inCyGhE8IB4CD6Hn0N2aDQjMp3wtvao6c9aZh9OqtCpyHX0W8EC27RiTOf\+kFy2XxGQ544nJo08g//6bwbZVTS5/Ta4OOjPu92pu40BwnQQyOpPz0FErEGujInFeHhSp3mF3/MFavn5PO8Ne8sZ3CAHg\+NIrQO7i94UYTpnM22E/Xjr3FhN7/qmjHAM/\ anand --maintenance-policy=MIGRATE --service-account=$serviceAccount --scopes=https://www.googleapis.com/auth/cloud-platform --tags=allowall,http-server,https-server --image=centos-7-v20190423 --image-project=centos-cloud --boot-disk-size=20GB --boot-disk-type=pd-standard --boot-disk-device-name=$fixedinstance

# wait until consensus instance is created
until gcloud compute instances list --project=$project | grep RUNNING > /dev/null; do
printf 'Creating Consensus Instance'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nConsensus Instance created..."

# waiting for startup script to finish
while ! ssh -q $regional_ip ls -lt /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt &> /dev/null; do
printf 'waiting for startup script to finish... "'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nConsensus Instance is now READY... "

#preparing wolk.toml + google.json credential
scp -q $regional_ip:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/wolk.toml-gc-template /tmp/wolk.toml
sed -i "s/_ConsensusIdx/$node/g" /tmp/wolk.toml
sed -i "s/_NodeType/consensus/g" /tmp/wolk.toml
sed -i "s/_Region/$region/g" /tmp/wolk.toml
sed -i "s/_GoogleDatastoreProject/$project/g" /tmp/wolk.toml
scp -q /tmp/wolk.toml $regional_ip:/root/go/src/github.com/wolkdb/cloudstore/wolk.toml
scp /tmp/google.json-$project $regional_ip:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json

# copying ~/.config/gcloud/configurations/config_default
ssh $regional_ip mkdir -p ~/.config/gcloud/configurations
scp ~/.config/gcloud/configurations/config_default $regional_ip:~/.config/gcloud/configurations/config_default

# stop fixedinstance to create image
zone=$(gcloud compute instances list --project=$project --filter="name=( '$fixedinstance' )" | grep $fixedinstance | awk '{print$2}')
gcloud -q compute instances stop $fixedinstance --zone=$zone --project=$project

# wait until instance is stopped
until ! gcloud compute instances list --project=$project | grep RUNNING > /dev/null; do
printf 'Stopping Consensus Instance'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nConsensus Instance stopped for image creation"

# check if a custom image already exists or not. if it does, deprecate it
if gcloud compute images list --project=$project | grep wolk; then
gcloud -q compute images delete wolk-$region-image --project=$project
fi

#create image
gcloud compute images create wolk-$region-image --project=$project --source-disk=$fixedinstance --source-disk-zone=$zone

until gcloud compute images list --project=$project | grep wolk | grep READY > /dev/null; do
printf 'Creating image'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nImage created"

# start consensus instance back up
gcloud -q compute instances start $fixedinstance --zone=$zone --project=$project

until gcloud compute instances list --project=$project | grep RUNNING > /dev/null; do
printf 'Starting Consensus Instance'
for ((i = 0; i < 5; ++i)); do
    for ((j = 0; j < 4; ++j)); do
        printf .
        sleep 1
    done

    printf '\b\b\b\b    \b\b\b\b'
done
printf '....done\n'
done
echo -e "\nConsensus Instance restarted after image creation"

# make wolk on consensus image
if ssh $regional_ip ls /root/scripts/make-wolk.sh; then
ssh $regional_ip chmod +x /root/scripts/make-wolk.sh;
ssh $regional_ip sh /root/scripts/make-wolk.sh &> /dev/null &
fi

# create template with ssh keys and startup script
gcloud compute --project=$project instance-templates create $region-template --machine-type=custom-2-4096 --network=projects/$project/global/networks/default --network-tier=PREMIUM --metadata=startup-script-url=gs://wolk-scripts/scripts/cloudstore/cloudstore-git-update-gcp-storage-new.sh,ssh-keys=root:ssh-rsa\ AAAAB3NzaC1yc2EAAAABIwAAAQEA0\+dWJfxKJKozhrHQ8Zn06CsIXg3nav5tBi5ojZUKjDrOI78P0BPwaIod48fj8er8Z/spyW/pQ5Ys/TAL739TWPMtTQwfsWvsD1B5chOVVWrb5BomcEqWzcg/u6vCUqYFfP8q2p5p46w5U41nO\+S1mO\+NjdWsNn4f2Gqg8xIXZs/BDb02\+AdBZ/DTwj12HAJHoBpUF7OBLVibJwDGX4xj1BQuYtg7\+iCeaW2aW0NDCifU5bPOCZGQ4gAWG7HLGoqEE\+EGvousqXIJ\+K58Ex/G\+21qgaMqLr4QL\+ZBkHGEZ6q72/ziz7Nz9Na3XpanUDYcdfC1ppXlydtxV8yXQgSfNQ==\ root$'\n'anand:ssh-rsa\ AAAAB3NzaC1yc2EAAAADAQABAAABAQDofwXiKi9VAXlXETZMX16aSu3xZRuwVBB8Jgd/JIT3wG0yfCSymSvSS1Cc8nMm3hwywL5IxwCiIDXxrHsIsBeTuyQWSMZRcNcbEOC4fQxyRNBi1Arqb3t5OPIEZD1Y57C42vV0Hpk08zfysveuww7vJD69inCyGhE8IB4CD6Hn0N2aDQjMp3wtvao6c9aZh9OqtCpyHX0W8EC27RiTOf\+kFy2XxGQ544nJo08g//6bwbZVTS5/Ta4OOjPu92pu40BwnQQyOpPz0FErEGujInFeHhSp3mF3/MFavn5PO8Ne8sZ3CAHg\+NIrQO7i94UYTpnM22E/Xjr3FhN7/qmjHAM/\ anand$'\n'tapas:ssh-rsa\ AAAAB3NzaC1yc2EAAAABJQAAAQEAwXK1uMe3EQBW/uaQP1EdQtMW65ELJFtQKZ7314fmB43IhAXmm8opWI6uylF4ELRz2fhZX/s2Pu/XIe3NXmzRAg9w5gNbMQWl8ezNOOxpNPsImbYiVGuQJ2ldFEp\+d80eToGrcAciTv1jP2YW91P9MJwZ5LdM0AvQuwBn89xYVCn5ASTTzXzuUwxh\+zAjZKbYYF6iCMVysm5PUYBbWcAS69X9E4LAZNffPFDiS\+hEwLQtpmCXYrVQ\+EnCUQihldhhDe4xfxHi8JHHLyzH80glE/HLer3ygmrs/FOTPKuOQZ9jhqT6mqWjRY8JSLhwyft9OzNdAOenRBEubDzRAnNqaw==\ tapas --maintenance-policy=MIGRATE --service-account=$serviceAccount --scopes=https://www.googleapis.com/auth/cloud-platform --tags=allowall,http-server,https-server --image=wolk-$region-image --image-project=$project --boot-disk-size=20GB --boot-disk-type=pd-standard --boot-disk-device-name=$region-template --project=$project


#Create Managed Instance Group
## gcloud compute instance-groups managed create
gcloud -q beta compute --project=$project instance-groups managed create $instance_group --base-instance-name=$instance_group --template=$region-template --size=1 --zones=$zones --initial-delay=300

## gcloud compute http-healsth-checks create
gcloud compute --project "$project" http-health-checks create "$healthcheck" --port "$port" --request-path "/healthcheck" --check-interval "5" --timeout "5" --unhealthy-threshold "2" --healthy-threshold "2"
gcloud compute --project "$project" https-health-checks create "$healthcheck" --port "$https_port" --request-path "/healthcheck" --check-interval "5" --timeout "5" --unhealthy-threshold "2" --healthy-threshold "2"

## Add/Associate named-ports with compute instance group (required by backend)
gcloud compute instance-groups set-named-ports --project $project $fixedinstance --named-ports "" --region $region
gcloud compute instance-groups set-named-ports --project $project --region=$region --named-ports=$portname:$port $instance_group
gcloud compute instance-groups set-named-ports --project $project --region=$region --named-ports=$https_portname:$https_port $instance_group

#### LOAD BALANCER ####
# add ssl certificate
gcloud compute ssl-certificates create wolk-ssl-cert --certificate=/etc/ssl/certs/wildcard/www.wolk.com.crt --private-key=/etc/ssl/certs/wildcard/www.wolk.com.key --project=$project
ssl_certificates=`gcloud compute ssl-certificates list --project=$project | tail -n1 | cut -d " " -f1`

# 1a. gcloud compute backend-services create
gcloud compute --project $project backend-services create $https_lbname --global --https-health-checks $healthcheck --load-balancing-scheme=EXTERNAL --port-name=$https_portname --protocol=HTTPS

# 1b. gcloud compute backend-services add-backend
gcloud compute backend-services add-backend $https_lbname --instance-group=$instance_group --instance-group-region=$region --balancing-mode=UTILIZATION --global --max-utilization=0.8 --max-rate-per-instance=1000 --project $project

# 2. Add URL MAP
gcloud compute url-maps create $urlmap --default-service $https_lbname --description "Backend Service for LB" --project $project

# 3. CREATE TARGET PROXY "cloudstore-sa-gc"/"$app-$region-$provider" USING THE ABOVE URL MAP
gcloud compute --project=$project target-https-proxies create $targetproxy --url-map=$urlmap --ssl-certificates=$ssl_certificates

# 4. Create Static IP
gcloud beta compute --project=$project addresses create $globalip --global --network-tier=PREMIUM

# 5. Create GLOBAL forwarding-rules
gcloud compute --project=$project forwarding-rules create $https_lbname --global --address=$globalip --ip-protocol=TCP --ports=$https_port --target-https-proxy=$targetproxy

# 6 Firewall
gcloud compute firewall-rules create allow-all --allow tcp,udp --source-ranges=0.0.0.0/0 --target-tags=allowall --project=$project

# starting wolk on consensus instance
ssh $regional_ip service wolk restart
