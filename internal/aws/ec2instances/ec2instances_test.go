package ec2instances

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func TestTags_Filter(t *testing.T) {
	tags := Tags{
		IncludeTags: []string{"kv:kv_value", "k", "mk:val1", "mk:val2", "mk:val3"},
		ExcludeTags: []string{"exclude"},
		Separator:   ":",
	}

	gotFilters, err := tags.IncludeFilter()

	if err != nil {
		t.Errorf("error creating AWS filters from tag config values: %s", err)
	}

	want := "tag:kv"
	got := *gotFilters[0].Name

	if got != want {
		t.Errorf("key value filter is incorrect: want %s: got %s", want, got)
	}

	want = "tag-key"
	got = *gotFilters[1].Name

	if got != want {
		t.Errorf("key only filter is incorrect: want %s: got %s", want, got)
	}

	want = "kv_value"
	got = *gotFilters[0].Values[0]

	if got != want {
		t.Errorf("key only filter is incorrect: want %s: got %s", want, got)
	}

	gotLen := len(gotFilters[2].Values)

	if gotLen != 3 {
		t.Errorf("length of multi value filter is incorrect: want 3: got %d", gotLen)
	}
}

type mockEC2Client struct {
	// aws interface for mocking the aws api
	ec2iface.EC2API
	// DescribeInstances returns this function
	describe func(input *ec2.DescribeInstancesInput, t *testing.T) (*ec2.DescribeInstancesOutput, error)
	t        *testing.T
}

func (m mockEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	// return the function assigned to mockEC2Client.describe
	return m.describe(input, m.t)
}

func (m mockEC2Client) GetPasswordData(_ *ec2.GetPasswordDataInput) (*ec2.GetPasswordDataOutput, error) {
	return &ec2.GetPasswordDataOutput{
		PasswordData: aws.String(`ITPAqsegJEBl8zUr6e5t80mGKdueK6nFusgfaUMp5IA5gaJfqnp5Q3w/HkKsffC7twg9EhwOfLElUU5U7
AOhJLVAKKK/6g1hu8Qo7w+ZY03qqvYJIlWltS/VHBSgAfPudJ6oQCQXfZtkc4vak2d0ttYY2nUuN/oVVU8eT744eFU0fQkMEMaVpA0GdIPyI5A9Qtm/o6V
pU+nI60J88l2vCcIgPSIPuwf1NkcJxUbB02VbXtCyU5znzXxIHuRYpLrmNwgTiFMFC1MfFrKYEEjMYrVsiklEGBJXJDfubLLqulUUIGuEAzceesHgdtRFI
rcJfX78bOwI67aVe0b2TaorzQ==`),
	}, nil
}

func testDescribeSingle(input *ec2.DescribeInstancesInput, t *testing.T) (*ec2.DescribeInstancesOutput, error) {
	// test input filter
	if len(input.Filters) == 0 || *input.Filters[0].Name != "instance-id" {
		t.Errorf("InstanceFromID sent AWS a request with no instance filter")
	}

	// return a valid response
	return outputSingle(), nil
}

func testDescribeMulti(_ *ec2.DescribeInstancesInput, _ *testing.T) (*ec2.DescribeInstancesOutput, error) {
	// return multiple instances
	return outputMulti(), nil
}

func testDescribeTagged(_ *ec2.DescribeInstancesInput, _ *testing.T) (*ec2.DescribeInstancesOutput, error) {
	return outputTagged(), nil
}

func TestInstanceFromID(t *testing.T) {
	svc := mockEC2Client{
		t: t,
	}

	// normal input and response
	svc.describe = testDescribeSingle
	_, err := InstanceFromID(svc, "i-07dd0954800821234")

	if err != nil {
		t.Errorf("InstanceFromID returned error for valid input: %s", err)
	}

	// receive multiple instances in response
	svc.describe = testDescribeMulti
	_, err = InstanceFromID(svc, "i-07dd0954800821234")

	if err == nil {
		t.Errorf("InstanceFromID failed to return an error when receiving more than one instance")
	}
}

func TestInstanceFromTagFilter(t *testing.T) {
	svc := mockEC2Client{
		t: t,
	}

	// normal input and response
	svc.describe = testDescribeTagged

	tags := Tags{
		IncludeTags: []string{"includeKey", "includeKeyValue:includeKeyValue-val"},
		ExcludeTags: []string{"excludeKey", "excludeKeyValue:excludeKeyValue-val"},
		Separator:   ":",
	}

	instance, err := InstanceFromTagFilter(svc, tags)

	if err != nil {
		t.Errorf("InstanceFromTagFilter returned an error for valid input: %s", err)
	}

	if *instance.InstanceId != "want" {
		t.Errorf("InstanceFromTagFilter returned the wrong instance")
	}
}

func TestGetPassword(t *testing.T) {
	svc := mockEC2Client{}
	got, err := GetPassword(svc, "", testPrivateKey())

	if err != nil {
		t.Errorf("GetPassword returned an error for valid input: %s", err)
	}

	want := "gFVI.q-;IkmI8Hi62lVqYx8Gg8KmCzW?"
	if got != want {
		t.Errorf("GetPassword returned the incorrect value: want %s: got %s", want, got)
	}
}

func instance(tags []*ec2.Tag, id string) *ec2.Instance {
	// return a normal instance
	return &ec2.Instance{
		PrivateIpAddress: aws.String("10.0.0.3"),
		KeyName:          aws.String("KEYNAME3.pem"),
		InstanceId:       aws.String(id),
		Tags:             tags,
	}
}

func outputSingle() *ec2.DescribeInstancesOutput {
	// return a single instance
	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					instance(nil, "i-07dd0954800829101"),
				},
			},
		},
	}
}

func outputMulti() *ec2.DescribeInstancesOutput {
	// return 3 instances
	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					instance(nil, "i-07dd0954800829101"),
					instance(nil, "i-07dd0954800829101"),
					instance(nil, "i-07dd0954800829101"),
				},
			},
		},
	}
}

func outputTagged() *ec2.DescribeInstancesOutput {
	// return 3 instances
	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					instance([]*ec2.Tag{
						{Key: aws.String("neutral"), Value: aws.String("neutral-val")},
						{Key: aws.String("includeKey"), Value: aws.String("includeKey-val")},
						{Key: aws.String("excludeKey"), Value: aws.String("excludeKey-val")},
					},
						"bad"),

					instance([]*ec2.Tag{
						{Key: aws.String("neutral"), Value: aws.String("neutral-val")},
						{Key: aws.String("includeKey"), Value: aws.String("includeKey-val")},
						{Key: aws.String("includeKeyValue"), Value: aws.String("includeKeyValue-val")},
					},
						"want"),

					instance([]*ec2.Tag{
						{Key: aws.String("includeKey"), Value: aws.String("includeValue")},
						{Key: aws.String("excludeKey"), Value: aws.String("excludeKey-val")},
					},
						"bad"),

					instance([]*ec2.Tag{
						{Key: aws.String("includeKey"), Value: aws.String("includeKey-val")},
						{Key: aws.String("excludeKeyValue"), Value: aws.String("excludeKeyValue-val")},
					},
						"bad"),
				},
			},
		},
	}
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
