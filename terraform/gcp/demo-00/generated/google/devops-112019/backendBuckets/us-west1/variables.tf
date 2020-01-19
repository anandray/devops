data "terraform_remote_state" "gcs" {
  backend = "local"

  config = {
    path = "../../../../../generated/google/devops-112019/gcs/us-west1/terraform.tfstate"
  }
}
