package qliksense

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
)

const (
	RSA_KEY_BITS   = 4096
	privateKeyPath = "privKey"
	publicKeyPath  = "pubKey"
)

// GenerateRSAEncryptionKeys is used to generate a new public-private key pair
func GenerateRSAEncryptionKeys(publicKeyFilePath, privateKeyFilePath string) error {
	LogDebugMessage("Generating new RSA key pair")
	privateKey, err := rsa.GenerateKey(rand.Reader, RSA_KEY_BITS)
	if err != nil {
		log.Printf("error generating RSA private keys: %v\n", err)
		return err
	}

	privateKeyPEM := PrivateKeyToBytes(privateKey)
	if err := writeContentToFile(privateKeyPEM, privateKeyFilePath); err != nil {
		return err
	}
	pubKeyPEM := PublicKeyToBytes(&privateKey.PublicKey)
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

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) []byte {
	// hash := sha512.New()
	// ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, msg)
	if err != nil {
		log.Println(err)
	}
	return ciphertext
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) []byte {
	// hash := sha512.New()
	// plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	if err != nil {
		log.Println(err)
	}
	return plaintext
}

// PrivateKeyToBytes private key to bytes
func PrivateKeyToBytes(priv *rsa.PrivateKey) []byte {
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	return privBytes
}

// PublicKeyToBytes public key to bytes
func PublicKeyToBytes(pub *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Println(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes
}

// BytesToPrivateKey bytes to private key
func BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Println(err)
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		log.Println(err)
	}
	return key
}

// BytesToPublicKey bytes to public key
func BytesToPublicKey(pub []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Println(err)
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		log.Println(err)
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		log.Println("not ok")
	}
	return key
}
