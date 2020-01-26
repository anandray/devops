#Set up the environment

org_id=`gcloud organizations list --format text | grep organizations | cut -d "/" -f2`
billing_account_id=`gcloud beta billing accounts list --format text | grep billingAccounts | cut -d "/" -f2`

#Export the following variables to your environment for use throughout the tutorial.
export TF_VAR_org_id=$org_id
export TF_VAR_billing_account=$billing_account_id
export TF_ADMIN=${USER}-terraform-admin
export TF_CREDS=~/.config/gcloud/${USER}-terraform-admin.json
#Note: The TF_ADMIN variable will be used for the name of the Terraform Admin Project and must be unique.

#Create a new project and link it to your billing account:
gcloud projects create ${TF_ADMIN} \
  --organization ${TF_VAR_org_id} \
  --set-as-default

gcloud beta billing projects link ${TF_ADMIN} \
  --billing-account ${TF_VAR_billing_account}

#Create the Terraform service account
#Create the service account in the Terraform admin project and download the JSON credentials:

gcloud iam service-accounts create terraform \
  --display-name "Terraform admin account"

gcloud iam service-accounts keys create ${TF_CREDS} \
  --iam-account terraform@${TF_ADMIN}.iam.gserviceaccount.com

#Grant the service account permission to view the Admin Project and manage Cloud Storage:

gcloud projects add-iam-policy-binding ${TF_ADMIN} \
  --member serviceAccount:terraform@${TF_ADMIN}.iam.gserviceaccount.com \
  --role roles/viewer

gcloud projects add-iam-policy-binding ${TF_ADMIN} \
  --member serviceAccount:terraform@${TF_ADMIN}.iam.gserviceaccount.com \
  --role roles/storage.admin

#Any actions that Terraform performs require that the API be enabled to do so. In this guide, Terraform requires the following:

gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable cloudbilling.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable compute.googleapis.com
gcloud services enable serviceusage.googleapis.com

#Add organization/folder-level permissions
#Grant the service account permission to create projects and assign billing accounts:

gcloud organizations add-iam-policy-binding ${TF_VAR_org_id} \
  --member serviceAccount:terraform@${TF_ADMIN}.iam.gserviceaccount.com \
  --role roles/resourcemanager.projectCreator

gcloud organizations add-iam-policy-binding ${TF_VAR_org_id} \
  --member serviceAccount:terraform@${TF_ADMIN}.iam.gserviceaccount.com \
  --role roles/billing.user

# Reason: forbidden, Message: Required 'compute.firewalls.create' permission for 'projects/anand-terraform-admin/global/firewalls/nginx-firewall 
# Reason: forbidden, Message: Required 'compute.networks.updatePolicy' permission for 'projects/anand-terraform-admin/global/networks/default
gcloud organizations add-iam-policy-binding ${TF_VAR_org_id} \
  --member serviceAccount:terraform@${TF_ADMIN}.iam.gserviceaccount.com \
  --role roles/compute.securityAdmin

