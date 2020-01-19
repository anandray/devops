data "terraform_remote_state" "instanceTemplates" {
  backend = "local"

  config = {
    path = "../../../../../generated/google/devops-112019/instanceTemplates/us-west1/terraform.tfstate"
  }
}
