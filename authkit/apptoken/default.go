package apptoken

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

var InvalidTokenError = errors.New("invalid token")

// Note: when used for oauth2 state, pid goes into subject and audience is empty
// when used for confirmation email link, email goes into audience and password hash - into subject
func newToken(
	tokenIssuer, audience, subject string,
	tokenExpiration time.Duration,
	tokenSignKey []byte) (string, error) {
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

func parseToken(
	tokenIssuer, tokenString string,
	tokenSignKey []byte) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return tokenSignKey, nil
		})
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return nil, errors.WithStack(InvalidTokenError)
	}
	if claims.Issuer != tokenIssuer {
		return nil, errors.WithStack(InvalidTokenError)
	}
	return claims, nil
}
