# RunRDP

RunRDP is a tool for configuring Windows Remote Desktop hosts in text files and starting RDP sessions from the command line.
It is not a standalone RDP client, it runs the default RDP client.

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
