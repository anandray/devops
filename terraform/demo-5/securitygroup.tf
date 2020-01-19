data "aws_ip_ranges" "american_ec2" {
  regions  = ["us-west-1", "us-central-1"]
  services = ["ec2"]
}

resource "aws_security_group" "from_us" {
  name = "from_us"

  ingress {
    from_port   = "443"
    to_port     = "443"
    protocol    = "tcp"
    cidr_blocks = data.aws_ip_ranges.american_ec2.cidr_blocks
  }
  ingress {
    from_port   = "80"
    to_port     = "80"
    protocol    = "tcp"
    cidr_blocks = data.aws_ip_ranges.american_ec2.cidr_blocks
  }
  tags = {
    CreateDate = data.aws_ip_ranges.american_ec2.create_date
    SyncToken  = data.aws_ip_ranges.american_ec2.sync_token
  }
}

