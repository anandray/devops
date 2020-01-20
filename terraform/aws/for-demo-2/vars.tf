variable "AWS_REGION" {
  type    = string
  default = "us-west-1"
}
variable "project_tags" {
  type          = map(string)
  default       = {
    Component   = "Frontend"
    Environment = "Production"
  }
}

