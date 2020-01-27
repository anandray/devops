variable "AWS_REGION" {
  default = "us-west-1"
}

variable "PATH_TO_PRIVATE_KEY" {
  default = "mykey"
}

variable "PATH_TO_PUBLIC_KEY" {
  default = "mykey.pub"
}

variable "AMIS" {
  type = map(string)
  default = {
    us-east-1 = "ami-13be557e"
    us-west-2 = "ami-0d1cd67c26f5fca19"
    us-west-1 = "ami-03ba3948f6c37a4b0"
  }
}

variable "RDS_PASSWORD" {
}

