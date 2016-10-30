package hydra

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/mendsley/gojwk"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/middleware"
)

func TestGetKey(t *testing.T) {
	defer gock.Off()

	assert := assert.New(t)

	testPrepareKeysResponder(t, 0)

	h := testCreateHydra()

	key, err := h.getKey("hydra.consent.response", "private")
	assert.NoError(err)
	assert.NotNil(key)

	key, err = h.getKey("hydra.consent.challenge", "public")
	assert.NoError(err)
	assert.NotNil(key)
}

func TestGetKeyCached(t *testing.T) {
	defer gock.Off()

	assert := assert.New(t)

	testPrepareKeysResponder(t, 1)

	h := testCreateHydra()

	for i := 0; i < 10; i++ {
		key, err := h.getConsentChallengePublicKey()
		assert.NoError(err)
		assert.NotNil(key)

		key, err = h.getConsentResponsePrivateKey()
		assert.NoError(err)
		assert.NotNil(key)
	}
}

func TestVerifyConsentChallenge(t *testing.T) {
	defer gock.Off()

	assert := assert.New(t)

	testPrepareKeysResponder(t, 0)

	h := testCreateHydra()

	consent, err := h.IssueConsentToken("client-id", []string{"some.scope"})
	assert.NoError(err)
	token, err := h.verifyConsentChallenge(consent)
	assert.NoError(err)
	assert.NotNil(token)
}

func TestGenerateConsentToken(t *testing.T) {
	defer gock.Off()

	assert := assert.New(t)

	testPrepareKeysResponder(t, 0)

	h := testCreateHydra()

	consent, err := h.IssueConsentToken("client-id", []string{"some.scope"})
	assert.NoError(err)
	token, err := h.GenerateConsentToken(
		"valid@email.ok",
		[]string{"some.scope"},
		consent)
	assert.NoError(err)
	assert.NotNil(token)
	token, err = h.GenerateConsentToken(
		"valid@email.ok",
		[]string{"some.scope, some.other.scope"},
		consent)
	assert.Error(err)
}

func TestIssueToken(t *testing.T) {
	defer gock.Off()

	assert := assert.New(t)

	prepareHTTPClientWithoutRedirects = func(h hydra) *http.Client {
		return &http.Client{
			Transport: gock.NewTransport(),
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	testPrepareKeysResponder(t, 0)

	gock.New("http://foo.com").
		Get("/auth/auth").
		Reply(302).Header.Set("Location", "http://foo.com/callback?code=xxx")

	h := testCreateHydra()

	token, err := h.IssueToken("valid@email.ok")
	assert.NoError(err)
	assert.NotNil(token)
}

func TestValidate(t *testing.T) {
	defer gock.Off()

	assert := assert.New(t)

	testPrepareKeysResponder(t, 0)

	gock.New("http://foo.com").
		Post("/warden/token/allowed").
		Reply(200).
		JSON(map[string]interface{}{"allowed": true})

	h := testCreateHydra()

	err := h.Validate("access-token", &middleware.DefaultPermission{
		Resource: "some-resource",
		Action:   "some-action",
		Scopes:   []string{"some.scope"},
	})
	assert.NoError(err)
}

func testPrepareKeysResponder(t *testing.T, timesForKeys int) {
	assert := assert.New(t)

	token := map[string]string{
		"access_token": "access-token",
		"token_type":   "Bearer",
	}

	k, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(err)
	privKey, err := gojwk.PrivateKey(k)
	assert.NoError(err)
	pubKey, err := gojwk.PublicKey(k.Public())
	assert.NoError(err)

	privKeys := map[string][]gojwk.Key{"keys": {*privKey}}
	pubKeys := map[string][]gojwk.Key{"keys": {*pubKey}}

	gock.New("http://foo.com").
		Post("/auth/token").
		Persist().
		Reply(200).
		JSON(token)
	gock.New("http://foo.com").
		Get("/keys/hydra.consent.response/private").
		Persist().
		Times(timesForKeys).
		Reply(200).
		JSON(privKeys)
	gock.New("http://foo.com").
		Get("/keys/hydra.consent.challenge/public").
		Persist().
		Times(timesForKeys).
		Reply(200).
		JSON(pubKeys)
}

func testCreateHydra() *hydra {
	return New(
		"http://foo.com",
		"some-provider-id",
		"some-trusted-provider-id",
		1*time.Hour,
		&oauth2.Config{
			ClientID: "client-id",
			Scopes:   []string{"some.scope"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "http://foo.com/auth/auth",
				TokenURL: "http://foo.com/auth/token",
			},
		},
		&clientcredentials.Config{
			TokenURL: "http://foo.com/auth/token",
		},
		authkit.OAuth2State{
			TokenIssuer:  "some-issuer",
			TokenSignKey: []byte("some-key"),
			Expiration:   1 * time.Hour,
		},
		authkit.DefaultContextCreator{},
		false).(*hydra)
}
