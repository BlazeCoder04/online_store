package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func generateRSAPrivateKeyBase64(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA private key: %v", err)
	}

	return base64.StdEncoding.EncodeToString(
		pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}),
	)
}

func generateRSAPublicKeyBase64(t *testing.T, privateKeyBase64 string) string {
	t.Helper()

	decodedPEM, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		t.Fatalf("failed to decode base64 private key: %v", err)
	}

	pemBlock, _ := pem.Decode(decodedPEM)
	if pemBlock == nil || pemBlock.Type != "RSA PRIVATE KEY" {
		t.Fatalf("failed to decode PEM block containing private key, got type: %v", pemBlock.Type)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal RSA public key: %v", err)
	}

	return base64.StdEncoding.EncodeToString(
		pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		}),
	)
}
