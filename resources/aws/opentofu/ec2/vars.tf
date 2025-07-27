variable "instance_type" {
  type        = string
  description = "AWS EC2 Instance type"
  default     = "t2.micro"
}


variable "ec2_count" {

}

variable "ami" {
  description = "ami Id from which instance needs to be created from"
  default     = "ami-0c7af5fe939f2677f"
}

variable "subnet_id" {
  description = "The subnet value to create this instance in"
  type = list(any)

}

variable "sg_id" {
  description = "The security group to associate this ec2 instance"

}

variable "key_name" {
  description = "The key name to associate this ec2 instance"

}

