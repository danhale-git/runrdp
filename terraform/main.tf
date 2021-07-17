terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

locals {
  my_cidr  = "81.97.232.136/32"
  vpc_cidr = "172.31.0.0/16"
}

provider "aws" {
  profile = "default"
  region  = "eu-west-2"
}

resource "aws_instance" "rdp_target" {
  ami           = "ami-0b75abf94620286c5"
  instance_type = "t2.micro"

  tags = {
    Name = "rdp-target"
  }

  vpc_security_group_ids = [aws_security_group.allow_rdp.id]
  key_name               = "VPC"

  get_password_data = true
}

resource "aws_instance" "rdp_proxy" {
  ami           = "ami-03ac5a9b225e99b02"
  instance_type = "t2.micro"

  tags = {
    Name = "rdp-proxy"
  }

  vpc_security_group_ids = [aws_security_group.allow_rdp.id]
  key_name               = "VPC"

  // https://www.bogotobogo.com/DevOps/Terraform/Terraform-terraform-userdata.php
  // https://superuser.com/questions/1531047/using-a-linux-server-to-forward-an-rdp-connection-between-lan-and-vpn-interface
  user_data = <<EOT
#!/bin/bash
sudo yum install -y socat

while true; do socat tcp4-listen:3389,reuseaddr tcp:${aws_instance.rdp_target.private_dns}:3389; done
	EOT
}


resource "aws_security_group" "allow_rdp" {
  name        = "allow_rdp"
  description = "Allow RDP inbound traffic"

  ingress {
    description = "RDP"
    from_port   = 3389
    to_port     = 3389
    protocol    = "TCP"
    cidr_blocks = [local.my_cidr, local.vpc_cidr]
  }

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "TCP"
    cidr_blocks = [local.my_cidr]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = "allow_rdp"
  }
}

output "rdp_target_password" {
  value = aws_instance.rdp_target.password_data
}

resource "local_file" "config" {
  content = templatefile("config.tpl",
    {
      region        = "eu-west-2"
      instance_name = aws_instance.rdp_target.tags["Name"]
      proxy_address = aws_instance.rdp_proxy.public_ip
    }
  )
  filename = "C:/Users/danha/.runrdp/terraformconfig.toml"
}