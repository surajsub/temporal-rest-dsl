resource "aws_subnet" "temporal" {
  count             = length(var.subnet_cidr_public)
  vpc_id            = var.vpc_id
  cidr_block        = var.subnet_cidr_public[count.index]
  availability_zone = var.availability_zone[count.index]
  tags = {
    "Name" = "app-1-public-${count.index + 1}"
  }
}


output "aws_subnet_public_ids" {
  description = "List of IDs of public subnets"
  value       = aws_subnet.temporal[*].id
}