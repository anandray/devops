#!/bin/bash
region=$1
resource_group=wolk-rg-$region
#node=$2
vmss_name=$(az vmss list -g $resource_group --query []."{id:name}" -o tsv)
#scale vmss to 0
az vmss scale -n $vmss_name -g $resource_group --new-capacity 0
#Delete vmss
az vmss delete -g $resource_group -n $vmss_name
lb_name=$(az network lb list -g $resource_group --query []."{id:name}" -o tsv)
#Delete Load Balancer
az network lb delete -g $resource_group -n $lb_name
#Delete PulicIP of vmss
vmss_publicIP=public_ip-$region-ss
az network public-ip delete -g $resource_group -n $vmss_publicIP
#delete Consensus vm
vm_name=$(az vm list -g $resource_group --query []."{id:name}" -o tsv)
vm_disk_name=$(az disk list -g $resource_group --query []."{id:name}" --output tsv | grep cosmos)
az vm delete -g $resource_group -n $vm_name --yes
az disk delete -g $resource_group -n $vm_disk_name --yes
#Delete NIC
vm_nic=$(az network nic list -g $resource_group --query []."{id:name}" -o tsv)
az network nic delete -g $resource_group -n $vm_nic
vm_publicIP=public_ip-$region-consensus
az network public-ip delete -g $resource_group -n $vm_publicIP
#delete Network Security Group
nsg_name=$(az network nsg list -g $resource_group --query []."{id:name}" -o tsv)
az network nsg delete -g $resource_group -n $nsg_name
#Delete vnet
vnet_name=$(az network vnet list -g $resource_group --query []."{id:name}" -o tsv)
az network vnet delete -g $resource_group -n $vnet_name
#Delete Storage Account
az storage account delete -n $(az storage account list -g $resource_group --query [*].name -o tsv) -g $resource_group --yes
