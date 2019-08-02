# Create project
gcloud projects create wolk-sa

# Set default project
## ~/.config/gcloud/configurations/config_default
gcloud config set project wolk-sa 

## Unset default PROJECT
# gcloud config unset project

## Enable Billing
# gcloud alpha billing accounts list
gcloud alpha billing projects link wolk-sa --billing-account 018164-166386-DAEEED

# gcloud iam service-accounts list
#gcloud iam service-accounts create
#gcloud projects add-iam-policy-binding crosschannel-1307 --member serviceAccount:cloudstore-wolk-sa@wolk-sa.iam.gserviceaccount.com --role roles/owner
#gcloud projects add-iam-policy-binding crosschannel-1307 --member serviceAccount:649137252491-compute@developer.gserviceaccount.com --role roles/owner

# New Service account created after Billing is enabled
serviceAccount=`gcloud projects describe wolk-t33sstt | grep projectNumber | cut -d "'" -f2 | awk '{print$1"-compute@developer.gserviceaccount.com"}'`
echo $serviceAccount

# Grant permission to new service account
gcloud projects add-iam-policy-binding crosschannel-1307 --member serviceAccount:$serviceAccount --role roles/owner

## gcloud compute instance-groups managed set-instance-template
gcloud beta compute --project=wolk-sa instance-templates create cloudstore-sa-gc --machine-type=custom-1-1024 --network=projects/wolk-sa/global/networks/default --network-tier=PREMIUM --metadata=startup-script-url=gs://startup_scripts_us/scripts/cloudstore/startup-script-cloudstore-repo.sh,ssh-keys=root:ssh-rsa\ AAAAB3NzaC1yc2EAAAABIwAAAQEA0\+dWJfxKJKozhrHQ8Zn06CsIXg3nav5tBi5ojZUKjDrOI78P0BPwaIod48fj8er8Z/spyW/pQ5Ys/TAL739TWPMtTQwfsWvsD1B5chOVVWrb5BomcEqWzcg/u6vCUqYFfP8q2p5p46w5U41nO\+S1mO\+NjdWsNn4f2Gqg8xIXZs/BDb02\+AdBZ/DTwj12HAJHoBpUF7OBLVibJwDGX4xj1BQuYtg7\+iCeaW2aW0NDCifU5bPOCZGQ4gAWG7HLGoqEE\+EGvousqXIJ\+K58Ex/G\+21qgaMqLr4QL\+ZBkHGEZ6q72/ziz7Nz9Na3XpanUDYcdfC1ppXlydtxV8yXQgSfNQ==\ root@www6002$'\n'anand:ssh-rsa\ AAAAB3NzaC1yc2EAAAADAQABAAABAQDofwXiKi9VAXlXETZMX16aSu3xZRuwVBB8Jgd/JIT3wG0yfCSymSvSS1Cc8nMm3hwywL5IxwCiIDXxrHsIsBeTuyQWSMZRcNcbEOC4fQxyRNBi1Arqb3t5OPIEZD1Y57C42vV0Hpk08zfysveuww7vJD69inCyGhE8IB4CD6Hn0N2aDQjMp3wtvao6c9aZh9OqtCpyHX0W8EC27RiTOf\+kFy2XxGQ544nJo08g//6bwbZVTS5/Ta4OOjPu92pu40BwnQQyOpPz0FErEGujInFeHhSp3mF3/MFavn5PO8Ne8sZ3CAHg\+NIrQO7i94UYTpnM22E/Xjr3FhN7/qmjHAM/\ anand@MdotMs-MacBook-Pro.local --maintenance-policy=MIGRATE --service-account=649137252491-compute@developer.gserviceaccount.com --scopes=https://www.googleapis.com/auth/cloud-platform --tags=allowall,http-server,https-server --image=centos-7-v20190116 --image-project=centos-cloud --boot-disk-size=20GB --boot-disk-type=pd-standard --boot-disk-device-name=cloudstore-sa-gc

## Create FIXED NODE:
## IP
gcloud beta compute --project=wolk-sa addresses create wolk-8-gc-sa-datastore-8-regional-ip --region=southamerica-east1 --network-tier=PREMIUM
regional_ip=`gcloud beta compute --project=wolk-sa addresses list`

