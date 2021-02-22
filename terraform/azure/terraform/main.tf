# We strongly recommend using the required_providers block to set the
# Azure Provider source and version being used
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=2.46.0"
    }
  }
}

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {}

  subscription_id = var.subscription_id
  tenant_id = var.tenant_id
}

# Create a resource group
resource "azurerm_resource_group" "network" {
  name = "reltio-infosec"
  location = var.location
}

# Build Virtual Network
module "network" {
    source              = "Azure/network/azurerm"
    resource_group_name = azurerm_resource_group.network.name
#    location            = "{var.location}"
    vnet_name           = "infosec-vnet"
    address_space       = "10.250.0.0/16"
    subnet_prefixes     = ["10.250.1.0/24", "10.250.2.0/24", "10.250.3.0/24"]
    subnet_names        = ["infosec-subnet-1", "infosec-subnet-2", "infosec-subnet-3"]

    tags                = {
                            name = "infosec"
                            owner = "user"
                            environment = "infosec"
                          }
    depends_on = [azurerm_resource_group.network]
}

module "network-security-group" {
  source                = "Azure/network-security-group/azurerm"
  resource_group_name   = azurerm_resource_group.network.name
#  location              = var.location # Optional; if not provided, will use Resource Group location
  security_group_name   = "infosec-nsg"
  source_address_prefixes = ["54.164.86.205/32", "35.238.34.205/32"]

  custom_rules = [
    {
      name                   = "infosec-inbound"
      priority               = 201
      direction              = "Inbound"
      access                 = "Allow"
      protocol               = "tcp"
      source_port_range      = "*"
      destination_port_ranges = ["22", "80", "443", "5432", "8443"]
      source_address_prefixes = ["54.164.86.205/32", "35.238.34.205/32"]
      description            = "Reltio InfoSec Inbound rules"
    },
    {
      name                    = "infosec-outbound"
      priority                = 200
      direction               = "Outbound"
      access                  = "Allow"
      protocol                = "tcp"
      source_port_range       = "*"
      destination_port_range  = "*"
      source_address_prefixes = ["54.164.86.205/32", "35.238.34.205/32"]
      description             = "Reltio InfoSec Outbound rules"
    },
  ]
    depends_on = [azurerm_resource_group.network]
}
