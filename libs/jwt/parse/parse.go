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
		return nil, err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(decoded)
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return nil, ErrTokenInvalid
		}

		return nil, err
	}

	return key, nil
}

func ParsePublicKey(publicKey string) (*rsa.PublicKey, error) {
	decoded, err := decodePEMFromBase64(publicKey)
	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(decoded)
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return nil, ErrTokenInvalid
		}

		return nil, err
	}

	return key, nil
}
