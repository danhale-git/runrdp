package secretsmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type SecretsManagerMockService struct {
	SecretValues map[string]string
}

func (s *SecretsManagerMockService) GetSecretValue(i *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	secret := s.SecretValues[*i.SecretId]
	return &secretsmanager.GetSecretValueOutput{SecretString: &secret}, nil
}

func TestGet(t *testing.T) {
	s := SecretsManagerMockService{SecretValues: map[string]string{
		"testUsername": "username",
		"testPassword": "password",
	}}

	testGet(s, "testUsername", "username", t)
	testGet(s, "testPassword", "password", t)
}

func testGet(s SecretsManagerMockService, id, exp string, t *testing.T) {
	u, err := Get(&s, id)
	if err != nil {
		t.Errorf("unexpected error returned getting %s", id)
	}
	if u != exp {
		t.Errorf("unexpected value '%s' returned: expected '%s'", u, exp)
	}
}
