![Go Report Card](https://goreportcard.com/badge/github.com/danhale-git/runrdp)
![golangci-lint](https://github.com/danhale-git/runrdp/actions/workflows/golangci-lint.yaml/badge.svg)
![test](https://github.com/danhale-git/craft/actions/workflows/go-test.yaml/badge.svg)
![coverage](https://img.shields.io/badge/coverage-66.0%25-yellow)

# RunRDP

RunRDP is a tool for launching MS RDP sessions from the command line based on a text configuration. It is not a standalone RDP client.

## Features
* SSH tunnel (SSH port forwarding) and proxy support
* Remote Secret Store integration
    * Thycotic Secret Server
    * AWS Secrets Manager
* AWS EC2 integration
    * Identify instances by ID or tag filter
    * Authenticate using shared credentials
    * EC2 _Get Password_ for RDP authentication
-------
# Configuration Reference

Configuration is in TOML format. All config objects consist of a heading and key/value pairs:
```
# Format
[<config type>[.<sub type>].<name>]
  string   = "abc"
  int      = 0
  bool     = false
```
`<name>` is the user defined label used to reference the object. Below is a host of type EC2.
```toml
# Example
[host.ec2.myhost]
  id = "i-abcde1234"
```

# Hosts
Hosts are remote computers. The label given to a host configuration object is used on the command line when connecting.
```toml
[host.ec2.myhost]
  id = "i-abcde1234"
```
```bash
$ runrdp myhost
```

## References To Other Objects 
Hosts commonly reference other configuration objects such as credentials or RDP settings.

### cred
Refers to a Credentials object used for RDP authentication.
```toml
[cred.<type>.mycred]
    # See Credential Types section below

[host.ec2.myhost]
    cred = "mycred"
    id = "i-abcde1234"
```

### proxy
Refers to another host configuration for a computer which is proxying RDP connections.
```toml
### proxy
[host.<type>.myproxyhost]
    # See Host Types section below

[host.ec2.myhost]
    proxy = "myproxyhost"
    id = "i-abcde1234"
```

### settings
Refers to an RDP session settings object to configure windows size. Naming a settings object `[settings.default]` will make it the default for hosts omitting the settings field. System defaults are used if no settings are defined.
```toml
[settings.mysettings]
  height      = 800   # Height of the window in pixels
  width       = 600   # Width of the window in pixels
  fullscreen  = false # Start the session in full-screen mode (might still start in full-screen if false)
  span        = false # Span multiple monitors with the setting

[host.ec2.myhost]
  settings = "mysettings"
  id = "i-abcde1234"
```

### tunnel
An intermediate host used for SSH forwarding. All fields must be defined.
```toml
[tunnel.mytunnel]
  host = "myhost"                 # Reference to a host config object used as the intermediate forwarding host
  localport = "3390"              # Port to connect to locally over localhost
  key = "C:/Users/me/.ssh/key"    # Full path to the SSH key used for authentication
  user = "ubuntu"                 # SSH Username for authentication

[host.<type>.myhost]
  mytunnel = "mytunnel"
```

## Literal Global Fields
These take precedence when conflicting with another configuration field.
```toml
[host.<any host type>.myhost]
  # Literal values which take precedence in the configuration
  address     = "1.2.3.4"         # Address for the RDP endpoint
  port        = "1234"            # Port for the RDP endpoint
  username    = "Administrator"   # Username for RDP authentication
```

## Host Types
Host types offer specific functionality with the exception of the basic host type which only uses global fields.

### host.basic
Basic does not have any fields, only global fields may be defined. A literal address is required.
```toml
[host.basic.mybasichost]
  address = "1.3.4.5" # This is a global field (see Global Fields), defined here as an example
```

### host.ec2
EC2 instance to connect to by getting its address from the AWS API. Either an ID or filterJSON is required unless the global _address_ field is defined.
```toml
[host.ec2.myec2host]
  private = true      # Connect to the private IP address of this EC2 host
  getcred = true      # Call the AWS EC2 _Get Password_ feature to get credentials for RDP authentication
  id = "i-abcde1234"  # Locate the EC2 host by instance ID
  profile = "default" # AWS Shared Credentials profile to use for authentication
  region = "eu-west"  # AWS region in which to operate

  # https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instances.html#options
  filterjson = """
      [
        {
          "Name": "tag:Name",
          "Values": ["rdp-target"]
        }
      ]
      """
```

## Credential Types

### cred.thycotic
Retrieve a secret based on its ID. The secret being retrieved from Thycotic must have _Username_ and _Password_ fields.
```toml
[cred.thycotic.mycred]
  secretid = "12345"

# This object must be defined to use thycotic
[thycotic.settings]
  thycotic-url = "testthycotic-url"         # URL of the Thycotic service
  thycotic-domain = "testthycotic-domain"   # Optional Active Directory domain name
```

### cred.awssm
Retrieve a username and password from AWS Secrets Manager. The _username_ and _password_ fields must be the ID of AWS Secrets Manager secrets of type string.
```toml
[cred.awssm.mycred]
  usernameid = "MyUsername" # The username to authenticate with
  passwordid = "MyPassword" # The password to authenticate with
  region = "eu-west-2"      # If omitted the profile default region will be used
  profile = "dev"
```