package service

import (
	"errors"
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte(string(os.Getenv("SECRET")))

type сlaims struct {
	UserID string `json:"uid"`
	UserIP string `json:"uip"`
	jwt.StandardClaims
}

func getClaims(token string) (*сlaims, error) {
	claims := &сlaims{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Signing method isn't HMAC: %v", token.Header["alg"])
		}

		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !tkn.Valid {
		return claims, errors.New("Token is not valid")
	}

	return claims, nil
}
