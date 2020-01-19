data "terraform_remote_state" "networks" {
  backend = "local"

  config = {
    path = "../../../../../generated/google/devops-112019/networks/us-west1/terraform.tfstate"
  }
}

data "terraform_remote_state" "subnetworks" {
  backend = "local"

  config = {
    path = "../../../../../generated/google/devops-112019/subnetworks/us-west1/terraform.tfstate"
  }
}
