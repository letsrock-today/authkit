package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
)

type testConfig struct {
	oauth2State           testOAuth2State
	oauth2Providers       []testOAuth2Provider
	privateOAuth2Provider testOAuth2Provider
	modTime               time.Time
	tlsInsecureSkipVerify bool
}

func (c testConfig) OAuth2Providers() chan authkit.OAuth2Provider {
	ch := make(chan authkit.OAuth2Provider)
	go func() {
		for _, p := range c.oauth2Providers {
			ch <- p
		}
		close(ch)
	}()
	return ch
}

func (c testConfig) OAuth2ProviderByID(id string) authkit.OAuth2Provider {
	for _, p := range c.oauth2Providers {
		if id == p.ID() {
			return p
		}
	}
	return nil
}

func (c testConfig) PrivateOAuth2Provider() authkit.OAuth2Provider {
	return c.privateOAuth2Provider
}

func (c testConfig) OAuth2State() authkit.OAuth2State {
	return c.oauth2State
}

func (c testConfig) AuthCookieName() string {
	return "xxx-auth-cookie"
}

func (c testConfig) ModTime() time.Time {
	return c.modTime
}

type testOAuth2State struct {
	tokenIssuer  string
	tokenSignKey []byte
	expiration   time.Duration
}

func (s testOAuth2State) TokenIssuer() string {
	return s.tokenIssuer
}

func (s testOAuth2State) TokenSignKey() []byte {
	return s.tokenSignKey
}

func (s testOAuth2State) Expiration() time.Duration {
	return s.expiration
}

type testOAuth2Provider struct {
	id               string
	name             string
	iconURL          string
	oauth2Config     authkit.OAuth2Config
	privOAuth2Config authkit.OAuth2Config
}

func (p testOAuth2Provider) ID() string {
	return p.id
}

func (p testOAuth2Provider) Name() string {
	return p.name
}

func (p testOAuth2Provider) IconURL() string {
	return p.iconURL
}

func (p testOAuth2Provider) OAuth2Config() authkit.OAuth2Config {
	return p.oauth2Config
}

func (p testOAuth2Provider) PrivateOAuth2Config() authkit.OAuth2Config {
	return p.privOAuth2Config
}

type testErrorCustomizer struct{}

func (testErrorCustomizer) InvalidRequestParameterError(error) interface{} {
	return struct {
		Code string
	}{
		"invalid req param",
	}
}

func (testErrorCustomizer) UserCreationError(e error) interface{} {
	var msg string
	switch e.(type) {
	case authkit.DuplicateUserError:
		msg = "dup user"
	case authkit.AccountDisabledError:
		msg = "acc disabled"
	case authkit.RequestConfirmationError:
		msg = "req confirm error"
	default:
		msg = "general error"
	}
	return struct {
		Code string
	}{
		msg,
	}
}

func (testErrorCustomizer) UserAuthenticationError(error) interface{} {
	return struct {
		Code string
	}{
		"user auth err",
	}
}

type testUser struct {
	login        string
	email        string
	passwordHash string
}

func (u testUser) Login() string {
	return u.login
}

func (u testUser) Email() string {
	return u.email
}

func (u testUser) PasswordHash() string {
	return u.passwordHash
}

type testProfile struct {
	login string
}

func (p testProfile) Login() string {
	return p.login
}

type testBodyEncoderFunc func(v url.Values) io.Reader

var testBodyEncoders = []struct {
	name        string
	contentType string
	invalid     bool
	encoder     testBodyEncoderFunc
}{
	{
		name:        "invalid payload",
		contentType: "application/json",
		invalid:     true,
		encoder: func(v url.Values) io.Reader {
			return strings.NewReader("zzz")
		},
	},
	{
		name:        "form payload",
		contentType: "application/x-www-form-urlencoded",
		encoder: func(v url.Values) io.Reader {
			return strings.NewReader(v.Encode())
		},
	},
	{
		name:        "multipart payload",
		contentType: "multipart/form-data; boundary=---",
		encoder: func(v url.Values) io.Reader {
			var b bytes.Buffer
			w := multipart.NewWriter(&b)
			w.SetBoundary("---")
			for n, v := range v {
				for _, v := range v {
					fw, err := w.CreateFormField(n)
					if err != nil {
						panic(err)
					}
					_, err = fw.Write([]byte(v))
					if err != nil {
						panic(err)
					}
				}
			}
			w.Close()
			return &b
		},
	},
}

func testNewEmailTokenString(
	t *testing.T,
	config authkit.Config,
	email, passwordHash string,
	expired ...bool) []string {
	exp := 1 * time.Hour
	if len(expired) > 0 && expired[0] {
		exp = -1 * time.Hour
	}
	s, err := apptoken.NewEmailTokenString(
		config.OAuth2State().TokenIssuer(),
		email,
		passwordHash,
		exp,
		config.OAuth2State().TokenSignKey())
	assert.NoError(t, err)
	return []string{s}
}

func testNewStateTokenString(
	t *testing.T,
	config authkit.Config,
	pid, login string,
	expired ...bool) []string {
	exp := 1 * time.Hour
	if len(expired) > 0 && expired[0] {
		exp = -1 * time.Hour
	}
	if login == "" {
		s, err := apptoken.NewStateTokenString(
			config.OAuth2State().TokenIssuer(),
			pid,
			exp,
			config.OAuth2State().TokenSignKey())
		assert.NoError(t, err)
		return []string{s}
	}
	s, err := apptoken.NewStateWithLoginTokenString(
		config.OAuth2State().TokenIssuer(),
		pid,
		login,
		exp,
		config.OAuth2State().TokenSignKey())
	assert.NoError(t, err)
	return []string{s}
}
