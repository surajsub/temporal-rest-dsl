

resource "aws_route_table" "public" {
  vpc_id = var.vpc_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = var.igw_id
  }


  tags = {
    Name = "temporal-public"
  }
}


// Since we are creating one public and one private subnet we will need two route associations

resource "aws_route_table_association" "public" {
  count          = var.rt_count
  //subnet_id      = element(aws_subnet.public.*.id, count.index)
  subnet_id        = var.subnet_cidr_public_ids[count.index]
  route_table_id = aws_route_table.public.id
}


/*resource "aws_route_table_association" "public-temporal" {
  subnet_id      = var.subnet_id
  route_table_id = aws_route_table.public.id
  depends_on     = [aws_route_table.public]
}

*/


output "rt_public_id" {
  //value = aws_route_table_association.public.id
  value       = aws_route_table_association.public[*].id
}