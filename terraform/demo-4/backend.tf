terraform {
  backend "s3" {
    bucket = "terraform-tfstate-2019"
    key    = "terraform/demo4"
    region = "us-west-1"
  }
}
