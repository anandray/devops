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

    tags                = {
                            name = "infosec"
                            owner = "user"
                            environment = "infosec"
                          }
}
