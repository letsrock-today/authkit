package apptoken

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// EmailToken represents token, used to be sent in confirmation email.
type (
	EmailToken interface {

		// Login returns user's login.
		Login() string

		// Email returns users's email which token was issued for.
		Email() string

		// PasswordHash returns the hash of the password.
		// Password hash goes into the token and then into the link, which is
		// included into the email for the user. When the user follow the link,
		// the token is returned to the server in the url param and the hash is
		// used to check if the password has not been changed yet.
		PasswordHash() string
	}

	mailTokenFields struct {
		Login        string `json:"login"`
		Email        string `json:"email"`
		PasswordHash string `json:"pwdh"`
	}

	mailToken struct {
		jwt.StandardClaims
		mailTokenFields
	}
)

// NewEmailTokenString creates new jwt token and converts it to signed string.
// It may be used to create password reset URL for confirmation email.
func NewEmailTokenString(
	issuer, login, email, passwordHash string,
	expiration time.Duration,
	signKey []byte) (string, error) {
	claims := mailToken{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
			Issuer:    issuer,
			Audience:  issuer,
		},
		mailTokenFields{
			login,
			email,
			passwordHash,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signKey)
}

// ParseEmailToken can parse jwt tokens from strings created by NewEmailTokenString.
func ParseEmailToken(
	issuer, token string,
	signKey []byte) (EmailToken, error) {
	t, err := jwt.ParseWithClaims(
		token,
		&mailToken{},
		func(token *jwt.Token) (interface{}, error) {
			return signKey, nil
		})
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}
	claims, ok := t.Claims.(*mailToken)
	if !ok || !t.Valid {
		return nil, errors.WithStack(ErrInvalidToken)
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return nil, errors.WithStack(ErrInvalidToken)
	}
	if !claims.VerifyIssuer(issuer, true) {
		return nil, errors.WithStack(ErrInvalidToken)
	}
	if !claims.VerifyAudience(issuer, true) {
		return nil, errors.WithStack(ErrInvalidToken)
	}
	return claims, nil
}

func (m *mailToken) Login() string {
	return m.mailTokenFields.Login
}

func (m *mailToken) Email() string {
	return m.mailTokenFields.Email
}

func (m *mailToken) PasswordHash() string {
	return m.mailTokenFields.PasswordHash
}
