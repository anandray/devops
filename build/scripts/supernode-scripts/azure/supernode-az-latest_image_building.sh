#!/bin/bash
region=$1
node=$2
# derivative inputs
app="cloudstore"
port=80
https_port=443
#resourceGroup="wolk-rg-$region"
provider="az"
fixedinstance="wolk-$node-$provider-$region-cosmos"
public_ip="wolk-$provider-$region-ip"
#lb_name="wolk-$provider-$region-lb"
healthprobe="wolk-$provider-$region-healthprobe"
portname="$app-$port"
#create public-ip for fixed instance
az network public-ip create -g wolk-rg-$region -n public_ip-$region-consensus --allocation-method static
# Ceate Network Security Group
az network nsg create -g wolk-rg-$region -n wolk-az-$region-nsg
# Create Network Security Group Rule
az network nsg rule create -g wolk-rg-$region --nsg-name wolk-az-$region-nsg -n wolk-az-$region-nsg-tcp_rule --protocol tcp --direction inbound --source-address-prefix '*' --source-port-range '*' --destination-address-prefix '*' --destination-port-range '*' --access allow --priority 100
az network nsg rule create -g wolk-rg-$region --nsg-name wolk-az-$region-nsg -n wolk-az-$region-nsg-udp_rule --protocol udp --direction inbound --source-address-prefix '*' --source-port-range '*' --destination-address-prefix '*' --destination-port-range '*' --access allow --priority 110
# Create Virtual Network
az network vnet create -g wolk-rg-$region -n wolk-az-$region-vnet --subnet-name wolk-az-$region-subnet
#Create NIC for fixed instance
az network nic create -g wolk-rg-$region -n wolk-az-$region-nic-consensus --subnet wolk-az-$region-subnet --vnet-name wolk-az-$region-vnet --public-ip-address public_ip-$region-consensus --network-security-group wolk-az-$region-nsg
#Create Instance with Centos Image
az vm create -g wolk-rg-$region -n wolk-image-$region-vm --image Centos --nics wolk-az-$region-nic-consensus
#scp startup script to vm
image_instance_ip=$(az network public-ip list -g wolk-rg-$region --query "[].{ipAddress:ipAddress}" --output tsv)
scp startup-script-cloudstore-repo-azure.sh wolk_azure@$image_instance_ip:/home/wolk_azure/
#Run startup-script on VM
az vm run-command invoke -g wolk-rg-$region -n wolk-image-$region-vm --command-id RunShellScript --scripts "sudo sh /home/wolk_azure/startup-script-cloudstore-repo-azure.sh"
#az vm run-command invoke -g wolk-rg-$region -n wolk-image-$region-vm --command-id RunShellScript --scripts "sudo waagent -force -deprovision+user && history -c"
ssh $image_instance_ip sudo waagent -force -deprovision+user
az vm deallocate -g wolk-rg-$region -n wolk-image-$region-vm
az vm generalize -g wolk-rg-$region -n wolk-image-$region-vm
az image create -g wolk-rg-$region -n wolk-image-$region --source wolk-image-$region-vm
az vm delete -g wolk-rg-$region -n wolk-image-$region-vm --yes
vm_disk_name=$(az disk list -g wolk-rg-$region --query []."{id:name}" --output tsv)
az disk delete -g wolk-rg-$region -n $vm_disk_name --yes
resourceId=$(az image list -g wolk-rg-$region --query "[].{id:id}" --output tsv)
#create Consensus box
az vm create -g wolk-rg-$region -n $fixedinstance --image $resourceId --nics wolk-az-$region-nic-consensus --size Standard_D2_v3
#Cr ate static Pulic IP for scale set
az network public-ip create -g wolk-rg-$region -n public_ip-$region-ss --allocation-method static
# Create Scale Set VMss
suffix=`date | sha256sum | head -c 4 ; echo`
az vmss create -g wolk-rg-$region -n wolk-az-ss-$region-$suffix --image $resourceId --instance-count 1 --upgrade-policy-mode manual --public-ip-per-vm --nsg wolk-az-$region-nsg --backend-pool-name wolk-az-$region-bep --lb wolk-az-$region-lb --public-ip-address public_ip-$region-ss
#Create Health Probe
az network lb probe create -g wolk-rg-$region --lb-name wolk-az-$region-lb -n $healthprobe --protocol tcp --port $https_port --interval 15
#Create Scale Set Load Balance Rule
az network lb rule create -g wolk-rg-$region -n wolk-az-$region-lbrule --lb-name wolk-az-$region-lb --backend-pool-name wolk-az-$region-bep --backend-port $https_port --frontend-ip-name loadBalancerFrontEnd --frontend-port $https_port --protocol tcp --probe-name $healthprobe
# add storage account and BLOB
az storage account create -g wolk-rg-$region -n wolkaz$region --default-action allow --kind Storage
AZURE_STORAGE_ACCOUNT="wolkaz$region"
AZURE_STORAGE_KEY=`az storage account keys list -g wolk-rg-$region -n wolkaz$region --query [].{keyName:value} -o tsv | head -n1`
AZURE_STORAGE_CONNECTION_STRING=`az storage account show-connection-string -n wolkaz$region -o tsv`
az storage container create --connection-string $AZURE_STORAGE_CONNECTION_STRING -n wolkdbmaster
# adding $AZURE_STORAGE_ACCOUNT and $AZURE_STORAGE_KEY to bashrc
instance_ip=`az network public-ip list -g wolk-rg-$region --query "[].{ipAddress:ipAddress}" --output tsv | head -n1`
storage_ip=( $(az vmss list-instance-public-ips -g wolk-rg-$region -n wolk-az-ss-$region-$suffix --query []."{ipAddress:ipAddress}" -o tsv) )

