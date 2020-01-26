
variable "path" {default = "/home/udemy/terraform/credentials"}

provider "google" {
    project = "devops-112019"
    region = "us-west1-a"
    credentials = "${file("${var.path}/secrets.json")}"
}