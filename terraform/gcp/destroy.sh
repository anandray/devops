#######################################################################################

#!/bin/bash

######################################################################################

# Author        : DevOps Made Easy
# Email         : devopsmadeeasy@gmail.com
# Description   : Terraform destroy script
# Documentation : https://www.terraform.io/docs/commands/destroy.html

######################################################################################

# Export Path Variable
export PATH=$PATH:/opt

######################################################################################

# If statement to ensure a user has provided a Terraform folder path
if [[ -z "$1" ]]; then
echo ""
echo "You have not provided a Terraform path."
echo "SYNTAX = ./destroy.sh <PATH>"
echo "EXAMPLE = ./destroy.sh terraform/instance"
echo ""
exit
fi

######################################################################################

# The Init command is used to initialize a working directory containing Terraform configuration files.
# This is the first command that should be run after writing a new Terraform configuration
terraform init $1

#The Get command is used to download and update modules mentioned in the root module.
terraform get $1

#The Plan command is used to create an execution plan
terraform destroy -auto-approve $1 

######################################################################################
