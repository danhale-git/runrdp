# RunRDP
RunRDP is a tool for running RDP sessions from the command line. It is not a stand alone RDP client, it runs your default RDP client.

The tool works by writing a .rdp file, running it then deleting it immediately.

Desktops can be defined in a configuration file or command line flags.
    
    # connect to the myDesktop configuration entry
    runrdp myDesktop
    
    # connect to an IP address or hostname with the given credentials
    runrdp 123.654.321.456 -u Administrator -p 'pa$$w0rd'

The following providers/features are supported:
#### AWS
* Supports EC2 feature to get initial administrator password using SSH Key.

        
        # connect to the AWS instance with the given ID, getting the password from AWS
        runrdp aws -i i-02d88385685f8a739 --awspass
    
* Secrets Manager support.
