package api

import (
	"encoding/base64"
	"log"
	"os"
	"testing"
)

func Test_generateRSAEncryptionKeys(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "valid case",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GenerateAndStoreSecretKeypair(os.TempDir()); (err != nil) != tt.wantErr {
				t.Errorf("generateRSAEncryptionKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_encryption_decryption(t *testing.T) {
	privKeyBytes := []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`)

	publicKeyBytes := []byte(`-----BEGIN RSA PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAwUCimKCidbF3UxEHPy8K
+hvhklRB9JYhj5sJy0if4lTVibkK1MrYCykOnmC40pPU9GLY1b8HxAg9tvyRn0YH
UxOra6vVQaVcOVJhTM8D18d+lSr3Lp1yiX+UGT4nzWI9+R1CCbwXrqeQVoZs6QZK
ynEXMkFI9/wNMOwPOvQFOSTuoEoCO+zyTyUWEkNbUq825ELUQdIsjgmlWUOONudx
sAr7ESRXW9QTHVh6uWmr3VRKZHby1JdU3I/wjdlGg5M2dDuXy5nQO9w/nYLjJXiw
+zzOetZ/+t7/VOkOpNTeJQhwTM1WF7Y2VLetbi9FHgyzHatrduh07+XEiTbgDf3G
Ix2bp2p6oh0G3N2zpiLcK/aZj8rouWWydfFfsU3MZ4FfJDP8I6b9awxjmKYqIr6h
iPQCJaLBED8mwK+I5evIbnKv6E6uK+BApWA/R7ElragoFYbqQ1VpvntVMtJt9Dy5
ZrI+IQARdXD3bb34oh0IPBhClnvvMUc1cWxDoXEX6oJ4I+LzxE87Zkwnan9qOwen
golMVKFwPx1o37qrbmrXID21kKt7FL6xN4HxHLkItr1fKzdyWDFRHgASTAWfx5BI
wvPuUW0vZHkvO80VyV2L63whVhPnPASmFkbviomrBttYfpr2aGQqF/qR1Nlxe834
MFxk1pS9LMa/WnzvFr0gWakCAwEAAQ==
-----END RSA PUBLIC KEY-----
`)
	origStr := "Value1234"

	pubKey, err := DecodeToPublicKey(publicKeyBytes)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	privKey, err := DecodeToPrivateKey(privKeyBytes)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	encData, err := Encrypt([]byte(origStr), pubKey)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	encDataStr := base64.StdEncoding.EncodeToString(encData)
	log.Printf("encrypted data: %s\n", encDataStr)
	dec, _ := base64.StdEncoding.DecodeString(encDataStr)
	data, err := Decrypt(dec, privKey)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(data) != origStr {
		t.Error("original string and decrypted string don't match")
		t.FailNow()
	}
	log.Printf("decrypted data: %s\n", data)
}
