##
#Variables (resource-group and location)
#resource-group=Wolk-AZ-RG location=westus
#creating one fixed instance for consensus
az vm create --resource-group Wolk-AZ-RG --name Wolk-AZ-FI --image /subscriptions/3615d31e-a8ac-4e88-acf2-1709ad2debab/resourceGroups/wolk-az-us-west/providers/Microsoft.Compute/images/wolk-az-us-west-image --admin-username wolk_azure --ssh-key-value ~/.ssh/id_rsa.pub
#1.Create Resource Group
az group create --name Wolk-AZ-RG --location westus
#list locations:az account list-locations --output table
##2.Create Public IP

az network public-ip create --resource-group Wolk-AZ-RG --name Wolk-AZ-IP --allocation-method static

##3.Create Load Balancer

az network lb create --resource-group Wolk-AZ-RG --name Wolk-AZ-LB --frontend-ip-name Wolk-AZ-FEP --backend-pool-name Wolk-AZ-BEP --public-ip-address Wolk-AZ-IP

##4.Create Health Probe

az network lb probe create --resource-group Wolk-AZ-RG --lb-name Wolk-AZ-LB --name Wolk-AZ-HealthProbe --protocol tcp --port 80

##5.Create load alancer rule

az network lb rule create --resource-group Wolk-AZ-RG --lb-name Wolk-AZ-LB --name Wolk-AZ-LB-Rule --protocol tcp --frontend-port 80 --backend-port 80 --frontend-ip-name Wolk-AZ-FEP --backend-pool-name Wolk-AZ-BEP --probe-name Wolk-AZ-HealthProbe

##7.Create Virtual Network

az network vnet create --resource-group Wolk-AZ-RG --name Wolk-AZ-Vnet --subnet-name Wolk-AZ-Subnet

##8.Ceate Network Security Group

az network nsg create --resource-group Wolk-AZ-RG --name Wolk-AZ-NSG
## Create Network Security Group Rule
az network nsg rule create --resource-group Wolk-AZ-RG --nsg-name Wolk-AZ-NSG --name Wolk-AZ-NSGRule --protocol tcp --direction inbound --source-address-prefix '*' --source-port-range '*' --destination-address-prefix '*' --destination-port-range '*' --access allow --priority 200
#Create Inbound Nat rule for ssh to VMs
for i in `seq 1 2`; do az network lb inbound-nat-rule create -g Wolk-AZ-RG --lb-name Wolk-AZ-LB -n Wolk-AZ-LBRuleSSH$i --protocol tcp --frontend-port 422$i --backend-port 22 --frontend-ip-name Wolk-AZ-FEP; done
##Create NICs
for i in `seq 1 2`; do az network nic create --resource-group Wolk-AZ-RG --name Wolk-AZ-Nic$i --vnet-name Wolk-AZ-Vnet --subnet Wolk-AZ-Subnet --network-security-group Wolk-AZ-NSG --lb-name Wolk-AZ-LB --lb-address-pools Wolk-AZ-BEP --lb-inbound-nat-rules Wolk-AZ-LBRuleSSH$i;done

##9.Create Availability Set

az vm availability-set create --resource-group Wolk-AZ-RG --name Wolk-AZ-AS

###10. Create VM for Availability Set
for i in `seq 1 2`; do az vm create --resource-group Wolk-AZ-RG --name Wolk-AZ-VM$i --availability-set Wolk-AZ-AS --nics Wolk-AZ-Nic$i --image /subscriptions/3615d31e-a8ac-4e88-acf2-1709ad2debab/resourceGroups/wolk-az-us-west/providers/Microsoft.Compute/images/wolk-az-us-west-image --admin-username wolk_azure --ssh-key-value ~/.ssh/id_rsa.pub;done

