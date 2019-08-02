#!/bin/bash

#region=`az account list-locations --query "[].{Region:name}" --output table | tail -n+3`
region=`az account list-locations --query "[].{Region:name}" --output table | grep -E -v "westcentralus|australiacentral|southafricawest|francesouth" | tail -n+3`
region_options=`echo $region`
options=($region_options)
prompt="Select an Azure location:"

PS3="$prompt "
select answer in "${options[@]}"; do
	echo 'region="'$answer'"' > supernode-az-$answer.conf
	echo "Enter node #:"
	read node
	    echo 'node="'$node'"' >> supernode-az-$answer.conf
      break 2
done

source supernode-az-$answer.conf

# derivative inputs
app="cloudstore"
port=80
resourceGroup="wolk-rg-$region"
provider="az"
fixedinstance="wolk-$node-$provider-$region-cosmos"
public_ip="wolk-$provider-$region-ip"
#lb_name="wolk-$provider-$region-lb"
healthprobe="wolk-$provider-$region-healthprobe"
portname="$app-$port"
#Added for assigning fixed ip and nsg rule for fixed instance
#create public-ip for fixed instance
az network public-ip create -g $resourceGroup -n public_ip-$region-fi --allocation-method static

# Ceate Network Security Group
az network nsg create -g $resourceGroup -n wolk-az-$region-nsg

# Create Network Security Group Rule
az network nsg rule create -g $resourceGroup --nsg-name wolk-az-$region-nsg -n wolk-az-$region-nsg-rule --protocol tcp --direction inbound --source-address-prefix '*' --source-port-range '*' --destination-address-prefix '*' --destination-port-range '*' --access allow --priority 200

# Create Virtual Network
az network vnet create -g $resourceGroup -n wolk-az-$region-vnet --subnet-name wolk-az-$region-subnet

#Create NIC for fixed instance
az network nic create -g $resourceGroup -n wolk-az-$region-nic-fi --subnet wolk-az-$region-subnet --vnet-name wolk-az-$region-vnet --public-ip-address public_ip-$region-fi --network-security-group wolk-az-$region-nsg

#Create Fixed Instance 
resourceId=`az image list -g $resourceGroup --query "[].{id:id}" --output tsv`

#az vm create -g $resourceGroup -n $fixedinstance --image $resourceId
az vm create -g $resourceGroup -n $fixedinstance --image $resourceId --nics wolk-az-$region-nic-fi
#Create static Pulic IP for scale set
az network public-ip create -g $resourceGroup -n public_ip-$region-ss --allocation-method static
# Create Scale Set VMss
suffix=`date | sha256sum | head -c 4 ; echo`
az vmss create -g $resourceGroup -n wolk-az-ss-$region-$suffix --image $resourceId --instance-count 1 --upgrade-policy-mode manual --public-ip-per-vm --nsg wolk-az-$region-nsg --backend-pool-name wolk-az-$region-bep --lb wolk-az-$region-lb --public-ip-address public_ip-$region-ss 
#Create Health Probe
az network lb probe create -g $resourceGroup --lb-name wolk-az-$region-lb -n $healthprobe --protocol tcp --port 443 --interval 15
#Create Scale Set Load Balance Rule
az network lb rule create -g $resourceGroup -n wolk-az-$region-lbrule --lb-name wolk-az-$region-lb --backend-pool-name wolk-az-$region-bep --backend-port 443 --frontend-ip-name loadBalancerFrontEnd --frontend-port 443 --protocol tcp --probe-name $healthprobe
# add storage account and BLOB
az storage account create -g $resourceGroup -n wolkaz$region --default-action allow --kind Storage
AZURE_STORAGE_ACCOUNT="wolkaz$region"
AZURE_STORAGE_KEY=`az storage account keys list -g wolk-rg-$region -n wolkaz$region --query [].{keyName:value} -o tsv | head -n1`
AZURE_STORAGE_CONNECTION_STRING=`az storage account show-connection-string -n wolkaz$region -o tsv`

az storage container create --connection-string $AZURE_STORAGE_CONNECTION_STRING -n wolkdbmaster