## Create FIXED COMPUTE INSTANCE
gcloud compute --project=wolk-sa instances create wolk-8-gc-sa-datastore-8 --zone=southamerica-east1-b --machine-type=custom-1-1024 --address $regional_ip --subnet=default --network-tier=PREMIUM --metadata=startup-script-url=gs://startup_scripts_us/scripts/cloudstore/startup-script-cloudstore-repo.sh,ssh-keys=root:ssh-rsa\ AAAAB3NzaC1yc2EAAAABIwAAAQEA0\+dWJfxKJKozhrHQ8Zn06CsIXg3nav5tBi5ojZUKjDrOI78P0BPwaIod48fj8er8Z/spyW/pQ5Ys/TAL739TWPMtTQwfsWvsD1B5chOVVWrb5BomcEqWzcg/u6vCUqYFfP8q2p5p46w5U41nO\+S1mO\+NjdWsNn4f2Gqg8xIXZs/BDb02\+AdBZ/DTwj12HAJHoBpUF7OBLVibJwDGX4xj1BQuYtg7\+iCeaW2aW0NDCifU5bPOCZGQ4gAWG7HLGoqEE\+EGvousqXIJ\+K58Ex/G\+21qgaMqLr4QL\+ZBkHGEZ6q72/ziz7Nz9Na3XpanUDYcdfC1ppXlydtxV8yXQgSfNQ==\ root@www6002$'\n'anand:ssh-rsa\ AAAAB3NzaC1yc2EAAAADAQABAAABAQDofwXiKi9VAXlXETZMX16aSu3xZRuwVBB8Jgd/JIT3wG0yfCSymSvSS1Cc8nMm3hwywL5IxwCiIDXxrHsIsBeTuyQWSMZRcNcbEOC4fQxyRNBi1Arqb3t5OPIEZD1Y57C42vV0Hpk08zfysveuww7vJD69inCyGhE8IB4CD6Hn0N2aDQjMp3wtvao6c9aZh9OqtCpyHX0W8EC27RiTOf\+kFy2XxGQ544nJo08g//6bwbZVTS5/Ta4OOjPu92pu40BwnQQyOpPz0FErEGujInFeHhSp3mF3/MFavn5PO8Ne8sZ3CAHg\+NIrQO7i94UYTpnM22E/Xjr3FhN7/qmjHAM/\ anand@MdotMs-MacBook-Pro.local --maintenance-policy=MIGRATE --service-account=649137252491-compute@developer.gserviceaccount.com --scopes=https://www.googleapis.com/auth/cloud-platform --tags=allowall,http-server,https-server --image=centos-7-v20190116 --image-project=centos-cloud --boot-disk-size=20GB --boot-disk-type=pd-standard --boot-disk-device-name=wolk-8-gc-sa-datastore-8


## gcloud compute instance-groups list
## gcloud compute instance-groups managed create
gcloud -q beta compute --project=wolk-sa instance-groups managed create wolk-8-gc-sa-datastore-8 --base-instance-name=wolk-8-gc-sa-datastore-8 --template=cloudstore-sa-gc --size=1 --zones=southamerica-east1-a,southamerica-east1-b,southamerica-east1-c --initial-delay=300

## gcloud compute instance-groups managed set-autoscaling
gcloud compute --project "wolk-sa" instance-groups managed set-autoscaling "wolk-8-gc-sa-datastore-8" --region "southamerica-east1" --cool-down-period "60" --max-num-replicas "1" --min-num-replicas "1" --target-cpu-utilization "0.6"

## gcloud compute http-health-checks list --project=wolk-sa
## gcloud compute http-health-checks create --project=wolk-sa
gcloud compute --project "wolk-sa" http-health-checks create "cloudstore-sa-healthcheck" --port "80" --request-path "/healthcheck" --check-interval "5" --timeout "5" --unhealthy-threshold "2" --healthy-threshold "2"

