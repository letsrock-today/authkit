package persisttoken

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/letsrock-today/hydra-sample/authkit"
)

func TestPersistToken(t *testing.T) {
	old := &oauth2.Token{
		AccessToken: "old",
		Expiry:      time.Now().Add(-1 * time.Hour)}
	new := &oauth2.Token{
		AccessToken: "new",
		Expiry:      time.Now().Add(1 * time.Hour)}
	ts := &testTokenStore{}
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

	testCfg := &testOAuth2Config{}
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

type testTokenStore struct {
	mock.Mock
}

func (m *testTokenStore) OAuth2Token(
	login, providerID string) (*oauth2.Token, authkit.UserServiceError) {
	args := m.Called(login, providerID)
	if err, ok := args.Error(1).(authkit.UserServiceError); ok {
		return nil, err
	}
	return args.Get(0).(*oauth2.Token), nil
}

func (m *testTokenStore) UpdateOAuth2Token(
	login, providerID string, token *oauth2.Token) authkit.UserServiceError {
	args := m.Called(login, providerID, token)
	if err, ok := args.Error(0).(authkit.UserServiceError); ok {
		return err
	}
	return nil
}

type testOAuth2Config struct {
	mock.Mock
}

func (m *testOAuth2Config) Client(
	ctx context.Context,
	t *oauth2.Token) *http.Client {
	args := m.Called(ctx, t)
	return args.Get(0).(*http.Client)
}

func (m *testOAuth2Config) TokenSource(
	ctx context.Context,
	t *oauth2.Token) oauth2.TokenSource {
	args := m.Called(ctx, t)
	return args.Get(0).(oauth2.TokenSource)
}

func (m *testOAuth2Config) AuthCodeURL(
	state string,
	opts ...oauth2.AuthCodeOption) string {
	args := m.Called(state, opts)
	return args.String(0)
}

func (m *testOAuth2Config) PasswordCredentialsToken(
	ctx context.Context,
	username, password string) (*oauth2.Token, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *testOAuth2Config) Exchange(
	ctx context.Context,
	code string) (*oauth2.Token, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(*oauth2.Token), args.Error(1)
}
