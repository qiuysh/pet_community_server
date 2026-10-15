package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
)

var rsaPrivateKey *rsa.PrivateKey
var rsaPublicKey *rsa.PublicKey

func InitRSA(bits int) error {
	var err error
	rsaPrivateKey, err = rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	rsaPublicKey = &rsaPrivateKey.PublicKey
	return nil
}

func GetPublicKey() string {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(rsaPublicKey)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(publicKeyBytes)
}

func DecryptRSA(ciphertext string) (string, error) {
	if rsaPrivateKey == nil {
		return "", errors.New("RSA not initialized")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, data)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
