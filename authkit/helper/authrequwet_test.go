package helper

import (
	"errors"
	"testing"

	"golang.org/x/oauth2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/letsrock-today/hydra-sample/authkit"
)

func TestWithOAuth2TokenDo(t *testing.T) {
	ts := &testTokenStore{}
	ts.On(
		"OAuth2Token",
		"valid@login.ok",
		mock.Anything).Return(&oauth2.Token{}, nil)
	ts.On(
		"UpdateOAuth2Token",
		"valid@login.ok",
		"update-token-provider",
		mock.Anything).Return(nil).Once()

	testError := errors.New("error in callback")

	cases := []struct {
		name       string
		login      string
		providerID string
		err        error
		do         func(*oauth2.Token) (*oauth2.Token, error)
	}{
		{
			name:       "All OK",
			login:      "valid@login.ok",
			providerID: "some-provider",
			do: func(t *oauth2.Token) (*oauth2.Token, error) {
				return t, nil
			},
		},
		{
			name:       "Update token - OK",
			login:      "valid@login.ok",
			providerID: "update-token-provider",
			do: func(t *oauth2.Token) (*oauth2.Token, error) {
				return &oauth2.Token{AccessToken: "new"}, nil
			},
		},
		{
			name:       "Error from callback",
			login:      "valid@login.ok",
			providerID: "some-provider",
			err:        testError,
			do: func(t *oauth2.Token) (*oauth2.Token, error) {
				return &oauth2.Token{AccessToken: "new2"}, testError
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(st *testing.T) {
			assert := assert.New(st)
			err := WithOAuth2TokenDo(
				ts,
				c.login,
				c.providerID,
				c.do)
			if c.err == nil {
				assert.NoError(err)
			} else {
				assert.Error(err)
				assert.Equal(c.err, err)
			}
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
