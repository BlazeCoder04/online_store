package parse

import (
	"crypto/rsa"
	"encoding/base64"

	"github.com/golang-jwt/jwt"
)

func decodePEMFromBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func ParsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	decoded, err := decodePEMFromBase64(privateKey)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPrivateKeyFromPEM(decoded)
}

func ParsePublicKey(publicKey string) (*rsa.PublicKey, error) {
	decoded, err := decodePEMFromBase64(publicKey)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(decoded)
}
