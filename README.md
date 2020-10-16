# RunRDP
RunRDP is a tool for running RDP sessions from the command line and configuring Desktops in a TOML file. It is not a stand alone RDP client, it runs your default RDP client.

The tool works by writing a .rdp file, running it then deleting it immediately.

The goal of the tool is to enable powerful provide specific Desktop configuration in text files and easy initiation of RDP connections from command line.

Desktops can be defined in a configuration file or command line flags.
    
    # connect to the myDesktop configuration entry
    runrdp myDesktop
    
    # connect to an IP address or hostname with the given credentials
    runrdp 123.654.321.456 -u Administrator -p 'pa$$w0rd'

The following providers/features are supported:

* SSH Tunnel support
       
        [host.basic.mytunnelhost]
            address = "123.456.789.123"
            tunnel = "mytunnel" # reference to tunnel config 
            
        [tunnel.mytunnel]
            host = "bastion" # reference to bastion host config 
            localport = "3390"
            key = "C:/Users/myuser/.ssh/id_rsa"
            user = "ubuntu"

        [host.awsec2.bastion]
            id = "i-0f4bd11234be468f6"
            private = false
            profile = "default"


#### AWS
* Define RDP Desktops by AWS tag filters

        [host.awsec2.myhost]
            private = false # Use public IP
            getcred = true # Get instance password from AWS
            profile = "default"
            region = "eu-west-2"
            includetags = ["Name;MyInstanceName", "env;production"] # Filter for instance using tags

* Supports EC2 feature to get initial administrator password using SSH Key.

        # connect to the AWS instance with the given ID, getting the password from AWS
        runrdp aws -i i-02d88385685f8a739 --awspass
    
* Secrets Manager support.

        # define RDP credentials by Secrets Manager secret key.
        [cred.awssm.mycred]
            usernameid = "MyUsernameSecret"
            passwordid = "MyPasswordSecret"
            region = "eu-west-2"
            profile = "default"
       
