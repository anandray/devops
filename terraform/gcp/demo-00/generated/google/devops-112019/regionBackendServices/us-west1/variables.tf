data "terraform_remote_state" "healthChecks" {
  backend = "local"

  config = {
    path = "../../../../../generated/google/devops-112019/healthChecks/us-west1/terraform.tfstate"
  }
}
