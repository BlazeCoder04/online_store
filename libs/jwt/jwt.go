package jwt

import (
	"errors"
	"time"

	"github.com/BlazeCoder04/online_store/libs/jwt/parse"
	"github.com/golang-jwt/jwt"
)

func Create(ttl time.Duration, userID, privateKey string) (string, error) {
	key, err := parse.ParsePrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		return "", err
	}

	return token, nil
}

func Validate(token, publicKey string) (string, error) {
	key, err := parse.ParsePublicKey(publicKey)
	if err != nil {
		return "", err
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New(t.Header["alg"].(string))
		}

		return key, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return "", errors.New(ErrTokenInvalid)
	}

	return claims["sub"].(string), nil
}
