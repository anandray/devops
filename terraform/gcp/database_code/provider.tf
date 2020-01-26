
variable "path" {
    default = "/Users/anand/.gcloud/Terraform.json"
}

provider "google" {
    project = "devops-112019"
    region = "us-west1"
    credentials = file(var.path)
}
