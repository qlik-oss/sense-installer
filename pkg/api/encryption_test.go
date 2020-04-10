package api

import (
	"testing"
)

func Test_encrypt_decrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	testData := "this is a secret value"
	enc, err := EncryptData([]byte(testData), key)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	dec, err := DecryptData(enc, key)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if testData != string(dec) {
		t.Log("expected: " + testData)
		t.Log("actual: " + string(dec))
		t.Fail()
	}
}
