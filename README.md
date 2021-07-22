![Go Report Card](https://goreportcard.com/badge/github.com/danhale-git/runrdp)
![golangci-lint](https://github.com/danhale-git/runrdp/actions/workflows/golangci-lint.yaml/badge.svg)
![test](https://github.com/danhale-git/craft/actions/workflows/go-test.yaml/badge.svg)
![coverage](https://img.shields.io/badge/coverage-72.3%25-yellow)

# RunRDP

RunRDP is a tool for launching MS RDP sessions from the command line based on a text configuration. It is not a standalone RDP client.

## Features
* SSH tunnel (SSH port forwarding) and proxy support
* Thycotic Secret Server credential storage
* AWS Secrets Manager credentials storage
* AWS EC2 integration
    * Define instances by inclusive/exclusive tag set or instance ID
    * AWS authentication from shared credentials file
    * EC2 _Get Password_ for RDP authentication


Configuration parsing and command line interface use github.com/spf13/cobra and github.com/spf13/viper.
_______
Configuration is in TOML format.

This is a host configuration entry of type AWS EC2 named 'myhost'.

    [host.awsec2.myhost]
        private = false                                     # Connect to the public IP of the EC2 instance
        profile = "dev"                                     # The AWS credentials profile name
        includetags = ["Name;my-instance-name", "keyonly"]  # The default key/value separator is ';'
        
The instance IP and credentials are obtained at runtime using the AWS profile 'dev'.
Call it with `runrdp myhost` or `runrdp find <pattern>` for a fuzzy search of hosts.
All host configurations below start authenticated sessions.

A host configuration of type basic (which is )just an IP address) named 'bastion'.

    [host.basic.bastion]
        address = "1.2.3.4"
        proxy = "myiphost"
        cred = "mycred"
        
The 'cred' field above refers to a credential entry of type AWS Secrets Manager.
        
    [cred.awssm.mycred]
        usernameid = "TestInstanceUsername" # The username to authenticate with
        passwordid = "TestInstancePassword" # The password to authenticate with
        region = "eu-west-2"                # If omitted the profile default region will be used
        profile = "dev"                     # The AWS credentials profile name
        
This is an SSH tunnel (SSH Port Forwarding) configuration entry.

`ssh -i <key> -N -L <localport>:<address>:<port> <username>@<host address>`
(address and port come from the host declaring the tunnel)

    [tunnel.mytunnel]
        host = "bastion"                # The intermediate host forwarding the connection.
        localport = "3390"              # The port to connect to locally.
        key = "C:/Users/me/.ssh/id_rsa" # SSH key for the intermediate host.
        user = "ubuntu"                 # SSH user for the intermediate host.

Tunneling to myhost via bastion. Tunnel (above) is declared in myhost and refers to bastion as the intermediate host.

    [host.awsec2.myhost]
        tunnel = "mytunnel"                                 # Open an SSH tunnel before connecting 
        private = true                                      # Use the private IP address to connect
        profile = "dev"
        includetags = ["Name;my-instance-name", "keyonly"]
        
Using bastion as a proxy to connect to myhost.
        
    [host.awsec2.myhost]
        proxy = "bastion"                                   # Connect to the IP of bastion
        profile = "dev"
        includetags = ["Name;my-instance-name", "keyonly"]
