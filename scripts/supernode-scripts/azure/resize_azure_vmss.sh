#!/bin/bash
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
vmscaleset=$(az vmss list -g $resourceGroup --query [*].name -o tsv)
#old_ip=$(az vmss list-instance-public-ips -g $resourceGroup -n $vmscaleset --query [*].ipAddress -o tsv)
#Present Capacity
echo Present Capacity $(az vmss show -g $resourceGroup -n $vmscaleset --query [sku.capacity] -o tsv)
#New Capacity Prompt
read -p "New Capacity:" number
#Name of vmscaleset
#vmscaleset=$(az vmss list -g $resourceGroup --query [*].name -o tsv)
az vmss scale -g $resourceGroup -n $vmscaleset --new-capacity $number
#Get storage box IPs
storage_ip=( $(az vmss list-instance-public-ips -g $resourceGroup -n $vmscaleset --query []."{ipAddress:ipAddress}" -o tsv) ) 
#scp /root/.bashrc
for storageIP in "${storage_ip[@]}";
do
sudo scp -q /tmp/bashrc ${storageIP}:/root/.bashrc
done
# scp wolk.toml to autoscaling instance
storage_ip=( $(az vmss list-instance-public-ips -g $resourceGroup -n $vmscaleset --query []."{ipAddress:ipAddress}" -o tsv) ) 
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
