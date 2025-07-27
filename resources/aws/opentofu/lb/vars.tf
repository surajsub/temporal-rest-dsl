variable "vpc_id" {}

variable "ec2_count"  {}

variable "instance_id" {
  type = list(any)
  description = "The instance id to be used"
}

variable "subnet_id" {
  type = list(any)
  description = "The subnet type to be used"
}

variable "sg_id" {
  description = "The security group to be used"
}