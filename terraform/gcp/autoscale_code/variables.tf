variable "path" {
    default = "/Users/anand/.gcloud/Terraform.json"
}

variable "project" {
    default = "devops-112019"
}

variable "region" {
    default = "us-west1"
}

#provider "google" {
#    project = "devops-112019"
#    region = "us-west1"
#    credentials = file(var.path)
#}
