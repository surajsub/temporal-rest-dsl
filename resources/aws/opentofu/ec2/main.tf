
resource "aws_key_pair" "temporal" {
  key_name   = var.key_name
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDKfoaw2sb/gQpvY3ctk8iGmdhFqMhpSCYZ0cd4GoBujgSZfcWtA9qFwrmZy2snihSE+Nv3hZJi4aGE5SC5UTas1OuHaH2MDQzmgE5x9Cz1L5oy3u4rW/lYtBIIXN/xHPPNkjeYYgQ4lVhHDi/EclPRihS4k7pzOR+S9m8mgqagqUSn5Q1kOOrIljy7oY73lHSEJrLL1nwoPy65jhpYJMDezj5y1yDXVP4mMpPfGDTMMUUnV2AH/YEe8Aswx4jqRShLpo4gUDGFOW9/3DQQ9G884iqezvJv/hjTG2oLjtUD2or8mT9yBNmz+nUj2v9Xyv1pngpCx38Q8t/bLggmxvDKEy+xYJxhjm3VQ1lJehHUIt6/NWKKab9cWV5i4B0OdYQeZ1lq9YcmodhLOOFnEBqcs8ttUyQnuHw8P8h5Wbk42Wbrf8AKUd/xNJomgpkMcAbyVsgAOShpBmopMYKtajafDJKbGGVxBkefPQ6BUDpZsYIGOZJ2zLaSBBhO3B/OqaE= suraj@Surajs-Macbook-Pro.local"
}

resource "aws_network_interface" "temporal" {
    count = var.ec2_count
   subnet_id        = var.subnet_id[count.index]
  security_groups = [var.sg_id]
}


resource "aws_instance" "temporal" {
  count = var.ec2_count
  ami                    = var.ami
  instance_type          = var.instance_type
  // subnet_id              = var.subnet_id
  subnet_id        = var.subnet_id[count.index]
  key_name               = aws_key_pair.temporal.key_name
  vpc_security_group_ids = [var.sg_id]



  associate_public_ip_address = true
  tags = {

    Name = "suraj-temporal-${count.index + 1}"


  }
  user_data = file("user_data/user_data.tpl")

}




output "instance_id" {
  value = aws_instance.temporal[*].id
}

output "instance_public_ip" {
  value = aws_instance.temporal[*].public_ip
}