## Add/Associate named-ports with compute instance group (required by backend)
#a. --named-ports=cloudstore:80
# gcloud compute instance-groups  set-named-ports --region=southamerica-east1 --named-ports=cloudstore-80:80 wolk-8-gc-sa-datastore-8
#b. --named-ports=http:8080
# gcloud compute instance-groups  set-named-ports --region=southamerica-east1 --named-ports=cloudstore-8080:8080 wolk-8-gc-sa-datastore-8
#c. --named-ports=cloudstore:8080,http:80 (CLEAR NAMED PORT FIRST)
gcloud compute instance-groups set-named-ports wolk-8-gc-sa-datastore-8 --named-ports "" --region southamerica-east1
gcloud compute instance-groups  set-named-ports --region=southamerica-east1 --named-ports=cloudstore-8080:8080,cloudstore-80:80 wolk-8-gc-sa-datastore-8
#d. To clear the list of named ports pass empty list as flag value. For example:
# gcloud compute instance-groups set-named-ports wolk-8-gc-sa-datastore-8 --named-ports "" --region southamerica-east1
#e. To GET NAMED PORTS
# gcloud compute instance-groups get-named-ports wolk-8-gc-sa-datastore-8 --region southamerica-east1

#### LOAD BALANCER ####
#1a. gcloud compute backend-services create - PORT 80
gcloud compute --project wolk-sa backend-services create cloudstore-sa-gc-80 --global --http-health-checks cloudstore-sa-healthcheck --load-balancing-scheme=EXTERNAL --port-name=cloudstore-80 --protocol=HTTP
#1b. gcloud compute backend-services add-backend - PORT 80
gcloud compute backend-services add-backend cloudstore-sa-gc-80 --instance-group=wolk-8-gc-sa-datastore-8 --instance-group-region=southamerica-east1 --balancing-mode=UTILIZATION --global --max-utilization=0.8 --max-rate-per-instance=1000

#1c. gcloud compute backend-services create - PORT 8080
# gcloud compute --project wolk-sa backend-services create cloudstore-sa-gc-8080 --global --http-health-checks cloudstore-sa-healthcheck --load-balancing-scheme=EXTERNAL --port-name=cloudstore-8080 --protocol=HTTP
#1d. gcloud compute backend-services add-backend - PORT 8080
# gcloud compute backend-services add-backend cloudstore-sa-gc-8080 --instance-group=wolk-8-gc-sa-datastore-8 --instance-group-region=southamerica-east1 --balancing-mode=UTILIZATION --global --max-utilization=0.8 --max-rate-per-instance=1000

#2a. Add URL MAP - cloudstore-sa-gc - PORT 80
gcloud compute url-maps create  cloudstore-sa-gc --default-service cloudstore-sa-gc-80 --description "Backend Service for LB"

#2b. Add URL MAP - cloudstore-sa-gc - PORT 8080
# gcloud compute url-maps create  cloudstore-sa-gc --default-service cloudstore-sa-gc-8080 --description "Backend Service for LB"

#3. CREATE TARGET PROXY "cloudstore-sa-gc" USING THE ABOVE URL MAP
gcloud compute --project=wolk-sa target-http-proxies create cloudstore-sa-gc-target-proxy-80 --url-map=cloudstore-sa-gc
# gcloud compute --project=wolk-sa target-http-proxies create cloudstore-sa-gc-target-proxy-8080 --url-map=cloudstore-sa-gc

#4a. Create Static IP - cloudstore-sa-gc-global
gcloud beta compute --project=wolk-sa addresses create cloudstore-sa-gc-global-ip --global --network-tier=PREMIUM

#4b. Create Static IP - cloudstore-sa-gc-regional
# gcloud beta compute --project=wolk-sa addresses create cloudstore-sa-gc-regional-ip --region=southamerica-east1 --network-tier=PREMIUM

#4c. List IP to use it in the next step for forwarding-rules
# gcloud beta compute --project=wolk-sa addresses list
# gcloud compute backend-services list --project=wolk-sa

#5a. Create GLOBAL forwarding-rules - cloudstore-sa-gc-80
gcloud compute --project=wolk-sa forwarding-rules create cloudstore-sa-gc-80 --global --address=cloudstore-sa-gc-global-ip --ip-protocol=TCP --ports=80 --target-http-proxy=cloudstore-sa-gc-target-proxy-80
#5b. Create GLOBAL forwarding-rules - cloudstore-sa-gc-8080
# gcloud compute --project=wolk-sa forwarding-rules create cloudstore-sa-gc-8080 --global --address=cloudstore-sa-gc-global-ip --ip-protocol=TCP --ports=8080 --target-http-proxy=cloudstore-sa-gc-target-proxy-8080