echo -e "\nAdding AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_KEY to bashrc"
#sudo scp -q $instance_ip:/root/.bashrc /tmp/bashrc
sudo azcopy --quiet --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/bashrc --destination /tmp/bashrc  --source-key $AZURE_STORAGE_KEY_MAIN 2> /dev/null
sudo sed -i "/AZURE_STORAGE_ACCOUNT/d" /tmp/bashrc
sudo sed -i "/AZURE_STORAGE_KEY/d" /tmp/bashrc
sudo sed -i "/AZURE_STORAGE_CONNECTION_STRING/d" /tmp/bashrc
sudo sed -i "/Azure environment variables/d" /tmp/bashrc

sudo su - <<EOF
echo -e "# Azure environment variables
export AZURE_STORAGE_ACCOUNT=$AZURE_STORAGE_ACCOUNT
export AZURE_STORAGE_KEY=$AZURE_STORAGE_KEY
export AZURE_STORAGE_CONNECTION_STRING='$AZURE_STORAGE_CONNECTION_STRING'" >> /tmp/bashrc
EOF
sudo scp -q /tmp/bashrc $instance_ip:/root/.bashrc

for storageIP in "${storage_ip[@]}";
do
sudo scp -q /tmp/bashrc ${storageIP}:/root/.bashrc
done
# scp wolk.toml to consensus instance
AZURE_STORAGE_KEY_MAIN="CG4pGq6GMTOoIvNiYWa5F0I4s2byQVAFxkExCVLRqlN09qP2C/MF7ATm3RTkMr60Og/thbSlbGnrf8+v3Ot7pQ=="
instance_ip=`az network public-ip list -g wolk-rg-$region --query "[].{ipAddress:ipAddress}" --output tsv | head -n1`
storage_ip=( $(az vmss list-instance-public-ips -g wolk-rg-$region -n wolk-az-ss-$region-$suffix --query []."{ipAddress:ipAddress}" -o tsv) )

echo -e "\nAdding ConsensusIdx = $node and AZURE KEYS to wolk.toml"

sudo ssh $instance_ip git --git-dir=/root/go/src/github.com/wolkdb/cloudstore/.git pull 2> /dev/null
sudo scp -q $instance_ip:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/wolk.toml-az-template /tmp/wolk.toml
sudo sed -i "s/_ConsensusIdx/$node/g" /tmp/wolk.toml
sudo sed -i "s/_NodeType/consensus/g" /tmp/wolk.toml
sudo sed -i "s/_Region/$region/g" /tmp/wolk.toml
sudo sed -i "s/_MicrosoftAzureAccountName/$AZURE_STORAGE_ACCOUNT/g" /tmp/wolk.toml
sudo sed -i "s|_MicrosoftAzureAccountKey|$AZURE_STORAGE_KEY|g" /tmp/wolk.toml

sudo scp -q /tmp/wolk.toml $instance_ip:/root/go/src/github.com/wolkdb/cloudstore/wolk.toml
echo -e "\nAdding the ssl certificates to Consensus node..."
sudo mkdir -p /root/azure/wildcard.wolk.com
sudo azcopy --quiet --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/wildcard.wolk.com/www.wolk.com.key --destination /root/azure/wildcard.wolk.com/www.wolk.com.key --source-key $AZURE_STORAGE_KEY_MAIN 2> /dev/null
sudo azcopy --quiet --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/wildcard.wolk.com/www.wolk.com.crt --destination /root/azure/wildcard.wolk.com/www.wolk.com.crt --source-key $AZURE_STORAGE_KEY_MAIN 2> /dev/null
sudo scp -r /root/azure/wildcard.wolk.com $instance_ip:/etc/ssl/certs/ 2> /dev/null
sudo ssh $instance_ip git --git-dir=/root/go/src/github.com/wolkdb/cloudstore/.git pull 2> /dev/null
sudo ssh -q $instance_ip make wolk -C /root/go/src/github.com/wolkdb/cloudstore
for i in {1..5};
do sudo ssh -q $instance_ip cp /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk$i;
done
sudo ssh -q $instance_ip service wolk restart 2> /dev/null
# scp wolk.toml to autoscaling instance
storage_ip=( $(az vmss list-instance-public-ips -g wolk-rg-$region -n wolk-az-ss-$region-$suffix --query []."{ipAddress:ipAddress}" -o tsv) )

for storageIP in "${storage_ip[@]}";
do
echo -e "\nAdding AZURE KEYS to wolk.toml on Storage nodes --> ${storageIP}"
sudo ssh -q ${storageIP} git --git-dir=/root/go/src/github.com/wolkdb/cloudstore/.git pull 2> /dev/null
sudo sed -i "s|consensus|storage|g" /tmp/wolk.toml
sudo scp -q /tmp/wolk.toml ${storageIP}:/root/go/src/github.com/wolkdb/cloudstore/wolk.toml
echo -e "\nAdding the ssl certificates to Storage nodes..."
sudo scp -q -r /root/azure/wildcard.wolk.com ${storageIP}:/etc/ssl/certs/
sudo ssh -q ${storageIP} service wolk restart
done
