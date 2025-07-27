/* Resource file for internet gateway creation */
resource "aws_internet_gateway" "temporal" {

 vpc_id = var.vpc_id
 tags = {
   Name = var.igw_name
 }

}

output "igw_id" {
  value = aws_internet_gateway.temporal.id
}

