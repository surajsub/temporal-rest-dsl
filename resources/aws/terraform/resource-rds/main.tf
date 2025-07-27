
#resource "aws_db_subnet_group" "my_db_subnet_group" {
#  count = 2
#  name       = "my-db-subnet-group"
#  subnet_ids = var.subnet_ids
#
#  tags = {
#    Name = "My DB Subnet Group"
#  }
#}


resource "aws_db_instance" "default" {
  allocated_storage = 10
  storage_type      = "gp2"
  engine            = "mysql"
  engine_version    = "5.7"
  instance_class    = "db.t2.micro"
  identifier        = "mydb"
  username          = "dbuser"
  password          = "dbpassword"

  vpc_security_group_ids =  [var.sg_id]

  skip_final_snapshot = true
}
