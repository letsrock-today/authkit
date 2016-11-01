package deezer

import (
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestAuthCodeURL(t *testing.T) {
	conf := Config{
		&oauth2.Config{
			Endpoint: oauth2.Endpoint{
				AuthURL: "server/auth",
			},
			RedirectURL: "REDIRECT_URL",
			ClientID:    "CLIENT_ID",
			Scopes:      []string{"scope1", "scope2"},
		},
	}
	assert := assert.New(t)
	url := conf.AuthCodeURL("foo")
	const want = "server/auth?app_id=CLIENT_ID&perms=scope1%2Cscope2&redirect_uri=REDIRECT_URL&response_type=code&state=foo"
	assert.Equal(want, url)
}

func TestExchange(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)
	gock.New("http://foo.com").
		Get("/token").
		MatchParams(map[string]string{
			"app_id": "CLIENT_ID",
			"code":   "exchange-code",
			"secret": "CLIENT_SECRET"}).
		Reply(200).
		Type("url").
		BodyString("access_token=42&token_type=bearer")
	conf := Config{
		&oauth2.Config{
			Endpoint: oauth2.Endpoint{
				TokenURL: "http://foo.com/token",
			},
			ClientID:     "CLIENT_ID",
			ClientSecret: "CLIENT_SECRET",
			Scopes:       []string{"scope1", "scope2"},
		},
	}
	tok, err := conf.Exchange(context.Background(), "exchange-code")
	assert.NoError(err)
	assert.True(tok.Valid())
	assert.Equal("42", tok.AccessToken)
	assert.Equal("bearer", tok.TokenType)
}
