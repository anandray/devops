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
lb_name="wolk-$provider-$region-lb"
healthprobe="wolk-$provider-$region-healthprobe"
portname="$app-$port"
vmss=$(az vmss list -g wolk-rg-$region --query [*].name -o tsv)
# Delete Fixed Instance 
az vm delete --ids $(az vm list -g $resourceGroup --query "[].id" -o tsv | grep cosmos) --yes

#Delete OS Disk of cosmos instance
az disk delete -g $resourceGroup -n $(az disk list -g $resourceGroup --query []."{id:name}" --output tsv | grep cosmos) --yes

#Delete Nic for cosmos Instance
az network nic delete -g $resourceGroup -n wolk-az-$region-nic-fi 

#Delete PublicIP attached with cosmos instance
az network public-ip delete -g $resourceGroup -n public_ip-$region-fi
#Delete VMSS
az vmss delete -g $resourceGroup -n $vmss
#Delete LB
az network lb delete -g $resourceGroup -n $(az network lb list -g $resourceGroup --query "[].{name:name}" --output tsv)
#Delete LB PulicIP
az network public-ip delete -g $resourceGroup -n public_ip-$region-ss
#Delete vnet
az network vnet delete -g $resourceGroup -n wolk-az-$region-vnet;
#delete NSG
az network nsg delete -g $resourceGroup -n wolk-az-$region-nsg
#Delete Storage Account
az storage account delete -n $(az storage account list -g wolk-rg-$region --query [*].name -o tsv) -g wolk-rg-$region --yes
