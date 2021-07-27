[settings.default]
    fullscreen = false
    width = 800
    height = 600

[cred.awssm.testawssm]
    usernameid = "${secret_u}"
    passwordid = "${secret_p}"

[host.awsec2.testec2hostgetcred]
    getcred = true
    profile = "default"
    region = "${region}"
    includetags = ["Name;${instance_name}"]

[host.awsec2.testec2hostawssm]
    profile = "default"
    region = "${region}"
    includetags = ["Name;${instance_name}"]
    cred = "testawssm"

[host.awsec2.testproxyec2host]
    getcred = true
    profile = "default"
    region = "${region}"
    includetags = ["Name;${instance_name}"]
    proxy = "testproxy"

[host.awsec2.testtunnelec2host]
    getcred = true
    profile = "default"
    region = "${region}"
    includetags = ["Name;${instance_name}"]
    private = true
    tunnel = "testtunnel"

[host.basic.testproxy]
    address = "${proxy_address}"

[tunnel.testtunnel]
    host = "testproxy" 
	localport = "55389" 
	key = "C:/Users/danha/.ssh/VPC"
	User = "ec2-user"