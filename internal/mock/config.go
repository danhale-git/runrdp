package mock

const Config = `[cred.awssm.awssmtest]
    usernameid = "TestInstanceUsername"
    passwordid = "TestInstancePassword"
    region = "eu-west-2"
    profile = "default"

[host.awsec2.awsec2test]
    tunnel = "mytunnel"
    private = true
    cred = "testcred"
    profile = "TESTVALUE"
    region = "eu-west-2"
    includetags = ["mytag;mytagvalue", "Name;MyInstanceName"]

[host.basic.basictest]
    address = "35.178.168.122"
    cred = "testcred"

[tunnel.tunneltest]
    host = "myiphost"
    localport = "3390"
    key = "C:/Users/danha/.ssh/vpc_key"
    user = "ubuntu"
`
