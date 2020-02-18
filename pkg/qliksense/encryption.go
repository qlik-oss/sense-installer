package qliksense

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	RSA_KEY_LENGTH = 4096
	privateKeyPath = "privKey"
	publicKeyPath  = "pubKey"
)

// GenerateRSAEncryptionKeys is used to generate a new public-private key pair
func GenerateRSAEncryptionKeys(publicKeyFilePath, privateKeyFilePath string) error {
	LogDebugMessage("Generating new RSA key pair")
	privateKey, err := rsa.GenerateKey(rand.Reader, RSA_KEY_LENGTH)
	if err != nil {
		log.Printf("error generating RSA private key: %v\n", err)
		return err
	}

	privateKeyPEM := EncodePrivateKey(privateKey)
	if err := writeContentToFile(privateKeyPEM, privateKeyFilePath); err != nil {
		return err
	}
	pubKeyPEM, err2 := EncodePublicKey(&privateKey.PublicKey)
	if err2 != nil {
		log.Printf("error occurred when encoding public key: %v\n", err2)
		return err2
	}
	if err := writeContentToFile(pubKeyPEM, publicKeyFilePath); err != nil {
		return err
	}
	return nil
}

// writeContentToFile writes keys to a file
func writeContentToFile(keyData []byte, fileName string) error {
	err := ioutil.WriteFile(fileName, keyData, 0600)
	if err != nil {
		log.Printf("error writing to file (%s): %v", fileName, err)
		return err
	}
	LogDebugMessage("Key saved: %s", fileName)
	return nil
}

// Encrypt encrypts data with public key
func Encrypt(pt []byte, pub *rsa.PublicKey) ([]byte, error) {
	// hash := sha512.New()
	// ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	ct, err := rsa.EncryptPKCS1v15(rand.Reader, pub, pt)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return ct, nil
}

// Decrypt decrypts data with private key
func Decrypt(ct []byte, priv *rsa.PrivateKey) ([]byte, error) {
	// hash := sha512.New()
	// plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	pt, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ct)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return pt, nil
}

// EncodePrivateKey private key to bytes
func EncodePrivateKey(priv *rsa.PrivateKey) []byte {
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	return privBytes
}

// EncodePublicKey public key to bytes
func EncodePublicKey(pub *rsa.PublicKey) ([]byte, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes, nil
}

// DecodePrivateKey bytes to private key
func DecodePrivateKey(priv []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return key, nil
}

// DecodePublicKey bytes to public key
func DecodePublicKey(pub []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	iface, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	key, ok := iface.(*rsa.PublicKey)
	if !ok {
		err := fmt.Errorf("Unable to decode public key")
		log.Println(err)
		return nil, err
	}
	return key, nil
}
