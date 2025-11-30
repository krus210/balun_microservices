package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// RSAKeyGenerator - генератор RSA ключей
type RSAKeyGenerator struct {
	bits int
}

func NewRSAKeyGenerator() *RSAKeyGenerator {
	return &RSAKeyGenerator{bits: 2048}
}

func (g *RSAKeyGenerator) Generate() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, g.bits)
}

// EncodePrivateKeyToPEM - кодирует приватный ключ в PEM формат
func EncodePrivateKeyToPEM(key *rsa.PrivateKey) (string, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

// EncodePublicKeyToPEM - кодирует публичный ключ в PEM формат
func EncodePublicKeyToPEM(key *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

// DecodePrivateKeyFromPEM - декодирует приватный ключ из PEM формата
func DecodePrivateKeyFromPEM(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block for private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA key")
	}

	return rsaKey, nil
}

// DecodePublicKeyFromPEM - декодирует публичный ключ из PEM формата
func DecodePublicKeyFromPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block for public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaKey, nil
}
