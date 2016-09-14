package handler

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func newStateToken(
	tokenSignKey []byte,
	tokenIssuer, pid string,
	tokenExpiration time.Duration) (string, error) {
	claims := stateClaims{
		pid,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenExpiration).Unix(),
			Issuer:    tokenIssuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tokenSignKey)
}

func parseStateToken(
	tokenSignKey []byte,
	tokenIssuer, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&stateClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return tokenSignKey, nil
		})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*stateClaims)
	if !ok || !token.Valid {
		return "", errors.New("Invalid token")
	}
	if claims.Issuer != tokenIssuer {
		return "", errors.New("Invalid issuer")
	}
	return claims.Pid, nil
}

//// Internals

type stateClaims struct {
	Pid string `json:"pid"`
	jwt.StandardClaims
}
