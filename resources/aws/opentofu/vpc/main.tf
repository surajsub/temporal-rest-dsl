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



resource "aws_vpc" "temporal" {

 cidr_block = var.cidr_block
  enable_dns_support = true
  enable_dns_hostnames = true

 tags = {

   Name = var.vpc_name

 }

}



output "vpc_id" {
  value = aws_vpc.temporal.id
}

output "vpc_cidr_block" {
  value = aws_vpc.temporal.cidr_block
}
