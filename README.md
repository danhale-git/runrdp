![Go Report Card](https://goreportcard.com/badge/github.com/danhale-git/runrdp)
![golangci-lint](https://github.com/danhale-git/runrdp/actions/workflows/golangci-lint.yaml/badge.svg)
![test](https://github.com/danhale-git/craft/actions/workflows/go-test.yaml/badge.svg)
![coverage](https://img.shields.io/badge/coverage-36.5%25-orange)

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
-------
# Configuration Reference

Configuration is in TOML format. All config entries consist of a heading and set of key/value assignments.
Entries take the format:
```toml
[<config type>[.sub type].<name>]
stringkey   = "value"
intkey      = 0
boolkey     = false
```

Not all config types have sub types. `<name>` is the label chosen and referenced by the user.

### [settings]
Settings are the RDP session settings, mostly relating to window size. Naming an entry `[settings.default]` will make it the default for all hosts that don't explicitly reference another settings entry.
```toml
[settings.mysettings]
height      = 800   # Height of the window in pixels
width       = 600   # Width of the window in pixels
fullscreen  = false # Start the session in full-screen mode (might still start in full-screen if false)
span        = false # Span multiple monitors with the setting
```

## Hosts

#### Global Fields
Global fields may be defined for hosts of any sub type (`[host.<sub type>.myhost]`) and will override values given by that host sub type. For example the EC2 sub type obtains the IP address from AWS. The `address` global field would override that IP. 
```toml
[host.<any host type>.myhost]
cred        = "mycred"          # Reference to a cred config entry used to authenticate (e.g. [cred.thycotic.mycred])
proxy       = "myproxy"         # Reference to a host config entry
address     = "1.2.3.4"         # Literal address for the RDP endpoint
port        = "1234"            # Literal port for the RDP endpoint
username    = "Administrator"   # Literal username for RDP authentication
tunnel      = "mytunnel"        # Reference to a tunnel config entry used to start an SSH tunnel (e.g. [cred.tunnel.mytunnel])
settings    = "mysettings"      # Reference to a settings config entry to define RDP settings (e.g. [settings.mysettings])
```

### [host.basic]
Basic does not have any fields, only global fields may be defined. A literal address must be given in order to connect to a host.
```toml
[host.basic.mybasichost]
address = "1.3.4.5" # This is a global field (see Global Fields), defined here as an example
```

### [host.ec2]

```toml
[host.ec2.myec2host]
private = true      # Connect to the private IP address of this EC2 host
getcred = true      # Call the AWS EC2 _Get Password_ feature to get credentials for RDP authentication
id = "i-abcde1234"  # Locate the EC2 host by instance ID
profile = "default" # AWS Shared Credentials profile to use for authentication
region = "eu-west"  # AWS region in which to operate

includetags = ["Name;rdp-target","env;dev"] # Locate the EC2 host by filtering for these tags
excludetags = ["env;prod"]                  # Filter out any hosts with these tags
```



// EC2 defines an AWS EC2 instance to connect to by getting it's address from the AWS API.
type EC2 struct {
Private     bool
GetCred     bool
ID          string
Profile     string
Region      string
IncludeTags []string
ExcludeTags []string

	svc ec2iface.EC2API
}

_______
## Configuration Guide

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
