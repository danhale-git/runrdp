package creds

import "testing"

func TestThycoticStruct(t *testing.T) {
	var i interface{} = ThycoticStruct()

	if _, ok := i.(*Thycotic); !ok {
		t.Errorf("ThycoticStruct return value cannot be cast to a Thycotic struct")
	}
}

func TestThycotic_Validate(t *testing.T) {
	zeroID := Thycotic{SecretID: 0}
	if err := zeroID.Validate(); err == nil {
		t.Errorf("no error returned when Thycotic.SecretID is 0")
	}
}
