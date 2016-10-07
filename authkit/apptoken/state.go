package apptoken

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// StateToken represents state token.
type StateToken interface {

	// ProviderID is an ID of OAuth2 provider, used to initiate auth request.
	ProviderID() string

	// User's login if known at state creation (like in case of form-based auth)
	Login() string
}

func NewStateTokenString(
	issuer, providerID string,
	expiration time.Duration,
	signKey []byte) (string, error) {
	return newToken(issuer, "", providerID, expiration, signKey)
}

func NewStateWithLoginTokenString(
	issuer, providerID, login string,
	expiration time.Duration,
	signKey []byte) (string, error) {
	return newToken(issuer, login, providerID, expiration, signKey)
}

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

type stateToken jwt.StandardClaims

func (s *stateToken) ProviderID() string {
	return s.Subject
}

func (s *stateToken) Login() string {
	return s.Audience
}
