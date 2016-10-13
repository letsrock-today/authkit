package apptoken

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// EmailToken represents token, used to be sent in confirmation email.
type (
	EmailToken interface {

		// Login returns user's login.
		Login() string

		// PasswordHash returns the hash of the password.
		// Password hash goes into the token and then into the link, which is
		// included into the email for the user. When the user follow the link,
		// the token is returned to the server in the url param and the hash is
		// used to check if the password has not been changed yet.
		PasswordHash() string
	}

	mailToken jwt.StandardClaims
)

// NewEmailTokenString creates new jwt token and converts it to signed string.
// It may be used to create password reset URL for confirmation email.
func NewEmailTokenString(
	issuer, email, passwordHash string,
	expiration time.Duration,
	signKey []byte) (string, error) {
	return newToken(issuer, email, passwordHash, expiration, signKey)
}

// ParseEmailToken can parse jwt tokens from strings created by NewEmailTokenString.
func ParseEmailToken(
	issuer, token string,
	signKey []byte) (EmailToken, error) {
	m, err := parseToken(issuer, token, signKey)
	if err != nil {
		return nil, err
	}
	t := mailToken(*m)
	return &t, nil
}

func (m *mailToken) Login() string {
	return m.Audience
}

func (m *mailToken) PasswordHash() string {
	return m.Subject
}
