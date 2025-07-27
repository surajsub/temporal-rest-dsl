variable "subnet_cidr_public" {
  description = "cidr blocks for the public subnets"
  //default     = ["10.20.20.0/28", "10.20.20.16/28", "10.20.20.32/28"]
  type        = list(any)
}

variable "availability_zone" {
  description = "availability zones for the public subnets"
  //default     = ["us-east-2a","us-east-2b", "us-east-2c"]
  type        = list(any)
}

variable "vpc_id" {
  description = "The id of the vpc"
}