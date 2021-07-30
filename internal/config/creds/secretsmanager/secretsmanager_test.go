package secretsmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type APIMock struct {
	secretsmanageriface.SecretsManagerAPI
	SecretValues map[string]string
}

func (s *APIMock) GetSecretValue(i *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	secret := s.SecretValues[*i.SecretId]
	return &secretsmanager.GetSecretValueOutput{SecretString: &secret}, nil
}

func TestGet(t *testing.T) {
	s := APIMock{SecretValues: map[string]string{
		"testUsername": "username",
		"testPassword": "password",
	}}

	testGet(s, "testUsername", "username", t)
	testGet(s, "testPassword", "password", t)
}

func testGet(s APIMock, id, exp string, t *testing.T) {
	u, err := Get(&s, id)
	if err != nil {
		t.Errorf("unexpected error returned getting %s", id)
	}
	if u != exp {
		t.Errorf("unexpected value '%s' returned: expected '%s'", u, exp)
	}
}
