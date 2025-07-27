resource "aws_lb_target_group" "temporal" {
  name     = "application-front"
  port     = 80
  protocol = "HTTP"
  vpc_id   = var.vpc_id
  health_check {
    enabled             = true
    healthy_threshold   = 3
    interval            = 10
    matcher             = 200
    path                = "/"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 3
    unhealthy_threshold = 2
  }
}

resource "aws_lb_target_group_attachment" "attach-app1" {
  count            = var.ec2_count
  target_group_arn = aws_lb_target_group.temporal.arn
  target_id = var.instance_id[count.index]
  port             = 80
}



resource "aws_lb" "temporal" {
  name               = "front"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [var.sg_id]
  subnets            = var.subnet_id

  enable_deletion_protection = false

  tags = {
    Environment = "front"
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.temporal.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.temporal.arn
  }
}


output "lb_dns" {
  value = aws_lb.temporal.dns_name
}

output "lb_arn" {
  value = aws_lb.temporal.arn
}