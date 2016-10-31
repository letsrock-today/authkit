package persisttoken

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/letsrock-today/authkit/authkit/mocks"
)

func TestPersistToken(t *testing.T) {
	old := &oauth2.Token{
		AccessToken: "old",
		Expiry:      time.Now().Add(-1 * time.Hour)}
	new := &oauth2.Token{
		AccessToken: "new",
		Expiry:      time.Now().Add(1 * time.Hour)}
	ts := &mocks.TokenStore{}
	ts.On(
		"OAuth2Token",
		"valid@login.ok",
		"provider-1").
		Return(old, nil)
	ts.On(
		"OAuth2Token",
		"valid@login.ok",
		"provider-2").
		Return(new, nil)
	ts.On(
		"OAuth2Token",
		"valid@login.ok",
		"provider-3").
		Return(new, nil)

	ts.On(
		"UpdateOAuth2Token",
		"valid@login.ok",
		mock.Anything,
		mock.Anything).Return(nil).Once()

	testCfg := &mocks.OAuth2Config{}
	testCfg.On(
		"TokenSource",
		mock.Anything,
		mock.Anything).
		Return(oauth2.StaticTokenSource(new))

	for i := 1; i < 3; i++ {
		providerID := fmt.Sprintf("provider-%d", i)
		t.Run(providerID, func(st *testing.T) {
			assert := assert.New(st)
			cfg := WrapOAuth2Config(
				testCfg,
				"valid@login.ok",
				providerID,
				ts)
			t, err := cfg.TokenSource(context.Background(), nil).Token()
			assert.NoError(err)
			assert.NotNil(t)
		})
	}

	ts.AssertNumberOfCalls(t, "UpdateOAuth2Token", 1)
}
