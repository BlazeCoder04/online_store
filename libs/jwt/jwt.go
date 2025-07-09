package jwt

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/BlazeCoder04/online_store/libs/jwt/parse"
	"github.com/golang-jwt/jwt"
)

var ErrTokenInvalid = errors.New("token.invalid")

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
		log.Println(fmt.Sprintf("[%s] %v", "parse public key", err))
		return nil, err
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New(t.Header["alg"].(string))
		}

		return key, nil
	})
	if err != nil {
		log.Println(fmt.Sprintf("[%s] %v", "jwt parse", err))
		return nil, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
