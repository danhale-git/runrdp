package thycotic

import (
	"testing"

	"github.com/danhale-git/tss-sdk-go/server"
)

type ThycoticMockServer struct {
	Fields []server.SecretField
}

func (s *ThycoticMockServer) Secret(_ int) (*server.Secret, error) {
	return &server.Secret{Name: "testSecret", Fields: s.Fields}, nil
}

func TestGetCredentials(t *testing.T) {
	uw := "Username_value"
	pw := "Password_value"

	s := ThycoticMockServer{
		Fields: []server.SecretField{
			{
				FieldName:  "Username",
				ItemValue:  uw,
				Slug:       "Username",
				IsPassword: true,
			},
			{
				FieldName:  "Password",
				ItemValue:  pw,
				Slug:       "Password",
				IsPassword: true,
			},
		},
	}

	u, p, err := GetCredentials(&s, 0)
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}

	if u != uw {
		t.Errorf("unexpected username value '%s': wanted %s", u, uw)
	}

	if p != pw {
		t.Errorf("unexpected username value '%s': wanted %s", p, pw)
	}

	s.Fields = s.Fields[:1]

	u, p, err = GetCredentials(&s, 0)
	if err == nil {
		t.Fatalf("no error returned when Password field is missing")
	}
}
