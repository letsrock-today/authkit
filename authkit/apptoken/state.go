package apptoken

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type (

	// StateToken represents state token.
	StateToken interface {

		// ProviderID is an ID of OAuth2 provider, used to initiate auth request.
		ProviderID() string

		// User's login if known at state creation (like in case of form-based auth)
		Login() string
	}

	stateTokenFields struct {
		Login      string `json:"login"`
		ProviderID string `json:"pid"`
	}

	stateToken struct {
		jwt.StandardClaims
		stateTokenFields
	}
)

// NewStateTokenString creates new jwt token and converts it to signed string.
// It can be used to create state token for OAuth2 code flow.
func NewStateTokenString(
	issuer, providerID string,
	expiration time.Duration,
	signKey []byte) (string, error) {
	claims := stateToken{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
			Issuer:    issuer,
			Audience:  issuer,
		},
		stateTokenFields{
			"",
			providerID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signKey)
}

// NewStateWithLoginTokenString creates new jwt token and converts it to signed string.
// It can be used to create state token for OAuth2 code flow.
// This method additionally packs login into token, which is useful in case of
// login/password authentication via application's own login form.
func NewStateWithLoginTokenString(
	issuer, providerID, login string,
	expiration time.Duration,
	signKey []byte) (string, error) {
	claims := stateToken{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
			Issuer:    issuer,
			Audience:  issuer,
		},
		stateTokenFields{
			login,
			providerID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signKey)
}

// ParseStateToken can parse jwt tokens from strings created by NewStateTokenString or NewStateWithLoginTokenString.
func ParseStateToken(
	issuer, token string,
	signKey []byte) (StateToken, error) {
	t, err := jwt.ParseWithClaims(
		token,
		&stateToken{},
		func(token *jwt.Token) (interface{}, error) {
			return signKey, nil
		})
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}
	claims, ok := t.Claims.(*stateToken)
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

func (s *stateToken) ProviderID() string {
	return s.stateTokenFields.ProviderID
}

func (s *stateToken) Login() string {
	return s.stateTokenFields.Login
}
