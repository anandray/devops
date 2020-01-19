#!/bin/bash

org_id=`gcloud organizations list --format text | grep organizations | cut -d "/" -f2`
billing_account_id=`gcloud beta billing accounts list --format text | grep billingAccounts | cut -d "/" -f2`

#Export the following variables to your environment for use throughout the tutorial.
export TF_VAR_org_id=$org_id
export TF_VAR_billing_account=$billing_account_id
export TF_ADMIN=${USER}-terraform-admin
export TF_CREDS=~/.config/gcloud/${USER}-terraform-admin.json

#export TF_ADMIN=devops-112019
#export TF_CREDS=~/.gcloud/Terraform.json
