resource "aws_instance" "example" {
  ami           = var.AMIS
  instance_type = "t2.micro"
}

