package apptoken

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// StateToken represents state token.
type (
	StateToken interface {

		// ProviderID is an ID of OAuth2 provider, used to initiate auth request.
		ProviderID() string

		// User's login if known at state creation (like in case of form-based auth)
		Login() string
	}

	stateToken jwt.StandardClaims
)

// NewStateTokenString creates new jwt token and converts it to signed string.
// It can be used to create state token for OAuth2 code flow.
func NewStateTokenString(
	issuer, providerID string,
	expiration time.Duration,
	signKey []byte) (string, error) {
	return newToken(issuer, "", providerID, expiration, signKey)
}

// NewStateWithLoginTokenString creates new jwt token and converts it to signed string.
// It can be used to create state token for OAuth2 code flow.
// This method additionally packs login into token, which is useful in case of
// login/password authentication via application's own login form.
func NewStateWithLoginTokenString(
	issuer, providerID, login string,
	expiration time.Duration,
	signKey []byte) (string, error) {
	return newToken(issuer, login, providerID, expiration, signKey)
}

// ParseStateToken can parse jwt tokens from strings created by NewStateTokenString or NewStateWithLoginTokenString.
func ParseStateToken(
	issuer, token string,
	signKey []byte) (StateToken, error) {
	s, err := parseToken(issuer, token, signKey)
	if err != nil {
		return nil, err
	}
	t := stateToken(*s)
	return &t, nil
}

func (s *stateToken) ProviderID() string {
	return s.Subject
}

func (s *stateToken) Login() string {
	return s.Audience
}
