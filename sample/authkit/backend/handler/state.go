package handler

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Note: when used for oaut2 state, pid goes into subject and audience is empty
// when used for confirmation email link, email goes into audience and password hash - into subject
func newStateToken(
	tokenSignKey []byte,
	tokenIssuer, audience, subject string,
	tokenExpiration time.Duration) (string, error) {
	claims :=
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenExpiration).Unix(),
			Issuer:    tokenIssuer,
			Audience:  audience,
			Subject:   subject,
		}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tokenSignKey)
}

func parseStateToken(
	tokenSignKey []byte,
	tokenIssuer, tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return tokenSignKey, nil
		})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token")
	}
	if claims.Issuer != tokenIssuer {
		return nil, errors.New("Invalid issuer")
	}
	return claims, nil
}
