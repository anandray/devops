variable "region" {
  default = "us-west1"
}

variable "region_zone" {
  default = "us-west1-a"
}

variable "org_id" {
  description = "The ID of the Google Cloud Organization."
}

variable "billing_account_id" {
  description = "The ID of the associated billing account (optional)."
}

variable "credentials_file_path" {
  description = "Location of the credentials to use."
  default     = "~/.gcloud/devops-112019-911b07b6a1a9.json"
#  default     = "~/.gcloud/Terraform.json"
}
