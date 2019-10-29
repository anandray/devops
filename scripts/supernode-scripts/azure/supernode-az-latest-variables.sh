#!/bin/bash

region=`az account list-locations --query "[].{Region:name}" --output table | tail -n+3`
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
autoscaledinstance="wolk-$node-autoscale-$region"
prefix="$app-$region-$provider"
instancetemplate="$prefix"
urlmap="$prefix"
instancetemplate="$prefix"
lbname="$prefix-$port"
globalip="$app-$region-$provider-global-ip"
targetproxy="$prefix-target-proxy-$port"
regionalipname="$app-$region-$provider-regional-ip"
healthcheck="$app-$region-healthcheck"
portname="$app-$port"


b=$(tput bold)
n=$(tput sgr0)

echo -e "\n${b}Region: ${n}$region"
echo -e "\n${b}ResourceGroup: ${n}$resourceGroup"
echo -e "\n${b}Provider: ${n}$provider"
echo -e "\n${b}FixedInstance: ${n}$fixedinstance"

echo -e "\n${b}Image Name:${n} ${n}wolk-az-$region-image"
echo -e "\n${b}Resource Group:${n} $resourceGroup"
echo -e "\n${b}Public IP:${n} wolk-az-$region-ip"
echo -e "\n${b}LB Name:${n} wolk-az-$region-lb"
echo -e "\n${b}Frontend IP Name:${n} wolk-az-$region-feip"
echo -e "\n${b}Backend Pool Name:${n} wolk-az-$region-bep"
echo -e "\n${b}HealthProbe Name:${n} wolk-az-$region-healthprobe"
echo -e "\n${b}LB Rule:${n} wolk-az-$region-lbrule"
echo -e "\n${b}Virtual Network Name:${n} wolk-az-$region-vnet"
echo -e "\n${b}Subnet Name:${n} wolk-az-$region-subnet"
echo -e "\n${b}Network Security Group:${n} wolk-az-$region-nsg"
echo -e "\n${b}NSG Rule:${n} wolk-az-$region-nsgrule"
for i in `seq 1 2`;
do
echo -e "\n${b}LB SSH Rule:${n} wolk-az-$region-lbrulessh$i";
done

for i in `seq 1 2`;
do
echo -e "\n${b}NIC:${n} wolk-az-$region-nic$i";
done

echo -e "\n${b}Availability Set:${n} wolk-az-$region-as"
for i in `seq 1 2`;
do
echo -e "\n${b}VM:${n} wolk-az-$region-vm-$node-$i";
done
echo -e "\n"
