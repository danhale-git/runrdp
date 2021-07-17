[host.awsec2.target]
    getcred = true
    profile = "default"
    region = "${region}"
    includetags = ["Name;${instance_name}"]
    proxy = "targetproxy"

[host.basic.targetproxy]
    address = "${proxy_address}"