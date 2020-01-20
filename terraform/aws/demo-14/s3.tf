resource "aws_s3_bucket" "b" {
  bucket = "mybucket-anand-012020"
  acl    = "private"

  tags = {
    Name = "mybucket-anand-012020"
  }
}

