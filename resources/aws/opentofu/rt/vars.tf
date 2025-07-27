variable vpc_id {
  description = "VPC ID"
}



variable "igw_id" {
  description = "Internet Gateway"
}


//variable "subnet_id" {
 // description = "THe public subnet id created for this resource"

//}

variable "rt_count" {
    description = "The number of route table associations to create"

}

variable "subnet_cidr_public_ids" {
  description = "cidr blocks for the public subnets"
  //default     = ["10.20.20.0/28", "10.20.20.16/28", "10.20.20.32/28"]
  type        = list(any)
}