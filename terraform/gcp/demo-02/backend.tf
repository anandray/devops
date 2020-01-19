terraform {
 backend "gcs" {
   bucket  = "anand-terraform-admin"
   prefix  = "terraform/state"
 }
}
