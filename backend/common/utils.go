package common

import (
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
)

type customClaims struct {
	Pid string `json:"pid"`
	jwt.StandardClaims
}

func CreateToken(tokenSignKey, tokenIssuer, pid string, tokenExpiration time.Duration) string {
	claims := customClaims{
		pid,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenExpiration).Unix(),
			Issuer:    tokenIssuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(tokenSignKey)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func ValidToken(tokenSignKey string, t interface{}) (string, bool) {
	if s, ok := t.(string); ok {
		token, err := jwt.ParseWithClaims(s, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
			return tokenSignKey, nil
		})
		if err != nil {
			log.Print(err)
			return "", false
		}
		if claims, ok := token.Claims.(*customClaims); ok && token.Valid {
			return claims.Pid, true
		}
	}
	return "", false
}
