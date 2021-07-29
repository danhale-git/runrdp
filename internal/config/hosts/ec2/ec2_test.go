package ec2

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/danhale-git/runrdp/internal/mock"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type APIMock struct {
	ec2iface.EC2API
	Instances []*ec2.Instance
}

// TestGetInstance

func TestGetInstance(t *testing.T) {
	svc := APIMock{Instances: mock.InstancesWithNames("instance1", "instance2", "instance3")}

	expectedLen := len(svc.Instances)

	instances, err := GetInstances(&svc, "", "")
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}

	if len(instances) != expectedLen {
		t.Errorf("unexpected number of instances returned: expected %d: got %d", expectedLen, len(instances))
	}
}

func (e *APIMock) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	var r []*ec2.Reservation

	// Try to return multiple reservations to test that it is supported
	if len(e.Instances) > 1 {
		r = []*ec2.Reservation{
			{
				Instances: e.Instances[:1],
			},
			{
				Instances: e.Instances[1:],
			},
		}
	} else {
		r = []*ec2.Reservation{{Instances: e.Instances[:1]}}
	}

	return &ec2.DescribeInstancesOutput{Reservations: r}, nil
}

// TestGetPassword

func TestGetPassword(t *testing.T) {
	svc := &APIMock{}
	got, err := GetPassword(svc, "", testPrivateKey())

	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
	}

	want := testPassword()
	if got != want {
		t.Errorf("unexpected value returned: want %s: got %s", want, got)
	}
}

func (e *APIMock) GetPasswordData(_ *ec2.GetPasswordDataInput) (*ec2.GetPasswordDataOutput, error) {
	return &ec2.GetPasswordDataOutput{
		PasswordData: aws.String(`ITPAqsegJEBl8zUr6e5t80mGKdueK6nFusgfaUMp5IA5gaJfqnp5Q3w/HkKsffC7twg9EhwOfLElUU5U7
AOhJLVAKKK/6g1hu8Qo7w+ZY03qqvYJIlWltS/VHBSgAfPudJ6oQCQXfZtkc4vak2d0ttYY2nUuN/oVVU8eT744eFU0fQkMEMaVpA0GdIPyI5A9Qtm/o6V
pU+nI60J88l2vCcIgPSIPuwf1NkcJxUbB02VbXtCyU5znzXxIHuRYpLrmNwgTiFMFC1MfFrKYEEjMYrVsiklEGBJXJDfubLLqulUUIGuEAzceesHgdtRFI
rcJfX78bOwI67aVe0b2TaorzQ==`),
	}, nil
}

func testPassword() string {
	return "gFVI.q-;IkmI8Hi62lVqYx8Gg8KmCzW?"
}

func testPrivateKey() []byte {
	return []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAytINa2Bvkj5ExQBD5byp5xrf4J1lgCSSoqEhJJ7zL4pUM2dw
zdNobYnR4b8Uw+DRazqfmeFqb6PYWNdH8pTlTXk6mLpEXTI3FaWI4QXq/c6L/vwt
qJGHTeMzWBzKN83K6antk3yMGKzraztIm2RlYTWie1tci9Lo9G3sHkmMDv7DHKxy
RYa0AcLXcRMrWm5kgoOrWxFvJ4Qez63OZNV5P54MgFMsnVh0lPZf0pY1VKqqLzOG
/dPSlfTjVofU0eOb3Pyi0nMHz/K7bVLRX4uFamyHwHttWn30EpZ1HxhmrNkKOBYf
XWrpmN6gqmPdBVZevzh+cWKPhA3tZzJbaqccRwIDAQABAoIBAHmWIuVUEo6hNajD
1/BJgbFBsyR8NvTy99T2s1+4yiCd1IxcXouFSP0hueiTHGewxtp5cmRVdfEnT/My
W7dY+33ORwp337Pe/pbDfaMoYQ92WwapNtXvKCXRJl6UI8YAYLxjWkEoUPQZ9ad7
CrPdI8l61cUNqBVKgszFeN9PR99UWAwQ3x+bejBpfJ8BytKhgPr0JNdJFJni5uTt
WpYmLS84UDVCo4juaLt7DFEALKy1D3zorhlvkbzNFAUDhgO0jmJm16QSSEbla3kO
1vth0Ix139E/rAP9RB0a10SmNlocWKY0TJZSzygQqAAxPg15izbDz2Pv8hPkFeKQ
GcRF6UECgYEA/JRB+u3DuFej7bPMqFYZcRFNIbfbYMVSk7HMpJJoS/zQRQH5KIsV
XZE3FyQeACFvbIw9f1gchM/VuJOv94uq7PrFJrenBkNUG/HV4ns6X+V/4cZjwxAj
CWSUDQFQlaYpY7OrHvB6nA12N5gWzspdTJPJNbIlX9GRgtwg3Ct8vy0CgYEAzZFF
ifMInTxQxEGWGWCEkoHppn55e/05nj+uz54wem1bQg/fxVaWy+LUcvukRWFNF+IW
Oa0iJTCgcFaTm/aNo1FsbvMQG92+3+wfDPSGB96lkfctNjrdZGIDhmfsO/2L0hRc
6LC1d4p4hHw3fG/ipsjDU4oFUXN5saeLFUtdkcMCgYBZBAMw5UTaHgEHEBvro9R5
lchiPsLRGxncNYhS48pgJWxdNbHTCRlxjXEl9bOhBieX0OEHlU0PvZOr5ljY3F9T
/5kl6QmzWl01MAjaNeW/0Ek+j8WvBGvkro7C+pik9ReXLMX9NHFxuAjW1QIMxSMW
jusVwoALgfdPcDcggS8IzQKBgBga8eGUSy1M9ledLUG6jLE1ZLWuXQaKEiiZZSFZ
dmvUyP+9JstYNQShi7IUChZMq6KiU2LeB4P+6MFjlZmTVtaQ5Ls562ipHwnZAWce
gV0I4bd1GasjSfTMfYdURmJef/fZhW+P0Se8aBd5DXSdFiHipuzz4V3Ewb9wWyHb
HZTLAoGBAIldvFbyeDuquYlXCgOR9QU/1YtTc5aHaxDbm4dxAp8HbD2mZhaaMvVW
Z6ovPDMDlCv3VHlUlLTi5WlYpcrxJBUNyeTHbxOTr1HWE18cbrcAWrbqbdevUWUi
u+eYWXZMNchaUW+a52/0KE7Q2pYZm70LiY41bDEe26epRk5+Dg4I
-----END RSA PRIVATE KEY-----`)
}

// TestChooseInstance

func TestChooseInstance(t *testing.T) {
	instances := mock.InstancesWithNames("a", "b", "c")

	buf := bytes.NewBuffer([]byte{})

	buf.Write([]byte("1\n"))

	i, err := ChooseInstance(instances, buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if i == nil {
		t.Fatal("nil instances returned with no error")
	}

	if *GetTag(i, "Name") != "a" {
		t.Errorf("unexpected instance '%s' returned: expected '%s'", *GetTag(i, "Name"), "a")
	}

	buf.Write([]byte("invalidinput\n"))

	i, _ = ChooseInstance(instances, buf)
	if i != nil {
		t.Fatalf("instance was not nil for invalid input")
	}

	// ChoseInstances prints a string that doesn't end in a newline
	fmt.Println()
}
