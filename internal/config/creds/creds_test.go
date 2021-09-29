package creds

import "testing"

func TestSecretsManagerStruct(t *testing.T) {
	var i interface{} = SecretsManagerStruct()

	if _, ok := i.(*SecretsManager); !ok {
		t.Errorf("SecretsManagerStruct return value cannot be cast to a SecretsManager struct")
	}
}

func TestSecretsManager_Validate(t *testing.T) {
	// Validate not yet implemented
}
