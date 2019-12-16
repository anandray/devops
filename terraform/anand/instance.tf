resource "aws_instance" "example" {
  ami           = var.AWS_AMIS
  instance_type = "t2.micro"
}
