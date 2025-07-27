variable "subnet_cidr_public" {
  description = "cidr blocks for the public subnets"
  //default     = ["10.20.20.0/28", "10.20.20.16/28", "10.20.20.32/28"]
   default = ["10.0.1.0/24", "10.0.2.0/24"]
  type        = list(any)
}

variable "availability_zone" {
  description = "availability zones for the public subnets"
  //default     = ["us-east-2a","us-east-2b", "us-east-2c"]
  default     = ["us-west-1a","us-west-1c"]
  type        = list(any)
}

variable "vpc_id" {
  description = "The id of the vpc"
}

variable "region" {
  description ="The region "
  default = "us-west-1"
}
