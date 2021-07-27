terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}


variable my_cidr {
  type = string
}
variable vpc_cidr {
  type = string
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

  vpc_security_group_ids = [aws_security_group.rdp_ssh.id]
  key_name               = "VPC"

  get_password_data = true
}

resource aws_secretsmanager_secret rdp_target_username {
  name = format("%sUsername", aws_instance.rdp_target.tags["Name"])

  provisioner "local-exec" {
    command = format("aws secretsmanager put-secret-value --secret-id %s --secret-string %s",
      format("%sUsername", aws_instance.rdp_target.tags["Name"]),
      "Administrator"
    )
  }

  depends_on = [
    aws_instance.rdp_target
  ]
}

resource aws_secretsmanager_secret rdp_target_password {
  name = format("%sPassword", aws_instance.rdp_target.tags["Name"])

  provisioner "local-exec" {
    command = format("aws secretsmanager put-secret-value --secret-id %s --secret-string %s",
      format("%sPassword", aws_instance.rdp_target.tags["Name"]),
      rsadecrypt(aws_instance.rdp_target.password_data, file(pathexpand("~/.ssh/VPC"))),
    )
  }

  depends_on = [
    aws_instance.rdp_target
  ]
}

resource "aws_instance" "rdp_proxy" {
  ami           = "ami-03ac5a9b225e99b02"
  instance_type = "t2.nano"

  tags = {
    Name = "rdp-proxy"
  }

  vpc_security_group_ids = [aws_security_group.rdp_ssh.id]
  key_name               = "VPC"

  // https://www.bogotobogo.com/DevOps/Terraform/Terraform-terraform-userdata.php
  // https://superuser.com/questions/1531047/using-a-linux-server-to-forward-an-rdp-connection-between-lan-and-vpn-interface
  user_data = <<EOT
#!/bin/bash
sudo yum install -y socat

while true; do socat tcp4-listen:3389,reuseaddr tcp:${aws_instance.rdp_target.private_dns}:3389; done
	EOT
}


resource "aws_security_group" "rdp_ssh" {
  name        = "rdp_ssh"
  description = "Allow RDP and SSH traffic"

  ingress {
    description = "RDP"
    from_port   = 3389
    to_port     = 3389
    protocol    = "TCP"
    cidr_blocks = [var.my_cidr, var.vpc_cidr]
  }

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "TCP"
    cidr_blocks = [var.my_cidr]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = "rdp_ssh"
  }
}

resource "local_file" "config" {
  content = templatefile("config.tpl",
    {
      secret_u = aws_secretsmanager_secret.rdp_target_username.name
      secret_p = aws_secretsmanager_secret.rdp_target_password.name
      region        = "eu-west-2"
      instance_name = aws_instance.rdp_target.tags["Name"]
      proxy_address = aws_instance.rdp_proxy.public_ip
    }
  )
  filename = pathexpand("~/.runrdp/terraformconfig.toml")
}

output "rdp_target_public_ip" {
  value = aws_instance.rdp_target.public_ip
}

output "rdp_target_private_dns" {
  value = aws_instance.rdp_target.private_dns
}

output "rdp_proxy_public_ip" {
  value = aws_instance.rdp_proxy.public_ip
}

output "rdp_proxy_private_dns" {
  value = aws_instance.rdp_proxy.private_dns
}
