resource "aws_instance" "instsance-terra-01" {
  ami           = var.AMIS[var.AWS_REGION]
  instance_type = "t2.micro"
}