# adding $AZURE_STORAGE_ACCOUNT and $AZURE_STORAGE_KEY to bashrc
instance_ip=`az network public-ip list -g $resourceGroup --query "[].{ipAddress:ipAddress}" --output tsv | head -n1`
#public_ip1=`az network public-ip list -g $resourceGroup --query "[].{ipAddress:ipAddress}" --output tsv | tail -n1`
storage_ip=( $(az vmss list-instance-public-ips -g $resourceGroup -n wolk-az-ss-$region-$suffix --query []."{ipAddress:ipAddress}" -o tsv) )

echo -e "\nAdding AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_KEY to bashrc"
sudo scp -q $instance_ip:/root/.bashrc /tmp/bashrc
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
instance_ip=`az network public-ip list -g $resourceGroup --query "[].{ipAddress:ipAddress}" --output tsv | head -n1`
#public_ip1=`az network public-ip list -g $resourceGroup --query "[].{ipAddress:ipAddress}" --output tsv | tail -n1`
storage_ip=( $(az vmss list-instance-public-ips -g $resourceGroup -n wolk-az-ss-$region-$suffix --query []."{ipAddress:ipAddress}" -o tsv) )

echo -e "\nAdding ConsensusIdx = $node and AZURE KEYS to wolk.toml"

#sudo ssh $instance_ip git --git-dir=/root/go/src/github.com/wolkdb/cloudstore/.git pull 2> /dev/null
sudo scp -q $instance_ip:/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/wolk.toml-az-template /tmp/wolk.toml
sudo sed -i "1 i\ConsensusIdx = $node" /tmp/wolk.toml
sudo sed -i "/MicrosoftAzureAccountName/d" /tmp/wolk.toml
sudo sed -i "/MicrosoftAzureAccountKey/d" /tmp/wolk.toml
sudo sed -i "/SSL/d" /tmp/wolk.toml

sudo su - << EOF
echo 'MicrosoftAzureAccountName = "$AZURE_STORAGE_ACCOUNT"
MicrosoftAzureAccountKey = "$AZURE_STORAGE_KEY"
SSLCertFile = "/etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt"
SSLKeyFile = "/etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key"' >> /tmp/wolk.toml
EOF

sudo scp -q /tmp/wolk.toml $instance_ip:/root/go/src/github.com/wolkdb/cloudstore/wolk.toml

echo -e "\nAdding the ssl certificates to Consensus node..."
sudo mkdir -p /root/azure/wildcard.wolk.com
sudo azcopy --quiet --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/wildcard.wolk.com/www.wolk.com.key --destination /root/azure/wildcard.wolk.com/www.wolk.com.key --source-key $AZURE_STORAGE_KEY_MAIN 2> /dev/null
sudo azcopy --quiet --source https://wolkaz.file.core.windows.net/wolk-startup-scripts/scripts/plasma/wildcard.wolk.com/www.wolk.com.crt --destination /root/azure/wildcard.wolk.com/www.wolk.com.crt --source-key $AZURE_STORAGE_KEY_MAIN 2> /dev/null

sudo scp -r /root/azure/wildcard.wolk.com $instance_ip:/etc/ssl/certs/ 2> /dev/null
sudo ssh -q $instance_ip service wolk restart 2> /dev/null

# scp wolk.toml to autoscaling instance
storage_ip=( $(az vmss list-instance-public-ips -g $resourceGroup -n wolk-az-ss-$region-$suffix --query []."{ipAddress:ipAddress}" -o tsv) )
sudo sed -i "/ConsensusIdx/d" /tmp/wolk.toml

for storageIP in "${storage_ip[@]}";
do
echo -e "\nAdding AZURE KEYS to wolk.toml on Storage nodes --> ${storageIP}"
sudo ssh -q ${storageIP} git --git-dir=/root/go/src/github.com/wolkdb/cloudstore/.git pull 2> /dev/null
sudo scp -q /tmp/wolk.toml ${storageIP}:/root/go/src/github.com/wolkdb/cloudstore/wolk.toml

echo -e "\nAdding the ssl certificates to Storage nodes..."
sudo scp -q -r /root/azure/wildcard.wolk.com ${storageIP}:/etc/ssl/certs/
sudo ssh -q ${storageIP} service wolk restart
done
