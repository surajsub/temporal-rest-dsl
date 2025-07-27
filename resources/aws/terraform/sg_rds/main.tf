resource "aws_security_group" "rds_sg" {
  name_prefix = "rds-"

  vpc_id = var.rds_vpc

  # Add any additional ingress/egress rules as needed
  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
