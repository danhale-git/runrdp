[settings.default]
    fullscreen = false
    width = 800
    height = 600

[cred.awssm.testawssm]
    usernameid = "${secret_u}"
    passwordid = "${secret_p}"

[host.awsec2.testec2hostgetcred]
    id = "${instance_id}"
    getcred = true
    profile = "default"
    region = "${region}"

[host.awsec2.testec2hostawssm]
    id = "${instance_id}"
    profile = "default"
    region = "${region}"
    cred = "testawssm"

[host.awsec2.testproxyec2host]
    id = "${instance_id}"
    getcred = true
    profile = "default"
    region = "${region}"
    proxy = "testproxy"

[host.awsec2.testtunnelec2host]
    id = "${instance_id}"
    getcred = true
    profile = "default"
    region = "${region}"
    private = true
    tunnel = "testtunnel"

[host.basic.testproxy]
    address = "${proxy_address}"

[tunnel.testtunnel]
    host = "testproxy"
	localport = "55389"
	key = "C:/Users/danha/.ssh/VPC"
	User = "ec2-user"

[host.awsec2.testec2hostfilter]
    getcred = true
    profile = "default"
    region = "${region}"
    proxy = "testproxy"
    filterjson = """
    [
      {
        "Name": "tag:Name",
        "Values": ["rdp-target"]
      }
    ]
    """