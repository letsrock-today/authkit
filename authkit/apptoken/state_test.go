package apptoken

import (
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestStateToken(t *testing.T) {
	cases := []struct {
		name    string
		issuer1 string
		issuer2 string
		pid     string
		login   string
		exp     time.Duration
		key     string
		err     error
	}{
		{
			"Correct args",
			"some issuer",
			"some issuer",
			"sompe pid",
			"xxx",
			1 * time.Hour,
			"some secret",
			nil,
		},
		{
			"Empty login",
			"some issuer",
			"some issuer",
			"sompe pid",
			"",
			1 * time.Hour,
			"some secret",
			nil,
		},
		{
			"Expired",
			"some issuer",
			"some issuer",
			"sompe pid",
			"",
			-1 * time.Hour,
			"some secret",
			&jwt.ValidationError{Errors: jwt.ValidationErrorExpired},
		},
		{
			"Incorrect issuer",
			"some issuer",
			"some other issuer",
			"sompe pid",
			"",
			1 * time.Hour,
			"some secret",
			ErrInvalidToken,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(st *testing.T) {
			assert := assert.New(st)

			key := []byte(c.key)
			var s string
			var err error
			if c.login == "" {
				s, err = NewStateTokenString(c.issuer1, c.pid, c.exp, key)
			} else {
				s, err = NewStateWithLoginTokenString(c.issuer1, c.pid, c.login, c.exp, key)
			}
			assert.NoError(err)
			assert.NotZero(s)

			token, err := ParseStateToken(c.issuer2, s, key)
			if c.err != nil {
				assert.Error(err)
				cause := errors.Cause(err)
				assert.IsType(c.err, cause)
				if e, ok := c.err.(*jwt.ValidationError); ok {
					assert.Equal(e.Errors, cause.(*jwt.ValidationError).Errors)
				}
			} else {
				assert.NoError(err)
				assert.NotNil(token)
				if token == nil {
					assert.FailNow("Token is nil")
				}
				assert.NotZero(token)
				assert.Equal(c.pid, token.ProviderID())
				assert.Equal(c.login, token.Login())
			}
		})
	}
}
