terraform {

  required_providers {

    aws = {

      source = "hashicorp/aws"

      version = "~> 4.16"

    }


  }

  required_version = ">= 1.2.0"

}

provider "aws" {

  region = var.region


}



resource "aws_subnet" "rds" {
  count             = length(var.subnet_cidr_public)
  vpc_id            = var.vpc_id
  cidr_block        = var.subnet_cidr_public[count.index]
  availability_zone = var.availability_zone[count.index]
  tags = {
    "Name" = "rds-1-public-${count.index + 1}"
  }
}


output "aws_subnet_public_ids" {
  description = "List of IDs of public subnets"
  value       = aws_subnet.rds[*].id
}
