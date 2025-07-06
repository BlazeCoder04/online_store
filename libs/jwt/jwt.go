package jwt

import (
	"errors"
	"time"

	"github.com/BlazeCoder04/online_store/libs/jwt/parse"
	"github.com/golang-jwt/jwt"
)

const ErrTokenInvalid = "token.invalid"

func Create(ttl time.Duration, userID, userRole, privateKey string) (string, error) {
	key, err := parse.ParsePrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub":  userID,
		"role": userRole,
		"exp":  time.Now().Add(ttl).Unix(),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		return "", err
	}

	return token, nil
}

func Verify(token string, publicKey string) (jwt.MapClaims, error) {
	key, err := parse.ParsePublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New(t.Header["alg"].(string))
		}

		return key, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, errors.New(ErrTokenInvalid)
	}

	return claims, nil
}
