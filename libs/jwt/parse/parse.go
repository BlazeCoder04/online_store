package parse

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"

	"github.com/golang-jwt/jwt"
)

var ErrTokenInvalid = errors.New("token.invalid")

func decodePEMFromBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func ParsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	decoded, err := decodePEMFromBase64(privateKey)
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return nil, ErrTokenInvalid
		}

		return nil, err
	}

	return jwt.ParseRSAPrivateKeyFromPEM(decoded)
}

func ParsePublicKey(publicKey string) (*rsa.PublicKey, error) {
	decoded, err := decodePEMFromBase64(publicKey)
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return nil, ErrTokenInvalid
		}

		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(decoded)
}
